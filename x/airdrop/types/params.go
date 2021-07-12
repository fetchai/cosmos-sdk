package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"strings"
)

var (
	// KeyWhiteList is store's key for WhiteList Params
	KeyWhiteList = []byte("AllowList")
)

func NewParams(whiteListClients ...string) Params {
	return Params{
		AllowList: whiteListClients,
	}
}

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyWhiteList, &p.AllowList, validateAllowList),
	}
}

func validateAllowList(i interface{}) error {
	clients, ok := i.([]string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	for _, address := range clients {
		_, err := sdk.AccAddressFromBech32(strings.TrimSpace(address))
		if err != nil {
			return fmt.Errorf("invalid addresss: %s", address)
		}
	}

	return nil
}

// IsAllowedSender checks if the given address can perform an airdrop
func (p Params) IsAllowedSender(sender sdk.AccAddress) bool {
	for _, address := range p.AllowList {
		accAddress, err := sdk.AccAddressFromBech32(strings.TrimSpace(address))
		if err != nil {
			continue
		}

		if sender.Equals(accAddress) {
			return true
		}
	}
	return false
}
