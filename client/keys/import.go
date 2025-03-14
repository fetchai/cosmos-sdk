package keys

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/input"
)

// ImportKeyCommand imports private keys from a keyfile.
func ImportKeyCommand() *cobra.Command {
	return &cobra.Command{
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

			passphrase, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
			if err != nil {
				return err
			}

			return clientCtx.Keyring.ImportPrivKey(args[0], string(bz), passphrase)
		},
	}
}

// ImportUnarmoredKeyCommand imports private keys from a keyfile.
func ImportUnarmoredKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-unarmored <name> [keyfile]",
		Short: "Imports unarmored private key into the local keybase",
		Long: `Imports hex encoded raw unarmored private key into the local keybase 

[keyfile] - Path to the file containing unarmored hex encoded private key.
            => *IF* this non-mandatory 2nd positional argument has been
               *PROVIDED*, then private key will be read from that file.
			
            => *ELSE* If this positional argument has been *OMITTED*, then
               user will be prompted on terminal to provide the private key
               at SECURE PROMPT = passed in characters of the key hex value
               will *not* be displayed on the terminal.

            File format: The only condition for the file format is, that
            the unarmored key must be on the first line (the file can also
            contain further lines, though they are ignored).

            The 1st line must contain only hex encoded unarmored raw value,
            serialised *exactly* as it is expected by given cryptographic
            algorithm specified by the '--unarmored-key-algo <algo>' flag
            (see the description of that flag).
            Hex key value can be preceded & followed by any number of any
            whitespace characters, they will be ignored.

Key value:
As mentioned above, key is expected to be hex encoded. Hex encoding can be
lowercase, uppercase or mixed case, it does not matter, and it can (but
does NOT need to) contain the '0x' or just 'x' prefix at the beginning of
the hex encoded value.

Output:
The command will print key info after the import, the same way as the
'keys add ...' command does. 
This is quite useful, since user will immediately see the address (and pub
key value) derived from the imported private key.`,
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
			acceptedHexValPrefixes := []string{"0x", "x"}
			for _, prefix := range acceptedHexValPrefixes {
				if strings.HasPrefix(privKeyHexLC, prefix) {
					privKeyHexLC = privKeyHexLC[len(prefix):]
					break
				}
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

	cmd.Flags().String(flagUnarmoredKeyAlgo, string(hd.Secp256k1Type), fmt.Sprintf(
		`Defines cryptographic scheme algorithm of the provided unarmored private key.
At the moment *ONLY* the "%s" and "%s" algorithms are supported.
Expected serialisation format of the raw unarmored key value:
* for "%s": 32 bytes raw private key (hex encoded)  
* for "%s": 32 bytes raw public key immediately followed by 32 bytes
                 private key = 64 bytes altogether (hex encoded)
`, hd.Secp256k1Type, hd.Ed25519Type, hd.Secp256k1Type, hd.Ed25519Type))

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
