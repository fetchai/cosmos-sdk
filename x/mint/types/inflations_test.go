package types_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"math"
	"math/rand"
	"strconv"
	"testing"
)

func resetSupply(app *simapp.SimApp, ctx sdk.Context, initSupply sdk.Coins, curSupply sdk.Coins) {
	err := app.BankKeeper.MintCoins(ctx, types.ModuleName, initSupply)
	if err != nil {
		panic(err)
	}
	err = app.BankKeeper.BurnCoins(ctx, types.ModuleName, curSupply)
	if err != nil {
		panic(err)
	}
}

func getTestingAccounts(r *rand.Rand, n int, ctx sdk.Context, app *simapp.SimApp) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	for _, account := range accounts {
		acc := app.AccountKeeper.NewAccountWithAddress(ctx, account.Address)
		app.AccountKeeper.SetAccount(ctx, acc)
	}

	return accounts
}

func TestHandleInflations(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	acc := auth.NewEmptyModuleAccount(types.ModuleName, auth.Minter, auth.Burner)
	app.AccountKeeper.SetModuleAccount(ctx, acc)

	s := rand.NewSource(1)
	r := rand.New(s)
	targetAccounts := getTestingAccounts(r, 3, ctx, app)

	minter := types.DefaultInitialMinter()
	params := types.DefaultParams()

	app.MintKeeper.SetParams(ctx, params)
	app.MintKeeper.SetMinter(ctx, minter)

	testDenom := "testDenom"
	initSupplyAmount, _ := sdk.NewIntFromString("1000000000000000000000000000")
	initSupplyCoin := sdk.NewCoin(testDenom, initSupplyAmount)

	tests := []struct {
		inflation       types.Inflation
		expectedBalance sdk.Int
	}{
		{types.NewInflation(testDenom, targetAccounts[0].Address.String(), sdk.NewDecWithPrec(1, 2)), sdk.NewInt(1576534791000000000)},
		{types.NewInflation(testDenom, targetAccounts[1].Address.String(), sdk.NewDecWithPrec(2, 2)), sdk.NewInt(3137536969000000000)},
		{types.NewInflation(testDenom, targetAccounts[2].Address.String(), sdk.NewDecWithPrec(3, 2)), sdk.NewInt(4683309617000000000)},
	}
	for _, tc := range tests {
		minter.Inflations = []*types.Inflation{&tc.inflation}
		// Mint initial supply
		resetSupply(app, ctx, sdk.NewCoins(initSupplyCoin), sdk.NewCoins(app.BankKeeper.GetSupply(ctx, testDenom)))

		// Calculate inflation
		inflationRatePerBlock, err := types.CalculateInflationPerBlock(&tc.inflation, params.BlocksPerYear)
		if err != nil {
			panic(err)
		}

		newCoinsToSend := types.CalculateInflationNewCoins(inflationRatePerBlock, app.BankKeeper.GetSupply(ctx, testDenom))
		require.Equal(t, tc.expectedBalance.Int64(), newCoinsToSend.AmountOf(testDenom).Int64())

		// Mint inflation returns
		require.NoError(t, app.BankKeeper.MintCoins(ctx, types.ModuleName, newCoinsToSend))

		// Convert targetAddress to sdk.AccAddress
		acc, err := sdk.AccAddressFromBech32(tc.inflation.TargetAddress)
		require.NoError(t, err)

		// Send new reward tokens to target address
		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, acc, newCoinsToSend)
		require.NoError(t, err)

		// Assert tokens reached account
		testAccountBalance := app.BankKeeper.GetBalance(ctx, acc, tc.inflation.Denom)
		require.Equal(t, tc.expectedBalance, testAccountBalance.Amount)
	}
}

func TestInflationsValidation(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	targetAccounts := simtypes.RandomAccounts(r, 1)
	minter := types.DefaultInitialMinter()

	tests := []struct {
		inflation      types.Inflation
		expectedToPass bool
	}{
		// Pass: 100% inflation
		{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.OneDec()), true},
		// Pass: 1% inflation
		{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(1, 2)), true},
		// Pass: -1% inflation
		{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(1, 2).Neg()), true},
		// Fail: -1.1% inflation
		{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(11, 3).Neg()), false},
		// Fail: invalid denom
		{types.NewInflation("!&£$%", targetAccounts[0].Address.String(), sdk.OneDec()), false},
		{types.NewInflation("", targetAccounts[0].Address.String(), sdk.OneDec()), false},
		// Fail: invalid targetAddress
		{types.NewInflation("stake", "fetch123abc", sdk.OneDec()), false},
		{types.NewInflation("stake", "", sdk.OneDec()), false},
	}
	for _, tc := range tests {
		minter.Inflations = []*types.Inflation{&tc.inflation}
		if tc.expectedToPass {
			require.NoError(t, types.ValidateInflations(minter.Inflations))
		} else {
			require.Error(t, types.ValidateInflations(minter.Inflations))
		}
	}
}

