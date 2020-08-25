package attestation

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
)

func generatePrivateKey(t *testing.T) crypto.PrivKey {

	// generate seed for the private key
	bz := make([]byte, 32)
	_, err := rand.Read(bz)
	require.NoError(t, err)

	// create the key
	return keys.SecpPrivKeyGen(bz)
}

func TestBasicAttestation(t *testing.T) {
	privKey := generatePrivateKey(t)

	// create the attestation
	att, err := NewAttestation(privKey)
	require.NoError(t, err)

	// verify that it is correct
	require.True(t, att.Verify(privKey.PubKey().Address()))
}

func TestBasicAttestationMarshalling(t *testing.T) {
	privKey := generatePrivateKey(t)

	// create the attestation
	att, err := NewAttestation(privKey)
	require.NoError(t, err)

	// marshall it to binary
	bz, err := MarshalBinaryBare(att)
	require.NoError(t, err)

	// recover the attestation
	recoveredAtt := &Attestation{}
	err = UnmarshalBinaryBare(bz, recoveredAtt)
	require.NoError(t, err)

	require.True(t, att.PublicKey.Equals(recoveredAtt.PublicKey))
	require.True(t, bytes.Equal(att.Signature, recoveredAtt.Signature))

	// check that it sis correct
	require.True(t, recoveredAtt.Verify(att.PublicKey.Address()))
}

func TestAttestationAsString(t *testing.T) {
	privKey := generatePrivateKey(t)

	// create the attestation
	att, err := NewAttestation(privKey)
	require.NoError(t, err)

	// recover the attestation
	recovered, err := NewAttestationFromString(att.String())
	require.NoError(t, err)

	require.True(t, att.PublicKey.Equals(recovered.PublicKey))
	require.True(t, bytes.Equal(att.Signature, recovered.Signature))
	require.True(t, recovered.Verify(privKey.PubKey().Address()))
}
