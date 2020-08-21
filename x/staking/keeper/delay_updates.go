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

// CheckValidatorUpdates determines whether block height is sufficiently close to the next aeon start
// to trigger dkg and consensus validator changes
func (k Keeper) CheckValidatorUpdates(ctx sdk.Context, header abci.Header) {
	// One block before a new aeon start need to compute validator updates for next dkg committee.
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
	// Calculate validator set changes.
	//
	// NOTE: ApplyAndReturnValidatorSetUpdates has to come before
	// UnbondAllMatureValidatorQueue.
	// This fixes a bug when the unbonding period is instant (is the case in
	// some of the tests). The test expected the validator to be completely
	// unbonded after the Endblocker (go from Bonded -> Unbonding during
	// ApplyAndReturnValidatorSetUpdates and then Unbonding -> Unbonded during
	// UnbondAllMatureValidatorQueue).
	updates := k.ApplyAndReturnValidatorSetUpdates(ctx)
	store.Set(validatorUpdatesKey, k.cdc.MustMarshalBinaryLengthPrefixed(updates))
	return updates
}

// ValidatorUpdates retrieve last saved updates from store when non-trivial update
// is triggered by BeginBlock,
func (k Keeper) ValidatorUpdates(ctx sdk.Context) []abci.ValidatorUpdate {
	store := ctx.KVStore(k.storeKey)
	if len(store.Get(computeValidatorUpdateKey)) == 0 {
		// Check mature items in queues every block, regardless of whether return validator updates
		// or not, in order for items to be removed as soon as possible
		k.RemoveMatureQueueItems(ctx)
		return []abci.ValidatorUpdate{}
	}
	store.Set(computeValidatorUpdateKey, []byte{})
	updateBytes := store.Get(validatorUpdatesKey)
	updates := []abci.ValidatorUpdate{}
	k.cdc.UnmarshalBinaryLengthPrefixed(updateBytes, &updates)
	k.ExecuteUnbonding(ctx, updates)
	k.RemoveMatureQueueItems(ctx)
	return updates
}
