package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgAirDrop      = "airdrop"
)

func NewMsgAirDrop(fromAddr sdk.AccAddress, amount sdk.Coin, dripRate sdk.Int) *MsgAirDrop {
	return &MsgAirDrop{FromAddress: fromAddr.String(), Fund: &Fund{
		Amount:          &amount,
		DripRate:        dripRate,
	}}
}

var _ sdk.Msg = &MsgAirDrop{}

func (msg MsgAirDrop) Route() string {
	return RouterKey
}

func (msg MsgAirDrop) Type() string {
	return TypeMsgAirDrop
}

func (msg MsgAirDrop) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	err = msg.Fund.ValidateBasic()
	if err != nil {
		return err
	}

	return nil
}

func (msg MsgAirDrop) GetSignBytes() []byte {
	panic("Airdrop messages do not support amino")
}

func (msg MsgAirDrop) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}
