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
	"math/rand"
	"testing"
)

var (
	onePercent              = sdk.NewDecWithPrec(1, 2)
	almostOne               = sdk.OneDec().Sub(sdk.NewDecWithPrec(1, sdk.Precision))
	allowedRelativeMulError = sdk.NewDecWithPrec(1, 9)
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

func TestCalculateInflationPerBlockAndIssuance(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	s := rand.NewSource(1)
	r := rand.New(s)
	targetAccounts := getTestingAccounts(r, 3, ctx, app)

	params := types.DefaultParams()

	supply, _ := sdk.NewIntFromString("1000000000000000000000000000")
	testDenom := "testDenom"

	tests := []struct {
		inflation              *types.Inflation
		expectedAnnualIssuance sdk.Int
	}{
		// Pass: 2 = 200% inflation
		{types.NewInflation(testDenom, targetAccounts[0].Address.String(), sdk.NewDec(2)), supply.MulRaw(2)},
		// Pass: 1 = 100% inflation
		{types.NewInflation(testDenom, targetAccounts[0].Address.String(), sdk.OneDec()), supply},
		// Pass: 0.5 = 50% inflation
		{types.NewInflation(testDenom, targetAccounts[0].Address.String(), sdk.NewDecWithPrec(5, 1)), supply.QuoRaw(2)},
		// Pass: 0.01 = 1% inflation
		{types.NewInflation(testDenom, targetAccounts[0].Address.String(), onePercent), supply.QuoRaw(100)},
		//// Pass: -0.01 = -1% inflation
		//{types.NewInflation(testDenom, targetAccounts[0].Address.String(), onePercent.Neg()), supply.QuoRaw(100).Neg()},
		//// Pass: -0.011 = -1.1% inflation
		//{types.NewInflation(testDenom, targetAccounts[0].Address.String(), sdk.NewDecWithPrec(11, 3).Neg()), supply.MulRaw(11).QuoRaw(1000).Neg()},
		//// Pass: -0.5 = -50% inflation
		//{types.NewInflation(testDenom, targetAccounts[0].Address.String(), sdk.NewDecWithPrec(5, 1).Neg()), supply.QuoRaw(2)},
		//// Pass: -0.999...9 = -99.999...9% inflation
		//{types.NewInflation(testDenom, targetAccounts[0].Address.String(), almostOne.Neg()), sdk.NewDecFromInt(supply).Mul(almostOne).TruncateInt().Neg()},
	}

	for _, tc := range tests {

		// Calculate inflation
		inflationRatePerBlock, err := types.CalculateInflationPerBlock(tc.inflation, params.BlocksPerYear)
		require.NoError(t, err)

		reconstitutedInflationPerAnnum := inflationRatePerBlock.Add(sdk.OneDec()).Power(params.BlocksPerYear).Sub(sdk.OneDec())

		mulErrorAfterReconstitution := reconstitutedInflationPerAnnum.Quo(tc.inflation.InflationRate).Sub(sdk.OneDec()).Abs()
		require.True(t, mulErrorAfterReconstitution.LT(allowedRelativeMulError))

		issuedTokensAnnually := types.CalculateInflationIssuance(reconstitutedInflationPerAnnum, sdk.Coin{Denom: testDenom, Amount: supply})
		issuanceRelativeMulError := sdk.NewDecFromInt(issuedTokensAnnually.AmountOf(testDenom)).Quo(sdk.NewDecFromInt(tc.expectedAnnualIssuance)).Sub(sdk.OneDec()).Abs()

		require.True(t, issuanceRelativeMulError.LT(allowedRelativeMulError))
	}
}

func TestValidationOfMunicipalInflation(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	targetAccounts := simtypes.RandomAccounts(r, 1)

	tests := []struct {
		inflation      *types.Inflation
		expectedToPass bool
	}{
		// Pass: 2 = 200% inflation
		{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDec(2)), true},
		// Pass: 1 = 100% inflation
		{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.OneDec()), true},
		// Pass: 0.5 = 50% inflation
		{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(5, 1)), true},
		// Pass: 0.01 = 1% inflation
		{types.NewInflation("stake", targetAccounts[0].Address.String(), onePercent), true},
		//// Pass: -0.01 = -1% inflation
		//{types.NewInflation("stake", targetAccounts[0].Address.String(), onePercent.Neg()), true},
		//// Pass: -0.011 = -1.1% inflation
		//{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(11, 3).Neg()), true},
		//// Pass: -0.5 = -50% inflation
		//{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(5, 1).Neg()), true},
		//// Pass: -0.999...9 = -99.999...9% inflation
		//{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.OneDec().Sub(sdk.NewDecWithPrec(1, sdk.Precision)).Neg()), true},
		//// Fail: -1 = -100% inflation
		//{types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.OneDec().Neg()), false},
		// Fail: invalid denom
		{types.NewInflation("!&£$%", targetAccounts[0].Address.String(), onePercent), false},
		{types.NewInflation("", targetAccounts[0].Address.String(), onePercent), false},
		// Fail: invalid targetAddress
		{types.NewInflation("stake", "fetch123abc", onePercent), false},
		{types.NewInflation("stake", "", onePercent), false},
	}
	for _, tc := range tests {
		err := tc.inflation.Validate()
		if tc.expectedToPass {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
}

func TestBulkValidationOfMunicipalInflations(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	targetAccounts := simtypes.RandomAccounts(r, 1)

	expectedToPass := []*types.Inflation{
		// Pass: 2 = 200% inflation
		types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDec(2)),
		// Pass: 1 = 100% inflation
		types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.OneDec()),
		// Pass: 0.5 = 50% inflation
		types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(5, 1)),
		// Pass: 0.01 = 1% inflation
		types.NewInflation("stake", targetAccounts[0].Address.String(), onePercent),
		//// Pass: -0.01 = -1% inflation
		//types.NewInflation("stake", targetAccounts[0].Address.String(), onePercent.Neg()),
		//// Pass: -0.011 = -1.1% inflation
		//types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(11, 3).Neg()),
		//// Pass: -0.5 = -50% inflation
		//types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(5, 1).Neg()),
		//// Pass: -0.999...9 = -99.999...9% inflation
		//types.NewInflation("stake", targetAccounts[0].Address.String(), sdk.OneDec().Sub(sdk.NewDecWithPrec(1, sdk.Precision)).Neg()),
	}

	// Expected Success:
	err := types.ValidateMunicipalInflations(expectedToPass)
	require.NoError(t, err)

	expectedToPass2 := append(expectedToPass, types.NewInflation("stake", targetAccounts[0].Address.String(), onePercent))
	err = types.ValidateMunicipalInflations(expectedToPass2)
	require.NoError(t, err)

	expectedToPass3 := []*types.Inflation{types.NewInflation("stake", targetAccounts[0].Address.String(), onePercent)}
	err = types.ValidateMunicipalInflations(expectedToPass3)
	require.NoError(t, err)

	// Expected Failures:
	expectedToFail := append(expectedToPass, types.NewInflation("stake", targetAccounts[0].Address.String(), onePercent.Neg()))
	err = types.ValidateMunicipalInflations(expectedToFail)
	require.Error(t, err)

	expectedToFail2 := []*types.Inflation{types.NewInflation("stake", targetAccounts[0].Address.String(), onePercent.Neg())}
	err = types.ValidateMunicipalInflations(expectedToFail2)
	require.Error(t, err)
}

