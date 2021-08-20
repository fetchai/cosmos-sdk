package types_test

import (
	"crypto/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/airdrop/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestNewGenesisState(t *testing.T) {
	p := types.NewParams()

	expectedFunds := []types.ActiveFund{
		{
			Sender: sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String(),
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
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	p1 := types.Params{
		AllowList: []string{
			addr.String(), // valid address
		},
	}

	invalidAddrBytes := make([]byte, address.MaxAddrLen+1)
	rand.Read(invalidAddrBytes)

	p2 := types.Params{
		AllowList: []string{
			sdk.AccAddress(invalidAddrBytes).String(), // invalid address
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
