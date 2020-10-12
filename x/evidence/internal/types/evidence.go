package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"gopkg.in/yaml.v2"
)

// Evidence type constants
const (
	RouteEquivocation    = "equivocation"
	TypeEquivocation     = "equivocation"
	TypeBeaconInactivity = "beacon_inactivity"
)

var _ exported.Evidence = (*Equivocation)(nil)

// Equivocation implements the Evidence interface and defines evidence of double
// signing misbehavior.
type Equivocation struct {
	Height           int64           `json:"height" yaml:"height"`
	Time             time.Time       `json:"time" yaml:"time"`
	Power            int64           `json:"power" yaml:"power"`
	ConsensusAddress sdk.ConsAddress `json:"consensus_address" yaml:"consensus_address"`
}

// Route returns the Evidence Handler route for an Equivocation type.
func (e Equivocation) Route() string { return RouteEquivocation }

// Type returns the Evidence Handler type for an Equivocation type.
func (e Equivocation) Type() string { return TypeEquivocation }

func (e Equivocation) String() string {
	bz, _ := yaml.Marshal(e)
	return string(bz)
}

// Hash returns the hash of an Equivocation object.
func (e Equivocation) Hash() tmbytes.HexBytes {
	return tmhash.Sum(ModuleCdc.MustMarshalBinaryBare(e))
}

// ValidateBasic performs basic stateless validation checks on an Equivocation object.
func (e Equivocation) ValidateBasic() error {
	if e.Time.IsZero() {
		return fmt.Errorf("invalid equivocation time: %s", e.Time)
	}
	if e.Height < 1 {
		return fmt.Errorf("invalid equivocation height: %d", e.Height)
	}
	if e.Power < 1 {
		return fmt.Errorf("invalid equivocation validator power: %d", e.Power)
	}
	if e.ConsensusAddress.Empty() {
		return fmt.Errorf("invalid equivocation validator consensus address: %s", e.ConsensusAddress)
	}

	return nil
}

// GetConsensusAddress returns the validator's consensus address at time of the
// Equivocation infraction.
func (e Equivocation) GetConsensusAddress() sdk.ConsAddress {
	return e.ConsensusAddress
}

// GetHeight returns the height at time of the Equivocation infraction.
func (e Equivocation) GetHeight() int64 {
	return e.Height
}

// GetTime returns the time at time of the Equivocation infraction.
func (e Equivocation) GetTime() time.Time {
	return e.Time
}

// GetValidatorPower returns the validator's power at time of the Equivocation
// infraction.
func (e Equivocation) GetValidatorPower() int64 {
	return e.Power
}

// GetTotalPower is a no-op for the Equivocation type.
func (e Equivocation) GetTotalPower() int64 { return 0 }

// ConvertDuplicateVoteEvidence converts a Tendermint concrete Evidence type to
// SDK Evidence using Equivocation as the concrete type.
func ConvertDuplicateVoteEvidence(dupVote abci.Evidence) exported.Evidence {
	return Equivocation{
		Height:           dupVote.Height,
		Power:            dupVote.Validator.Power,
		ConsensusAddress: sdk.ConsAddress(dupVote.Validator.Address),
		Time:             dupVote.Time,
	}
}

//-------------------------------------------------------------------------------

var _ exported.Evidence = (*BeaconInactivity)(nil)

// BeaconInactivity implements the Evidence interface and defines evidence of
// inactivity in the random beacon
type BeaconInactivity struct {
	Height           int64           `json:"height" yaml:"height"`
	Time             time.Time       `json:"time" yaml:"time"`
	Power            int64           `json:"power" yaml:"power"`
	ConsensusAddress sdk.ConsAddress `json:"consensus_address" yaml:"consensus_address"`
	Threshold        int64           `json:"threshold" yaml:"threshold"`
}

// Route returns the Evidence Handler route for an BeaconInactivity type. We do not
// allow BeaconInactivity evidence to be submitted in transaction form. Should only
// be included in block evidence
func (e BeaconInactivity) Route() string { return "unregistered_route" }

// Type returns the Evidence Handler type for an BeaconInactivity type.
func (e BeaconInactivity) Type() string { return TypeBeaconInactivity }

func (e BeaconInactivity) String() string {
	bz, _ := yaml.Marshal(e)
	return string(bz)
}

// Hash returns the hash of an Equivocation object.
func (e BeaconInactivity) Hash() tmbytes.HexBytes {
	return tmhash.Sum(ModuleCdc.MustMarshalBinaryBare(e))
}

// ValidateBasic performs basic stateless validation checks on an Equivocation object.
func (e BeaconInactivity) ValidateBasic() error {
	if e.Time.IsZero() {
		return fmt.Errorf("invalid complaint time: %s", e.Time)
	}
	if e.Height < 1 {
		return fmt.Errorf("invalid complaint height: %d", e.Height)
	}
	if e.Power < 1 {
		return fmt.Errorf("invalid complaint validator power: %d", e.Power)
	}
	if e.ConsensusAddress.Empty() {
		return fmt.Errorf("invalid complaint validator consensus address: %s", e.ConsensusAddress)
	}
	if e.Threshold < 0 {
		return fmt.Errorf("invalid complaint threshold: %v", e.Threshold)
	}

	return nil
}

// GetConsensusAddress returns the validator's consensus address at time of the
// Equivocation infraction.
func (e BeaconInactivity) GetConsensusAddress() sdk.ConsAddress {
	return e.ConsensusAddress
}

// GetHeight returns the height at time of the Equivocation infraction.
func (e BeaconInactivity) GetHeight() int64 {
	return e.Height
}

// GetTime returns the time at time of the Equivocation infraction.
func (e BeaconInactivity) GetTime() time.Time {
	return e.Time
}

// GetValidatorPower returns the validator's power at time of the Equivocation
// infraction.
func (e BeaconInactivity) GetValidatorPower() int64 {
	return e.Power
}

// GetTotalPower is a no-op for the BeaconInactivity type.
func (e BeaconInactivity) GetTotalPower() int64 { return 0 }

// ConvertBeaconInactivityEvidence converts a Tendermint concrete Evidence type to
// SDK Evidence using BeaconInactivity as the concrete type.
func ConvertBeaconInactivityEvidence(ev abci.Evidence) exported.Evidence {
	return BeaconInactivity{
		Height:           ev.Height,
		Power:            ev.Validator.Power,
		ConsensusAddress: sdk.ConsAddress(ev.Validator.Address),
		Time:             ev.Time,
		Threshold:        ev.Threshold,
	}
}
