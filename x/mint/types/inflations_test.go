package types

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"math/rand"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *simapp.SimApp
}

func (suite *SimTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)
	suite.app = app
	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{})
}

func (suite *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	for _, account := range accounts {
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
	}

	return accounts
}

func (suite *SimTestSuite) testHandleInflations() {
	s := rand.NewSource(1)
	r := rand.New(s)

	targetAccounts := suite.getTestingAccounts(r, 3)
	minter := suite.app.MintKeeper.GetMinter(suite.ctx)
	params := suite.app.MintKeeper.GetParams(suite.ctx)
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
		require.NoError(suite.T(), suite.app.BankKeeper.MintCoins(suite.ctx, ModuleName, tc.coins))

		newCoinsToSend, err := CalculateInflation(&tc.inflation, params.BlocksPerYear, suite.app.BankKeeper.GetSupply(suite.ctx, tc.inflation.Denom))
		require.NoError(suite.T(), err)

		err = suite.app.BankKeeper.SendCoinsFromModuleToAccount(suite.ctx, ModuleName, sdk.AccAddress(tc.inflation.TargetAddress), newCoinsToSend)
		require.NoError(suite.T(), err)

		testAccountBalance := suite.app.BankKeeper.GetBalance(suite.ctx, targetAccounts[i].Address, tc.inflation.Denom)
		require.Equal(suite.T(), tc.expectedBalance, testAccountBalance.Amount)
	}
}

func (suite *SimTestSuite) testInflationsValidation() {
	s := rand.NewSource(1)
	r := rand.New(s)

	targetAccounts := suite.getTestingAccounts(r, 3)
	minter := suite.app.MintKeeper.GetMinter(suite.ctx)

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
			require.NoError(suite.T(), ValidateInflations(minter.Inflations))
		} else {
			require.Error(suite.T(), ValidateInflations(minter.Inflations))
		}
	}
}

// TODO(JS): different approaches of calculating inflations - timings + precision, test many different numbers
