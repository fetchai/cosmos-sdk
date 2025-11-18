package types_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

var (
	onePercent              = math.LegacyNewDecWithPrec(1, 2)
	almostOne               = math.LegacyOneDec().Sub(math.LegacyNewDecWithPrec(1, math.LegacyPrecision))
	allowedRelativeMulError = math.LegacyNewDecWithPrec(1, 9)
	testAcc                 = "TEST_ACC"
)

// func resetSupply(app *simapp.SimApp, ctx sdk.Context, initSupply sdk.Coins, curSupply sdk.Coins) {
//	err := app.BankKeeper.MintCoins(ctx, types.ModuleName, initSupply)
//	if err != nil {
//		panic(err)
//	}
//	err = app.BankKeeper.BurnCoins(ctx, types.ModuleName, curSupply)
//	if err != nil {
//		panic(err)
//	}
//}
//
//func getTestingAccounts(r *rand.Rand, n int, ctx sdk.Context, app *simapp.SimApp) []simtypes.Account {
//	accounts := simtypes.RandomAccounts(r, n)
//
//	for _, account := range accounts {
//		acc := app.AccountKeeper.NewAccountWithAddress(ctx, account.Address)
//		app.AccountKeeper.SetAccount(ctx, acc)
//	}
//
//	return accounts
//}

func TestCalculateInflationPerBlockAndIssuance(t *testing.T) {
	params := types.DefaultParams()

	supply, _ := math.NewIntFromString("1000000000000000000000000000")
	testDenom := "testDenom"

	tests := []struct {
		inflation              *types.MunicipalInflation
		expectedAnnualIssuance math.Int
	}{
		// Pass: 2 = 200% inflation
		{types.NewMunicipalInflation(testAcc, math.LegacyNewDec(2)), supply.MulRaw(2)},
		// Pass: 1 = 100% inflation
		{types.NewMunicipalInflation(testAcc, math.LegacyOneDec()), supply},
		// Pass: 0.5 = 50% inflation
		{types.NewMunicipalInflation(testAcc, math.LegacyNewDecWithPrec(5, 1)), supply.QuoRaw(2)},
		// Pass: 0.01 = 1% inflation
		{types.NewMunicipalInflation(testAcc, onePercent), supply.QuoRaw(100)},
		//// Pass: -0.01 = -1% inflation
		// {types.NewMunicipalInflation(testDenom, testAcc, onePercent.Neg()), supply.QuoRaw(100).Neg()},
		//// Pass: -0.011 = -1.1% inflation
		//{types.NewMunicipalInflation(testDenom, testAcc, math.LegacyNewDecWithPrec(11, 3).Neg()), supply.MulRaw(11).QuoRaw(1000).Neg()},
		//// Pass: -0.5 = -50% inflation
		//{types.NewMunicipalInflation(testDenom, testAcc, math.LegacyNewDecWithPrec(5, 1).Neg()), supply.QuoRaw(2)},
		//// Pass: -0.999...9 = -99.999...9% inflation
		//{types.NewMunicipalInflation(testDenom, testAcc, almostOne.Neg()), math.LegacyNewDecFromInt(supply).Mul(almostOne).TruncateInt().Neg()},
	}

	for _, tc := range tests {

		// Calculate inflation
		inflationRatePerBlock, err := types.CalculateInflationPerBlock(tc.inflation.Value, params.BlocksPerYear)
		require.NoError(t, err)

		reconstitutedInflationPerAnnum := inflationRatePerBlock.Add(math.LegacyOneDec()).Power(params.BlocksPerYear).Sub(math.LegacyOneDec())

		mulErrorAfterReconstitution := reconstitutedInflationPerAnnum.Quo(tc.inflation.Value).Sub(math.LegacyOneDec()).Abs()
		require.True(t, mulErrorAfterReconstitution.LT(allowedRelativeMulError))

		issuedTokensAnnually := types.CalculateInflationIssuance(reconstitutedInflationPerAnnum, sdk.Coin{Denom: testDenom, Amount: supply})
		issuanceRelativeMulError := math.LegacyNewDecFromInt(issuedTokensAnnually.AmountOf(testDenom)).Quo(math.LegacyNewDecFromInt(tc.expectedAnnualIssuance)).Sub(math.LegacyOneDec()).Abs()

		require.True(t, issuanceRelativeMulError.LT(allowedRelativeMulError))
	}
}

