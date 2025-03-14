package keys

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runImportCmd(t *testing.T) {
	testCases := []struct {
		name           string
		keyringBackend string
		userInput      string
		expectError    bool
	}{
		{
			name:           "test backend success",
			keyringBackend: keyring.BackendTest,
			// key armor passphrase
			userInput: "123456789\n",
		},
		{
			name:           "test backend fail with wrong armor pass",
			keyringBackend: keyring.BackendTest,
			userInput:      "987654321\n",
			expectError:    true,
		},
		{
			name:           "file backend success",
			keyringBackend: keyring.BackendFile,
			// key armor passphrase + keyring password x2
			userInput: "123456789\n12345678\n12345678\n",
		},
		{
			name:           "file backend fail with wrong armor pass",
			keyringBackend: keyring.BackendFile,
			userInput:      "987654321\n12345678\n12345678\n",
			expectError:    true,
		},
		{
			name:           "file backend fail with wrong keyring pass",
			keyringBackend: keyring.BackendFile,
			userInput:      "123465789\n12345678\n87654321\n",
			expectError:    true,
		},
		{
			name:           "file backend fail with no keyring pass",
			keyringBackend: keyring.BackendFile,
			userInput:      "123465789\n",
			expectError:    true,
		},
	}

	armoredKey := `-----BEGIN TENDERMINT PRIVATE KEY-----
salt: A790BB721D1C094260EA84F5E5B72289
kdf: bcrypt

HbP+c6JmeJy9JXe2rbbF1QtCX1gLqGcDQPBXiCtFvP7/8wTZtVOPj8vREzhZ9ElO
3P7YnrzPQThG0Q+ZnRSbl9MAS8uFAM4mqm5r/Ys=
=f3l4
-----END TENDERMINT PRIVATE KEY-----
`

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := ImportKeyCommand()
			cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
			mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

			// Now add a temporary keybase
			kbHome := t.TempDir()
			kb, err := keyring.New(sdk.KeyringServiceName(), tc.keyringBackend, kbHome, nil)

			clientCtx := client.Context{}.
				WithKeyringDir(kbHome).
				WithKeyring(kb).
				WithInput(mockIn)
			ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

			require.NoError(t, err)
			t.Cleanup(func() {
				kb.Delete("keyname1") // nolint:errcheck
			})

			keyfile := filepath.Join(kbHome, "key.asc")

			require.NoError(t, os.WriteFile(keyfile, []byte(armoredKey), 0o644))

			defer func() {
				_ = os.RemoveAll(kbHome)
			}()

			mockIn.Reset(tc.userInput)
			cmd.SetArgs([]string{
				"keyname1", keyfile,
				fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, tc.keyringBackend),
			})

			err = cmd.ExecuteContext(ctx)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

type UnarmoredKeyTestConfig struct {
	desc        string
	key         string
	expectError bool
}

func createTestConfigs(descrPreface string, hexKeyWithout0x string, expectError bool) []UnarmoredKeyTestConfig {
	if len(descrPreface) > 0 {
		descrPreface = descrPreface + ", "
	}

	testsConfigs := []UnarmoredKeyTestConfig{
		{descrPreface + "lower case, prefix 0x", "0x" + strings.ToLower(hexKeyWithout0x), expectError},
		{descrPreface + "upper case, prefix 0x", "0x" + strings.ToUpper(hexKeyWithout0x), expectError},
		{descrPreface + "mixed case, prefix 0x", "0x" + hexKeyWithout0x, expectError},
		{descrPreface + "lower case, prefix x", "x" + strings.ToLower(hexKeyWithout0x), expectError},
		{descrPreface + "upper case, prefix x", "x" + strings.ToUpper(hexKeyWithout0x), expectError},
		{descrPreface + "mixed case, prefix x", "x" + hexKeyWithout0x, expectError},
		{descrPreface + "lower case", strings.ToLower(hexKeyWithout0x), expectError},
		{descrPreface + "upper case", strings.ToUpper(hexKeyWithout0x), expectError},
		{descrPreface + "mixed case", hexKeyWithout0x, expectError},
		{descrPreface + "leading&trailing whitespaces", "\t   \t " + hexKeyWithout0x + "   \t \t", expectError},
		{
			descrPreface + "multi-line leading&trailing whitespaces",
			"\t   \t " + hexKeyWithout0x + "   \t \t\nbla bla\nsome more bla\n",
			expectError,
		},
	}
	return testsConfigs
}

func Test_runImportUnarmoredCmd(t *testing.T) {
	keyringBackend := keyring.BackendTest
	unarmoredKeyMixedCase := "8BdFbD2eaad5dc4324d19fAbed72882709dc080b39e61044d51b91A6e38F6871"
	// expectedAddress := "fetch1wurz7uwmvchhc8x0yztc7220hxs9jxdjdsrqmn"

	unarmoredKeyMixedCaseTooLong := unarmoredKeyMixedCase + "Aa"
	unarmoredKeyMixedCaseTooShort := unarmoredKeyMixedCase[:len(unarmoredKeyMixedCase)-1]

	testConfigs := createTestConfigs("[Expected Pass]", unarmoredKeyMixedCase, false)
	testConfigs = append(testConfigs, createTestConfigs("[Expected Failure] too short", unarmoredKeyMixedCaseTooShort, true)...)
	testConfigs = append(testConfigs, createTestConfigs("[Expected Failure] too long", unarmoredKeyMixedCaseTooLong, true)...)

	for _, tc := range testConfigs {
		runImportUnarmoredCmd := func(commandFnc func(cmd *cobra.Command, testConfig *UnarmoredKeyTestConfig, mockInBufferReader *testutil.BufferReader, kbHome string) error, keyringBackend string, cleanup func(kbHome string)) {
			cmd := ImportUnarmoredKeyCommand()
			cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
			mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

			// Now add a temporary keybase
			kbHome := t.TempDir()
			kb, err := keyring.New(sdk.KeyringServiceName(), keyringBackend, kbHome, nil)

			clientCtx := client.Context{}.
				WithKeyringDir(kbHome).
				WithKeyring(kb).
				WithInput(mockIn)
			ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

			require.NoError(t, err)
			t.Cleanup(func() {
				kb.Delete("keyname1") // nolint:errcheck
			})

			defer cleanup(kbHome)
			err = commandFnc(cmd, &tc, &mockIn, kbHome)
			require.NoError(t, err)

			err = cmd.ExecuteContext(ctx)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		}

		t.Run(tc.desc, func(t *testing.T) {
			runImportUnarmoredCmd(func(cmd *cobra.Command, testConfig *UnarmoredKeyTestConfig, mockInBufferReader *testutil.BufferReader, kbHome string) error {
				keyfile := filepath.Join(kbHome, "key.asc")

				err := os.WriteFile(keyfile, []byte(tc.key), 0o644)
				if err != nil {
					return err
				}

				cmd.SetArgs([]string{
					"keyname1", keyfile,
					fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyringBackend),
				})

				return nil
			}, keyringBackend, func(kbHome string) { _ = os.RemoveAll(kbHome) })
		})

		t.Run(tc.desc, func(t *testing.T) {
			runImportUnarmoredCmd(func(cmd *cobra.Command, testConfig *UnarmoredKeyTestConfig, mockInBufferReader *testutil.BufferReader, kbHome string) error {
				(*mockInBufferReader).Reset(tc.key)
				cmd.SetArgs([]string{
					"keyname1",
					fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyringBackend),
				})
				return nil
			}, keyringBackend, func(_ string) {})
		})
	}
}
