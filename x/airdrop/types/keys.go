package types

import (
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "airdrop"
	StoreKey   = ModuleName
	RouterKey  = ModuleName
)

var (
	ActiveFundKeyPrefix = []byte{0x01}
)

func GetActiveFundKey(address sdk.AccAddress) []byte {
	return append(ActiveFundKeyPrefix, address.Bytes()...)
}

func GetAddressFromActiveFundKey(key []byte) (sdk.AccAddress, error) {
	addr := key[1:]
	if err := sdk.VerifyAddressFormat(addr); err != nil {
		return nil, fmt.Errorf("invalid address: %v", err)
	}

	return sdk.AccAddress(addr), nil
}
