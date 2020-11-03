package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/bls12_381"
)

// nolint:deadcode,unused
var (
	pk1      = bls12_381.GenPrivKey().PubKey()
	pk2      = bls12_381.GenPrivKey().PubKey()
	pk3      = bls12_381.GenPrivKey().PubKey()
	addr1    = pk1.Address()
	addr2    = pk2.Address()
	addr3    = pk3.Address()
	valAddr1 = sdk.ValAddress(addr1)
	valAddr2 = sdk.ValAddress(addr2)
	valAddr3 = sdk.ValAddress(addr3)

	emptyAddr   sdk.ValAddress
	emptyPubkey crypto.PubKey
)
