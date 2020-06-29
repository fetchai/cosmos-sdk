package attestation

import (
	"bytes"
	"encoding/hex"
	"github.com/tendermint/tendermint/crypto"
)

type Attestation struct {
	PublicKey crypto.PubKey
	Signature []byte
}

func NewAttestation(key crypto.PrivKey) (*Attestation, error) {

	// create the basic attestation
	att := &Attestation{
		PublicKey: key.PubKey(),
		Signature: []byte{},
	}

	// sign the attestation
	err := att.sign(key)
	if err != nil {
		return nil, err
	}

	return att, nil
}

func NewAttestationFromString(encoded string) (*Attestation, error) {

	// decode the string
	bz, err := hex.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	// unmarshall the attestation
	att := &Attestation{}
	err = UnmarshalBinaryBare(bz, att)
	if err != nil {
		return nil, err
	}

	return att, nil
}

func (at *Attestation) sign(key crypto.PrivKey) error {

	// sign the payload
	signature, err := key.Sign(at.PublicKey.Address().Bytes())
	if err != nil {
		return err
	}

	// update the signature
	at.Signature = signature

	return nil
}

func (at *Attestation) Verify(address crypto.Address) bool {

	// ensure that the address derived from the public key matches the required address
	if !bytes.Equal(at.PublicKey.Address().Bytes(), address.Bytes()) {
		return false
	}

	// validate the signature present matches the public key
	return at.PublicKey.VerifyBytes(at.PublicKey.Address().Bytes(), at.Signature)
}

func (at *Attestation) Bytes() []byte {
	return AttestationCdc.MustMarshalBinaryBare(at)
}

func (at *Attestation) String() string {
	return hex.EncodeToString(at.Bytes())
}
