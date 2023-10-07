package abci

import (
	"cosmossdk.io/log"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ciprianmuja/weight-shift/weightskeeper"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
)

type VoteExtHandler struct {
	logger       log.Logger
	currentBlock int64               // current block height
	provider     map[string]Provider // provider from which get the external weight data

	Keeper    weightskeeper.WeightsKeeper
	GovKeeper govkeeper.Keeper
}

func NewVoteExtensionHandler(
	logger log.Logger,
	keeper weightskeeper.WeightsKeeper,
	govKeeper govkeeper.Keeper,
) *VoteExtHandler {
	return &VoteExtHandler{
		logger:    logger,
		Keeper:    keeper,
		GovKeeper: govKeeper,
	}
}

// WeightedVotingPowerVoteExtension defines the canonical vote extension structure.
type WeightedVotingPowerVoteExtension struct {
	Weights map[string]int64
}

func (h *VoteExtHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		h.logger.Info(fmt.Sprintf("🗳️ :: Extending Vote"))
		h.currentBlock = req.Height

		computedWeights, err := h.Keeper.GetWeights(ctx)
		if err != nil || computedWeights == nil {
			computedWeights = make(map[string]int64)
		}
		var maxPercentage int64 = 55 // this could be set by a governance proposal

		provider := Provider{}
		uptimePercentages := provider.GetValidatorsUptime(ctx, h.GovKeeper)
		governancePercentages := provider.GetValidatorsUptime(ctx, h.GovKeeper)
		githubContributions := provider.GetValidatorsGitHubContributions(ctx)
		if uptimePercentages == nil || governancePercentages == nil || githubContributions == nil {
			return nil, errors.New("invalid weights")
		}
		for validatorAddress, _ := range uptimePercentages {
			computedWeights[validatorAddress] = governancePercentages[validatorAddress] + uptimePercentages[validatorAddress] +
				githubContributions[validatorAddress]
			// eventually add other params like from more sources
		}

		// Calculate the maximum and minimum values in the current computedWeights map
		var maxWeight int64 = 0
		var minWeight int64 = 0
		for _, value := range computedWeights {
			if value > maxWeight {
				maxWeight = value
			}
			if value < minWeight {
				minWeight = value
			}
		}

		// Calculate the scaling factor to map values to the custom range
		var scalingFactor int64 = 1 // fallback value, no changes
		if maxWeight-minWeight > 0 {
			scalingFactor = maxPercentage / (maxWeight - minWeight)
		}

		// Scale the computedWeights values to the custom range
		for validatorAddress, _ := range uptimePercentages {
			computedWeights[validatorAddress] = (computedWeights[validatorAddress] - minWeight) * scalingFactor
			// Now, the values in computedWeights are scaled to the range [0, maxPercentage]
		}

		// produce a canonical vote extension
		voteExt := WeightedVotingPowerVoteExtension{
			Weights: computedWeights,
		}

		bz, err := json.Marshal(voteExt)
		if err != nil {
			h.logger.Error(err.Error())
			return nil, fmt.Errorf("failed to marshal vote extension: %w", err)
		}

		return &abci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

func (h *VoteExtHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *abci.RequestVerifyVoteExtension) (*abci.ResponseVerifyVoteExtension, error) {
		h.logger.Info(fmt.Sprintf(" :: Verifying Extended Votes"))
		var voteExt WeightedVotingPowerVoteExtension

		err := json.Unmarshal(req.VoteExtension, &voteExt)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal vote extension: %w", err)
		}

		// verify if they are valid
		if err := h.verifyWeights(ctx, voteExt.Weights); err != nil {
			return nil, fmt.Errorf("failed to verify weights from validator %X: %w", req.ValidatorAddress, err)
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

func (h *VoteExtHandler) verifyWeights(ctx sdk.Context, weights map[string]int64) error {
	return nil
}
