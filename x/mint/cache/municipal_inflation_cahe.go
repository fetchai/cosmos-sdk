package cache

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
	"sync"
)

type MunicipalInflationCacheItem struct {
	PerBlockInflation sdk.Dec
	AnnualInflation   *types.MunicipalInflation
}

type MunicipalInflationCache struct {
	blocksPerYear uint64
	original      *[]*types.MunicipalInflationPair
	inflations    map[string]*MunicipalInflationCacheItem // {denom: inflationPerBlock}
	mu            sync.RWMutex
}

var (
	// GMunicipalInflationCache Thread safety:
	// As the things stand now from design & impl. perspective:
	//  1. This global variable is supposed to be initialised(= its value set)
	//     just *ONCE* here, at this place,
	//  2. This global variable is *NOT* used anywhere else in the initialisation
	//     context of the *global scope* - e.g. as input for initialisation of
	//     another global variable, etc. ...
	//  3. All *exported* methods of `MunicipalInflationCache` type *ARE* thread safe,
	//     and so can be called from anywhere, *EXCEPT* from the initialisation context
	//     of the global scope(implication of the point 2 above)!
	GMunicipalInflationCache = MunicipalInflationCache{}
)

func (cache *MunicipalInflationCache) Refresh(inflations *[]*types.MunicipalInflationPair, blocksPerYear uint64) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.refresh(inflations, blocksPerYear)
}

func (cache *MunicipalInflationCache) RefreshIfNecessary(inflations *[]*types.MunicipalInflationPair, blocksPerYear uint64) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.refreshIfNecessary(inflations, blocksPerYear)
}

func (cache *MunicipalInflationCache) IsRefreshRequired(blocksPerYear uint64) bool {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return cache.isRefreshRequired(blocksPerYear)
}

func (cache *MunicipalInflationCache) GetInflation(denom string) (MunicipalInflationCacheItem, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	infl, exists := cache.inflations[denom]
	return *infl, exists
}

func (cache *MunicipalInflationCache) GetOriginal() *[]*types.MunicipalInflationPair {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	// NOTE(pb): Mutex locking might not be necessary here since we are returning by pointer
	return cache.original
}

// NOTE(pb): *NOT* thread safe
func (cache *MunicipalInflationCache) refresh(inflations *[]*types.MunicipalInflationPair, blocksPerYear uint64) {
	if err := types.ValidateMunicipalInflations(inflations); err != nil {
		panic(err)
	}

	cache.blocksPerYear = blocksPerYear
	cache.original = inflations
	cache.inflations = map[string]*MunicipalInflationCacheItem{}

	for _, pair := range *inflations {
		inflationPerBlock, err := types.CalculateInflationPerBlock(pair.Inflation.Value, blocksPerYear)

		if err != nil {
			panic(err)
		}

		cache.inflations[pair.Denom] = &MunicipalInflationCacheItem{
			inflationPerBlock,
			pair.Inflation,
		}
	}
}

// NOTE(pb): *NOT* thread safe
func (cache *MunicipalInflationCache) refreshIfNecessary(inflations *[]*types.MunicipalInflationPair, blocksPerYear uint64) {
	if cache.isRefreshRequired(blocksPerYear) {
		cache.refresh(inflations, blocksPerYear)
	}
}

// NOTE(pb): *NOT* thread safe
func (cache *MunicipalInflationCache) isRefreshRequired(blocksPerYear uint64) bool {
	return cache.blocksPerYear != blocksPerYear
}