func TestInflationsCalculations(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	acc := auth.NewEmptyModuleAccount(types.ModuleName, auth.Minter, auth.Burner)
	app.AccountKeeper.SetModuleAccount(ctx, acc)
	keeper := app.MintKeeper

	s := rand.NewSource(1)
	r := rand.New(s)
	targetAccounts := getTestingAccounts(r, 6, ctx, app)

	minter := types.DefaultInitialMinter()
	params := types.DefaultParams()

	testDenom := "testDenom"
	initSupplyAmount, _ := sdk.NewIntFromString("1000000000000000000000000000")
	initSupplyCoin := sdk.NewCoin(testDenom, initSupplyAmount)
	params.BlocksPerYear = 10000
	keeper.SetParams(ctx, params)

	tests := []struct {
		inflation types.Inflation
	}{
		{types.NewInflation(testDenom, targetAccounts[0].Address.String(), sdk.NewDecWithPrec(1, 2))},
		{types.NewInflation(testDenom, targetAccounts[1].Address.String(), sdk.NewDecWithPrec(5, 2))},
		{types.NewInflation(testDenom, targetAccounts[2].Address.String(), sdk.NewDecWithPrec(25, 2))},
		{types.NewInflation(testDenom, targetAccounts[3].Address.String(), sdk.NewDecWithPrec(50, 2))},
		{types.NewInflation(testDenom, targetAccounts[4].Address.String(), sdk.NewDecWithPrec(75, 2))},
		{types.NewInflation(testDenom, targetAccounts[5].Address.String(), sdk.NewDecWithPrec(100, 2))},
	}
	for _, tc := range tests {
		resetSupply(app, ctx, sdk.NewCoins(initSupplyCoin), sdk.NewCoins(keeper.BankKeeper.GetSupply(ctx, testDenom)))
		minter.Inflations = []*types.Inflation{&tc.inflation}
		keeper.SetMinter(ctx, minter)

		for i := 0; i < int(params.BlocksPerYear); i++ {
			mint.HandleInflations(ctx, keeper)
		}

		inflatedSupplyA := (keeper.BankKeeper.GetSupply(ctx, testDenom).Amount).Sub(initSupplyAmount)
		inflatedSupplyB := tc.inflation.InflationRate.MulInt64(int64(params.BlocksPerYear))

		inflationPerYear := tc.inflation.InflationRate
		inflationPerBlock, err := types.CalculateInflationPerBlock(&tc.inflation, params.BlocksPerYear)
		if err != nil {
			panic(err)
		}

		acquiredInfPerYear := math.Pow(1+inflationPerBlock.MustFloat64(), float64(params.BlocksPerYear)) - 1
		acquiredInfPerYearStr := strconv.FormatFloat(acquiredInfPerYear, 'f', 18, 64)
		acquiredInfPerYearDec, err := sdk.NewDecFromStr(acquiredInfPerYearStr)
		if err != nil {
			panic(err)
		}

		account, err := sdk.AccAddressFromBech32(tc.inflation.TargetAddress)
		if err != nil {
			panic(err)
		}

		testAccountBalance := app.BankKeeper.GetBalance(ctx, account, tc.inflation.Denom).Amount

		deltaInflation := (acquiredInfPerYearDec.Quo(inflationPerYear)).Sub(sdk.OneDec()).Abs()
		deltaSupply := inflatedSupplyB.QuoInt(inflatedSupplyA)

		// ensure target address is funded appropriately
		require.True(t, testAccountBalance.Equal(inflatedSupplyA))
		// ensure difference between actual inflated supply and theoretical is acceptable
		require.True(t, deltaSupply.LT(sdk.NewDecWithPrec(1, 8)))
		// ensure difference between actual inflation and theoretical is acceptable
		require.True(t, deltaInflation.LT(sdk.NewDecWithPrec(1, 8)))
	}
}
