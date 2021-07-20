package keeper_test

import (
	gocontext "context"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var (
	queryAddr1 = sdk.AccAddress([]byte("addr1_______________"))
	queryAddr2 = sdk.AccAddress([]byte("addr2_______________"))
)

type KeeperGrpcQueryTestSuite struct {
	suite.Suite

	app         *simapp.SimApp
	ctx         sdk.Context
	queryClient types.QueryClient
}

func (s *KeeperGrpcQueryTestSuite) SetupTest() {
	s.app = simapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.app.AirdropKeeper.SetParams(s.ctx, types.NewParams(queryAddr1.String(), queryAddr2.String()))

	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.AirdropKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)
}

func (s *KeeperGrpcQueryTestSuite) TestQueryFund() {
	var (
		req         *types.QueryFundRequest
		expResponse types.QueryFundResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"without an address being specified",
			func() {
				req = &types.QueryFundRequest{}
				expResponse = types.QueryFundResponse{}
			},
			false,
		},
		{
			"with an address being specified",
			func() {
				amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 2000)
				fund := types.Fund{
					Amount:     &amount,
					DripAmount: sdk.NewInt(40),
				}
				err := s.app.BankKeeper.SetBalance(s.ctx, queryAddr1, amount)
				if err != nil {
					panic(err)
				}
				err = s.app.AirdropKeeper.AddFund(s.ctx, queryAddr1, fund)
				if err != nil {
					panic(err)
				}

				req = &types.QueryFundRequest{
					Address: queryAddr1.String(),
				}
				expResponse = types.QueryFundResponse{Fund: &fund}
			},
			true,
		},
		{
			"with an address being specified but fund not present",
			func() {
				req = &types.QueryFundRequest{
					Address: queryAddr1.String(),
				}
				expResponse = types.QueryFundResponse{}
			},
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			s.SetupTest() // reset

			tc.malleate()

			res, err := s.queryClient.Fund(gocontext.Background(), req)

			if tc.expPass {
				s.Require().NoError(err)
				s.Require().NotNil(res)
				s.Require().Equal(&expResponse, res)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func (s *KeeperGrpcQueryTestSuite) TestQueryAllFund() {
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 2000)
	fund := types.Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(40),
	}

	var (
		req         *types.QueryAllFundsRequest
		expResponse types.QueryAllFundsResponse
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"when no funds are present",
			func() {
				req = &types.QueryAllFundsRequest{}
				expResponse = types.QueryAllFundsResponse{
					Pagination: &query.PageResponse{
						NextKey: nil,
						Total:   0,
					},
				}
			},
			true,
		},
		{
			"when funds are present",
			func() {
				err := s.app.BankKeeper.SetBalance(s.ctx, queryAddr1, *fund.Amount)
				if err != nil {
					panic(err)
				}
				err = s.app.BankKeeper.SetBalance(s.ctx, queryAddr2, *fund.Amount)
				if err != nil {
					panic(err)
				}
				err = s.app.AirdropKeeper.AddFund(s.ctx, queryAddr1, fund)
				if err != nil {
					panic(err)
				}
				err = s.app.AirdropKeeper.AddFund(s.ctx, queryAddr2, fund)
				if err != nil {
					panic(err)
				}

				req = &types.QueryAllFundsRequest{}
				expResponse = types.QueryAllFundsResponse{
					Funds: []*types.ActiveFund{
						{
							Sender: queryAddr1.String(),
							Fund:   &fund,
						},
						{
							Sender: queryAddr2.String(),
							Fund:   &fund,
						},
					},
					Pagination: &query.PageResponse{
						NextKey: nil,
						Total:   2,
					},
				}
			},
			true,
		},
		{
			"when funds are present with page 1",
			func() {
				err := s.app.BankKeeper.SetBalance(s.ctx, queryAddr1, *fund.Amount)
				if err != nil {
					panic(err)
				}
				err = s.app.BankKeeper.SetBalance(s.ctx, queryAddr2, *fund.Amount)
				if err != nil {
					panic(err)
				}
				err = s.app.AirdropKeeper.AddFund(s.ctx, queryAddr1, fund)
				if err != nil {
					panic(err)
				}
				err = s.app.AirdropKeeper.AddFund(s.ctx, queryAddr2, fund)
				if err != nil {
					panic(err)
				}

				req = &types.QueryAllFundsRequest{
					Pagination: &query.PageRequest{
						Key:        nil,
						Offset:     0,
						Limit:      1,
						CountTotal: false,
					},
				}
				expResponse = types.QueryAllFundsResponse{
					Funds: []*types.ActiveFund{
						{
							Sender: queryAddr1.String(),
							Fund:   &fund,
						},
					},
					Pagination: &query.PageResponse{
						NextKey: queryAddr2,
					},
				}
			},
			true,
		},
		{
			"when funds are present with page 2",
			func() {
				err := s.app.BankKeeper.SetBalance(s.ctx, queryAddr1, *fund.Amount)
				if err != nil {
					panic(err)
				}
				err = s.app.BankKeeper.SetBalance(s.ctx, queryAddr2, *fund.Amount)
				if err != nil {
					panic(err)
				}
				err = s.app.AirdropKeeper.AddFund(s.ctx, queryAddr1, fund)
				if err != nil {
					panic(err)
				}
				err = s.app.AirdropKeeper.AddFund(s.ctx, queryAddr2, fund)
				if err != nil {
					panic(err)
				}

				req = &types.QueryAllFundsRequest{
					Pagination: &query.PageRequest{
						Key:        nil,
						Offset:     1,
						Limit:      1,
						CountTotal: false,
					},
				}
				expResponse = types.QueryAllFundsResponse{
					Funds: []*types.ActiveFund{
						{
							Sender: queryAddr2.String(),
							Fund:   &fund,
						},
					},
					Pagination: &query.PageResponse{},
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			s.SetupTest() // reset

			tc.malleate()

			res, err := s.queryClient.AllFunds(gocontext.Background(), req)

			if tc.expPass {
				s.Require().NoError(err)
				s.Require().Equal(&expResponse, res)
			} else {
				s.Require().Error(err)
			}
		})
	}
}

func TestKeeperGrpcQueryTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperGrpcQueryTestSuite))
}
