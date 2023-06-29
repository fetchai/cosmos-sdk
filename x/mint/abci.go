package mint

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
	"time"
)

type MunicipalInflationCache struct {
	blocksPerYear      uint64
	perBlockInflations map[string]sdk.Dec // {denom: inflationPerBlock}
}

var (
	// NOTE(pb): This is *NOT* thread safe.
	//			 However, in our case, this global variable is by design
	//			 *NOT* supposed to be accessed simultaneously in multiple
	//			 different threads, or in global scope somewhere else.
	//			 Once such requirements arise, concept of this global variable
	//			 needs to be changed to something what is thread safe.
	infCache = MunicipalInflationCache{0, nil}
)

// NOTE(pb): Not thread safe, as per comment above.
func (cache *MunicipalInflationCache) refresh(minter *types.Minter, blocksPerYear uint64) {
	if err := types.ValidateMunicipalInflations(minter.Inflations); err != nil {
		panic(err)
	}

	cache.blocksPerYear = blocksPerYear
	cache.perBlockInflations = map[string]sdk.Dec{}

	for _, inflation := range minter.Inflations {
		inflationPerBlock, err := types.CalculateInflationPerBlock(inflation, blocksPerYear)
		if err != nil {
			panic(err)
		}

		cache.perBlockInflations[inflation.Denom] = inflationPerBlock
	}
}

// NOTE(pb): Not thread safe, as per comment above.
func (cache *MunicipalInflationCache) refreshIfNecessary(minter *types.Minter, blocksPerYear uint64) {
	if infCache.blocksPerYear != blocksPerYear {
		cache.refresh(minter, blocksPerYear)
	}
}

// HandleMunicipalInflation iterates through all other native tokens specified in the Minter.inflations structure, and processes
// the minting of new coins in line with the respective inflation rate of each denomination
func HandleMunicipalInflation(ctx sdk.Context, k keeper.Keeper) {
	minter := k.GetMinter(ctx)
	params := k.GetParams(ctx)

	infCache.refreshIfNecessary(&minter, params.BlocksPerYear)

	// iterate through native denominations
	for _, inflation := range minter.Inflations {
		denom := inflation.Denom
		targetAddress := inflation.TargetAddress

		// gather supply value & calculate number of new tokens created from relevant inflation
		totalDenomSupply := k.BankKeeper.GetSupply(ctx, denom)
		coinsToMint := types.CalculateInflationIssuance(infCache.perBlockInflations[inflation.Denom], totalDenomSupply)

		err := k.MintCoins(ctx, coinsToMint)

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

		err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, acc, coinsToMint)
		if err != nil {
			panic(err)
		}
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeMunicipalMint,
				sdk.NewAttribute(types.AttributeKeyDenom, inflation.Denom),
				sdk.NewAttribute(types.AttributeKeyInflation, inflation.InflationRate.String()),
				sdk.NewAttribute(types.AttributeKeyTargetAddr, inflation.TargetAddress),
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
