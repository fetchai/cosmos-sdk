package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// Simulation parameter constants
const (
	Inflation     = "inflation"
	InflationRate = "inflation_rate"
)

// GenInflation randomized Inflation
func GenInflation(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenInflationRate randomized InflationRate
func GenInflationRate(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
}

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(simState *module.SimulationState) {
	// minter
	var inflation sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, Inflation, &inflation, simState.Rand,
		func(r *rand.Rand) { inflation = GenInflation(r) },
	)

	// params
	var inflationRateChange sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, InflationRate, &inflationRateChange, simState.Rand,
		func(r *rand.Rand) { inflationRateChange = GenInflationRate(r) },
	)

	mintDenom := sdk.DefaultBondDenom
	blocksPerYear := uint64(60 * 60 * 8766 / 5)
	params := types.NewParams(mintDenom, inflationRateChange, blocksPerYear)

	mintGenesis := types.NewGenesisState(types.InitialMinter(inflation), params)

	bz, err := json.MarshalIndent(&mintGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(mintGenesis)
}
