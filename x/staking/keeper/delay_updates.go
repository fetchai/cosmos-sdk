package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	validatorUpdateKey = []byte("validatorUpdateKey")
)

// Check whether entropy generation round corresponds to validator changeover height
func (k Keeper) CheckValidatorUpdates(ctx sdk.Context, header abci.Header) {
	// If two blocks before an next aeon start, need to return new validator set of next aeon
	if header.Entropy.GetRound() == header.Entropy.GetAeonLength()-2 {
		store := ctx.KVStore(k.storeKey)
		store.Set(validatorUpdateKey, []byte{0})
	}
}

// Tells EndBlocker whether to compute validator updates using variable set in BeginBlocker
func (k Keeper) PerformValidatorUpdates(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	if len(store.Get(validatorUpdateKey)) == 0 {
		return false
	}
	store.Set(validatorUpdateKey, []byte{})
	return true
}