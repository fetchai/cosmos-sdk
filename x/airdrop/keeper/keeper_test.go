package keeper_test

import (
	"sort"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/airdrop/keeper"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	moduleAddress       = authtypes.NewModuleAddress(types.ModuleName)
	feeCollectorAddress = authtypes.NewModuleAddress(authtypes.FeeCollectorName)

	keeperAddr1 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	keeperAddr2 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	keeperAddr3 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	keeperAddr4 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
)

// sendCoinToAccount is a helper to replace the deprecated bankKeeper.SetBalance
// it does mint the coins on the mintmodule before transfering them to the destination account
func sendCoinToAccount(t *testing.T, app *simapp.SimApp, ctx sdk.Context, destAddr sdk.AccAddress, amount sdk.Coin) {
	err := app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(amount))
	require.NoError(t, err)

	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, destAddr, sdk.NewCoins(amount))
	require.NoError(t, err)
}

type KeeperTestSuite struct {
	suite.Suite

	app *simapp.SimApp
	ctx sdk.Context
}

func (s *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(false)
	s.app = app
	s.ctx = app.BaseApp.NewContext(false, tmproto.Header{
		Time:   time.Now(),
		Height: 10,
	})

	s.app.AirdropKeeper.SetParams(s.ctx, types.NewParams(keeperAddr1.String(), keeperAddr2.String(), keeperAddr3.String(), keeperAddr4.String()))
}

func (s *KeeperTestSuite) TestAddNewFund() {
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 4000)
	sendCoinToAccount(s.T(), s.app, s.ctx, keeperAddr1, amount)

	addrBalance := s.app.BankKeeper.GetBalance(s.ctx, keeperAddr1, sdk.DefaultBondDenom)
	moduleBalance := s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, sdk.DefaultBondDenom)

	// sanity check
	s.Require().Equal(addrBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(4000)))
	s.Require().Equal(moduleBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)))

	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(40),
	}
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr1, fund))

	addrBalance = s.app.BankKeeper.GetBalance(s.ctx, keeperAddr1, sdk.DefaultBondDenom)
	moduleBalance = s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, sdk.DefaultBondDenom)

	s.Require().Equal(addrBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)))
	s.Require().Equal(moduleBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(4000)))
}

func (s *KeeperTestSuite) TestCreateRetrieveFund() {
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 4000)
	sendCoinToAccount(s.T(), s.app, s.ctx, keeperAddr1, amount)

	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(40),
	}
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr1, fund))

	recoveredFund, err := s.app.AirdropKeeper.GetFund(s.ctx, keeperAddr1)
	s.Require().NoError(err)
	s.Require().Equal(*recoveredFund, fund)
}

func (s *KeeperTestSuite) TestUnableToCreateDuplicateFunds() {
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 4000)
	addrAmount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 8000)

	sendCoinToAccount(s.T(), s.app, s.ctx, keeperAddr1, addrAmount)

	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(40),
	}
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr1, fund))
	s.Require().Error(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr1, fund)) // this should fail
}

func (s *KeeperTestSuite) TestUnableToCreateFundWithNecessaryFunds() {
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 4000)
	addrAmount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 2000) // not enough

	sendCoinToAccount(s.T(), s.app, s.ctx, keeperAddr1, addrAmount)

	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(40),
	}
	s.Require().Error(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr1, fund)) // this should fail - user doesn't have enough funds
}

func (s *KeeperTestSuite) TestQueryOfAllFunds() {
	amount1 := sdk.NewInt64Coin(sdk.DefaultBondDenom, 4000)
	fund1 := types.Fund{
		Amount:     &amount1,
		DripAmount: sdk.NewInt(40),
	}
	amount2 := sdk.NewInt64Coin(sdk.DefaultBondDenom, 5000)
	fund2 := types.Fund{
		Amount:     &amount2,
		DripAmount: sdk.NewInt(50),
	}

	sendCoinToAccount(s.T(), s.app, s.ctx, keeperAddr1, amount1)
	sendCoinToAccount(s.T(), s.app, s.ctx, keeperAddr2, amount2)

	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr1, fund1))
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr2, fund2))

	funds, err := s.app.AirdropKeeper.GetAllFunds(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(len(funds), 2)

	expectedFunds := []keeper.FundPair{
		{Account: keeperAddr1, Fund: fund1},
		{Account: keeperAddr2, Fund: fund2},
	}

	// sort the funds to have a consistent comparison
	sort.SliceStable(expectedFunds, func(i, j int) bool {
		return expectedFunds[i].Account.String() < expectedFunds[j].Account.String()
	})
	sort.SliceStable(funds, func(i, j int) bool {
		return funds[i].Account.String() < funds[j].Account.String()
	})
	s.Require().Equal(expectedFunds, funds)

	moduleBalance := s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, sdk.DefaultBondDenom)
	s.Require().Equal(moduleBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(9000)))
}

