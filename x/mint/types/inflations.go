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

func CalculateInflation(inflation *Inflation, blocksPerYear uint64, supply sdk.Coin) (result sdk.Coins, err error) {
	inflationPerBlockPlusOne, err := inflation.InflationRate.Add(sdk.OneDec()).ApproxRoot(blocksPerYear)
	//fmt.Println("inflationPerBlockPlusOne := " + inflationPerBlockPlusOne.String())
	if err != nil {
		return nil, fmt.Errorf("CalculateInflationError: %s", err)
	}
	inflationPerBlock := inflationPerBlockPlusOne.Sub(sdk.OneDec())
	//fmt.Println("inflationPerBlock := " + inflationPerBlock.String())
	newCoinAmounts := (inflationPerBlock.MulInt(supply.Amount)).TruncateInt()
	//fmt.Println("newCoinAmounts := " + newCoinAmounts.String())
	return sdk.NewCoins(sdk.NewCoin(inflation.Denom, newCoinAmounts)), nil
}

// ValidateInflation ensures validity of Inflation object fields
// TODO(JS): potentially to introduce negative inflations
func ValidateInflation(inflation Inflation) error {
	if inflation.InflationRate.IsNegative() {
		return fmt.Errorf("inflation object param, inflation_rate, should be positive, is %s",
			inflation.InflationRate.String())
	}
	//if inflation.InflationRate.GT(sdk.OneDec()) {
	//	return fmt.Errorf("inflation object param, inflation_rate, cannot be more than 100%%, is %s",
	//		inflation.InflationRate.String())
	//}

	if inflation.TargetAddress == "" {
		return fmt.Errorf("inflation object param, target_address, cannot be empty")
	}

	err := sdk.ValidateDenom(inflation.Denom)
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
