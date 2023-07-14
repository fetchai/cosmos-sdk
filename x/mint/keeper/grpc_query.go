package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

var _ types.QueryServer = Keeper{}

// Params returns params of the mint module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Inflation returns minter.Inflation of the mint module.
func (k Keeper) Inflation(c context.Context, _ *types.QueryInflationRequest) (*types.QueryInflationResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minter := k.GetMinter(ctx)

	return &types.QueryInflationResponse{Inflation: minter.Inflation}, nil
}

// MunicipalInflation returns minter.MunicipalInflation of the mint module.
func (k Keeper) MunicipalInflation(c context.Context, req *types.QueryMunicipalInflationRequest) (*types.QueryMunicipalInflationResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minter := k.GetMinter(ctx)
	denom := req.GetDenom()

	if len(denom) == 0 {
		return &types.QueryMunicipalInflationResponse{Inflations: minter.MunicipalInflation}, nil
	}

	infl, exists := minter.MunicipalInflation[denom]
	if !exists {
		return nil, fmt.Errorf("there is no municipal inflation defined for requested \"%s\" denomination", denom)
	}

	return &types.QueryMunicipalInflationResponse{Inflations: map[string]*types.MunicipalInflation{denom: infl}}, nil
}

// AnnualProvisions returns minter.AnnualProvisions of the mint module.
func (k Keeper) AnnualProvisions(c context.Context, _ *types.QueryAnnualProvisionsRequest) (*types.QueryAnnualProvisionsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	minter := k.GetMinter(ctx)

	return &types.QueryAnnualProvisionsResponse{AnnualProvisions: minter.AnnualProvisions}, nil
}
