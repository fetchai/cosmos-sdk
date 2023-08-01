package cli

import (
	"encoding/base64"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/x/verification/types"
)

// NewTxCmd returns a root CLI command handler for all x/bank transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Verification transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(NewSendTxCmd())

	return txCmd
}

// NewSendTxCmd returns a CLI command handler for creating a MsgSend transaction.
func NewSendTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "send [from_key_or_address] [base64_encoded_data]",
		Short: `send verification message from account with arbitrary data. Note, the'--from' flag is
ignored as it is implied from [from_key_or_address].`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			decoded_data, err := base64.StdEncoding.DecodeString(args[1])

			msg := &types.MsgSignData{
				FromAddress: clientCtx.GetFromAddress().String(),
				Data:        decoded_data,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
