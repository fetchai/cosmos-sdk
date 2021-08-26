package server

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/bls12381"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

func VerifyAggSignature(msgs [][]byte, msgCheck bool, sig []byte, pkss [][]cryptotypes.PubKey) error {
	pkssBls := make([][]*bls12381.PubKey, len(pkss))
	for i, pks := range pkss {
		for _, pk := range pks {
			pkBls, ok := pk.(*bls12381.PubKey)
			if !ok {
				return fmt.Errorf("only support bls public key")
			}
			pkssBls[i] = append(pkssBls[i], pkBls)
		}
	}

	return bls12381.VerifyAggregateSignature(msgs, msgCheck, sig, pkssBls)
}
