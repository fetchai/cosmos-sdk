package server

import (
	"github.com/cosmos/cosmos-sdk/regen/orm"
	servermodule "github.com/cosmos/cosmos-sdk/regen/types/module/server"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/exported"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Group Table
	GroupTablePrefix        byte = 0x0
	GroupTableSeqPrefix     byte = 0x1
	GroupByAdminIndexPrefix byte = 0x2

	// Group Member Table
	GroupMemberTablePrefix         byte = 0x10
	GroupMemberByGroupIndexPrefix  byte = 0x11
	GroupMemberByMemberIndexPrefix byte = 0x12

	// Group Account Table
	GroupAccountTablePrefix        byte = 0x20
	GroupAccountTableSeqPrefix     byte = 0x21
	GroupAccountByGroupIndexPrefix byte = 0x22
	GroupAccountByAdminIndexPrefix byte = 0x23

	// Proposal Table
	ProposalTablePrefix               byte = 0x30
	ProposalTableSeqPrefix            byte = 0x31
	ProposalByGroupAccountIndexPrefix byte = 0x32
	ProposalByProposerIndexPrefix     byte = 0x33

	// Vote Table
	VoteTablePrefix           byte = 0x40
	VoteByProposalIndexPrefix byte = 0x41
	VoteByVoterIndexPrefix    byte = 0x42

	// Poll Table
	PollTablePrefix          byte = 0x90
	PollTableSeqPrefix       byte = 0x91
	PollByGroupIndexPrefix   byte = 0x92
	PollByCreatorIndexPrefix byte = 0x93

	// VotePoll Table
	VotePollTablePrefix        byte = 0xa0
	VotePollByPollIndexPrefix  byte = 0xa1
	VotePollByVoterIndexPrefix byte = 0xa2
)

type serverImpl struct {
	key servermodule.RootModuleKey

	accKeeper  exported.AccountKeeper
	bankKeeper exported.BankKeeper

	// Group Table
	groupTable        orm.AutoUInt64Table
	groupByAdminIndex orm.Index

	// Group Member Table
	groupMemberTable         orm.PrimaryKeyTable
	groupMemberByGroupIndex  orm.UInt64Index
	groupMemberByMemberIndex orm.Index

	// Group Account Table
	groupAccountSeq          orm.Sequence
	groupAccountTable        orm.PrimaryKeyTable
	groupAccountByGroupIndex orm.UInt64Index
	groupAccountByAdminIndex orm.Index

	// Proposal Table
	proposalTable               orm.AutoUInt64Table
	proposalByGroupAccountIndex orm.Index
	proposalByProposerIndex     orm.Index

	// Vote Table
	voteTable           orm.PrimaryKeyTable
	voteByProposalIndex orm.UInt64Index
	voteByVoterIndex    orm.Index

	// Poll Table
	pollTable          orm.AutoUInt64Table
	pollByGroupIndex   orm.UInt64Index
	pollByCreatorIndex orm.Index

	// VotePoll Table
	votePollTable        orm.PrimaryKeyTable
	votePollByPollIndex  orm.UInt64Index
	votePollByVoterIndex orm.Index
}

