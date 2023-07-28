package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// bank message types
const (
	TypeMsgSignData = "sigdata"
)

var _ sdk.Msg = &MsgSignData{}

/*
// NewMsgSignData - construct a msg to send coins from one account to another.
//
//nolint:interfacer
func NewMsgSignData(fromAddr, data []byte) *MsgSignData {
	return &MsgSignData{FromAddress: fromAddr.String(), Data: data}
}
*/

// Route Implements Msg.
func (msg MsgSignData) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSignData) Type() string { return TypeMsgSignData }

// ValidateBasic Implements Msg.
func (msg MsgSignData) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgSignData) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners Implements Msg.
func (msg MsgSignData) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}
