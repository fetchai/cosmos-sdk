package types_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types/time"
	"testing"
)

var (
	allowedAddress  = sdk.AccAddress([]byte("allowed_____________"))
	blockedAddress1 = sdk.AccAddress([]byte("blocked_____________"))
	blockedAddress2	= sdk.AccAddress([]byte(""))
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
}

func (s *ParamsTestSuite) TestIsAllowedSender () {
	correct := types.NewParams(allowedAddress.String())			// New set of parameters with a new AllowList
	s.Require().True(correct.IsAllowedSender(allowedAddress))		// Address is in AllowList and has correct format
	s.Require().False(correct.IsAllowedSender(blockedAddress1))	// Address is not in AllowList and has correct format
	s.Require().False(correct.IsAllowedSender(blockedAddress2))	// Address is not in AllowList and has incorrect format
}

func (s *ParamsTestSuite) TestValidateAllowList () {
	correct := types.NewParams(allowedAddress.String()) 										// Allow list contains address with correct format
	incorrect := types.NewParams(allowedAddress.String(), blockedAddress2.String())				// Allow list contains address with incorrect format
	for _, paramPairs := range correct.ParamSetPairs() {
		s.Require().NoError(paramPairs.ValidatorFn(correct.AllowList))
	}
	for _, paramPairs := range incorrect.ParamSetPairs() {
		s.Require().Error(paramPairs.ValidatorFn(incorrect.AllowList))
	}
}

func TestParamsTestSuite(t *testing.T) {
	suite.Run(t, new(ParamsTestSuite))
}
