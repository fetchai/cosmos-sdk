package v039

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	ModuleName = "mint"
)

type (
	// Minter represents the minting state.
	Minter struct {
		Inflation        sdk.Dec `json:"inflation" yaml:"inflation"`                 // current annual inflation rate
		AnnualProvisions sdk.Dec `json:"annual_provisions" yaml:"annual_provisions"` // current annual expected provisions
	}

	// mint parameters
	Params struct {
		MintDenom     string  `json:"mint_denom" yaml:"mint_denom"`           // type of coin to mint
		InflationRate sdk.Dec `json:"inflation_rate" yaml:"inflation_rate"`   // maximum annual change in inflation rate
		BlocksPerYear uint64  `json:"blocks_per_year" yaml:"blocks_per_year"` // expected blocks per year
	}

	// GenesisState - minter state
	GenesisState struct {
		Minter Minter `json:"minter" yaml:"minter"` // minter object
		Params Params `json:"params" yaml:"params"` // inflation params
	}
)