func newServer(storeKey servermodule.RootModuleKey, accKeeper exported.AccountKeeper, bankKeeper exported.BankKeeper, cdc codec.Codec) serverImpl {
	s := serverImpl{key: storeKey, accKeeper: accKeeper, bankKeeper: bankKeeper}

	// Group Table
	groupTableBuilder := orm.NewAutoUInt64TableBuilder(GroupTablePrefix, GroupTableSeqPrefix, storeKey, &group.GroupInfo{}, cdc)
	s.groupByAdminIndex = orm.NewIndex(groupTableBuilder, GroupByAdminIndexPrefix, func(val interface{}) ([]orm.RowID, error) {
		addr, err := sdk.AccAddressFromBech32(val.(*group.GroupInfo).Admin)
		if err != nil {
			return nil, err
		}
		return []orm.RowID{addr.Bytes()}, nil
	})
	s.groupTable = groupTableBuilder.Build()

	// Group Member Table
	groupMemberTableBuilder := orm.NewPrimaryKeyTableBuilder(GroupMemberTablePrefix, storeKey, &group.GroupMember{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc)
	s.groupMemberByGroupIndex = orm.NewUInt64Index(groupMemberTableBuilder, GroupMemberByGroupIndexPrefix, func(val interface{}) ([]uint64, error) {
		group := val.(*group.GroupMember).GroupId
		return []uint64{group}, nil
	})
	s.groupMemberByMemberIndex = orm.NewIndex(groupMemberTableBuilder, GroupMemberByMemberIndexPrefix, func(val interface{}) ([]orm.RowID, error) {
		memberAddr := val.(*group.GroupMember).Member.Address
		addr, err := sdk.AccAddressFromBech32(memberAddr)
		if err != nil {
			return nil, err
		}
		return []orm.RowID{addr.Bytes()}, nil
	})
	s.groupMemberTable = groupMemberTableBuilder.Build()

	// Group Account Table
	s.groupAccountSeq = orm.NewSequence(storeKey, GroupAccountTableSeqPrefix)
	groupAccountTableBuilder := orm.NewPrimaryKeyTableBuilder(GroupAccountTablePrefix, storeKey, &group.GroupAccountInfo{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc)
	s.groupAccountByGroupIndex = orm.NewUInt64Index(groupAccountTableBuilder, GroupAccountByGroupIndexPrefix, func(value interface{}) ([]uint64, error) {
		group := value.(*group.GroupAccountInfo).GroupId
		return []uint64{group}, nil
	})
	s.groupAccountByAdminIndex = orm.NewIndex(groupAccountTableBuilder, GroupAccountByAdminIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		admin := value.(*group.GroupAccountInfo).Admin
		addr, err := sdk.AccAddressFromBech32(admin)
		if err != nil {
			return nil, err
		}
		return []orm.RowID{addr.Bytes()}, nil
	})
	s.groupAccountTable = groupAccountTableBuilder.Build()

	// Proposal Table
	proposalTableBuilder := orm.NewAutoUInt64TableBuilder(ProposalTablePrefix, ProposalTableSeqPrefix, storeKey, &group.Proposal{}, cdc)
	s.proposalByGroupAccountIndex = orm.NewIndex(proposalTableBuilder, ProposalByGroupAccountIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		account := value.(*group.Proposal).Address
		addr, err := sdk.AccAddressFromBech32(account)
		if err != nil {
			return nil, err
		}
		return []orm.RowID{addr.Bytes()}, nil
	})
	s.proposalByProposerIndex = orm.NewIndex(proposalTableBuilder, ProposalByProposerIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		proposers := value.(*group.Proposal).Proposers
		r := make([]orm.RowID, len(proposers))
		for i := range proposers {
			addr, err := sdk.AccAddressFromBech32(proposers[i])
			if err != nil {
				return nil, err
			}
			r[i] = addr.Bytes()
		}
		return r, nil
	})
	s.proposalTable = proposalTableBuilder.Build()

	// Vote Table
	voteTableBuilder := orm.NewPrimaryKeyTableBuilder(VoteTablePrefix, storeKey, &group.Vote{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc)
	s.voteByProposalIndex = orm.NewUInt64Index(voteTableBuilder, VoteByProposalIndexPrefix, func(value interface{}) ([]uint64, error) {
		return []uint64{value.(*group.Vote).ProposalId}, nil
	})
	s.voteByVoterIndex = orm.NewIndex(voteTableBuilder, VoteByVoterIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		addr, err := sdk.AccAddressFromBech32(value.(*group.Vote).Voter)
		if err != nil {
			return nil, err
		}
		return []orm.RowID{addr.Bytes()}, nil
	})
	s.voteTable = voteTableBuilder.Build()

	// Poll Table
	pollTableBuilder := orm.NewAutoUInt64TableBuilder(PollTablePrefix, PollTableSeqPrefix, storeKey, &group.Poll{}, cdc)
	s.pollByGroupIndex = orm.NewUInt64Index(pollTableBuilder, PollByGroupIndexPrefix, func(value interface{}) ([]uint64, error) {
		group := value.(*group.Poll).GroupId
		return []uint64{group}, nil
	})
	s.pollByCreatorIndex = orm.NewIndex(pollTableBuilder, PollByCreatorIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		creator := value.(*group.Poll).Creator
		addr, err := sdk.AccAddressFromBech32(creator)
		if err != nil {
			return nil, err
		}
		return []orm.RowID{addr.Bytes()}, nil
	})
	s.pollTable = pollTableBuilder.Build()

	// VotePoll Table
	votePollTableBuilder := orm.NewPrimaryKeyTableBuilder(VotePollTablePrefix, storeKey, &group.VotePoll{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc)
	s.votePollByPollIndex = orm.NewUInt64Index(votePollTableBuilder, VotePollByPollIndexPrefix, func(value interface{}) ([]uint64, error) {
		return []uint64{value.(*group.VotePoll).PollId}, nil
	})
	s.votePollByVoterIndex = orm.NewIndex(votePollTableBuilder, VotePollByVoterIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		addr, err := sdk.AccAddressFromBech32(value.(*group.VotePoll).Voter)
		if err != nil {
			return nil, err
		}
		return []orm.RowID{addr.Bytes()}, nil
	})
	s.votePollTable = votePollTableBuilder.Build()

	return s
}

func RegisterServices(configurator servermodule.Configurator, accountKeeper exported.AccountKeeper, bankKeeper exported.BankKeeper) {
	impl := newServer(configurator.ModuleKey(), accountKeeper, bankKeeper, configurator.Marshaler())
	group.RegisterMsgServer(configurator.MsgServer(), impl)
	group.RegisterQueryServer(configurator.QueryServer(), impl)
	configurator.RegisterInvariantsHandler(impl.RegisterInvariants)
	configurator.RegisterGenesisHandlers(impl.InitGenesis, impl.ExportGenesis)
	configurator.RegisterWeightedOperationsHandler(impl.WeightedOperations)
}