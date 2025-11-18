package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/mint/cache"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

var _ types.QueryServer = queryServer{}

func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// Params returns params of the mint module.
func (q queryServer) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := q.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

// Inflation returns minter.Inflation of the mint module.
func (q queryServer) Inflation(ctx context.Context, _ *types.QueryInflationRequest) (*types.QueryInflationResponse, error) {
	minter, err := q.k.Minter.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryInflationResponse{Inflation: minter.Inflation}, nil
}

// MunicipalInflation returns minter.MunicipalInflation of the mint module.
func (q queryServer) MunicipalInflation(ctx context.Context, req *types.QueryMunicipalInflationRequest) (*types.QueryMunicipalInflationResponse, error) {
	denom := req.GetDenom()

	if len(denom) == 0 {
		return &types.QueryMunicipalInflationResponse{Inflations: *cache.GMunicipalInflationCache.GetOriginal()}, nil
	}

	infl := cache.GMunicipalInflationCache.GetInflation(denom)
	if infl == nil {
		return nil, fmt.Errorf("there is no municipal inflation defined for requested \"%s\" denomination", denom)
	}

	return &types.QueryMunicipalInflationResponse{Inflations: []*types.MunicipalInflationPair{{Denom: denom, Inflation: infl.AnnualInflation}}}, nil
}

// AnnualProvisions returns minter.AnnualProvisions of the mint module.
func (q queryServer) AnnualProvisions(ctx context.Context, _ *types.QueryAnnualProvisionsRequest) (*types.QueryAnnualProvisionsResponse, error) {
	minter, err := q.k.Minter.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryAnnualProvisionsResponse{AnnualProvisions: minter.AnnualProvisions}, nil
}
