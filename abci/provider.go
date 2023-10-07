package abci

import (
	_ "github.com/ciprianmuja/weight-shift/weightskeeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
)

// Provider defines an interface for interacting with the external data sources
type Provider struct {
}

// GetValidatorsUptime gets the validator percentage uptime based on the latest N blocks.
func (receiver Provider) GetValidatorsUptime(ctx sdk.Context, gov govkeeper.Keeper) map[string]int64 {
	// mocked
	uptimes := map[string]int64{}
	uptimes["val1"] = 40
	uptimes["val2"] = 10
	return uptimes
}

// GetValidatorsProposalsVotePercentage gets the percentage of all the voted proposals which a validator have voted on,
func (receiver Provider) GetValidatorsProposalsVotePercentage(ctx sdk.Context, gov govkeeper.Keeper) map[string]int64 {
	// mocked
	proposals := map[string]int64{}
	proposals["val1"] = 40
	proposals["val2"] = 10
	return proposals
}

// GetValidatorsGitHubContributions gets the amount of contributions that a validator have made on the main repo,
func (receiver Provider) GetValidatorsGitHubContributions(ctx sdk.Context) map[string]int64 {
	// mocked
	activity := map[string]int64{}
	activity["val1"] = 2
	activity["val2"] = 100
	return activity
}
