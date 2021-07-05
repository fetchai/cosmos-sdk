package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (fund Fund) ValidateBasic() error {
	validDripRate := !fund.DripAmount.IsNil() && fund.DripAmount.IsPositive()
	if !validDripRate {
		return sdkerrors.Wrapf(sdkerrors.ErrConflict, "Invalid drip rate")
	}
	validAmount := !fund.Amount.Amount.IsNil() && !fund.DripAmount.IsNegative()
	if !validAmount {
		return sdkerrors.Wrapf(sdkerrors.ErrConflict, "Invalid amount")
	}

	return nil
}

// Drip the funds that should be dripped for this block
func (fund Fund) Drip() (Fund, sdk.Coin) {
	amount := fund.Amount.Amount
	if amount.GTE(fund.DripAmount) {
		amount = fund.DripAmount
	}

	drip := sdk.NewCoin(fund.Amount.Denom, amount)
	remainingAmount := fund.Amount.Sub(drip)

	return Fund{Amount: &remainingAmount, DripAmount: fund.DripAmount}, drip
}
