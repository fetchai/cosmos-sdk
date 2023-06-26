package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewInflation returns a new Inflation object with the given denom, target_address
// and inflation_rate
func NewInflation(denom string, targetAddress string, inflationRate sdk.Dec) Inflation {
	return Inflation{
		Denom:         denom,
		TargetAddress: targetAddress,
		InflationRate: inflationRate,
	}
}

func CalculateInflationPerBlock(inflation *Inflation, blocksPerYear uint64) (result sdk.Dec, err error) {
	inflationPerBlockPlusOne, err := inflation.InflationRate.Add(sdk.OneDec()).ApproxRoot(blocksPerYear)
	if err != nil {
		panic(err)
	}
	inflationPerBlock := inflationPerBlockPlusOne.Sub(sdk.OneDec())
	return inflationPerBlock, nil
}

func CalculateInflationNewCoins(inflationPerBlock sdk.Dec, supply sdk.Coin) (result sdk.Coins) {
	newCoinAmounts := (inflationPerBlock.MulInt(supply.Amount)).TruncateInt()
	return sdk.NewCoins(sdk.NewCoin(supply.Denom, newCoinAmounts))
}

// ValidateInflation ensures validity of Inflation object fields
func ValidateInflation(inflation Inflation) error { // TODO(JS): potentially allow inflations less than -1
	if inflation.InflationRate.LT(sdk.NewDecWithPrec(1, 2).Neg()) {
		return fmt.Errorf("inflation object param, inflation_rate, cannot be less than -1, value: %s",
			inflation.InflationRate.String())
	}

	_, err := sdk.AccAddressFromBech32(inflation.TargetAddress)
	if err != nil {
		return fmt.Errorf("inflation object param, target_address, is invalid: %s",
			inflation.TargetAddress)
	}

	err = sdk.ValidateDenom(inflation.Denom)
	if err != nil {
		return fmt.Errorf("inflation object param, denom: %s", err)
	}
	return nil
}

func ValidateInflations(i interface{}) error {
	v, ok := i.([]*Inflation)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	for _, key := range v {
		err := ValidateInflation(*key)
		if err != nil {
			return fmt.Errorf("inflation params for %s are invalid: %s", key.Denom, err)
		}
	}

	return nil
}
