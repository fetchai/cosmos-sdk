package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/cache"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// HandleMunicipalInflation iterates through all other native tokens specified in the minter.MunicipalInflation structure, and processes
// the minting of new coins in line with the respective inflation rate of each denomination
func HandleMunicipalInflation(minter *types.Minter, params *types.Params, ctx *sdk.Context, k *Keeper) {
	cache.GMunicipalInflationCache.RefreshIfNecessary(&minter.MunicipalInflation, params.BlocksPerYear)

	// iterate through native denominations
	for _, pair := range minter.MunicipalInflation {
		targetAddress := pair.Inflation.TargetAddress

		// gather supply value & calculate number of new tokens created from relevant inflation
		totalDenomSupply := k.bankKeeper.GetSupply(*ctx, pair.Denom)

		cacheItem := cache.GMunicipalInflationCache.GetInflation(pair.Denom)

		if cacheItem == nil {
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

		err = k.bankKeeper.SendCoinsFromModuleToAccount(*ctx, types.ModuleName, acc, coinsToMint)
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

// MintFn defines the function that needs to be implemented in order to customize the minting process.
type MintFn func(ctx sdk.Context, k *Keeper) error

// MintFn runs the mintFn of the keeper.
func (k *Keeper) MintFn(ctx sdk.Context) error {
	return k.mintFn(ctx, k)
}

// DefaultMintFn returns a default mint function.
// The default MintFn has a requirement on staking as it uses bond to calculate inflation.
func DefaultMintFn(ic types.InflationCalculationFn) MintFn {
	return func(ctx sdk.Context, k *Keeper) error {
		// fetch stored minter & params
		minter, err := k.Minter.Get(ctx)
		if err != nil {
			return err
		}

		params, err := k.Params.Get(ctx)
		if err != nil {
			return err
		}

		HandleMunicipalInflation(&minter, &params, &ctx, k)

		// recalculate inflation rate
		totalStakingSupply, err := k.StakingTokenSupply(ctx)
		if err != nil {
			return err
		}

		bondedRatio, err := k.BondedRatio(ctx)
		if err != nil {
			return err
		}

		minter.Inflation = ic(ctx, minter, params, bondedRatio)
		minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalStakingSupply)
		if err = k.Minter.Set(ctx, minter); err != nil {
			return err
		}

		// mint coins, update supply
		mintedCoin := minter.BlockProvision(params)
		mintedCoins := sdk.NewCoins(mintedCoin)

		err = k.MintCoins(ctx, mintedCoins)
		if err != nil {
			return err
		}

		// send the minted coins to the fee collector account
		err = k.AddCollectedFees(ctx, mintedCoins)
		if err != nil {
			return err
		}

		if mintedCoin.Amount.IsInt64() {
			defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMint,
				sdk.NewAttribute(types.AttributeKeyBondedRatio, bondedRatio.String()),
				sdk.NewAttribute(types.AttributeKeyInflation, minter.Inflation.String()),
				sdk.NewAttribute(types.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
			),
		)

		return nil
	}
}
