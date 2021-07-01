package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (fund Fund) ValidateBasic() error {
	validDripRate := !fund.DripRate.IsNil() && fund.DripRate.IsPositive()
	if !validDripRate {
		return sdkerrors.Wrapf(sdkerrors.ErrConflict, "Invalid drip rate")
	}
	validAmount := !fund.Amount.Amount.IsNil() && !fund.DripRate.IsNegative()
	if !validAmount {
		return sdkerrors.Wrapf(sdkerrors.ErrConflict, "Invalid amount")
	}

	return nil
}

// Drip the funds that should be dripped for this block
func (fund Fund) Drip() (Fund, sdk.Coin) {
	amount := fund.Amount.Amount
	if amount.GTE(fund.DripRate) {
		amount = fund.DripRate
	}

	drip := sdk.NewCoin(fund.Amount.Denom, amount)
	remainingAmount := fund.Amount.Sub(drip)

	return Fund{Amount: &remainingAmount, DripRate: fund.DripRate}, drip
}
