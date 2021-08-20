package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	genesisAddr = sdk.AccAddress([]byte("addr________________"))
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

	s.app.AirdropKeeper.SetParams(s.ctx, types.NewParams(genesisAddr.String()))
}

func (s *GenesisTestSuite) TestInitAndExportGenesis() {
	p := types.Params{
		AllowList: []string{
			genesisAddr.String(),
		},
	}

	coinAmount := sdk.Coin{
		Denom:  "test",
		Amount: sdk.NewInt(10),
	}

	sendCoinToAccount(s.T(), s.app, s.ctx, genesisAddr, coinAmount)

	funds := []types.ActiveFund{
		{
			Sender: genesisAddr.String(),
			Fund: &types.Fund{
				Amount:     &coinAmount,
				DripAmount: sdk.NewInt(1),
			},
		},
	}
	genState := types.NewGenesisState(p, funds)
	s.app.AirdropKeeper.InitGenesis(s.ctx, genState)
	actualFunds, _ := s.app.AirdropKeeper.GetActiveFunds(s.ctx)
	s.Require().Equal(s.app.AirdropKeeper.GetParams(s.ctx), p)
	s.Require().Equal(funds, actualFunds)
	s.Require().Equal(s.app.AirdropKeeper.ExportGenesis(s.ctx), genState)
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