func TestValidationOfMunicipalInflation(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	targetAccounts := simtypes.RandomAccounts(r, 1)

	tests := []struct {
		inflation      *types.MunicipalInflation
		expectedToPass bool
	}{
		// Pass: 2 = 200% inflation
		{types.NewMunicipalInflation(targetAccounts[0].Address.String(), math.LegacyNewDec(2)), true},
		// Pass: 1 = 100% inflation
		{types.NewMunicipalInflation(targetAccounts[0].Address.String(), math.LegacyOneDec()), true},
		// Pass: 0.5 = 50% inflation
		{types.NewMunicipalInflation(targetAccounts[0].Address.String(), math.LegacyNewDecWithPrec(5, 1)), true},
		// Pass: 0.01 = 1% inflation
		{types.NewMunicipalInflation(targetAccounts[0].Address.String(), onePercent), true},
		//// Pass: -0.01 = -1% inflation
		// {types.NewMunicipalInflation(targetAccounts[0].Address.String(), onePercent.Neg()), true},
		//// Pass: -0.011 = -1.1% inflation
		//{types.NewMunicipalInflation(targetAccounts[0].Address.String(), math.LegacyNewDecWithPrec(11, 3).Neg()), true},
		//// Pass: -0.5 = -50% inflation
		//{types.NewMunicipalInflation(targetAccounts[0].Address.String(), math.LegacyNewDecWithPrec(5, 1).Neg()), true},
		//// Pass: -0.999...9 = -99.999...9% inflation
		//{types.NewMunicipalInflation(targetAccounts[0].Address.String(), math.LegacyOneDec().Sub(math.LegacyNewDecWithPrec(1, sdk.Precision)).Neg()), true},
		//// Fail: -1 = -100% inflation
		//{types.NewMunicipalInflation(targetAccounts[0].Address.String(), math.LegacyOneDec().Neg()), false},
		// Fail: invalid targetAddress
		{types.NewMunicipalInflation("fetch123abc", onePercent), false},
		{types.NewMunicipalInflation("", onePercent), false},
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

	expectedToPass := []*types.MunicipalInflationPair{
		// Pass: 2 = 200% inflation
		{"stake0", types.NewMunicipalInflation(targetAccounts[0].Address.String(), math.LegacyNewDec(2))},
		// Pass: 1 = 100% inflation
		{"stake1", types.NewMunicipalInflation(targetAccounts[0].Address.String(), math.LegacyOneDec())},
		// Pass: 0.5 = 50% inflation
		{"stake2", types.NewMunicipalInflation(targetAccounts[0].Address.String(), math.LegacyNewDecWithPrec(5, 1))},
		// Pass: 0.01 = 1% inflation
		{"stake3", types.NewMunicipalInflation(targetAccounts[0].Address.String(), onePercent)},
		//// Pass: -0.01 = -1% inflation
		// {"stake4", types.NewMunicipalInflation(targetAccounts[0].Address.String(), onePercent.Neg())},
		//// Pass: -0.011 = -1.1% inflation
		//{"stake5", types.NewMunicipalInflation(targetAccounts[0].Address.String(), math.LegacyNewDecWithPrec(11, 3).Neg())},
		//// Pass: -0.5 = -50% inflation
		//{"stake6", types.NewMunicipalInflation(targetAccounts[0].Address.String(), math.LegacyNewDecWithPrec(5, 1).Neg())},
		//// Pass: -0.999...9 = -99.999...9% inflation
		//{"stake7", types.NewMunicipalInflation(targetAccounts[0].Address.String(), math.LegacyOneDec().Sub(math.LegacyNewDecWithPrec(1, sdk.Precision)).Neg())},
	}

	// Expected Success:
	err := types.ValidateMunicipalInflations(&expectedToPass)
	require.NoError(t, err)

	expectedToPass2 := append(expectedToPass, &types.MunicipalInflationPair{"stake8", types.NewMunicipalInflation(targetAccounts[0].Address.String(), onePercent)})
	err = types.ValidateMunicipalInflations(&expectedToPass2)
	require.NoError(t, err)

	expectedToPass3 := []*types.MunicipalInflationPair{{"stake0", types.NewMunicipalInflation(targetAccounts[0].Address.String(), onePercent)}}
	err = types.ValidateMunicipalInflations(&expectedToPass3)
	require.NoError(t, err)

	// Expected Failures:
	expectedToFail_NegativeInfl := append(expectedToPass, &types.MunicipalInflationPair{"stake8", types.NewMunicipalInflation(targetAccounts[0].Address.String(), onePercent.Neg())})
	err = types.ValidateMunicipalInflations(&expectedToFail_NegativeInfl)
	require.Error(t, err)

	expectedToFail2_NegativeInfl := []*types.MunicipalInflationPair{{"stake0", types.NewMunicipalInflation(targetAccounts[0].Address.String(), onePercent.Neg())}}
	err = types.ValidateMunicipalInflations(&expectedToFail2_NegativeInfl)
	require.Error(t, err)

	expectedToFail2_DuplicatedDenom := append(expectedToPass, &types.MunicipalInflationPair{"stake0", types.NewMunicipalInflation(targetAccounts[0].Address.String(), onePercent)})
	err = types.ValidateMunicipalInflations(&expectedToFail2_DuplicatedDenom)
	require.Error(t, err)
}

// func TestHandleMunicipalInflation(t *testing.T) {
//	app := simapp.Setup(t, false)
//	ctx := app.BaseApp.NewContext(false)
//	acc := auth.NewEmptyModuleAccount(types.ModuleName, auth.Minter, auth.Burner)
//	app.AccountKeeper.SetModuleAccount(ctx, acc)
//	keeper := app.MintKeeper
//	bankKeeper := app.BankKeeper
//
//	s := rand.NewSource(1)
//	r := rand.New(s)
//	targetAccounts := getTestingAccounts(r, 6, ctx, app)
//
//	minter := types.DefaultInitialMinter()
//	params := types.DefaultParams()
//
//	initSupplyAmount, _ := math.NewIntFromString("1000000000000000000000000000")
//	params.BlocksPerYear = 10000
//	err := keeper.Params.Set(ctx, params)
//	if err != nil {
//		panic(err)
//	}
//
//	definedInfations := []struct {
//		denom                  string
//		annualInflation        math.LegacyDec
//		expectedAnnualIssuance math.Int
//	}{
//		{"denom0", math.LegacyNewDecWithPrec(1, 2), initSupplyAmount.QuoRaw(100)},
//		{"denom1", math.LegacyNewDecWithPrec(5, 2), initSupplyAmount.MulRaw(5).QuoRaw(100)},
//		{"denom2", math.LegacyNewDecWithPrec(25, 2), initSupplyAmount.QuoRaw(4)},
//		{"denom3", math.LegacyNewDecWithPrec(50, 2), initSupplyAmount.QuoRaw(2)},
//		{"denom4", math.LegacyNewDecWithPrec(75, 2), initSupplyAmount.MulRaw(3).QuoRaw(4)},
//		{"denom5", math.LegacyNewDecWithPrec(100, 2), initSupplyAmount},
//	}
//
//	testMunicipalInflations := make([]*types.MunicipalInflationPair, len(definedInfations))
//	for i, infl := range definedInfations {
//		testMunicipalInflations[i] = &types.MunicipalInflationPair{infl.denom, &types.MunicipalInflation{targetAccounts[i].Address.String(), infl.annualInflation}}
//	}
//
//	// Configure/initialize Minter with MunicipalInflation:
//	minter.MunicipalInflation = testMunicipalInflations
//	err = keeper.Minter.Set(ctx, minter)
//	if err != nil {
//		panic(err)
//	}
//
//	// Reset supplies for each denomination to the same `initSupplyAmount` amount:
//	for _, infl := range definedInfations {
//		resetSupply(app, ctx, sdk.NewCoins(sdk.NewCoin(infl.denom, initSupplyAmount)), sdk.NewCoins(bankKeeper.GetSupply(ctx, infl.denom)))
//	}
//
//	// Recording starting balances for all test accounts:
//	startingTestAccountBalances := map[string]math.Int{}
//	for _, tc := range testMunicipalInflations {
//		account, _ := sdk.AccAddressFromBech32(tc.Inflation.TargetAddress)
//		startingTestAccountBalances[tc.Denom] = app.BankKeeper.GetBalance(ctx, account, tc.Denom).Amount
//	}
//
//	// vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv
//	// TEST SUBJECT: Calling production code for as many times as there is number of blocks in a year
//
//	for i := 0; i < int(params.BlocksPerYear); i++ {
//		mintKeeper.HandleMunicipalInflation(&minter, &params, &ctx, &keeper)
//	}
//	// ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
//
//	for i, tc := range testMunicipalInflations {
//		issuedAmount := (bankKeeper.GetSupply(ctx, tc.Denom).Amount).Sub(initSupplyAmount)
//		account, _ := sdk.AccAddressFromBech32(tc.Inflation.TargetAddress)
//		currentTestAccountBalance := app.BankKeeper.GetBalance(ctx, account, tc.Denom).Amount
//		require.True(t, issuedAmount.Equal(currentTestAccountBalance.Sub(startingTestAccountBalances[tc.Denom])))
//
//		issuanceRelativeMulError := math.LegacyNewDecFromInt(issuedAmount).QuoInt(definedInfations[i].expectedAnnualIssuance).Sub(math.LegacyOneDec())
//		require.True(t, issuanceRelativeMulError.LT(allowedRelativeMulError))
//	}
//}
