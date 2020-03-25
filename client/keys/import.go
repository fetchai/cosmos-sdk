package keys

import (
	"bufio"
	"encoding/hex"
	"errors"
	"io/ioutil"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mintkey"
)

const (
	flagAgentRawKey = "agent-raw-key"
)

// ImportKeyCommand imports private keys from a keyfile.
func ImportKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <name> <keyfile>",
		Short: "Import private keys into the local keybase",
		Long:  "Import a ASCII armored private key into the local keybase.",
		Args:  cobra.ExactArgs(2),
		RunE:  runImportCmd,
	}
	cmd.Flags().Bool(flagAgentRawKey, false, "Signal that you want to import an key from the agent framework")

	return cmd
}

func runImportCmd(cmd *cobra.Command, args []string) error {
	buf := bufio.NewReader(cmd.InOrStdin())
	kb, err := keyring.NewKeyring(sdk.KeyringServiceName(), viper.GetString(flags.FlagKeyringBackend), viper.GetString(flags.FlagHome), buf)
	if err != nil {
		return err
	}

	bz, err := ioutil.ReadFile(args[1])
	if err != nil {
		return err
	}

	// if the user has requested that a raw agent key is imported then we must read the file as such
	if viper.GetBool(flagAgentRawKey) {
		// hex decode the input string
		rawPrivKey, err := hex.DecodeString(string(bz))
		if err != nil {
			return err
		}

		// check the size of the binary data
		if len(rawPrivKey) != 32 {
			return errors.New("Incorrect raw key size. Please check input path")
		}

		// create the underlying private key
		var priv secp256k1.PrivKeySecp256k1
		copy(priv[:], rawPrivKey)

		// armor and encrypt the key with a dummy password (to reuse existing private key import path)
		passPhrase := "dummy-not-really-used"
		armored := mintkey.EncryptArmorPrivKey(priv, passPhrase, "secp256k1")

		return kb.ImportPrivKey(args[0], armored, passPhrase)
	}

	passphrase, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
	if err != nil {
		return err
	}

	return kb.ImportPrivKey(args[0], string(bz), passphrase)
}
