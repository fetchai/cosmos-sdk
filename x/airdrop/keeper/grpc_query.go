package keeper

import (
	"context"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) AllFunds(ctx context.Context, req *types.QueryAllFundsRequest) (*types.QueryAllFundsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	store := sdkCtx.KVStore(k.storeKey)
	activeFundStore := prefix.NewStore(store, types.ActiveFundKeyPrefix)

	var activeFunds []*types.ActiveFund
	pageRes, err := query.Paginate(activeFundStore, req.Pagination, func(key, value []byte) error {
		account := sdk.AccAddress(key)

		var fund types.Fund
		err := k.cdc.UnmarshalBinaryBare(value, &fund)
		if err != nil {
			return err
		}

		activeFunds = append(activeFunds, &types.ActiveFund{
			Sender: account.String(),
			Fund: &fund,
		})
		return nil
	})

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "paginate: %v", err)
	}

	return &types.QueryAllFundsResponse{Funds: activeFunds, Pagination: pageRes}, nil
}

func (k Keeper) Fund(ctx context.Context, req *types.QueryFundRequest) (*types.QueryFundResponse, error) {
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	fund, err := k.GetFund(sdkCtx, address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to lookup funds: %s", err.Error())
	}

	if fund == nil {
		return nil, status.Errorf(codes.NotFound, "unable to find fund")
	}

	resp := &types.QueryFundResponse{
		Fund: fund,
	}

	return resp, nil
}
