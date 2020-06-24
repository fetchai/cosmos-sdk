package attestation

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var AttestationCdc *codec.Codec

func init() {
	AttestationCdc = codec.New()
	codec.RegisterCrypto(AttestationCdc)
	AttestationCdc.RegisterConcrete(Attestation{}, "cosmos-sdk/Attestation", nil)
	AttestationCdc.Seal()
}

// marshal
func MarshalBinaryBare(o interface{}) ([]byte, error) {
	return AttestationCdc.MarshalBinaryBare(o)
}

// unmarshal
func UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	return AttestationCdc.UnmarshalBinaryBare(bz, ptr)
}
