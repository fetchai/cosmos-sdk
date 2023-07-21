package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewMunicipalInflation returns a new AnnualInflation object with the given denom, target_address
// and inflation_rate
func NewMunicipalInflation(targetAddress string, inflation sdk.Dec) *MunicipalInflation {
	return &MunicipalInflation{
		TargetAddress: targetAddress,
		Inflation:     inflation,
	}
}

func CalculateInflationPerBlock(inflation sdk.Dec, blocksPerYear uint64) (result sdk.Dec, err error) {
	inflationPerBlockPlusOne, err := inflation.Add(sdk.OneDec()).ApproxRoot(blocksPerYear)
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

// Validate ensures validity of AnnualInflation object fields

func (inflation *MunicipalInflation) Validate() error {
	// NOTE(pb): Algebraically speaking, negative inflation >= -1 is logically
	//			 valid, however it would cause issues once balance on
	//			 target_address runs out (we would need to burn tokens from all
	//			 addresses with non-zero token balance of given denomination,
	//			 what is politically & performance wise unfeasible.
	//		     To avoid issues for now, negative inflation is not allowed.
	if inflation.Inflation.IsNegative() {
		return fmt.Errorf("inflation object param, inflation_rate, cannot be negative, value: %s",
			inflation.Inflation.String())
	}

	_, err := sdk.AccAddressFromBech32(inflation.TargetAddress)
	if err != nil {
		return fmt.Errorf("inflation object param, target_address, is invalid: %s",
			inflation.TargetAddress)
	}

	return nil
}

func ValidateMunicipalInflations(inflations *[]*MunicipalInflationPair) (err error) {
	var denoms map[string]struct{}
	for _, pair := range *inflations {

		_, exists := denoms[pair.Denom]
		if exists {
			return fmt.Errorf("municipal inflation: denomination \"%s\" defined more than once", pair.Denom)
		}

		err = sdk.ValidateDenom(pair.Denom)
		if err != nil {
			return fmt.Errorf("inflation object param, denom: %s", err)
		}

		err = pair.Inflation.Validate()
		if err != nil {
			return
		}
	}

	return
}
