package bls12381_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/bls12381"
	bench "github.com/cosmos/cosmos-sdk/crypto/keys/internal/benchmarking"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"testing"
)

func TestSignAndValidateBls12381(t *testing.T) {
	privKey := bls12381.GenPrivKey()
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(1000)
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifySignature(msg, sig))

}


func BenchmarkSignBls(b *testing.B) {
	privKey := bls12381.GenPrivKey()

	bench.BenchmarkSigning(b, privKey)

}


func BenchmarkVerifyBls(b *testing.B) {
	privKey := bls12381.GenPrivKey()

	bench.BenchmarkVerification(b, privKey)
}