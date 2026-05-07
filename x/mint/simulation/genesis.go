package simulation

import (
	"math/rand"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// Simulation parameter constants
const (
	Inflation     = "inflation"
	InflationRate = "inflation_rate"
)

// GenInflation randomized Inflation
func GenInflation(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenInflationRate randomized InflationRate
func GenInflationRate(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(int64(r.Intn(99)), 2)
}

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(simState *module.SimulationState) {
	// minter
	var inflation math.LegacyDec
	simState.AppParams.GetOrGenerate(Inflation, &inflation, simState.Rand, func(r *rand.Rand) { inflation = GenInflation(r) })

	// params
	var inflationRate math.LegacyDec
	simState.AppParams.GetOrGenerate(InflationRate, &inflationRate, simState.Rand, func(r *rand.Rand) { inflationRate = GenInflationRate(r) })

	mintDenom := simState.BondDenom
	blocksPerYear := uint64(60 * 60 * 8766 / 5)
	params := types.NewParams(mintDenom, inflationRate, blocksPerYear)

	mintGenesis := types.NewGenesisState(types.InitialMinter(inflation), params)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(mintGenesis)
}
