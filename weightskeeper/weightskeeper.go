package weightskeeper

import (
	"context"
	weight_shift "github.com/ciprianmuja/weight-shift"
	"github.com/cosmos/cosmos-sdk/x/gov/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
)

type WeightsKeeper struct {
	cdc          codec.BinaryCodec
	addressCodec address.Codec
	authority    string

	// state management
	Schema        collections.Schema
	Weights       collections.Map[string, int64]
	stakingKeeper types.StakingKeeper
}

// NewWeightsKeeper creates a new Keeper instance
func NewWeightsKeeper(cdc codec.BinaryCodec, addressCodec address.Codec, storeService storetypes.KVStoreService) WeightsKeeper {

	sb := collections.NewSchemaBuilder(storeService)
	k := WeightsKeeper{
		cdc:          cdc,
		addressCodec: addressCodec,
		Weights:      collections.NewMap(sb, weight_shift.WeightsKey, "weights", collections.StringKey, collections.Int64Value),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}

	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k WeightsKeeper) GetAuthority() string {
	return k.authority
}

func (k WeightsKeeper) GetWeights(ctx context.Context) (map[string]int64, error) {
	var weights map[string]int64
	err := k.Weights.Walk(ctx, nil, func(key string, value int64) (bool, error) {
		weights[key] = value
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return weights, nil
}

func (k WeightsKeeper) SetWeights(ctx context.Context, weights map[string]int64) error {
	for b, q := range weights {
		err := k.Weights.Set(ctx, b, q)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k WeightsKeeper) GetValidatorsUptime(ctx context.Context) {

}
