package abci

import (
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ciprianmuja/weight-shift/weightskeeper"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// WeightedVotingPower defines the structure a proposer should use to calculate
// and submit the weighted voting power for the given validator set
type WeightedVotingPower struct {
	StakeWeightedWeighted map[string]int64
	ExtendedCommitInfo    abci.ExtendedCommitInfo
}

type ProposalHandler struct {
	logger        log.Logger
	keeper        weightskeeper.WeightsKeeper
	stakingKeeper *stakingkeeper.Keeper
	valStore      baseapp.ValidatorStore
}

func NewPrepareProposalHandler(logger log.Logger, keeper weightskeeper.WeightsKeeper, valStore baseapp.ValidatorStore,
	stakingKeeper *stakingkeeper.Keeper) *ProposalHandler {
	return &ProposalHandler{
		logger:        logger,
		keeper:        keeper,
		valStore:      valStore,
		stakingKeeper: stakingKeeper,
	}
}

func processVoteExtensions(req *abci.RequestPrepareProposal, log log.Logger) (WeightedVotingPowerVoteExtension, error) {
	log.Info(fmt.Sprintf("🛠️ :: Process Vote Extensions"))

	// Create empty response
	st := WeightedVotingPowerVoteExtension{
		map[string]int64{},
	}

	// Get Vote Ext for H-1 from Req
	voteExt := req.GetLocalLastCommit()
	votes := voteExt.Votes

	// Iterate through votes
	var ve WeightedVotingPowerVoteExtension
	for _, vote := range votes {
		err := json.Unmarshal(vote.VoteExtension, &ve)
		if err != nil {
			log.Error(fmt.Sprintf("❌ :: Error unmarshalling Vote Extension"))
		}

		// If Bids in VE, append to Special Transaction
		if len(ve.Weights) > 0 {
			log.Info("🛠️ :: Weights in VE")
			st.Weights = ve.Weights
		}
	}

	return st, nil
}

func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		var proposalTxs [][]byte
		if req.Height > 2 {

			// Get Special Transaction
			ve, err := processVoteExtensions(req, h.logger)
			if err != nil {
				h.logger.Error(fmt.Sprintf("❌️ :: Unable to process Vote Extensions: %v", err))
			}

			// Marshal Special Transaction
			bz, err := json.Marshal(ve)
			if err != nil {
				h.logger.Error(fmt.Sprintf("❌️ :: Unable to marshal Vote Extensions: %v", err))
			}

			// Append Special Transaction to proposal
			proposalTxs = append(proposalTxs, bz)
		}

		// if the current height does not have vote extensions enabled, skip it

		if req.Height >= ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {
			h.logger.Info(fmt.Sprintf("⚙️ :: Prepare Proposal"))

			// compute the weighted voting power
			weightedsVotingPower, err := h.processWeightedVotingPowerVoteExtensions(ctx, req.LocalLastCommit)
			if err != nil {
				return nil, err
			}

			for key, value := range weightedsVotingPower {
				h.logger.Info(fmt.Sprintf("%s:%d", key, value))
			}

			if err != nil {
				return nil, errors.New("failed to compute weights")
			}

			injectedVoteExtTx := WeightedVotingPower{
				StakeWeightedWeighted: weightedsVotingPower,
				ExtendedCommitInfo:    req.LocalLastCommit,
			}

			bz, err := json.Marshal(injectedVoteExtTx)
			//h.logger.Info(fmt.Sprint(bz))

			if err != nil {
				h.logger.Error("failed to encode injected vote extension tx", "err", err)
				return nil, errors.New("failed to encode injected vote extension tx")
			}

			// Inject the vote extension
			proposalTxs = append(proposalTxs, bz)
		}

		// keep the original txs
		return &abci.ResponsePrepareProposal{
			Txs: proposalTxs,
		}, nil
	}
}

func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		h.logger.Info(fmt.Sprintf("⚙️ :: Process Proposal"))
		if len(req.Txs) == 0 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		}

		var injectedVoteExtTx WeightedVotingPower
		if err := json.Unmarshal(req.Txs[0], &injectedVoteExtTx); err != nil {
			h.logger.Error("failed to decode injected vote extension tx", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}
		//h.logger.Info(fmt.Sprint(injectedVoteExtTx))

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}
func (h *ProposalHandler) processWeightedVotingPowerVoteExtensions(ctx sdk.Context, ci abci.ExtendedCommitInfo) (map[string]int64, error) {

	weightedVoting := make(map[string]int64)

	h.logger.Info(fmt.Sprintf("found %d votes", len(ci.Votes)))

	for _, v := range ci.Votes {
		if v.BlockIdFlag != cmtproto.BlockIDFlagCommit {
			h.logger.Info("skipping BlockIDFlagCommit")
			continue
		}

		h.logger.Info(fmt.Sprintf("Vote Extension: %s", v.String()))
		h.logger.Info(fmt.Sprint(v.VoteExtension))

		if len(v.VoteExtension) <= 0 {
			//h.logger.Error("no vote extensions")
			return nil, nil
		}

		var voteExt WeightedVotingPower
		if err := json.Unmarshal(v.VoteExtension, &voteExt); err != nil {
			h.logger.Error(err.Error())
			h.logger.Error("failed to decode vote extension", "err", err, "validator", fmt.Sprintf("%x", v.Validator.Address))
			return nil, err
		}
	}

	return weightedVoting, nil
}

func (h *ProposalHandler) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	res := &sdk.ResponsePreBlock{}
	if len(req.Txs) == 0 {
		return res, nil
	}

	var injectedVoteExtTx WeightedVotingPowerVoteExtension
	if err := json.Unmarshal(req.Txs[0], &injectedVoteExtTx); err != nil {
		h.logger.Error("failed to decode injected vote extension tx", "err", err)
		return nil, err
	}

	// set weights using the passed in context, which will make these weighted voting power available in the current block
	if err := h.keeper.SetWeights(ctx, injectedVoteExtTx.Weights); err != nil {
		return nil, err
	}

	// handle the weights logic to increase and decrease the voting power of the validators
	for valAddress, weight := range injectedVoteExtTx.Weights {
		h.logger.Info(fmt.Sprintf("%s: %d", valAddress, weight))
		if h.stakingKeeper != nil {
			h.stakingKeeper.SetLastTotalPower(ctx, math.NewInt(weight))
		}
	}

	return res, nil
}
