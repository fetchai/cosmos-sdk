package mint

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
	"time"
)

// HandleInflations iterates through all other native tokens specified in the Minter.inflations structure, and processes
// the minting of new coins in line with the respective inflation rate of each denomination
func HandleInflations(ctx sdk.Context, k keeper.Keeper) {
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// validate inflations
	err := types.ValidateInflations(minter.Inflations)
	if err != nil {
		panic(err)
	}

	// iterate through other native denominations
	for _, inflation := range minter.Inflations {
		denom := inflation.Denom
		targetAddress := inflation.TargetAddress

		// gather supply value & calculate number of new tokens created from relevant inflation
		totalDenomSupply := k.BankKeeper.GetSupply(ctx, denom)

		// mint these new tokens
		mintedCoins, err := types.CalculateInflation(inflation, params.BlocksPerYear, totalDenomSupply)
		if err != nil {
			panic(err)
		}

		err = k.MintCoins(ctx, mintedCoins)
		if err != nil {
			panic(err)
		}

		// send these new tokens to respective target address
		// TODO(JS): investigate whether this should be carried out in distribution module or not
		err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.AccAddress(targetAddress), mintedCoins)
		if err != nil {
			panic(err)
		}
	}
}

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	// recalculate inflation rate
	totalStakingSupply := k.StakingTokenSupply(ctx)
	minter.Inflation = minter.NextInflationRate(params)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalStakingSupply)
	k.SetMinter(ctx, minter)

	// mint coins, update supply
	mintedCoin := minter.BlockProvision(params)
	mintedCoins := sdk.NewCoins(mintedCoin)

	err := k.MintCoins(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	// send the minted coins to the fee collector account
	err = k.AddCollectedFees(ctx, mintedCoins)
	if err != nil {
		panic(err)
	}

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.AttributeKeyInflation, minter.Inflation.String()),
			sdk.NewAttribute(types.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)
}