func (s *KeeperTestSuite) TestFeeDrip() {
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 4000)
	fund1 := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(40),
	}
	fund2 := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(4000), // will only last one block
	}

	sendCoinToAccount(s.T(), s.app, s.ctx, keeperAddr1, amount)
	sendCoinToAccount(s.T(), s.app, s.ctx, keeperAddr2, amount)

	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr1, fund1))
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr2, fund2))

	// check the balances
	moduleBalance := s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, sdk.DefaultBondDenom)
	feeCollectorBalance := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddress, sdk.DefaultBondDenom)

	s.Require().Equal(moduleBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(8000)))
	s.Require().Equal(feeCollectorBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)))

	// test case - drip the funds

	_, err := s.app.AirdropKeeper.DripAllFunds(s.ctx)
	s.Require().NoError(err)

	// check that the fees have been transferred
	moduleBalance = s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, sdk.DefaultBondDenom)
	feeCollectorBalance = s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddress, sdk.DefaultBondDenom)

	s.Require().Equal(moduleBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(3960))) // 40 fund1 4000 fund2
	s.Require().Equal(feeCollectorBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(4040)))
}

func (s *KeeperTestSuite) TestAddNewFundWithDiffDenom() {
	amount := sdk.NewInt64Coin("denom", 4000)
	sendCoinToAccount(s.T(), s.app, s.ctx, keeperAddr1, amount)

	addrBalance := s.app.BankKeeper.GetBalance(s.ctx, keeperAddr1, "denom")
	moduleBalance := s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, "denom")

	// sanity check
	s.Require().Equal(addrBalance, sdk.NewCoin("denom", sdk.NewInt(4000)))
	s.Require().Equal(moduleBalance, sdk.NewCoin("denom", sdk.NewInt(0)))

	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(40),
	}
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr1, fund))

	addrBalance = s.app.BankKeeper.GetBalance(s.ctx, keeperAddr1, "denom")
	moduleBalance = s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, "denom")

	s.Require().Equal(addrBalance, sdk.NewCoin("denom", sdk.NewInt(0)))
	s.Require().Equal(moduleBalance, sdk.NewCoin("denom", sdk.NewInt(4000)))
}

func (s *KeeperTestSuite) TestMultiDenomFeeDrip() {
	amountStake := sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)
	amountDenom := sdk.NewInt64Coin("denom", 1000)
	fund1 := types.Fund{
		Amount:     &amountStake,
		DripAmount: sdk.NewInt(500)}

	fund2 := types.Fund{
		Amount:     &amountStake,
		DripAmount: sdk.NewInt(800),
	}

	fund3 := types.Fund{
		Amount:     &amountDenom,
		DripAmount: sdk.NewInt(100)}

	fund4 := types.Fund{
		Amount:     &amountDenom,
		DripAmount: sdk.NewInt(300),
	}
	sendCoinToAccount(s.T(), s.app, s.ctx, keeperAddr1, amountStake)
	sendCoinToAccount(s.T(), s.app, s.ctx, keeperAddr2, amountStake)
	sendCoinToAccount(s.T(), s.app, s.ctx, keeperAddr3, amountDenom)
	sendCoinToAccount(s.T(), s.app, s.ctx, keeperAddr4, amountDenom)

	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr1, fund1))
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr2, fund2))
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr3, fund3))
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, keeperAddr4, fund4)) // check the balances

	moduleBalanceStake := s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, sdk.DefaultBondDenom)
	moduleBalanceDenom := s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, "denom")
	feeCollectorBalanceStake := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddress, sdk.DefaultBondDenom)
	feeCollectorBalanceDenom := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddress, "denom")

	s.Require().Equal(moduleBalanceStake, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(2000)))
	s.Require().Equal(moduleBalanceDenom, sdk.NewCoin("denom", sdk.NewInt(2000)))
	s.Require().Equal(feeCollectorBalanceStake, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)))
	s.Require().Equal(feeCollectorBalanceDenom, sdk.NewCoin("denom", sdk.NewInt(0)))

	_, err := s.app.AirdropKeeper.DripAllFunds(s.ctx) // test case - drip the funds
	s.Require().NoError(err)                          // check that the fees have been transferred

	moduleBalanceStake = s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, sdk.DefaultBondDenom)
	moduleBalanceDenom = s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, "denom")
	feeCollectorBalanceStake = s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddress, sdk.DefaultBondDenom)
	feeCollectorBalanceDenom = s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddress, "denom")

	s.Require().Equal(feeCollectorBalanceStake, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1300))) // 800 drip fund 1, 500 drip fund 2
	s.Require().Equal(feeCollectorBalanceDenom, sdk.NewCoin("denom", sdk.NewInt(400)))               // 100 drip fund 3, 300 drip fund 4
	s.Require().Equal(moduleBalanceStake, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(700)))        // remainder 500 from fund 1, remainder 200 from fund 2
	s.Require().Equal(moduleBalanceDenom, sdk.NewCoin("denom", sdk.NewInt(1600)))                    // remainder 900 from fund 3, remainder 700 from fund 4

}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
