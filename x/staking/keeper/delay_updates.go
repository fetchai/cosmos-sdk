package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	validatorUpdateKey = []byte("validatorUpdateKey")
)

func (k Keeper) CheckValidatorUpdates(ctx sdk.Context, header abci.Header) {
	// If two blocks before an next aeon start, need to return new validator set of next aeon
	fmt.Printf("beginBlock: entropy round %v, aeon length %v \n", header.Entropy.GetRound(), header.Entropy.GetAeonLength())
	if header.Entropy.GetRound() == header.Entropy.GetAeonLength()-2 {
		store := ctx.KVStore(k.storeKey)
		store.Set(validatorUpdateKey, []byte{0})
		fmt.Printf("beginBlock: changeover validators set \n")
	}
}

func (k Keeper) PerformValidatorUpdates(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	if len(store.Get(validatorUpdateKey)) == 0 {
		return false
	}
	store.Set(validatorUpdateKey, []byte{})
	fmt.Printf("endBlock: compute validator updates \n")
	return true
}