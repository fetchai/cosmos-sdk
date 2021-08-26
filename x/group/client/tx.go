package client

import (
	"bytes"
	"encoding/base64"
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/bls12381"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	"github.com/gogo/protobuf/types"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/x/group"
	gogotypes "github.com/gogo/protobuf/types"
)

const (
	FlagExec = "exec"
	ExecTry  = "try"
)

// TxCmd returns a root CLI command handler for all x/group transaction commands.
func TxCmd(name string) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        name,
		Short:                      "Group transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		MsgCreateGroupCmd(),
		MsgUpdateGroupAdminCmd(),
		MsgUpdateGroupMetadataCmd(),
		MsgUpdateGroupMembersCmd(),
		MsgCreateGroupAccountCmd(),
		MsgUpdateGroupAccountAdminCmd(),
		MsgUpdateGroupAccountDecisionPolicyCmd(),
		MsgUpdateGroupAccountMetadataCmd(),
		MsgCreateProposalCmd(),
		MsgVoteCmd(),
		MsgExecCmd(),
		MsgVoteAggCmd(),
		GetVoteBasicCmd(),
		GetVerifyVoteBasicCmd(),
	)

	return txCmd
}

// MsgCreateGroupCmd creates a CLI command for Msg/CreateGroup.
func MsgCreateGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create-group [admin] [metadata] [members-json-file]",
		Short: "Create a group which is an aggregation " +
			"of member accounts with associated weights and " +
			"an administrator account. Note, the '--from' flag is " +
			"ignored as it is implied from [admin].",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a group which is an aggregation of member accounts with associated weights and
an administrator account. Note, the '--from' flag is ignored as it is implied from [admin].
Members accounts can be given through a members JSON file that contains an array of members.

Example:
$ %s tx group create-group [admin] [metadata] [members-json-file]

Where members.json contains:

