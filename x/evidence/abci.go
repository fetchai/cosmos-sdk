package evidence

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// BeginBlocker iterates through and handles any newly discovered evidence of
// misbehavior submitted by Tendermint. Currently, only equivocation is handled.
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k Keeper) {
	defer telemetry.ModuleMeasureSince(ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	for _, tmEvidence := range req.ByzantineValidators {
		switch tmEvidenceType := tmEvidence.Type; tmEvidenceType {
		case tmtypes.ABCIEvidenceTypeDuplicateVote:
			evidence := ConvertDuplicateVoteEvidence(tmEvidence)
			k.HandleDoubleSign(ctx, evidence.(Equivocation))
		case tmtypes.ABCIEvidenceTypeBeaconInactivity:
			fallthrough
		case tmtypes.ABCIEvidenceTypeDKG:
			evidence := ConvertBeaconEvidence(tmEvidence, tmEvidenceType)
			k.HandleBeaconInfraction(ctx, evidence.(BeaconInfraction))
		default:
			k.Logger(ctx).Error(fmt.Sprintf("ignored unknown evidence type: %s", tmEvidence.Type))
		}
	}

	k.PruneBeaconEvidence(ctx, req.Header.Height)
}
