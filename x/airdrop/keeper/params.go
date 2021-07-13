package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
)

func (k Keeper) GetWhiteListClients(ctx sdk.Context) []string {
	var res []string
	k.paramSpace.GetIfExists(ctx, types.KeyWhiteList, &res)	// Allow empty AllowList
	return res
}

// GetParams returns the total set of ibc-transfer parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(k.GetWhiteListClients(ctx)...)
}

// SetParams sets the total set of ibc-transfer parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
