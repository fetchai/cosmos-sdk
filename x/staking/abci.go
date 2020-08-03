package staking

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// BeginBlocker will persist the current header and validator set as a historical entry
// and prune the oldest entry based on the HistoricalEntries parameter
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k *keeper.Keeper) {
	k.TrackHistoricalInfo(ctx)
	// If two blocks before an next aeon start, need to return new validator set of next aeon
	if req.Header.Entropy.GetRound() == req.Header.Entropy.GetAeonLength()-2 {
		k.ChangeoverValidators = true
	}
}

// Called every block, update validator set
func EndBlocker(ctx sdk.Context, k *keeper.Keeper) []abci.ValidatorUpdate {
	return k.BlockValidatorUpdates(ctx)
}
