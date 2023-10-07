package abci

import (
	"cosmossdk.io/log"
	"encoding/json"
	"fmt"
	"github.com/ciprianmuja/weight-shift/weightskeeper"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type VoteExtHandler struct {
	logger       log.Logger
	currentBlock int64               // current block height
	provider     map[string]Provider // provider from which get the external weight data

	Keeper weightskeeper.WeightsKeeper
}

func NewVoteExtensionHandler(
	logger log.Logger,
	//providers map[string]Provider,
	keeper weightskeeper.WeightsKeeper,
) *VoteExtHandler {
	return &VoteExtHandler{
		logger: logger,
		//providers:       providers,
		Keeper: keeper,
	}
}

// WeightedVotingPowerVoteExtension defines the canonical vote extension structure.
type WeightedVotingPowerVoteExtension struct {
	Height  int64
	Weights map[string]int64
}

func (h *VoteExtHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *abci.RequestExtendVote) (*abci.ResponseExtendVote, error) {
		h.logger.Info(fmt.Sprintf("!! :: Extending Vote"))
		h.currentBlock = req.Height

		h.logger.Info("computing weighted voting power")

		//TODO: Add external sources here

		// TODO: if the new compute height is reached compute, otherwise use the already existing weights
		computedWeights, err := h.Keeper.GetWeights(ctx)
		if err != nil || len(computedWeights) <= 0 {
			computedWeights["val1"] = 0
			computedWeights["val2"] = 1
		}

		// produce a canonical vote extension
		voteExt := WeightedVotingPowerVoteExtension{
			Height:  7,
			Weights: computedWeights,
		}

		h.logger.Info("computed weights", "weights", computedWeights)

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
			// NOTE: It is safe to return an error as the Cosmos SDK will capture all
			// errors, log them, and reject the proposal.
			return nil, fmt.Errorf("failed to unmarshal vote extension: %w", err)
		}

		if voteExt.Height != req.Height {
			return nil, fmt.Errorf("vote extension height does not match request height; expected: %d, got: %d", req.Height, voteExt.Height)
		}

		// verify if they are valid
		if err := h.verifyWeights(ctx, voteExt.Weights); err != nil {
			return nil, fmt.Errorf("failed to verify oracle prices from validator %X: %w", req.ValidatorAddress, err)
		}

		return &abci.ResponseVerifyVoteExtension{Status: abci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}

func (h *VoteExtHandler) verifyWeights(ctx sdk.Context, prices map[string]int64) error {
	return nil
}
