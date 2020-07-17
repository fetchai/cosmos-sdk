package staking

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker will persist the current header and validator set as a historical entry
// and prune the oldest entry based on the HistoricalEntries parameter
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k Keeper) {
	defer telemetry.ModuleMeasureSince(ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	k.TrackHistoricalInfo(ctx)
	k.CheckValidatorUpdates(ctx, req.Header)
}

// Called every block, update validator set
func EndBlocker(ctx sdk.Context, k Keeper) ([]abci.ValidatorUpdate, []abci.ValidatorUpdate) {
	defer telemetry.ModuleMeasureSince(ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	// If dkg and validator updates are triggered at the same time then dkg validator updates
	// must be computed first
	dkgUpdates := k.DKGValidatorUpdates(ctx)
	return k.ValidatorUpdates(ctx), dkgUpdates
}
