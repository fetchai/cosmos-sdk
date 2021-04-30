package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter store keys
var (
	KeyMintDenom     = []byte("MintDenom")
	KeyInflationRate = []byte("InflationRate")
	KeyInflationMax  = []byte("InflationMax")
	KeyInflationMin  = []byte("InflationMin")
	KeyGoalBonded    = []byte("GoalBonded")
	KeyBlocksPerYear = []byte("BlocksPerYear")
)

// mint parameters
type Params struct {
	MintDenom     string  `json:"mint_denom" yaml:"mint_denom"`           // type of coin to mint
	InflationRate sdk.Dec `json:"inflation_rate" yaml:"inflation_rate"`   // the fixed inflation rate
	BlocksPerYear uint64  `json:"blocks_per_year" yaml:"blocks_per_year"` // expected blocks per year
}

// ParamTable for minting module.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string, inflationRate sdk.Dec, blocksPerYear uint64,
) Params {

	return Params{
		MintDenom:     mintDenom,
		InflationRate: inflationRate,
		BlocksPerYear: blocksPerYear,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:     sdk.DefaultBondDenom,
		InflationRate: sdk.NewDecWithPrec(3, 2),
		BlocksPerYear: uint64(60 * 60 * 8766 / 5), // assuming 5 second block times
	}
}

// validate params
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

func (p Params) String() string {
	return fmt.Sprintf(`Minting Params:
  Mint Denom:       %s
  Inflation Rate:   %s
  Blocks Per Year:  %d
`,
		p.MintDenom, p.InflationRate, p.BlocksPerYear,
	)
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		params.NewParamSetPair(KeyInflationRate, &p.InflationRate, validateInflationRate),
		params.NewParamSetPair(KeyBlocksPerYear, &p.BlocksPerYear, validateBlocksPerYear),
	}
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateInflationRate(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("max inflation cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("max inflation too large: %s", v)
	}

	return nil
}

func validateBlocksPerYear(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", v)
	}

	return nil
}
