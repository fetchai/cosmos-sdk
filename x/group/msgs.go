package group

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/bls12381"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/regen/types/math"
	"github.com/cosmos/cosmos-sdk/regen/types/module/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	proto "github.com/gogo/protobuf/proto"
)

var _ sdk.Msg = &MsgCreateGroup{}
var _ legacytx.LegacyMsg = &MsgCreateGroup{}

// Route Implements Msg.
func (m MsgCreateGroup) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgCreateGroup) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgCreateGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroup.
func (m MsgCreateGroup) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateGroup) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	members := Members{Members: m.Members}
	if err := members.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "members")
	}
	for i := range m.Members {
		member := m.Members[i]
		if _, err := math.ParsePositiveDecimal(member.Weight); err != nil {
			return sdkerrors.Wrap(err, "member weight")
		}
	}
	return nil
}

func (m Member) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "address")
	}
	if _, err := math.ParseNonNegativeDecimal(m.Weight); err != nil {
		return sdkerrors.Wrap(err, "weight")
	}

	return nil
}

var _ sdk.Msg = &MsgUpdateGroupAdmin{}
var _ legacytx.LegacyMsg = &MsgUpdateGroupAdmin{}

// Route Implements Msg.
func (m MsgUpdateGroupAdmin) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateGroupAdmin) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupAdmin) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupAdmin.
func (m MsgUpdateGroupAdmin) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupAdmin) ValidateBasic() error {
	if m.GroupId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "group")
	}

	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	newAdmin, err := sdk.AccAddressFromBech32(m.NewAdmin)
	if err != nil {
		return sdkerrors.Wrap(err, "new admin")
	}

	if admin.Equals(newAdmin) {
		return sdkerrors.Wrap(ErrInvalid, "new and old admin are the same")
	}
	return nil
}

func (m *MsgUpdateGroupAdmin) GetGroupID() uint64 {
	return m.GroupId
}

var _ sdk.Msg = &MsgUpdateGroupMetadata{}
var _ legacytx.LegacyMsg = &MsgUpdateGroupMetadata{}

// Route Implements Msg.
func (m MsgUpdateGroupMetadata) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateGroupMetadata) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupMetadata) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupMetadata.
func (m MsgUpdateGroupMetadata) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupMetadata) ValidateBasic() error {
	if m.GroupId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "group")

	}
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}
	return nil
}

func (m *MsgUpdateGroupMetadata) GetGroupID() uint64 {
	return m.GroupId
}

var _ sdk.Msg = &MsgUpdateGroupMembers{}
var _ legacytx.LegacyMsg = &MsgUpdateGroupMembers{}

// Route Implements Msg.
func (m MsgUpdateGroupMembers) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateGroupMembers) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupMembers) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

var _ sdk.Msg = &MsgUpdateGroupMembers{}
var _ legacytx.LegacyMsg = &MsgUpdateGroupMembers{}

// GetSigners returns the expected signers for a MsgUpdateGroupMembers.
func (m MsgUpdateGroupMembers) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupMembers) ValidateBasic() error {
	if m.GroupId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "group")

	}
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	if len(m.MemberUpdates) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "member updates")
	}
	members := Members{Members: m.MemberUpdates}
	if err := members.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "members")
	}
	return nil
}

func (m *MsgUpdateGroupMembers) GetGroupID() uint64 {
	return m.GroupId
}

var _ sdk.Msg = &MsgCreateGroupAccount{}
var _ legacytx.LegacyMsg = &MsgCreateGroupAccount{}

// Route Implements Msg.
func (m MsgCreateGroupAccount) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgCreateGroupAccount) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgCreateGroupAccount) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateGroupAccount.
func (m MsgCreateGroupAccount) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateGroupAccount) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}
	if m.GroupId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "group")
	}

	policy := m.GetDecisionPolicy()
	if policy == nil {
		return sdkerrors.Wrap(ErrEmpty, "decision policy")
	}

	if err := policy.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "decision policy")
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateGroupAccountAdmin{}
var _ legacytx.LegacyMsg = &MsgUpdateGroupAccountAdmin{}

