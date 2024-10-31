package keys

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"os"
	"strings"

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
			passphrase, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
			if err != nil {
				return err
			}

			bz, err := os.ReadFile(args[1])
			if err != nil {
				return err
			}

			return clientCtx.Keyring.ImportPrivKey(args[0], string(bz), passphrase)
		},
	}

	return cmd
}

// ImportUnarmoredKeyCommand imports private keys from a keyfile.
func ImportUnarmoredKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-unarmored <name> [keyfile]",
		Short: "Import unarmored private key into the local keybase",
		Long: `Import hex encoded unarmored private key into the local keybase 

Key must be hex encoded, and can be passed in either via file, or via
user password prompt. 
If the 2nd positional argument [keyfile] has been provided, private key
will be read from that file. The keyfile must contain hex encoded
unarmored raw private key on the very 1st line, and that line must
contain only the private key.
Otherwise, if the [keyfile] is not provided, the private key will be
requested via password prompt where it will be read from stdin. 
At the moment, only the secp256k1 curve/algo is supported.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			buf := bufio.NewReader(clientCtx.Input)
			var privKeyHex string
			if len(args) == 1 {
				privKeyHex, err = input.GetPassword("Enter hex encoded private key:", buf)
				if err != nil {
					return err
				}
			} else {
				filename := args[1]
				f, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
				if err != nil {
					return fmt.Errorf("open file \"%s\" error: %w", filename, err)
				}
				defer f.Close()

				sc := bufio.NewScanner(f)
				if sc.Scan() {
					firstLine := sc.Text()
					privKeyHex = strings.TrimSpace(firstLine)
				} else {
					return fmt.Errorf("unable to read 1st line from the \"%s\" file", filename)
				}

				if err := sc.Err(); err != nil {
					return fmt.Errorf("error while scanning the \"%s\" file: %w", filename, err)
				}
			}

			algo, _ := cmd.Flags().GetString(flagUnarmoredKeyAlgo)

			privKeyHexLC := strings.ToLower(privKeyHex)
			if strings.HasPrefix(privKeyHexLC, "0x") {
				privKeyHexLC = privKeyHexLC[2:]
			} else if strings.HasPrefix(privKeyHexLC, "x") {
				privKeyHexLC = privKeyHexLC[1:]
			}

			privKeyRaw, err := hex.DecodeString(privKeyHexLC)
			if err != nil {
				return fmt.Errorf("failed to decode provided hex value of private key: %w", err)
			}

			info, err := clientCtx.Keyring.ImportUnarmoredPrivKey(args[0], privKeyRaw, algo)
			if err != nil {
				return fmt.Errorf("importing unarmored private key: %w", err)

			}

			if err := printCreateUnarmored(cmd, info, clientCtx.OutputFormat); err != nil {
				return fmt.Errorf("printing private key info: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().String(flagUnarmoredKeyAlgo, string(hd.Secp256k1Type), fmt.Sprintf("defines cryptographic scheme algorithm of the private key (\"%s\", \"%s\"). At the moent *ONLY* the \"%s\" is supported. Defaults to \"%s\".", hd.Secp256k1Type, hd.Ed25519Type, hd.Secp256k1Type, hd.Secp256k1Type))

	return cmd
}

func printCreateUnarmored(cmd *cobra.Command, info keyring.Info, outputFormat string) error {
	switch outputFormat {
	case OutputFormatText:
		cmd.PrintErrln()
		printKeyInfo(cmd.OutOrStdout(), info, keyring.MkAccKeyOutput, outputFormat)
	case OutputFormatJSON:
		out, err := keyring.MkAccKeyOutput(info)
		if err != nil {
			return err
		}

		jsonString, err := KeysCdc.MarshalJSON(out)
		if err != nil {
			return err
		}

		cmd.Println(string(jsonString))

	default:
		return fmt.Errorf("invalid output format %s", outputFormat)
	}

	return nil
}
