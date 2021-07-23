package bls12381

import (
	"encoding/base64"
	"fmt"
	"os"

	blst "github.com/supranational/blst/bindings/go"
)

func aggregatePublicKey(pks []*PubKey) *blst.P1Affine {
	pubkeys := make([]*blst.P1Affine, len(pks))
	for i, pk := range pks {
		pubkeys[i] = new(blst.P1Affine).Deserialize(pk.Key)
		if pubkeys[i] == nil {
			panic("Failed to deserialize public key")
		}
	}

	aggregator := new(blst.P1Aggregate)
	b := aggregator.Aggregate(pubkeys, false)
	if !b {
		panic("Failed to aggregate public keys")
	}
	apk := aggregator.ToAffine()

	return apk
}

// AggregateSignature combines a set of verified signatures into a single bls signature
func AggregateSignature(sigs [][]byte) []byte {
	sigmas := make([]*blst.P2Affine, len(sigs))
	for i, sig := range sigs {
		sigmas[i] = new(blst.P2Affine).Uncompress(sig)
		if sigmas[i] == nil {
			panic("Failed to deserialize signature")
		}
	}

	aggregator := new(blst.P2Aggregate)
	b := aggregator.Aggregate(sigmas, false)
	if !b {
		panic("Failed to aggregate signatures")
	}
	aggSigBytes := aggregator.ToAffine().Compress()
	return aggSigBytes
}

// VerifyMultiSignature assumes public key is already validated
func VerifyMultiSignature(msg []byte, sig []byte, pks []*PubKey) bool {
	return VerifyAggregateSignature([][]byte{msg}, sig, [][]*PubKey{pks})
}

func Unique(msgs [][]byte) bool {
	if len(msgs) <= 1 {
		return true
	}
	msgMap := make(map[string]bool, len(msgs))
	for _, msg := range msgs {
		s := base64.StdEncoding.EncodeToString(msg)
		if _, ok := msgMap[s]; ok {
			return false
		}
		msgMap[s] = true
	}
	return true
}

func VerifyAggregateSignature(msgs [][]byte, sig []byte, pkss [][]*PubKey) bool {
	// messages must be pairwise distinct
	if !Unique(msgs) {
		fmt.Fprintf(os.Stdout, "messages must be pairwise distinct")
		return false
	}

	apks := make([]*blst.P1Affine, len(pkss))
	for i, pks := range pkss {
		apks[i] = aggregatePublicKey(pks)
	}

	sigma := new(blst.P2Affine).Uncompress(sig)
	if sigma == nil {
		panic("Failed to deserialize signature")
	}

	dst := []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")
	return sigma.AggregateVerify(true, apks, false, msgs, dst)
}
