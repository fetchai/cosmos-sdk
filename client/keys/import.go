package keys

import (
	"bufio"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"os"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
)

// ImportKeyCommand imports private keys from a keyfile.
func ImportKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <name> <keyfile>",
		Short: "Import private keys into the local keybase",
		Long:  "Import a ASCII armored private key into the local keybase.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			buf := bufio.NewReader(clientCtx.Input)

			bz, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}

			unarmored, _ := cmd.Flags().GetBool(flagUnarmoredHex)
			algo, _ := cmd.Flags().GetString(flagUnarmoredKeyAlgo)

			if unarmored {
				return clientCtx.Keyring.ImportUnarmoredPrivKey(args[0], string(bz), algo)
			}

			passphrase, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
			if err != nil {
				return err
			}

			return clientCtx.Keyring.ImportPrivKey(args[0], string(bz), passphrase)
		},
	}

	cmd.Flags().Bool(flagUnarmoredHex, false, "Import unarmored hex privkey")
	cmd.Flags().String(flagUnarmoredKeyAlgo, string(hd.Secp256k1Type), fmt.Sprintf("defines cryptographic scheme algorithm of the private key (%s, %s, %s, %s)", hd.Secp256k1Type, hd.Ed25519Type, hd.Sr25519Type, hd.MultiType))

	return cmd
}
