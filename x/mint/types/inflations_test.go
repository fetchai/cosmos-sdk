package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
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

	minter := DefaultInitialMinter()
	params := DefaultParams()

	app.MintKeeper.SetParams(ctx, params)
	app.MintKeeper.SetMinter(ctx, minter)

	var testSupply int64 = 1000000

	tests := []struct {
		coins           sdk.Coins
		inflation       Inflation
		expectedBalance sdk.Int
	}{
		{sdk.NewCoins(sdk.NewCoin("one", sdk.NewInt(testSupply))), NewInflation("one", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(1, 2)), sdk.NewInt(0)},
		{sdk.NewCoins(sdk.NewCoin("two", sdk.NewInt(testSupply))), NewInflation("two", targetAccounts[1].Address.String(), sdk.NewDecWithPrec(2, 2)), sdk.NewInt(0)},
		{sdk.NewCoins(sdk.NewCoin("three", sdk.NewInt(testSupply))), NewInflation("three", targetAccounts[2].Address.String(), sdk.NewDecWithPrec(3, 2)), sdk.NewInt(0)},
	}
	for i, tc := range tests {
		minter.Inflations = []*Inflation{&tc.inflation}
		require.NoError(t, app.BankKeeper.MintCoins(ctx, ModuleName, tc.coins))

		newCoinsToSend, err := CalculateInflation(&tc.inflation, params.BlocksPerYear, app.BankKeeper.GetSupply(ctx, tc.inflation.Denom))
		require.NoError(t, err)

		err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, ModuleName, sdk.AccAddress(tc.inflation.TargetAddress), newCoinsToSend)
		require.NoError(t, err)

		testAccountBalance := app.BankKeeper.GetBalance(ctx, targetAccounts[i].Address, tc.inflation.Denom)
		require.Equal(t, tc.expectedBalance, testAccountBalance.Amount)
	}
}

func TestInflationsValidation(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	targetAccounts := simtypes.RandomAccounts(r, 1)
	minter := DefaultInitialMinter()

	tests := []struct {
		inflation      Inflation
		expectedToPass bool
	}{
		{NewInflation("stake", targetAccounts[0].Address.String(), sdk.OneDec()), true},
		{NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(1, 2)), true},
		{NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDec(-1)), false},
		{NewInflation("!&£$%", targetAccounts[0].Address.String(), sdk.OneDec()), false},
		{NewInflation("stake", "", sdk.OneDec()), false},
	}
	for _, tc := range tests {
		minter.Inflations = []*Inflation{&tc.inflation}
		if tc.expectedToPass {
			require.NoError(t, ValidateInflations(minter.Inflations))
		} else {
			require.Error(t, ValidateInflations(minter.Inflations))
		}
	}
}

func calculateInflationB(inflation *Inflation, blocksPerYear uint64, supply sdk.Coin) (result sdk.Coins, err error) {
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

	minter := DefaultInitialMinter()
	params := DefaultParams()

	app.MintKeeper.SetParams(ctx, params)
	app.MintKeeper.SetMinter(ctx, minter)

	inflation := NewInflation(
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
			result, err := CalculateInflation(&inflation, blocksPerYear, supply)
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
