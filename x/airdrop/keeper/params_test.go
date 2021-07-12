package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
	"time"
)

var (
	AllowedAddress  = sdk.AccAddress([]byte("allowed_____________"))
	BlockedAddress1 = sdk.AccAddress([]byte("blocked_____________"))
	BlockedAddress2	= sdk.AccAddress([]byte(""))
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

	s.app.AirdropKeeper.SetParams(s.ctx, types.NewParams(AllowedAddress.String()))
}

func (s *ParamsTestSuite) TestIsAllowedSender () {
	s.Require().True(s.app.AirdropKeeper.GetParams(s.ctx).IsAllowedSender(AllowedAddress))		// Address is in AllowList and has correct format
	s.Require().False(s.app.AirdropKeeper.GetParams(s.ctx).IsAllowedSender(BlockedAddress1))	// Address is not in AllowList and has correct format
	s.Require().False(s.app.AirdropKeeper.GetParams(s.ctx).IsAllowedSender(BlockedAddress2))	// Address is not in AllowList and has incorrect format
}


func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}
