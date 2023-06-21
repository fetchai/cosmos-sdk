package types_test

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/simapp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"math"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

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

	s := rand.NewSource(1)
	r := rand.New(s)
	targetAccounts := getTestingAccounts(r, 3, ctx, app)

	minter := types.DefaultInitialMinter()
	params := types.DefaultParams()

	app.MintKeeper.SetParams(ctx, params)
	app.MintKeeper.SetMinter(ctx, minter)

	var testSupply int64 = 1000000

	tests := []struct {
		coins           sdk.Coins
		inflation       types.Inflation
		expectedBalance sdk.Int
	}{
		{sdk.NewCoins(sdk.NewCoin("one", sdk.NewInt(testSupply))), types.NewInflation("one", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(1, 2)), sdk.NewInt(0)},
		{sdk.NewCoins(sdk.NewCoin("two", sdk.NewInt(testSupply))), types.NewInflation("two", targetAccounts[1].Address.String(), sdk.NewDecWithPrec(2, 2)), sdk.NewInt(0)},
		{sdk.NewCoins(sdk.NewCoin("three", sdk.NewInt(testSupply))), types.NewInflation("three", targetAccounts[2].Address.String(), sdk.NewDecWithPrec(3, 2)), sdk.NewInt(0)},
	}
	for i, tc := range tests {
		minter.Inflations = []*types.Inflation{&tc.inflation}
		require.NoError(t, app.BankKeeper.MintCoins(ctx, types.ModuleName, tc.coins))

		newCoinsToSend, err := types.CalculateInflation(&tc.inflation, params.BlocksPerYear, app.BankKeeper.GetSupply(ctx, tc.inflation.Denom))
		require.NoError(t, err)

		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.AccAddress(tc.inflation.TargetAddress), newCoinsToSend)
		require.NoError(t, err)

		testAccountBalance := app.BankKeeper.GetBalance(ctx, targetAccounts[i].Address, tc.inflation.Denom)
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
		{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.OneDec()), true},
		{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(1, 2)), true},
		{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDec(-1)), false},
		{types.NewInflation("!&£$%", targetAccounts[0].Address.String(), sdk.OneDec()), false},
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

func calculateInflationB(inflation *types.Inflation, blocksPerYear uint64, supply sdk.Coin) (result sdk.Coins, err error) {
	inflationPerBlock := math.Pow((inflation.InflationRate.Add(sdk.OneDec())).MustFloat64(), float64(1/blocksPerYear)) - 1
	s, err := sdk.NewDecFromStr(strconv.FormatFloat(inflationPerBlock, 'g', -1, 64))
	reward := s.MulInt(supply.Amount).TruncateInt()
	result = sdk.NewCoins(sdk.NewCoin("stake", reward))
	return result, err
}

func TestInflationsCalculationsDuration(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	s := rand.NewSource(1)
	r := rand.New(s)
	targetAccounts := simtypes.RandomAccounts(r, 1)
	denom := "stake"

	minter := types.DefaultInitialMinter()
	params := types.DefaultParams()

	app.MintKeeper.SetParams(ctx, params)
	app.MintKeeper.SetMinter(ctx, minter)

	inflation := types.NewInflation(
		denom,
		targetAccounts[0].Address.String(),
		sdk.NewDecWithPrec(5, 2))
	supply := app.BankKeeper.GetSupply(ctx, denom)
	blocksPerYear := params.BlocksPerYear

	tests := []struct {
		calcType string
		calc     func() (coins sdk.Coins, err error)
	}{
		{"ApproxRoot()", func() (coins sdk.Coins, err error) {
			result, err := types.CalculateInflation(&inflation, blocksPerYear, supply)
			return result, err
		}},
		{"Math.Pow & strconv", func() (coins sdk.Coins, err error) {
			result, err := calculateInflationB(&inflation, blocksPerYear, supply)
			return result, err
		}},
	}
	for _, tc := range tests {
		fmt.Println("Calculation approach: " + tc.calcType)
		singleIterExactValue := sdk.NewInt(1000) // Exact inflated value, without type precision loss
		manyIterExactValue := sdk.NewInt(10000)

		start := time.Now()
		value, _ := tc.calc()
		elapsed := time.Since(start)
		precisionLoss := singleIterExactValue.Sub(value.AmountOf(denom))

		fmt.Println("Time elapsed to run once: " + elapsed.String())
		fmt.Println("Precision loss: " + precisionLoss.String())

		start = time.Now()
		for i := 0; i < 10000; i++ {
			value, _ = tc.calc()
		}
		elapsed = time.Since(start)
		precisionLoss = manyIterExactValue.Sub(value.AmountOf(denom))

		fmt.Println("Time elapsed to run 10,000 times: " + elapsed.String())
		fmt.Println("Precision loss: " + precisionLoss.String())
	}
}
