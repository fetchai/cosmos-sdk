package cache

import (
	"sync/atomic"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

type MunicipalInflationCacheItem struct {
	PerBlockInflation sdk.Dec
	AnnualInflation   *types.MunicipalInflation
}

type MunicipalInflationCacheInternal struct {
	blocksPerYear uint64
	original      *[]*types.MunicipalInflationPair
	inflations    map[string]*MunicipalInflationCacheItem // {denom: inflationPerBlock}
}

type MunicipalInflationCache struct {
	internal atomic.Pointer[MunicipalInflationCacheInternal]
}

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
var GMunicipalInflationCache = MunicipalInflationCache{}

func (cache *MunicipalInflationCache) Refresh(inflations *[]*types.MunicipalInflationPair, blocksPerYear uint64) {
	newCache := MunicipalInflationCacheInternal{}
	newCache.refresh(inflations, blocksPerYear)
	cache.internal.Store(&newCache)
}

// RefreshIfNecessary
// IMPORTANT: Assuming *NO* concurrent writes, designed with emphasis for concurrent reads. This requirement
// guaranteed, since this method is called exclusively from call contexts which are never concurrent.
// This approach will guarantee the most possible effective cache operation in heavily concurrent
// read environment with minimum possible blocking for concurrent read operations, but with slight
// limitation for write operations (there should not be concurrent write operations).
// Most of the read operations are assumed to be done from RPC (querying municipal inflation).
func (cache *MunicipalInflationCache) RefreshIfNecessary(inflations *[]*types.MunicipalInflationPair, blocksPerYear uint64) {
	current := cache.internal.Load()
	if current.isRefreshRequired(blocksPerYear) {
		cache.Refresh(inflations, blocksPerYear)
	}
}

func (cache *MunicipalInflationCache) GetInflation(denom string) (MunicipalInflationCacheItem, bool) {
	current := cache.internal.Load()
	infl, exists := current.inflations[denom]
	return *infl, exists
}

func (cache *MunicipalInflationCache) GetOriginal() *[]*types.MunicipalInflationPair {
	current := cache.internal.Load()
	return current.original
}

// NOTE(pb): *NOT* thread safe
func (cache *MunicipalInflationCacheInternal) refresh(inflations *[]*types.MunicipalInflationPair, blocksPerYear uint64) {
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
func (cache *MunicipalInflationCacheInternal) isRefreshRequired(blocksPerYear uint64) bool {
	return cache.blocksPerYear != blocksPerYear
}