// Route Implements Msg.
func (m MsgUpdateGroupAccountAdmin) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateGroupAccountAdmin) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupAccountAdmin) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupAccountAdmin.
func (m MsgUpdateGroupAccountAdmin) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupAccountAdmin) ValidateBasic() error {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	newAdmin, err := sdk.AccAddressFromBech32(m.NewAdmin)
	if err != nil {
		return sdkerrors.Wrap(err, "new admin")
	}

	_, err = sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group account")
	}

	if admin.Equals(newAdmin) {
		return sdkerrors.Wrap(ErrInvalid, "new and old admin are the same")
	}
	return nil
}

var _ sdk.Msg = &MsgUpdateGroupAccountDecisionPolicy{}
var _ legacytx.LegacyMsg = &MsgUpdateGroupAccountDecisionPolicy{}
var _ types.UnpackInterfacesMessage = MsgUpdateGroupAccountDecisionPolicy{}

func NewMsgUpdateGroupAccountDecisionPolicyRequest(admin sdk.AccAddress, address sdk.AccAddress, decisionPolicy DecisionPolicy) (*MsgUpdateGroupAccountDecisionPolicy, error) {
	m := &MsgUpdateGroupAccountDecisionPolicy{
		Admin:   admin.String(),
		Address: address.String(),
	}
	err := m.SetDecisionPolicy(decisionPolicy)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *MsgUpdateGroupAccountDecisionPolicy) SetDecisionPolicy(decisionPolicy DecisionPolicy) error {
	msg, ok := decisionPolicy.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	m.DecisionPolicy = any
	return nil
}

// Route Implements Msg.
func (m MsgUpdateGroupAccountDecisionPolicy) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateGroupAccountDecisionPolicy) Type() string {
	return sdk.MsgTypeURL(&m)
}

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupAccountDecisionPolicy) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupAccountDecisionPolicy.
func (m MsgUpdateGroupAccountDecisionPolicy) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupAccountDecisionPolicy) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	_, err = sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group account")
	}

	policy := m.GetDecisionPolicy()
	if policy == nil {
		return sdkerrors.Wrap(ErrEmpty, "decision policy")
	}

	if err := policy.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "decision policy")
	}

	return nil
}

func (m *MsgUpdateGroupAccountDecisionPolicy) GetDecisionPolicy() DecisionPolicy {
	decisionPolicy, ok := m.DecisionPolicy.GetCachedValue().(DecisionPolicy)
	if !ok {
		return nil
	}
	return decisionPolicy
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgUpdateGroupAccountDecisionPolicy) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var decisionPolicy DecisionPolicy
	return unpacker.UnpackAny(m.DecisionPolicy, &decisionPolicy)
}

var _ sdk.Msg = &MsgUpdateGroupAccountMetadata{}
var _ legacytx.LegacyMsg = &MsgUpdateGroupAccountMetadata{}

// Route Implements Msg.
func (m MsgUpdateGroupAccountMetadata) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgUpdateGroupAccountMetadata) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgUpdateGroupAccountMetadata) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateGroupAccountMetadata.
func (m MsgUpdateGroupAccountMetadata) GetSigners() []sdk.AccAddress {
	admin, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{admin}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgUpdateGroupAccountMetadata) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "admin")
	}

	_, err = sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group account")
	}

	return nil
}

var _ sdk.Msg = &MsgCreateGroupAccount{}
var _ legacytx.LegacyMsg = &MsgCreateGroupAccount{}
var _ types.UnpackInterfacesMessage = MsgCreateGroupAccount{}

// NewMsgCreateGroupAccount creates a new MsgCreateGroupAccount.
func NewMsgCreateGroupAccount(admin sdk.AccAddress, group uint64, metadata []byte, decisionPolicy DecisionPolicy) (*MsgCreateGroupAccount, error) {
	m := &MsgCreateGroupAccount{
		Admin:    admin.String(),
		GroupId:  group,
		Metadata: metadata,
	}
	err := m.SetDecisionPolicy(decisionPolicy)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (m *MsgCreateGroupAccount) GetAdmin() string {
	return m.Admin
}

func (m *MsgCreateGroupAccount) GetGroupID() uint64 {
	return m.GroupId
}

func (m *MsgCreateGroupAccount) GetMetadata() []byte {
	return m.Metadata
}

func (m *MsgCreateGroupAccount) GetDecisionPolicy() DecisionPolicy {
	decisionPolicy, ok := m.DecisionPolicy.GetCachedValue().(DecisionPolicy)
	if !ok {
		return nil
	}
	return decisionPolicy
}

func (m *MsgCreateGroupAccount) SetDecisionPolicy(decisionPolicy DecisionPolicy) error {
	msg, ok := decisionPolicy.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	m.DecisionPolicy = any
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgCreateGroupAccount) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var decisionPolicy DecisionPolicy
	return unpacker.UnpackAny(m.DecisionPolicy, &decisionPolicy)
}

