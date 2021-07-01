package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	ModuleName = "airdrop"
	StoreKey = ModuleName
	RouterKey = ModuleName
)

var (
	ActiveFundKeyPrefix = []byte{0x02}
)

func GetActiveFundKey(address sdk.AccAddress) []byte {
	return append(ActiveFundKeyPrefix, address.Bytes()...)
}

func GetAddressFromActiveFundKey(key []byte) sdk.AccAddress {
	addr := key[1:]
	if len(addr) != sdk.AddrLen {
		panic("unexpected key length")
	}
	return sdk.AccAddress(addr)
}

