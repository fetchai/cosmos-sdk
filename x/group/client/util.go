package client

import (
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/group"
)

func parseMembers(clientCtx client.Context, membersFile string) ([]group.Member, error) {
	members := group.Members{}

	if membersFile == "" {
		return members.Members, nil
	}

	contents, err := ioutil.ReadFile(membersFile)
	if err != nil {
		return nil, err
	}

	err = clientCtx.Codec.UnmarshalJSON(contents, &members)
	if err != nil {
		return nil, err
	}

	return members.Members, nil
}

func execFromString(execStr string) group.Exec {
	exec := group.Exec_EXEC_UNSPECIFIED
	switch execStr {
	case ExecTry:
		exec = group.Exec_EXEC_TRY
	}
	return exec
}

func parseGroupMembers(clientCtx client.Context, membersFile string) ([]*group.GroupMember, error) {
	res := group.QueryGroupMembersResponse{}

	if membersFile == "" {
		return res.Members, nil
	}

	contents, err := ioutil.ReadFile(membersFile)
	if err != nil {
		return  nil, err
	}
	err = clientCtx.Codec.UnmarshalJSON(contents, &res)
	if err != nil {
		return  nil, err
	}

	return res.Members, nil
}

func parseVoteBasic(clientCtx client.Context, voteFile string) (group.MsgVoteBasicResponse, error) {
	vote := group.MsgVoteBasicResponse{}

	if voteFile == "" {
		return vote, nil
	}

	contents, err := ioutil.ReadFile(voteFile)
	if err != nil {
		return vote, err
	}

	err = clientCtx.Codec.UnmarshalJSON(contents, &vote)
	if err != nil {
		return vote, err
	}

	return vote, nil

}