{
	"members": [
		{
			"address": "addr1",
			"weight": "1",
			"metadata": "some metadata"
		},
		{
			"address": "addr2",
			"weight": "1",
			"metadata": "some metadata"
		}
	]
}
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			members, err := parseMembers(clientCtx, args[2])
			if err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[1])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			msg := &group.MsgCreateGroup{
				Admin:    clientCtx.GetFromAddress().String(),
				Members:  members,
				Metadata: b,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupMembersCmd creates a CLI command for Msg/UpdateGroupMembers.
func MsgUpdateGroupMembersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-members [admin] [group-id] [members-json-file]",
		Short: "Update a group's members. Set a member's weight to \"0\" to delete it.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update a group's members

Example:
$ %s tx group update-group-members [admin] [group-id] [members-json-file]

Where members.json contains:

{
	"members": [
		{
			"address": "addr1",
			"weight": "1",
			"metadata": "some new metadata"
		},
		{
			"address": "addr2",
			"weight": "0",
			"metadata": "some metadata"
		}
	]
}

Set a member's weight to "0" to delete it.
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			members, err := parseMembers(clientCtx, args[2])
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := &group.MsgUpdateGroupMembers{
				Admin:         clientCtx.GetFromAddress().String(),
				MemberUpdates: members,
				GroupId:       groupID,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupAdminCmd creates a CLI command for Msg/UpdateGroupAdmin.
func MsgUpdateGroupAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-admin [admin] [group-id] [new-admin]",
		Short: "Update a group's admin",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := &group.MsgUpdateGroupAdmin{
				Admin:    clientCtx.GetFromAddress().String(),
				NewAdmin: args[2],
				GroupId:  groupID,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupMetadataCmd creates a CLI command for Msg/UpdateGroupMetadata.
func MsgUpdateGroupMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-metadata [admin] [group-id] [metadata]",
		Short: "Update a group's metadata",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			msg := &group.MsgUpdateGroupMetadata{
				Admin:    clientCtx.GetFromAddress().String(),
				Metadata: b,
				GroupId:  groupID,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgCreateGroupAccountCmd creates a CLI command for Msg/CreateGroupAccount.
func MsgCreateGroupAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create-group-account [admin] [group-id] [metadata] [decision-policy]",
		Short: "Create a group account which is an account " +
			"associated with a group and a decision policy. " +
			"Note, the '--from' flag is " +
			"ignored as it is implied from [admin].",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a group account which is an account associated with a group and a decision policy.
Note, the '--from' flag is ignored as it is implied from [admin].

Example:
$ %s tx group create-group-account [admin] [group-id] [metadata] \
'{"@type":"/regen.group.v1alpha1.ThresholdDecisionPolicy", "threshold":"1", "timeout":"1s"}'
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			var policy group.DecisionPolicy
			if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(args[3]), &policy); err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			msg, err := group.NewMsgCreateGroupAccount(
				clientCtx.GetFromAddress(),
				groupID,
				b,
				policy,
			)
			if err != nil {
				return err
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupAccountAdminCmd creates a CLI command for Msg/UpdateGroupAccountAdmin.
func MsgUpdateGroupAccountAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-account-admin [admin] [group-account] [new-admin]",
		Short: "Update a group account admin",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &group.MsgUpdateGroupAccountAdmin{
				Admin:    clientCtx.GetFromAddress().String(),
				Address:  args[1],
				NewAdmin: args[2],
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupAccountDecisionPolicyCmd creates a CLI command for Msg/UpdateGroupAccountDecisionPolicy.
func MsgUpdateGroupAccountDecisionPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-account-policy [admin] [group-account] [decision-policy]",
		Short: "Update a group account decision policy",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var policy group.DecisionPolicy
			if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(args[2]), &policy); err != nil {
				return err
			}

			accountAddress, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg, err := group.NewMsgUpdateGroupAccountDecisionPolicyRequest(
				clientCtx.GetFromAddress(),
				accountAddress,
				policy,
			)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupAccountMetadataCmd creates a CLI command for Msg/MsgUpdateGroupAccountMetadata.
func MsgUpdateGroupAccountMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-account-metadata [admin] [group-account] [new-metadata]",
		Short: "Update a group account metadata",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			msg := &group.MsgUpdateGroupAccountMetadata{
				Admin:    clientCtx.GetFromAddress().String(),
				Address:  args[1],
				Metadata: b,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgCreateProposalCmd creates a CLI command for Msg/CreateProposal.
func MsgCreateProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-proposal [group-account] [proposer[,proposer]*] [msg_tx_json_file] [metadata]",
		Short: "Submit a new proposal",
		Long: `Submit a new proposal.

Parameters:
			group-account: address of the group account
			proposer: comma separated (no spaces) list of proposer account addresses. Example: "addr1,addr2" 
			Metadata: metadata for the proposal
			msg_tx_json_file: path to json file with messages that will be executed if the proposal is accepted.
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			proposers := strings.Split(args[1], ",")
			for i := range proposers {
				proposers[i] = strings.TrimSpace(proposers[i])
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			theTx, err := authclient.ReadTxFromFile(clientCtx, args[2])
			if err != nil {
				return err
			}
			msgs := theTx.GetMsgs()

			b, err := base64.StdEncoding.DecodeString(args[3])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			execStr, _ := cmd.Flags().GetString(FlagExec)

			msg, err := group.NewMsgCreateProposalRequest(
				args[0],
				proposers,
				msgs,
				b,
				execFromString(execStr),
			)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagExec, "", "Set to 1 to try to execute proposal immediately after creation (proposers signatures are considered as Yes votes)")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgVoteCmd creates a CLI command for Msg/Vote.
func MsgVoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [proposal-id] [voter] [choice] [metadata]",
		Short: "Vote on a proposal",
		Long: `Vote on a proposal.

Parameters:
			proposal-id: unique ID of the proposal
			voter: voter account addresses.
			choice: choice of the voter(s)
				CHOICE_UNSPECIFIED: no-op
				CHOICE_NO: no
				CHOICE_YES: yes
				CHOICE_ABSTAIN: abstain
				CHOICE_VETO: veto
			Metadata: metadata for the vote
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			choice, err := group.ChoiceFromString(args[2])
			if err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[3])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			execStr, _ := cmd.Flags().GetString(FlagExec)

			msg := &group.MsgVote{
				ProposalId: proposalID,
				Voter:      args[1],
				Choice:     choice,
				Metadata:   b,
				Exec:       execFromString(execStr),
			}
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagExec, "", "Set to 1 to try to execute proposal immediately after voting")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgExecCmd creates a CLI command for Msg/MsgExec.
func MsgExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [proposal-id]",
		Short: "Execute a proposal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := &group.MsgExec{
				ProposalId: proposalID,
				Signer:     clientCtx.GetFromAddress().String(),
			}
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func MsgVoteAggCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-agg [sender] [propsal_id] [timeout] [group-members-json-file] [[vote-json-file]...]",
		Short: "Aggregate signatures of basic votes into aggregated signature and submit the combined votes",
		Long: `Aggregate signatures of basic votes into aggregated signature and submit the combined votes.

Parameters:
			sender: sender's account address
			proposal-id: unique ID of the proposal
			timeout: UTC time for the submission deadline of the aggregated vote, e.g., 2021-08-15T12:00:00Z
			group-members-json-file: path to json file that contains group members
			vote-json-file: path to json file that contains a basic vote with a verified signature
`,
		Args: cobra.MinimumNArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			timeString := fmt.Sprintf("\"%s\"", args[2])
			var timeout types.Timestamp
			err = clientCtx.Codec.UnmarshalJSON([]byte(timeString), &timeout)
			if err != nil {
				return err
			}
			timeNow := gogotypes.TimestampNow()
			if timeout.Compare(timeNow) <= 0 {
				return fmt.Errorf("deadline for submitting the vote has passed")
			}

			groupMembers, err := parseGroupMembers(clientCtx, args[3])
			if err != nil {
				return err
			}
			for _, mem := range groupMembers {
				if err = mem.ValidateBasic(); err != nil {
					return err
				}
			}

			index := make(map[string]int, len(groupMembers))
			for i, mem := range groupMembers {
				addr := mem.Member.Address
				if _, exists := index[addr]; exists {
					return fmt.Errorf("duplicate address: %s", addr)
				}
				index[addr] = i
			}

			votes := make([]group.Choice, len(groupMembers))
			for i := range votes {
				votes[i] = group.Choice_CHOICE_UNSPECIFIED
			}

			var sigs [][]byte
			for i := 4; i < len(args); i++ {
				vote, err := parseVoteBasic(clientCtx, args[i])
				if err != nil {
					return err
				}

				if vote.ProposalId != proposalID || !vote.Timeout.Equal(timeout) {
					return fmt.Errorf("invalid vote from %s: expected proposal id %d and timeout %s", vote.Voter, proposalID, timeout.String())
				}

				memIndex, ok := index[vote.Voter]
				if !ok {
					return fmt.Errorf("invalid voter")
				}

				votes[memIndex] = vote.Choice
				sigs = append(sigs, vote.Sig)
			}

			sigma, err := bls12381.AggregateSignature(sigs)
			if err != nil {
				return err
			}

			msg := &group.MsgVoteAggRequest{
				Sender: args[0],
				ProposalId: proposalID,
				Votes: votes,
				Timeout: timeout,
				AggSig: sigma,
				Metadata: nil,
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}


// GetVoteBasicCmd creates a CLI command for Msg/VoteBasic.
func GetVoteBasicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-basic [voter] [proposal-id] [choice] [timeout]",
		Short: "Vote on a proposal",
		Long: `Vote on a proposal and the vote will be aggregated with other votes.

Parameters:
			voter: voter account addresses.
			proposal-id: unique ID of the proposal
			choice: choice of the voter(s)
				CHOICE_UNSPECIFIED: no-op
				CHOICE_NO: no
				CHOICE_YES: yes
				CHOICE_ABSTAIN: abstain
				CHOICE_VETO: veto
			timeout: UTC time for the submission deadline of the vote, e.g., 2021-08-15T12:00:00Z
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			choice, err := group.ChoiceFromString(args[2])
			if err != nil {
				return err
			}

			timeString := fmt.Sprintf("\"%s\"", args[3])
			var timeout types.Timestamp
			err = clientCtx.Codec.UnmarshalJSON([]byte(timeString), &timeout)
			if err != nil {
				return err
			}

			timeNow := gogotypes.TimestampNow()
			if timeout.Compare(timeNow) <= 0 {
				return fmt.Errorf("deadline for submitting the vote has passed")
			}

			msg := &group.MsgVoteBasicRequest{
				ProposalId: proposalID,
				Choice:     choice,
				Timeout:	timeout,
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			bytesToSign := msg.GetSignBytes()
			sigBytes, pubKey, err := clientCtx.Keyring.Sign(clientCtx.GetFromName(), bytesToSign)

			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			if err != nil {
				return err
			}

			vote := &group.MsgVoteBasicResponse{
				ProposalId: proposalID,
				Choice:     choice,
				Timeout:    timeout,
				Voter:      args[0],
				PubKey:     pubKeyAny,
				Sig:        sigBytes,
			}

			return clientCtx.PrintProto(vote)
		},
	}

	cmd.Flags().String(flags.FlagFrom, "", "Name or address of private key with which to sign")
	cmd.Flags().StringP(tmcli.OutputFlag, "o", "text", "Output format (text|json)")

	return cmd
}



// GetVerifyVoteBasicCmd creates a CLI command for aggregating basic votes.
func GetVerifyVoteBasicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-vote-basic [file]",
		Short: "Verify signature for a basic vote",
		Long: `Verify signature for a basic vote.

Parameters:
			file: a basic vote with signature
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			vote, err := parseVoteBasic(clientCtx, args[0])
			if err != nil {
				return err
			}

			if err = vote.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			msgBytes := vote.GetSignBytes()

			pubKey, ok := vote.PubKey.GetCachedValue().(cryptotypes.PubKey)
			if !ok {
				return fmt.Errorf("failed to get public key")
			}

			voterAddress, err := sdk.AccAddressFromBech32(vote.Voter)
			if err != nil {
				return err
			}
			if !bytes.Equal(pubKey.Address(), voterAddress) {
				return fmt.Errorf("public key does not match the voter's address %s", vote.Voter)
			}

			if !pubKey.VerifySignature(msgBytes, vote.Sig) {
				return fmt.Errorf("verification failed")
			}

			cmd.Println("Verification Successful!")

			return nil
		},
	}

	return cmd
}



