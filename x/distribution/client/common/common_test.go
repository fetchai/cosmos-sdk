package common

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestQueryDelegationRewardsAddrValidation(t *testing.T) {
	clientCtx := client.Context{}.WithJSONMarshaler(types.ModuleCdc).WithTrustNode(true)

	type args struct {
		delAddr string
		valAddr string
	}

	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"invalid delegator address", args{"invalid", ""}, nil, true},
		{"empty delegator address", args{"", ""}, nil, true},
		{"invalid validator address", args{"fetch1zxcsu7l5qxs53lvp0fqgd09a9r2g6kqr9h3fr2", "invalid"}, nil, true},
		{"empty validator address", args{"fetch1zxcsu7l5qxs53lvp0fqgd09a9r2g6kqr9h3fr2", ""}, nil, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := QueryDelegationRewards(clientCtx, "", tt.args.delAddr, tt.args.valAddr)
			require.True(t, err != nil, tt.wantErr)
		})
	}
}
