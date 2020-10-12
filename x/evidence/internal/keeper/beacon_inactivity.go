package keeper

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"

	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	beaconEvidenceCountKey = []byte("beaconEvidenceCountKey")
	keysListKey            = []byte("keysListKey")
)

type BeaconEvidenceInfo struct {
	Threshold int64
	Count     int64
}

// HandleBeaconInactivity keeps track of the number of complaints against a validator for each height
// and triggers slashing of the validators stake if over threshold number of complaints is reached
func (k Keeper) HandleBeaconInactivity(ctx sdk.Context, evidence types.BeaconInactivity) {
	logger := k.Logger(ctx)
	consAddr := evidence.GetConsensusAddress()
	infractionHeight := evidence.GetHeight()

	// Add to evidence count and check threshold
	evidenceInfo := k.getBeaconEvidenceCount(ctx, infractionHeight, consAddr)
	if evidenceInfo.Count == 0 {
		evidenceInfo.Threshold = evidence.Threshold
	}
	evidenceInfo.Count++
	k.setBeaconEvidenceCount(ctx, infractionHeight, consAddr, evidenceInfo)
	if evidenceInfo.Count <= evidenceInfo.Threshold {
		logger.Info("BeaconEvidence: insufficient complaints", "address", fmt.Sprintf("%s", consAddr), "count",
			evidenceInfo.Count, "required", evidenceInfo.Threshold+1)
		return
	}
	// Delete evidence count info
	k.deleteBeaconEvidenceCount(ctx, infractionHeight, consAddr)

	validator := k.stakingKeeper.ValidatorByConsAddr(ctx, consAddr)
	if validator == nil || validator.IsUnbonded() {
		return
	}

	if ok := k.slashingKeeper.HasValidatorSigningInfo(ctx, consAddr); !ok {
		panic(fmt.Sprintf("expected signing info for validator %s but not found", consAddr))
	}

	// ignore if the validator is already tombstoned
	if k.slashingKeeper.IsTombstoned(ctx, consAddr) {
		logger.Info(
			fmt.Sprintf(
				"ignored double sign from %s at height %d, validator already tombstoned",
				consAddr, infractionHeight,
			),
		)
		return
	}

	logger.Info("BeaconEvidence: complaint successful", "height", infractionHeight, "address", fmt.Sprintf("%s", consAddr))

	// We need to retrieve the stake distribution which signed the block, so we
	// subtract ValidatorUpdateDelay from the evidence height.
	distributionHeight := infractionHeight - sdk.ValidatorUpdateDelay

	k.slashingKeeper.Slash(
		ctx,
		consAddr,
		k.slashingKeeper.SlashFractionBeaconInactivity(ctx),
		evidence.GetValidatorPower(), distributionHeight,
	)
}

func (k Keeper) getKeysList(ctx sdk.Context) []string {
	store := ctx.KVStore(k.storeKey)
	keysListBytes := store.Get(keysListKey)
	keysList := []string{}
	k.cdc.UnmarshalBinaryLengthPrefixed(keysListBytes, &keysList)
	return keysList
}

func (k Keeper) setKeysList(ctx sdk.Context, keysList []string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(keysListKey, k.cdc.MustMarshalBinaryLengthPrefixed(keysList))
}

func (k Keeper) getBeaconEvidenceCount(ctx sdk.Context, height int64, address sdk.ConsAddress) BeaconEvidenceInfo {
	store := ctx.KVStore(k.storeKey)
	countBytes := store.Get(key(height, address))
	evidenceCount := BeaconEvidenceInfo{}
	k.cdc.UnmarshalBinaryLengthPrefixed(countBytes, &evidenceCount)
	return evidenceCount
}

func (k Keeper) setBeaconEvidenceCount(ctx sdk.Context, height int64, address sdk.ConsAddress,
	newInfo BeaconEvidenceInfo) {
	store := ctx.KVStore(k.storeKey)
	infoKey := key(height, address)
	// If key is new then save it
	if !store.Has(infoKey) {
		keysList := k.getKeysList(ctx)
		keysList = append(keysList, string(infoKey))
		k.setKeysList(ctx, keysList)
	}
	store.Set(key(height, address), k.cdc.MustMarshalBinaryLengthPrefixed(newInfo))
}

func (k Keeper) deleteBeaconEvidenceCount(ctx sdk.Context, height int64, address sdk.ConsAddress) {
	store := ctx.KVStore(k.storeKey)
	infoKey := key(height, address)
	store.Delete(infoKey)
	// Remove from keys list
	keysList := k.getKeysList(ctx)
	for index, key := range keysList {
		if key == string(infoKey) {
			keysList[index] = keysList[len(keysList)-1]
			keysList = keysList[:len(keysList)-1]
			break
		}
	}
}

func key(height int64, address sdk.ConsAddress) []byte {
	return []byte(fmt.Sprintf("%s/%v/%s", beaconEvidenceCountKey, height, address))
}

// PruneBeaconEvidence prunes the evidence counts stored
func (k Keeper) PruneBeaconEvidence(ctx sdk.Context, height int64) {
	logger := k.Logger(ctx)
	store := ctx.KVStore(k.storeKey)

	// Get max evidence age either from consensus params or default values
	var maxEvidenceAge int64
	if consensusParams := ctx.ConsensusParams(); consensusParams != nil {
		maxEvidenceAge = consensusParams.Evidence.GetMaxAgeNumBlocks()
	} else {
		defaultMaxAge := tmtypes.DefaultEvidenceParams().MaxAgeNumBlocks
		logger.Error(fmt.Sprintf("PruneBeaconEvidence could not get consensus params. Using default %v",
			defaultMaxAge), "height", height)
		maxEvidenceAge = defaultMaxAge
	}

	// Find keys in keysList which have heights that are too old and remove them
	keysList := k.getKeysList(ctx)
	for i := 0; i < len(keysList); i++ {
		key := keysList[i]
		height, err := strconv.ParseInt(strings.Split(key, "/")[1], 10, 64)
		if err != nil {
			logger.Error("PruneBeaconEvidence: could not extract height from key", "key", key)
			continue
		}
		if height+maxEvidenceAge <= height {
			// Delete evidence info with this key from store and remove from keys list
			store.Delete([]byte(key))
			keysList = append(keysList[:i], keysList[i+1:]...)
		}
	}
}
