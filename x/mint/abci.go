package mint

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/cache"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// HandleMunicipalInflation iterates through all other native tokens specified in the minter.MunicipalInflation structure, and processes
// the minting of new coins in line with the respective inflation rate of each denomination
func HandleMunicipalInflation(minter *types.Minter, params *types.Params, ctx *sdk.Context, k *keeper.Keeper) {
	cache.GMunicipalInflationCache.RefreshIfNecessary(&minter.MunicipalInflation, params.BlocksPerYear)

	// iterate through native denominations
	for _, pair := range minter.MunicipalInflation {
		targetAddress := pair.Inflation.TargetAddress

		// gather supply value & calculate number of new tokens created from relevant inflation
		totalDenomSupply := k.BankKeeper.GetSupply(*ctx, pair.Denom)

		cacheItem, exists := cache.GMunicipalInflationCache.GetInflation(pair.Denom)

		if !exists {
			panic(fmt.Errorf("numicipal inflation: missing cache item for the \"%s\" denomination", pair.Denom))
		}

		coinsToMint := types.CalculateInflationIssuance(cacheItem.PerBlockInflation, totalDenomSupply)

		err := k.MintCoins(*ctx, coinsToMint)
		if err != nil {
			panic(err)
		}

		// send these new tokens to respective target address
		// TODO(JS): investigate whether this should be carried out in distribution module or not

		// Convert targetAddress to sdk.AccAddress
		acc, err := sdk.AccAddressFromBech32(targetAddress)
		if err != nil {
			panic(err)
		}

		err = k.BankKeeper.SendCoinsFromModuleToAccount(*ctx, types.ModuleName, acc, coinsToMint)
		if err != nil {
			panic(err)
		}
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMunicipalMint,
				sdk.NewAttribute(types.AttributeKeyDenom, pair.Denom),
				sdk.NewAttribute(types.AttributeKeyInflation, pair.Inflation.Value.String()),
				sdk.NewAttribute(types.AttributeKeyTargetAddr, pair.Inflation.TargetAddress),
				sdk.NewAttribute(sdk.AttributeKeyAmount, coinsToMint.String()),
			),
		)
	}
}

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored minter & params
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	HandleMunicipalInflation(&minter, &params, &ctx, &k)

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
