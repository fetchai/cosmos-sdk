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

// MunicipalInflationCache Cache optimised for concurrent reading performance.
// *NO* support for concurrent writing operations.
type MunicipalInflationCache struct {
	internal atomic.Value
}

// GMunicipalInflationCache Thread safety:
// As the things stand now from design & impl. perspective:
//  1. This global variable is supposed to be initialised(= its value set)
//     just *ONCE* here, at this place,
//  2. This global variable shall *NOT* be used anywhere else in the initialisation
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
// IMPORTANT: Assuming *NO* concurrent writes. This requirement is guaranteed given the *current*
// usage of this component = this method is called exclusively from non-concurrent call contexts.
// This approach will guarantee the most possible effective cache operation in heavily concurrent
// read environment = with minimum possible blocking for concurrent read operations, but with slight
// limitation for write operations (= no concurrent write operations).
// Most of the read operations are assumed to be done from RPC (querying municipal inflation),
// and since threading models of the RPC implementation is not know, the worst scenario(= heavily
// concurrent threading model) for read operation is assumed.
func (cache *MunicipalInflationCache) RefreshIfNecessary(inflations *[]*types.MunicipalInflationPair, blocksPerYear uint64) {
	if val := cache.internal.Load(); val == nil || val.(*MunicipalInflationCacheInternal).isRefreshRequired(blocksPerYear) {
		cache.Refresh(inflations, blocksPerYear)
	}
}

func (cache *MunicipalInflationCache) GetInflation(denom string) *MunicipalInflationCacheItem {
	val := cache.internal.Load()
	if val == nil {
		return nil
	}

	infl, exists := val.(*MunicipalInflationCacheInternal).inflations[denom]

	if exists {
		return infl
	}

	return nil
}

func (cache *MunicipalInflationCache) GetOriginal() *[]*types.MunicipalInflationPair {
	val := cache.internal.Load()
	if val == nil {
		return &[]*types.MunicipalInflationPair{}
	}

	current := val.(*MunicipalInflationCacheInternal)
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
