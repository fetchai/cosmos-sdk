package keys

import (
	"bufio"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/bech32"

	"github.com/cosmos/cosmos-sdk/client/attestation"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func AttestationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attestation",
		Short: "Create or verify attestations",
		Long:  `Create or verify offline proofs to demonstrate ownership of keys`,
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "create [name]",
			Short: "Create an attestation",
			Long:  "Create and attestation for one of the keys present in the store",
			Args:  cobra.ExactArgs(1),
			RunE:  runAttestationCreate,
		},
		&cobra.Command{
			Use:   "verify [address] [attestation]",
			Short: "Verify an attestation",
			Long:  "Given an attestation and address, verify that one proves ownership of the other",
			Args:  cobra.ExactArgs(2),
			RunE:  runAttestationVerify,
		},
	)

	return cmd
}

func runAttestationCreate(cmd *cobra.Command, args []string) error {
	buf := bufio.NewReader(cmd.InOrStdin())
	kb, err := keys.NewKeyring(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), buf)
	if err != nil {
		return err
	}

	decryptPassword, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
	if err != nil {
		return err
	}

	privKey, err := kb.ExportPrivateKeyObject(args[0], decryptPassword)
	if err != nil {
		return err
	}

	att, err := attestation.NewAttestation(privKey)
	if err != nil {
		return err
	}

	cmd.Println(att.String())

	return nil
}

func runAttestationVerify(cmd *cobra.Command, args []string) error {
	_, bz, err := bech32.DecodeAndConvert(args[0])
	if err != nil {
		return err
	}

	var address crypto.Address = bz

	// create the attestation
	att, err := attestation.NewAttestationFromString(args[1])
	if err != nil {
		return err
	}

	// verification check
	verified := att.Verify(address)
	if !verified {
		return errors.New("verification failed")
	}

	fmt.Println("verification successful")

	return nil
}
