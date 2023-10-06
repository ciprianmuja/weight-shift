package abci

import (
	"cosmossdk.io/log"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ciprianmuja/weight-shift/weightskeeper"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WeightedVotingPower defines the structure a proposer should use to calculate
// and submit the weighted voting power for the given validator set
type WeightedVotingPower struct {
	StakeWeightedPrices map[string]int64
	ExtendedCommitInfo  abci.ExtendedCommitInfo
}

type ProposalHandler struct {
	logger   log.Logger
	keeper   weightskeeper.WeightsKeeper
	valStore baseapp.ValidatorStore
}

func NewPrepareProposalHandler(logger log.Logger, keeper weightskeeper.WeightsKeeper, valStore baseapp.ValidatorStore) *ProposalHandler {
	return &ProposalHandler{
		logger:   logger,
		keeper:   keeper,
		valStore: valStore,
	}
}

func (h *ProposalHandler) PrepareProposal() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		h.logger.Info(fmt.Sprintf("⚙️ :: Prepare Proposal"))
		//TODO: commented for the moment, still have to understand how to pass the h.valStore
		/*err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), req.LocalLastCommit)
		if err != nil {
			return nil, err
		}*/

		proposalTxs := req.Txs

		// if the current height does not have vote extensions enabled, skip it
		if req.Height >= ctx.ConsensusParams().Abci.VoteExtensionsEnableHeight {

			// compute the weighted voting power
			weightedsVotingPower, err := h.processWeightedVotingPowerVoteExtensions(ctx, req.LocalLastCommit)
			if err != nil {
				return nil, err
			}

			for key, value := range weightedsVotingPower {
				h.logger.Info(fmt.Sprintf("%s:%d", key, value))
			}

			if err != nil {
				return nil, errors.New("failed to compute stake-weighted oracle prices")
			}

			injectedVoteExtTx := WeightedVotingPower{
				StakeWeightedPrices: weightedsVotingPower,

				ExtendedCommitInfo: req.LocalLastCommit,
			}

			bz, err := json.Marshal(injectedVoteExtTx)
			if err != nil {
				h.logger.Error("failed to encode injected vote extension tx", "err", err)
				return nil, errors.New("failed to encode injected vote extension tx")
			}

			// Inject a "fake" tx into the proposal s.t. validators can decode, verify,
			// and store the canonical stake-weighted average prices.
			proposalTxs = append(proposalTxs, bz)
		}

		// keep the original txs
		return &abci.ResponsePrepareProposal{
			Txs: req.Txs,
		}, nil
	}
}

func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		if len(req.Txs) == 0 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
		}

		var injectedVoteExtTx WeightedVotingPower
		if err := json.Unmarshal(req.Txs[0], &injectedVoteExtTx); err != nil {
			h.logger.Error("failed to decode injected vote extension tx", "err", err)
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}

		err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), injectedVoteExtTx.ExtendedCommitInfo)
		if err != nil {
			return nil, err
		}

		// Verify the proposer's stake-weighted oracle prices by computing the same
		// calculation and comparing the results. We omit verification for brevity
		// and demo purposes.
		//stakeWeightedPrices, err := h.computeStakeWeightedOraclePrices(ctx, injectedVoteExtTx.ExtendedCommitInfo)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}
		/*if err := compareOraclePrices(injectedVoteExtTx., nil); err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
		}*/

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

func (h *ProposalHandler) processWeightedVotingPowerVoteExtensions(ctx sdk.Context, ci abci.ExtendedCommitInfo) (map[string]int64, error) {

	stakeWeightedPrices := make(map[string]int64)

	for _, v := range ci.Votes {
		if v.BlockIdFlag != cmtproto.BlockIDFlagCommit {
			continue
		}

		h.logger.Info(fmt.Sprint(v.VoteExtension))

		var voteExt WeightedVotingPowerVoteExtension
		if err := json.Unmarshal(v.VoteExtension, &voteExt); err != nil {
			h.logger.Error("failed to decode vote extension", "err", err, "validator", fmt.Sprintf("%x", v.Validator.Address))
			return nil, err
		}

		// Compute stake-weighted average of prices for each supported pair, i.e.
		// (P1)(W1) + (P2)(W2) + ... + (Pn)(Wn) / (W1 + W2 + ... + Wn)
		//
		// NOTE: These are the prices computed at the PREVIOUS height, i.e. H-1
		for base, price := range voteExt.Weights {
			// Only compute stake-weighted average for supported pairs.
			//
			// NOTE: VerifyVoteExtension should be sufficient to ensure that only
			// supported pairs are supplied, but we add this here for demo purposes.
			if _, ok := stakeWeightedPrices[base]; ok {
				stakeWeightedPrices[base] = price
			}
		}
	}

	return stakeWeightedPrices, nil
}
