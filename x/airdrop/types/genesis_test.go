package types_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/airdrop/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	addr = sdk.AccAddress([]byte("addr________________"))
	//verboseAddr = sdk.AccAddress([]byte("\n\n\n\n\taddr________________\t\n\n\n\n\n"))
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
	s.app.AirdropKeeper.SetParams(s.ctx, types.NewParams(addr.String()))
}

func (s *GenesisTestSuite) TestNewGenesisState() {
	p := s.app.AirdropKeeper.GetParams(s.ctx)
	funds, err := s.app.AirdropKeeper.GetActiveFunds(s.ctx)
	if err != nil {
		s.Require().FailNow("Failed to get active funds")
	}
	s.Require().IsType(types.GenesisState{}, *types.NewGenesisState(p, funds))
}

func (s *GenesisTestSuite) TestValidateGenesisState() {
	p := s.app.AirdropKeeper.GetParams(s.ctx)
	activeFunds, err := s.app.AirdropKeeper.GetActiveFunds(s.ctx)
	if err != nil {
		s.Require().FailNow("Failed to get active funds")
	}
	genesis1 := types.NewGenesisState(p, activeFunds) // First genesis with no active funds
	s.Require().NoError(genesis1.Validate())
	amount := sdk.NewCoin("test", sdk.NewInt(1000))
	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(1),
	}
	s.Require().NoError(s.app.BankKeeper.SetBalance(s.ctx, addr, amount))
	s.Require().NoError(s.app.AirdropKeeper.AddFund(s.ctx, addr, fund))
	genesis2 := types.NewGenesisState(p, activeFunds) // Second genesis with valid active funds
	s.Require().NoError(genesis2.Validate())

}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
