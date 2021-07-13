package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

func NewGenesisState(params Params, funds []ActiveFund) *GenesisState {
	return &GenesisState{params, funds}
}

func (gs *GenesisState) Validate() error {
	allowListAddresses := []sdk.AccAddress{}

	for _, address := range gs.Params.AllowList {
		accAddress, err := sdk.AccAddressFromBech32(strings.TrimSpace(address))
		if err != nil {
			return err
		}
		allowListAddresses = append(allowListAddresses, accAddress)
	}

	return nil
}
