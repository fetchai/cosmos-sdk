package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	addr        = sdk.AccAddress([]byte("addr________________"))
	invalidAddr = sdk.AccAddress([]byte("\n\n\n\n\taddr________________\t\n\n\n\n\n"))
)

func TestNewGenesisState(t *testing.T) {
	p := types.NewParams()
	expectedFunds := []types.ActiveFund{
		{
			Sender: addr.String(),
			Fund: &types.Fund{
				Amount:     &sdk.Coin{Denom: "test", Amount: sdk.NewInt(10)},
				DripAmount: sdk.NewInt(1),
			},
		},
	}
	expectedState := &types.GenesisState{
		Params: p,
		Funds:  expectedFunds,
	}
	require.Equal(t, expectedState.GetFunds(), expectedFunds)
	require.Equal(t, expectedState.Params, p)
}

func TestValidateGenesisState(t *testing.T) {
	p1 := types.Params{
		AllowList: []string{
			addr.String(), // valid address
		},
	}
	p2 := types.Params{
		AllowList: []string{
			invalidAddr.String(), // invalid address
		},
	}
	funds := []types.ActiveFund{
		{
			Sender: addr.String(),
			Fund: &types.Fund{
				Amount: &sdk.Coin{
					Denom:  "test",
					Amount: sdk.NewInt(10),
				},
				DripAmount: sdk.NewInt(1),
			},
		},
	}
	gen1 := types.NewGenesisState(p1, funds)
	gen2 := types.NewGenesisState(p2, funds)
	require.NoError(t, gen1.Validate())
	require.Error(t, gen2.Validate())
}