func TestHandleMunicipalInflation(t *testing.T) {
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
		inflation              *types.Inflation
		expectedAnnualIssuance sdk.Int
	}{
		{types.NewInflation(testDenom, targetAccounts[0].Address.String(), sdk.NewDecWithPrec(1, 2)), initSupplyAmount.QuoRaw(100)},
		{types.NewInflation(testDenom, targetAccounts[1].Address.String(), sdk.NewDecWithPrec(5, 2)), initSupplyAmount.MulRaw(5).QuoRaw(100)},
		{types.NewInflation(testDenom, targetAccounts[2].Address.String(), sdk.NewDecWithPrec(25, 2)), initSupplyAmount.QuoRaw(4)},
		{types.NewInflation(testDenom, targetAccounts[3].Address.String(), sdk.NewDecWithPrec(50, 2)), initSupplyAmount.QuoRaw(2)},
		{types.NewInflation(testDenom, targetAccounts[4].Address.String(), sdk.NewDecWithPrec(75, 2)), initSupplyAmount.MulRaw(3).QuoRaw(4)},
		{types.NewInflation(testDenom, targetAccounts[5].Address.String(), sdk.NewDecWithPrec(100, 2)), initSupplyAmount},
	}

	for _, tc := range tests {
		resetSupply(app, ctx, sdk.NewCoins(initSupplyCoin), sdk.NewCoins(keeper.BankKeeper.GetSupply(ctx, testDenom)))
		minter.Inflations = []*types.Inflation{tc.inflation}
		keeper.SetMinter(ctx, minter)

		account, _ := sdk.AccAddressFromBech32(tc.inflation.TargetAddress)
		startingTestAccountBalance := app.BankKeeper.GetBalance(ctx, account, tc.inflation.Denom).Amount

		for i := 0; i < int(params.BlocksPerYear); i++ {
			mint.HandleMunicipalInflation(ctx, keeper)
		}

		issuedAmount := (keeper.BankKeeper.GetSupply(ctx, testDenom).Amount).Sub(initSupplyAmount)

		currentTestAccountBalance := app.BankKeeper.GetBalance(ctx, account, tc.inflation.Denom).Amount
		require.True(t, issuedAmount.Equal(currentTestAccountBalance.Sub(startingTestAccountBalance)))

		issuanceRelativeMulError := sdk.NewDecFromInt(issuedAmount).QuoInt(tc.expectedAnnualIssuance).Sub(sdk.OneDec())
		require.True(t, issuanceRelativeMulError.LT(allowedRelativeMulError))
	}
}
