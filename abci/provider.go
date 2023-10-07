package abci

import (
	_ "github.com/ciprianmuja/weight-shift/weightskeeper"
)

// Provider defines an interface for interacting with the external data sources
type Provider struct {
}

// GetValidatorsUptime gets the validator percentage uptime based on the latest N blocks.
func (receiver Provider) GetValidatorsUptime() map[string]int64 {
	// mocked
	return map[string]int64{}
}

// GetValidatorsProposalsVotePercentage gets the percentage of all the voted proposals which a validator have voted on,
// considering the bonding time
func (receiver Provider) GetValidatorsProposalsVotePercentage() map[string]int64 {
	// mocked
	return map[string]int64{}
}
