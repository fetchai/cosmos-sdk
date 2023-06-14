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

// ValidateInflation ensures validity of Inflation object fields
func ValidateInflation(inflation Inflation) error {
	if inflation.InflationRate.IsNegative() {
		return fmt.Errorf("inflation object param, inflation_rate, should be positive, is %s",
			inflation.InflationRate.String())
	}
	if inflation.InflationRate.GT(sdk.OneDec()) {
		return fmt.Errorf("inflation object param, inflation_rate, cannot be more than 100%%, is %s",
			inflation.InflationRate.String())
	}
	if inflation.TargetAddress == "" {
		return fmt.Errorf("inflation object param, target_address, cannot be empty")
	}
	if inflation.Denom == "" {
		return fmt.Errorf("inflation object param, denom, cannot be empty")
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
