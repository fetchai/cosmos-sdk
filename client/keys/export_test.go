package keys

import (
	"bufio"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/testutil"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_runExportCmd(t *testing.T) {
	cmd := ExportKeyCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)

	// Now add a temporary keybase
	kbHome := t.TempDir()

	// create a key
	kb, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendTest, kbHome, mockIn)
	require.NoError(t, err)
	t.Cleanup(func() {
		kb.Delete("keyname1") // nolint:errcheck
	})

	path := sdk.GetConfig().GetFullBIP44Path()
	_, err = kb.NewAccount("keyname1", testutil.TestMnemonic, "", path, hd.Secp256k1)
	require.NoError(t, err)

	// Now enter password
	args := []string{
		"keyname1",
		fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
		fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kbHome := t.TempDir()
			defaultArgs := []string{
				"keyname1",
				fmt.Sprintf("--%s=%s", flags.FlagHome, kbHome),
				fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, tc.keyringBackend),
			}

			cmd := ExportKeyCommand()
			cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())

			cmd.SetArgs(append(defaultArgs, tc.extraArgs...))
			mockIn, mockOut := testutil.ApplyMockIO(cmd)

			mockIn.Reset(tc.userInput)
			mockInBuf := bufio.NewReader(mockIn)

			// create a key
			kb, err := keyring.New(sdk.KeyringServiceName(), tc.keyringBackend, kbHome, bufio.NewReader(mockInBuf))
			require.NoError(t, err)
			t.Cleanup(func() {
				kb.Delete("keyname1") // nolint:errcheck
			})

			path := sdk.GetConfig().GetFullFundraiserPath()
			_, err = kb.NewAccount("keyname1", testutil.TestMnemonic, "", path, hd.Secp256k1)
			require.NoError(t, err)

			clientCtx := client.Context{}.
				WithKeyringDir(kbHome).
				WithKeyring(kb).
				WithInput(mockInBuf)
			ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

			err = cmd.ExecuteContext(ctx)
			if tc.mustFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedOutput, mockOut.String())
			}
		})
	}
}
