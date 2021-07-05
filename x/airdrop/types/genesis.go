package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

func NewGenesisState(params Params, funds []ActiveFund) *GenesisState {
	return &GenesisState{params, funds}
}

func (gs *GenesisState) Validate() error {
	whiteListAddresses := []sdk.AccAddress{}

	for _, address := range gs.Params.AllowList {
		accAddress, err := sdk.AccAddressFromBech32(strings.TrimSpace(address))
		if err != nil {
			return err
		}
		whiteListAddresses = append(whiteListAddresses, accAddress)
	}

	return nil
}
