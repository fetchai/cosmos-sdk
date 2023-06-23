package types_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"math"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func CalculateInflationB(inflation *types.Inflation, blocksPerYear uint64, supply sdk.Coin) (result sdk.Coins, err error) {
	inflationPerBlock := math.Pow((inflation.InflationRate.Add(sdk.OneDec())).MustFloat64(), 1/float64(blocksPerYear)) - 1
	inflationPerBlockStr := strconv.FormatFloat(inflationPerBlock, 'f', 18, 64)
	inflationPerBlockDec, err := sdk.NewDecFromStr(inflationPerBlockStr)
	if err != nil {
		return nil, err
	}
	reward := inflationPerBlockDec.MulInt(supply.Amount)
	result = sdk.NewCoins(sdk.NewCoin(inflation.Denom, reward.TruncateInt()))
	return result, err
}

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
	initSupplyAmount, _ := sdk.NewIntFromString("1000000000000000000")
	initSupplyCoin := sdk.NewCoin(testDenom, initSupplyAmount)

	tests := []struct {
		inflation       types.Inflation
		expectedBalance sdk.Int
	}{
		{types.NewInflation(testDenom, targetAccounts[0].Address.String(), sdk.NewDecWithPrec(1, 2)), sdk.NewInt(1576534791)},
		{types.NewInflation(testDenom, targetAccounts[1].Address.String(), sdk.NewDecWithPrec(2, 2)), sdk.NewInt(3137536969)},
		{types.NewInflation(testDenom, targetAccounts[2].Address.String(), sdk.NewDecWithPrec(3, 2)), sdk.NewInt(4683309617)},
	}
	for i, tc := range tests {
		minter.Inflations = []*types.Inflation{&tc.inflation}
		// Mint initial supply
		resetSupply(app, ctx, sdk.NewCoins(initSupplyCoin), sdk.NewCoins(app.BankKeeper.GetSupply(ctx, testDenom)))

		// Calculate inflation
		newCoinsToSend, err := types.CalculateInflation(&tc.inflation, params.BlocksPerYear, app.BankKeeper.GetSupply(ctx, tc.inflation.Denom))
		require.NoError(t, err)
		require.Equal(t, tc.expectedBalance.Int64(), newCoinsToSend.AmountOf(testDenom).Int64())

		// Mint inflation returns
		require.NoError(t, app.BankKeeper.MintCoins(ctx, types.ModuleName, newCoinsToSend))

		// Send new reward tokens to target address
		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, targetAccounts[i].Address, newCoinsToSend)
		require.NoError(t, err)

		// Assert tokens reached account
		testAccountBalance := app.BankKeeper.GetBalance(ctx, targetAccounts[i].Address, tc.inflation.Denom)
		require.Equal(t, tc.expectedBalance.Int64(), testAccountBalance.Amount.Int64())
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
		// Fail: -1% inflation
		{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDec(-1)), false},
		// Fail: invalid denom
		{types.NewInflation("!&£$%", targetAccounts[0].Address.String(), sdk.OneDec()), false},
		// Fail: invalid targetAddress
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

	s := rand.NewSource(1)
	r := rand.New(s)
	targetAccounts := simtypes.RandomAccounts(r, 1) // create test account

	minter := types.DefaultInitialMinter()
	params := types.DefaultParams()

	app.MintKeeper.SetParams(ctx, params)
	app.MintKeeper.SetMinter(ctx, minter)

	testDenom := "testDenom"
	initSupplyAmount, _ := sdk.NewIntFromString("1000000000000000000")
	initSupplyCoin := sdk.NewCoin(testDenom, initSupplyAmount)

	blocksPerYear := params.BlocksPerYear

	inflation := types.NewInflation(
		testDenom,
		targetAccounts[0].Address.String(),
		sdk.ZeroDec())

	tests := []struct {
		calcType string
		calc     func(supply sdk.Coin) (coins sdk.Coins, err error)
	}{
		{"ApproxRoot()", func(supply sdk.Coin) (coins sdk.Coins, err error) {
			result, err := types.CalculateInflation(&inflation, blocksPerYear, supply)
			return result, err
		}},
		{"Math.Pow & strconv", func(supply sdk.Coin) (coins sdk.Coins, err error) {
			result, err := CalculateInflationB(&inflation, blocksPerYear, supply)
			return result, err
		}},
	}
	for _, tc := range tests {
		for _, rate := range []int64{5, 50, 100} {
			inflation.InflationRate = sdk.NewDecWithPrec(rate, 2)
			resetSupply(app, ctx, sdk.NewCoins(initSupplyCoin), sdk.NewCoins(app.BankKeeper.GetSupply(ctx, testDenom)))

			fmt.Println("==> Calculation approach: " + tc.calcType)
			fmt.Println("Inflation at: " + strconv.FormatInt(rate, 10) + "%\n")

			start := time.Now()

			value, _ := tc.calc(app.BankKeeper.GetSupply(ctx, testDenom))
			err := app.MintKeeper.MintCoins(ctx, value)
			if err != nil {
				panic(err)
			}
			err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.AccAddress(inflation.TargetAddress), value)
			if err != nil {
				panic(err)
			}

			elapsed := time.Since(start)

			fmt.Println("Time elapsed to run once: " + elapsed.String())
			fmt.Println("Supply inflated by: " + (app.BankKeeper.GetSupply(ctx, testDenom).Amount).Sub(initSupplyAmount).String() + testDenom + " \n")

			start = time.Now()

			resetSupply(app, ctx, sdk.NewCoins(initSupplyCoin), sdk.NewCoins(app.BankKeeper.GetSupply(ctx, testDenom)))

			for i := 0; i < 10000; i++ {
				value, _ = tc.calc(app.BankKeeper.GetSupply(ctx, testDenom))
				err := app.MintKeeper.MintCoins(ctx, value)
				if err != nil {
					panic(err)
				}
				err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.AccAddress(inflation.TargetAddress), value)
				if err != nil {
					panic(err)
				}
			}

			elapsed = time.Since(start)

			fmt.Println("Time elapsed to run 10,000 times: " + elapsed.String())
			fmt.Println("Supply inflated by: " + (app.BankKeeper.GetSupply(ctx, testDenom).Amount).Sub(initSupplyAmount).String() + testDenom)
			fmt.Printf("===\n\n")
		}
	}
}
