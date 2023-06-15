package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/mint"
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
	var testSupply int64 = 1000000

	tests := []struct {
		coins           sdk.Coins
		inflation       Inflation
		expectedBalance sdk.Int
		expectedToPass  bool
	}{
		{sdk.NewCoins(sdk.NewCoin("denom-1", sdk.NewInt(testSupply))), NewInflation("denom-1", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(1, 2)), sdk.NewInt(0), true},
		{sdk.NewCoins(sdk.NewCoin("denom-2", sdk.NewInt(testSupply))), NewInflation("denom-2", targetAccounts[1].Address.String(), sdk.NewDecWithPrec(2, 2)), sdk.NewInt(0), true},
		{sdk.NewCoins(sdk.NewCoin("denom-3", sdk.NewInt(testSupply))), NewInflation("denom-3", targetAccounts[2].Address.String(), sdk.NewDecWithPrec(3, 2)), sdk.NewInt(0), true},
	}
	for i, tc := range tests {
		minter.Inflations = []*Inflation{&tc.inflation}
		require.NoError(suite.T(), suite.app.BankKeeper.MintCoins(suite.ctx, ModuleName, tc.coins))
		require.NotPanics(suite.T(), func() { mint.HandleInflations(suite.ctx, suite.app.MintKeeper) })

		testAccountBalance := suite.app.BankKeeper.GetBalance(suite.ctx, targetAccounts[i].Address, fmt.Sprintf("denom-%d", i+1))
		require.Equal(suite.T(), tc.expectedBalance, testAccountBalance.Amount)
	}
}

// TODO(JS): add test asserting validation boundaries
