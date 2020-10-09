package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"
)

const (
	beaconEvidenceCountKey = "beaconEvidenceCountKey"
)

type BeaconEvidenceInfo struct {
	Threshold int64
	Count     int64
}

func (k Keeper) HandleBeaconInactivity(ctx sdk.Context, evidence types.BeaconInactivity) {
	logger := k.Logger(ctx)
	consAddr := evidence.GetConsensusAddress()
	infractionHeight := evidence.GetHeight()

	// calculate the age of the evidence
	blockTime := ctx.BlockHeader().Time
	age := blockTime.Sub(evidence.GetTime())

	if _, err := k.slashingKeeper.GetPubkey(ctx, consAddr.Bytes()); err != nil {
		// Ignore evidence that cannot be handled.
		return
	}

	// reject evidence it is too old
	if age > k.MaxEvidenceAge(ctx) {
		logger.Info(
			fmt.Sprintf(
				"ignored double sign from %s at height %d, age of %d past max age of %d",
				consAddr, infractionHeight, age, k.MaxEvidenceAge(ctx),
			),
		)
		return
	}

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
	k.deleteBeaconEvidenceCount(height, consAddr)

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

	logger.Info(fmt.Sprintf("confirmed beacon inactivity from %s at height %d, age of %d", consAddr, infractionHeight, age))

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
	store.Set(key(height, address), k.cdc.MustMarshalBinaryLengthPrefixed(newInfo))
}

func (k Keeper) deleteBeaconEvidenceCount(ctx sdk.Context, height int64, address, sdk.ConsAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(key(height,address))
}

func key(height int64, address sdk.ConsAddress) []byte {
	return []byte(fmt.Sprintf("%s/%0.16X/%s", beaconEvidenceCountKey, height, address))
}
