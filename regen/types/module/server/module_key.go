package server

import (
	"github.com/cosmos/cosmos-sdk/regen/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ModuleKey interface {
	types.InvokerConn

	ModuleID() types.ModuleID
	Address() sdk.AccAddress
}

type InvokerFactory func(callInfo CallInfo) (types.Invoker, error)

type CallInfo struct {
	Method string
	Caller types.ModuleID
}
