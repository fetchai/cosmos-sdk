package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	moduleAddress       = authtypes.NewModuleAddress(types.ModuleName)
	feeCollectorAddress = authtypes.NewModuleAddress(authtypes.FeeCollectorName)
	addr1               = sdk.AccAddress([]byte("addr1_______________"))
	addr2               = sdk.AccAddress([]byte("addr2_______________"))
	addr3               = sdk.AccAddress([]byte("addr3_______________"))
	addr4               = sdk.AccAddress([]byte("addr4_______________"))
)

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

	s.app.AirdropKeeper.SetParams(s.ctx, types.NewParams(addr1.String(), addr2.String(), addr3.String(), addr4.String()))
}

func (s *KeeperTestSuite) TestAddNewFund() {
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 4000)
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr1, amount)) // ensure the account is funded

	addrBalance := s.app.BankKeeper.GetBalance(s.ctx, addr1, sdk.DefaultBondDenom)
	moduleBalance := s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, sdk.DefaultBondDenom)

	// sanity check
	s.Require().Equal(addrBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(4000)))
	s.Require().Equal(moduleBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)))

	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(40),
	}
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr1, fund))

	addrBalance = s.app.BankKeeper.GetBalance(s.ctx, addr1, sdk.DefaultBondDenom)
	moduleBalance = s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, sdk.DefaultBondDenom)

	s.Require().Equal(addrBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)))
	s.Require().Equal(moduleBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(4000)))
}

func (s *KeeperTestSuite) TestCreateRetrieveFund() {
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 4000)
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr1, amount)) // ensure the account is funded

	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(40),
	}
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr1, fund))

	recoveredFund, err := s.app.AirdropKeeper.GetFund(s.ctx, addr1)
	s.Require().NoError(err)
	s.Require().Equal(*recoveredFund, fund)
}

func (s *KeeperTestSuite) TestUnableToCreateDuplicateFunds() {
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 4000)
	addrAmount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 8000)
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr1, addrAmount)) // ensure the account is funded for the two funds

	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(40),
	}
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr1, fund))
	s.Require().Error(s.app.AirdropKeeper.AddFund(s.ctx, addr1, fund)) // this should fail
}

func (s *KeeperTestSuite) TestUnableToCreateFundWithNecessaryFunds() {
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 4000)
	addrAmount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 2000)                 // not enough
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr1, addrAmount)) // ensure the account is funded for the two funds

	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(40),
	}
	s.Require().Error(s.app.AirdropKeeper.AddFund(s.ctx, addr1, fund)) // this should fail - user doesn't have enough funds
}

func (s *KeeperTestSuite) TestQueryOfAllFunds() {
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 4000)
	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(40),
	}
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr1, amount))
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr2, amount))

	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr1, fund))
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr2, fund))

	funds, err := s.app.AirdropKeeper.GetAllFunds(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(len(funds), 2)

	s.Require().Equal(funds[0].Fund, fund)
	s.Require().Equal(funds[0].Account, addr1)

	s.Require().Equal(funds[1].Fund, fund)
	s.Require().Equal(funds[1].Account, addr2)

	moduleBalance := s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, sdk.DefaultBondDenom)
	s.Require().Equal(moduleBalance, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(8000)))
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

	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr1, amount))
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr2, amount))

	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr1, fund1))
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr2, fund2))

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
	amount := sdk.NewInt64Coin("afet", 4000)
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr1, amount)) // ensure the account is funded

	addrBalance := s.app.BankKeeper.GetBalance(s.ctx, addr1, "afet")
	moduleBalance := s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, "afet")

	// sanity check
	s.Require().Equal(addrBalance, sdk.NewCoin("afet", sdk.NewInt(4000)))
	s.Require().Equal(moduleBalance, sdk.NewCoin("afet", sdk.NewInt(0)))

	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(40),
	}
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr1, fund))

	addrBalance = s.app.BankKeeper.GetBalance(s.ctx, addr1, "afet")
	moduleBalance = s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, "afet")

	s.Require().Equal(addrBalance, sdk.NewCoin("afet", sdk.NewInt(0)))
	s.Require().Equal(moduleBalance, sdk.NewCoin("afet", sdk.NewInt(4000)))
}

func (s *KeeperTestSuite) TestMultiDenomFeeDrip() {
	amount_stake := sdk.NewInt64Coin(sdk.DefaultBondDenom, 4000)
	amount_afet := sdk.NewInt64Coin("afet", 4000)
	fund1 := types.Fund{
		Amount:     &amount_stake,
		DripAmount: sdk.NewInt(40)}

	fund2 := types.Fund{
		Amount:     &amount_stake,
		DripAmount: sdk.NewInt(4000), // will only last one block
	}

	fund3 := types.Fund{
		Amount:     &amount_afet,
		DripAmount: sdk.NewInt(40)}

	fund4 := types.Fund{
		Amount:     &amount_afet,
		DripAmount: sdk.NewInt(4000), // will only last one block
	}
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr1, amount_stake))
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr2, amount_stake))
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr3, amount_afet))
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr4, amount_afet))

	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr1, fund1))
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr2, fund2))
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr3, fund3))
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr4, fund4)) // check the balances

	moduleBalance_stake := s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, sdk.DefaultBondDenom)
	moduleBalance_afet := s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, "afet")
	feeCollectorBalance_stake := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddress, sdk.DefaultBondDenom)
	feeCollectorBalance_afet := s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddress, "afet")

	s.Require().Equal(moduleBalance_stake, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(8000)))
	s.Require().Equal(moduleBalance_afet, sdk.NewCoin("afet", sdk.NewInt(8000)))
	s.Require().Equal(feeCollectorBalance_stake, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)))
	s.Require().Equal(feeCollectorBalance_afet, sdk.NewCoin("afet", sdk.NewInt(0))) // test case - drip the funds

	_, err := s.app.AirdropKeeper.DripAllFunds(s.ctx)
	s.Require().NoError(err) // check that the fees have been transferred

	moduleBalance_stake = s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, sdk.DefaultBondDenom)
	moduleBalance_afet = s.app.BankKeeper.GetBalance(s.ctx, moduleAddress, "afet")
	feeCollectorBalance_stake = s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddress, sdk.DefaultBondDenom)
	feeCollectorBalance_afet = s.app.BankKeeper.GetBalance(s.ctx, feeCollectorAddress, "afet")

	s.Require().Equal(moduleBalance_stake, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(3960))) // 40 fund1 4000 fund2
	s.Require().Equal(moduleBalance_afet, sdk.NewCoin("afet", sdk.NewInt(3960)))                // 40 fund1 4000 fund2
	s.Require().Equal(feeCollectorBalance_stake, sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(4040)))
	s.Require().Equal(feeCollectorBalance_afet, sdk.NewCoin("afet", sdk.NewInt(4040)))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