var _ sdk.Msg = &MsgCreateProposal{}
var _ legacytx.LegacyMsg = &MsgCreateProposal{}

// NewMsgCreateProposalRequest creates a new MsgCreateProposal.
func NewMsgCreateProposalRequest(address string, proposers []string, msgs []sdk.Msg, metadata []byte, exec Exec) (*MsgCreateProposal, error) {
	m := &MsgCreateProposal{
		Address:   address,
		Proposers: proposers,
		Metadata:  metadata,
		Exec:      exec,
	}
	err := m.SetMsgs(msgs)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Route Implements Msg.
func (m MsgCreateProposal) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgCreateProposal) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgCreateProposal) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgCreateProposal.
func (m MsgCreateProposal) GetSigners() []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, len(m.Proposers))
	for i, proposer := range m.Proposers {
		addr, err := sdk.AccAddressFromBech32(proposer)
		if err != nil {
			panic(err)
		}
		addrs[i] = addr
	}
	return addrs
}

// ValidateBasic does a sanity check on the provided data
func (m MsgCreateProposal) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Address)
	if err != nil {
		return sdkerrors.Wrap(err, "group account")
	}

	if len(m.Proposers) == 0 {
		return sdkerrors.Wrap(ErrEmpty, "proposers")
	}
	addrs := make([]sdk.AccAddress, len(m.Proposers))
	for i, proposer := range m.Proposers {
		addr, err := sdk.AccAddressFromBech32(proposer)
		if err != nil {
			return sdkerrors.Wrap(err, "proposers")
		}
		addrs[i] = addr
	}
	if err := AccAddresses(addrs).ValidateBasic(); err != nil {
		return sdkerrors.Wrap(err, "proposers")
	}

	msgs := m.GetMsgs()
	for i, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return sdkerrors.Wrapf(err, "msg %d", i)
		}
	}
	return nil
}

// SetMsgs packs msgs into Any's
func (m *MsgCreateProposal) SetMsgs(msgs []sdk.Msg) error {
	anys, err := server.SetMsgs(msgs)
	if err != nil {
		return err
	}
	m.Msgs = anys
	return nil
}

// GetMsgs unpacks m.Msgs Any's into sdk.Msg's
func (m MsgCreateProposal) GetMsgs() []sdk.Msg {
	msgs, err := server.GetMsgs(m.Msgs)
	if err != nil {
		panic(err)
	}
	return msgs
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgCreateProposal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	return server.UnpackInterfaces(unpacker, m.Msgs)
}

var _ sdk.Msg = &MsgVote{}
var _ legacytx.LegacyMsg = &MsgVote{}

// Route Implements Msg.
func (m MsgVote) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgVote) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVote) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgVote.
func (m MsgVote) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgVote) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		return sdkerrors.Wrap(err, "voter")
	}
	if m.ProposalId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "proposal")
	}
	if m.Choice == Choice_CHOICE_UNSPECIFIED {
		return sdkerrors.Wrap(ErrEmpty, "choice")
	}
	if _, ok := Choice_name[int32(m.Choice)]; !ok {
		return sdkerrors.Wrap(ErrInvalid, "choice")
	}
	return nil
}

var _ sdk.Msg = &MsgVoteBasic{}
var _ legacytx.LegacyMsg = &MsgVoteBasic{}

// Route Implements Msg.
func (m MsgVoteBasic) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgVoteBasic) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVoteBasic) GetSignBytes() []byte {
	req := MsgVoteBasic{
		ProposalId: m.ProposalId,
		Choice:     m.Choice,
		Timeout:    m.Timeout,
	}
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&req))
}

// GetSigners returns the expected signers for a MsgVoteRequest.
func (m MsgVoteBasic) GetSigners() []sdk.AccAddress {
	panic("this message does not include signers")
}

