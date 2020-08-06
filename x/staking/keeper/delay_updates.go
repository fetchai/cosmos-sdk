package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	delayUpdates = true
	validatorUpdateKey = []byte("validatorUpdateKey")
)

func (k Keeper) CheckValidatorUpdates(ctx sdk.Context, header abci.Header) {
	// If two blocks before an next aeon start, need to return new validator set of next aeon
	if !delayUpdates || header.Entropy.GetRound() == header.Entropy.GetAeonLength()-2 {
		store := ctx.KVStore(k.storeKey)
		store.Set(validatorUpdateKey, []byte{0})
	}
}

func (k Keeper) PerformValidatorUpdates(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	if len(store.Get(validatorUpdateKey)) == 0 {
		return false
	}
	store.Set(validatorUpdateKey, []byte{})
	return true
}