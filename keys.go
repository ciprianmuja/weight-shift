package weight_shift

import "cosmossdk.io/collections"

const (
	ModuleName = "ws"
	StoreKey   = "ws"
)

var (
	WeightsKey = collections.NewPrefix(0)
)