// ValidateBasic does a sanity check on the provided data
func (m MsgVoteBasic) ValidateBasic() error {
	if m.ProposalId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "proposal")
	}
	if m.Choice == Choice_CHOICE_UNSPECIFIED {
		return sdkerrors.Wrap(ErrEmpty, "choice")
	}
	if _, ok := Choice_name[int32(m.Choice)]; !ok {
		return sdkerrors.Wrap(ErrInvalid, "choice")
	}
	return nil
}

var _ sdk.Msg = &MsgVoteBasicResponse{}
var _ legacytx.LegacyMsg = &MsgVoteBasicResponse{}
var _ types.UnpackInterfacesMessage = MsgVoteBasicResponse{}

// Route Implements Msg.
func (m MsgVoteBasicResponse) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgVoteBasicResponse) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVoteBasicResponse) GetSignBytes() []byte {
	res := MsgVoteBasic{
		ProposalId: m.ProposalId,
		Choice:     m.Choice,
		Timeout:    m.Timeout,
	}
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&res))
}

// GetSigners returns the expected signers for a MsgVoteRequest.
func (m MsgVoteBasicResponse) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Voter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgVoteBasicResponse) ValidateBasic() error {
	if m.ProposalId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "proposal")
	}
	if m.Choice == Choice_CHOICE_UNSPECIFIED {
		return sdkerrors.Wrap(ErrEmpty, "choice")
	}
	if _, ok := Choice_name[int32(m.Choice)]; !ok {
		return sdkerrors.Wrap(ErrInvalid, "choice")
	}
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgVoteBasicResponse) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var pk cryptotypes.PubKey
	return unpacker.UnpackAny(m.PubKey, &pk)
}

var _ sdk.Msg = &MsgVoteAgg{}
var _ legacytx.LegacyMsg = &MsgVoteAgg{}

// Route Implements Msg.
func (m MsgVoteAgg) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgVoteAgg) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgVoteAgg) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgVoteRequest.
func (m MsgVoteAgg) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgVoteAgg) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrap(err, "sender")
	}
	if m.ProposalId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "proposal")
	}
	for _, c := range m.Votes {
		if _, ok := Choice_name[int32(c)]; !ok {
			return sdkerrors.Wrap(ErrInvalid, "choice")
		}
	}
	return nil
}

func (m MsgVoteAgg) VerifyAggSig(pks []cryptotypes.PubKey) error {
	n := len(m.Votes)
	if n <= 0 {
		return fmt.Errorf("empty votes")
	}
	if len(pks) != n {
		return fmt.Errorf("the length of public keys and votes must equal")
	}

	pkeys := make([][]*bls12381.PubKey, len(Choice_name))
	for i, pk := range pks {
		pkey, ok := pk.(*bls12381.PubKey)
		if !ok {
			return fmt.Errorf("only support bls12381 public key")
		}
		c := m.Votes[i]
		pkeys[c] = append(pkeys[c], pkey)
	}

	var votes [][]byte
	var pubkeys [][]*bls12381.PubKey

	// skip the unspecified choice
	for i := 1; i < len(Choice_name); i++ {
		if len(pkeys[i]) != 0 {
			vote := MsgVoteBasic{m.ProposalId, Choice(i), m.Timeout}
			voteBytes := vote.GetSignBytes()
			votes = append(votes, voteBytes)
			pubkeys = append(pubkeys, pkeys[i])
		}
	}

	return bls12381.VerifyAggregateSignature(votes, false, m.AggSig, pubkeys)
}

var _ sdk.Msg = &MsgExec{}
var _ legacytx.LegacyMsg = &MsgExec{}

// Route Implements Msg.
func (m MsgExec) Route() string { return sdk.MsgTypeURL(&m) }

// Type Implements Msg.
func (m MsgExec) Type() string { return sdk.MsgTypeURL(&m) }

// GetSignBytes Implements Msg.
func (m MsgExec) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgExec.
func (m MsgExec) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// ValidateBasic does a sanity check on the provided data
func (m MsgExec) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		return sdkerrors.Wrap(err, "signer")
	}
	if m.ProposalId == 0 {
		return sdkerrors.Wrap(ErrEmpty, "proposal")
	}
	return nil
}
