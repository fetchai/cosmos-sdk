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

	initAmt := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, 200)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, account.Address)
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
		suite.Require().NoError(simapp.FundAccount(suite.app.BankKeeper, suite.ctx, account.Address, initCoins))
	}

	return accounts
}

func (suite *SimTestSuite) testHandleInflations() {
	s := rand.NewSource(1)
	r := rand.New(s)
	targetAccounts := suite.getTestingAccounts(r, 3)

	minter := suite.app.MintKeeper.GetMinter(suite.ctx)

	tests := []struct {
		inflation      Inflation
		expectedToPass bool
	}{
		{NewInflation("test-denom-1", targetAccounts[0].Address.String(), sdk.NewDecWithPrec(1, 2)), true},
		{NewInflation("test-denom-2", targetAccounts[1].Address.String(), sdk.NewDecWithPrec(2, 2)), true},
		{NewInflation("test-denom-3", targetAccounts[2].Address.String(), sdk.NewDecWithPrec(3, 2)), true},
	}
	for i, tc := range tests {
		minter.Inflations = []*Inflation{&tc.inflation}
		require.NotPanics(suite.T(), func() { mint.HandleInflations(suite.ctx, suite.app.MintKeeper) })

		testAccountBalance := suite.app.BankKeeper.GetBalance(suite.ctx, targetAccounts[i].Address, fmt.Sprintf("test-denom-%d", i+1))
		require.Equal(suite.T(), sdk.ZeroDec(), testAccountBalance)
	}
}
