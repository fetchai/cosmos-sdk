package types

import (
	"errors"
	"fmt"
	"math"
	"strings"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewParams returns Params instance with the given values.
func NewParams(mintDenom string, inflationRate sdkmath.LegacyDec, blocksPerYear uint64) Params {
	return Params{
		MintDenom:     mintDenom,
		InflationRate: inflationRate,
		BlocksPerYear: blocksPerYear,
	}
}

// DefaultParams returns default x/mint module parameters.
func DefaultParams() Params {
	return Params{
		MintDenom:     sdk.DefaultBondDenom,
		InflationRate: sdkmath.LegacyNewDecWithPrec(3, 2),
		BlocksPerYear: uint64(60 * 60 * 8766 / 5), // assuming 5 second block times
	}
}

// Validate does the sanity check on the params.
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateInflationRate(p.InflationRate); err != nil {
		return err
	}
	if err := validateBlocksPerYear(p.BlocksPerYear); err != nil {
		return err
	}

	return nil
}

func validateMintDenom(denom string) error {
	if strings.TrimSpace(denom) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(denom); err != nil {
		return err
	}

	return nil
}

func validateInflationRate(inflationRate sdkmath.LegacyDec) error {
	if inflationRate.IsNil() {
		return fmt.Errorf("inflation rate cannot be nil: %s", inflationRate)
	}
	if inflationRate.IsNegative() {
		return fmt.Errorf("inflation rate cannot be negative: %s", inflationRate)
	}
	if inflationRate.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("inflation rate too large: %s", inflationRate)
	}

	return nil
}

func validateBlocksPerYear(blocksPerYear uint64) error {
	if blocksPerYear == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", blocksPerYear)
	}

	if blocksPerYear > math.MaxInt64 {
		return fmt.Errorf("blocks per year too large: %d, maximum value is: %d", blocksPerYear, math.MaxInt64)
	}

	return nil
}
