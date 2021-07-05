package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) AirDrop(goCtx context.Context, msg *types.MsgAirDrop) (*types.MsgAirDropResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate the from address
	sender, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return nil, err
	}

	// validate the fund
	err = msg.Fund.ValidateBasic()
	if err != nil {
		return nil, err
	}

	// add the fund to our keeper
	err = m.AddFund(ctx, sender, *msg.Fund)
	if err != nil {
		return nil, err
	}

	return &types.MsgAirDropResponse{}, nil
}
