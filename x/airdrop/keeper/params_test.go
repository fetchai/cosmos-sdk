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
	addr    = sdk.AccAddress([]byte("addr________________"))
	addrSet = sdk.AccAddress([]byte("addrSet_____________"))
)

type ParamsTestSuite struct {
	suite.Suite

	app *simapp.SimApp
	ctx sdk.Context
}

func (s *ParamsTestSuite) SetupTest() {
	app := simapp.Setup(false)
	s.app = app
	s.ctx = app.BaseApp.NewContext(false, tmproto.Header{
		Time:   time.Now(),
		Height: 10,
	})

	s.app.AirdropKeeper.SetParams(s.ctx, types.NewParams(addr.String()))
}

func (s *ParamsTestSuite) TestGetAllowListClients() {
	list := []string{addr.String()}
	s.Require().Equal(s.app.AirdropKeeper.GetAllowListClients(s.ctx), list)
}

func (s *ParamsTestSuite) TestGetParams() {
	p := types.Params{
		AllowList: []string{addr.String()},
	}
	s.Require().Equal(s.app.AirdropKeeper.GetParams(s.ctx), p)
}

func (s *ParamsTestSuite) TestSetParams() {
	p := types.Params{
		AllowList: []string{addr.String(), addrSet.String()},
	}
	s.app.AirdropKeeper.SetParams(s.ctx, p)
	s.Require().Equal(s.app.AirdropKeeper.GetParams(s.ctx), p)
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}
