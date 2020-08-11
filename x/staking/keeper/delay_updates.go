package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	computeValidatorUpdateKey    = []byte("computeValidatorUpdateKey")
	computeDKGValidatorUpdateKey = []byte("computeDKGValidatorUpdateKey")
	validatorUpdatesKey          = []byte("validatorUpdatesKey")
)

// Check whether entropy generation round corresponds to validator changeover height
func (k Keeper) CheckValidatorUpdates(ctx sdk.Context, header abci.Header) {
	// One blocks before a new aeon start need to compute validator updates for next dkg committee.
	// Two blocks before a new aeon start need to update consensus committee to those which ran dkg
	nextAeonStart := header.Entropy.NextAeonStart
	if !k.delayValidatorUpdates || header.Height == nextAeonStart-1 {
		store := ctx.KVStore(k.storeKey)
		store.Set(computeDKGValidatorUpdateKey, []byte{0})
	}
	if !k.delayValidatorUpdates || header.Height == nextAeonStart-2 {
		store := ctx.KVStore(k.storeKey)
		store.Set(computeValidatorUpdateKey, []byte{0})
	}
}

// DKGValidatorUpdates returns dkg validator updates to EndBlock at block height set by BeginBlock and
// saves them to store for retrieval by ValidatorUpdates
func (k Keeper) DKGValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate {
	store := ctx.KVStore(k.storeKey)
	if len(store.Get(computeDKGValidatorUpdateKey)) == 0 {
		return []abci.ValidatorUpdate{}
	}
	store.Set(computeDKGValidatorUpdateKey, []byte{})
	updates := k.BlockValidatorUpdates(ctx)
	store.Set(validatorUpdatesKey, k.cdc.MustMarshalBinaryLengthPrefixed(updates))
	return updates
}

// ValidatorUpdates retrieve last saved updates from store when non-trivial update
// is triggered by BeginBlock,
func (k Keeper) ValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate {
	store := ctx.KVStore(k.storeKey)
	if len(store.Get(computeValidatorUpdateKey)) == 0 {
		return []abci.ValidatorUpdate{}
	}
	store.Set(computeValidatorUpdateKey, []byte{})
	updateBytes := store.Get(validatorUpdatesKey)
	updates := []abci.ValidatorUpdate{}
	k.cdc.UnmarshalBinaryLengthPrefixed(updateBytes, &updates)
	return updates
}
