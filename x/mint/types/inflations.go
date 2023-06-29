package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewInflation returns a new Inflation object with the given denom, target_address
// and inflation_rate
func NewInflation(denom string, targetAddress string, inflationRate sdk.Dec) *Inflation {
	return &Inflation{
		Denom:         denom,
		TargetAddress: targetAddress,
		InflationRate: inflationRate,
	}
}

func CalculateInflationPerBlock(inflation *Inflation, blocksPerYear uint64) (result sdk.Dec, err error) {
	inflationPerBlockPlusOne, err := inflation.InflationRate.Add(sdk.OneDec()).ApproxRoot(blocksPerYear)
	if err != nil {
		return
	}
	result = inflationPerBlockPlusOne.Sub(sdk.OneDec())
	return
}

func CalculateInflationIssuance(inflation sdk.Dec, supply sdk.Coin) (result sdk.Coins) {
	issuedAmount := (inflation.MulInt(supply.Amount)).TruncateInt()
	return sdk.NewCoins(sdk.NewCoin(supply.Denom, issuedAmount))
}

// Validate ensures validity of Inflation object fields

func (inflation *Inflation) Validate() error {
	// NOTE(pb): Algebraically speaking, negative inflation >= -1 is logically
	//			 valid, however it would cause issues once balance on
	//			 target_address runs out (we would need to burn tokens from all
	//			 addresses with non-zero token balance of given denomination,
	//			 what is politically & performance wise unfeasible.
	//		     To avoid issues for now, negative inflation is not allowed.
	if inflation.InflationRate.IsNegative() {
		return fmt.Errorf("inflation object param, inflation_rate, cannot be negative, value: %s",
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

func ValidateMunicipalInflations(i interface{}) (err error) {
	v, ok := i.([]*Inflation)
	if !ok {
		err = fmt.Errorf("invalid parameter type: %T", i)
		return
	}

	for _, inflation := range v {
		err = inflation.Validate()
		if err != nil {
			return
		}
	}

	return
}
