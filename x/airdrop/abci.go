package airdrop

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/airdrop/keeper"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	"time"
)

func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	_, err := k.DripAllFunds(ctx)
	if err != nil {
		ctx.Logger().Error("Unable to perform airdrop drip", "err", err.Error())
	}
}
