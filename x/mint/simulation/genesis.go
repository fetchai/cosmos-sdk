package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/bank/simulation"
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

// GenMunicipalInflation randomized Municipal Inflation configuration
func GenMunicipalInflation(simState *module.SimulationState) []*types.MunicipalInflationPair {
	r := simState.Rand

	coins := make([]*sdk.Coin, len(simulation.AdditionalTestSupply))
	for i := 0; i < len(simulation.AdditionalTestSupply); i++ {
		coins[i] = &simulation.AdditionalTestSupply[i]
	}

	len_ := r.Intn(len(coins) + 1)
	municipalInflation := make([]*types.MunicipalInflationPair, len_)
	for i := 0; i < len_; i++ {
		lenCoins := len(coins)
		lastIdx := lenCoins - 1
		rndIdx := r.Intn(lenCoins)
		fmt.Println(">>>>>>>>>>>>>>>", coins, "rndIdx:", rndIdx)
		c := coins[rndIdx]
		coins[rndIdx] = coins[lastIdx]
		coins = coins[:lastIdx]

		acc := &simState.Accounts[r.Intn(len(simState.Accounts))]
		infl := sdk.NewDecWithPrec(r.Int63n(201), 2)
		municipalInflation[i] = &types.MunicipalInflationPair{Denom: c.Denom, Inflation: types.NewMunicipalInflation(acc.Address.String(), infl)}
	}

	return municipalInflation
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

	minter := types.InitialMinter(inflation)
	minter.MunicipalInflation = GenMunicipalInflation(simState)
	mintGenesis := types.NewGenesisState(minter, params)

	bz, err := json.MarshalIndent(&mintGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(mintGenesis)
}
