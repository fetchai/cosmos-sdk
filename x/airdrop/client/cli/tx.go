package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	"github.com/spf13/cobra"
)

func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Airdrop transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(NewCreateTxCmd())

	return txCmd
}

func NewCreateTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create [from_key_or_address] [amount] [drip_amount]",
		Short: `Creates a new airdrop fund. Note, the '--from' flag is
ignored as it is implied from [from_key_or_address].`,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Creates a new airdrop fund.

When creating an airdrop fund, the sender transfers a specified amount of funds to the block chain along with
the drip amount. Every block up to a maximum of drip_amount of funds are added to the rewards of the current block.
The maximum duration of the airdrop is therefore calculated as amount / drip_amount blocks. 

Example:
  $ %s tx %s create [address] [amount] [drip_amount]
  $ %s tx %s create fetch1se8mjg4mtvy8zaf4599m84xz4atn59dlqmwhnl 200000000000000000000afet 2000000000000000000
`,
				version.AppName, types.ModuleName, version.AppName, types.ModuleName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Flags().Set(flags.FlagFrom, args[0])
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			dripAmount, okay := sdk.NewIntFromString(args[2])
			if !okay || !dripAmount.IsPositive() {
				return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "%s is not a valid drip rate", args[2])
			}

			msg := types.NewMsgAirDrop(clientCtx.GetFromAddress(), coin, dripAmount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
