package types_test

import (
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	addr = sdk.AccAddress([]byte("addr________________"))
	recv = sdk.AccAddress([]byte("recv________________"))
)

type GenesisTestSuite struct {
	suite.Suite

	app *simapp.SimApp
	ctx sdk.Context
}

func (s *GenesisTestSuite) SetupTest() {
	app := simapp.Setup(false)
	s.app = app
	s.ctx = app.BaseApp.NewContext(false, tmproto.Header{
		Time:   time.Now(),
		Height: 10,
	})
}

func (s *GenesisTestSuite) TestNewGenesisState() {
	p := s.app.AirdropKeeper.GetParams(s.ctx)
	funds, err := s.app.AirdropKeeper.GetActiveFunds(s.ctx)
	if err != nil {
		s.Require().FailNow("Failed getting Active funds from keeper")
	}
	s.Require().IsType(types.GenesisState{}, *types.NewGenesisState(p, funds))
}

func (s *GenesisTestSuite) TestValidateGenesisState() {
	p := s.app.AirdropKeeper.GetParams(s.ctx)
	amount := sdk.NewCoin("test", sdk.NewInt(1000))
	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(1),
	}
	s.app.AirdropKeeper.SetParams(s.ctx, types.NewParams(addr.String()))
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr, amount))
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr, fund))
	activeFunds, err := s.app.AirdropKeeper.GetActiveFunds(s.ctx)
	if err == nil {
		genesis := types.NewGenesisState(p, activeFunds)
		s.Require().NoError(genesis.Validate())
	}
	return
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
