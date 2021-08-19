package types

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/require"
)

func TestFundValidateBasic(t *testing.T) {
	amount := sdk.NewInt64Coin("test", 1)
	fundPosDrip := Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(1),
	}
	fundNilDrip := Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(0),
	}
	fundNegDrip := Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(-1),
	}
	require.NoError(t, fundPosDrip.ValidateBasic())
	require.Error(t, fundNilDrip.ValidateBasic())
	require.Error(t, fundNegDrip.ValidateBasic())
}

func TestFundDrip(t *testing.T) {
	amount := sdk.NewInt64Coin("test", 10)
	fund := Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(2),
	}
	fundHighDrip := Fund{
		Amount:     &amount,
		DripAmount: sdk.NewInt(100),
	}
	fund, _ = fund.Drip()
	fundHighDrip, _ = fundHighDrip.Drip()
	fmt.Println(fundHighDrip.Amount.Amount)
	require.Equal(t, fund.Amount.Amount, sdk.NewInt(8))
	require.Equal(t, fundHighDrip.Amount.Amount.Int64(), int64(0))
}

func TestMsgAirdropValidateBasic(t *testing.T) {
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	invalidAddrBytes := make([]byte, address.MaxAddrLen+1)
	rand.Read(invalidAddrBytes)
	addrInvalid := sdk.AccAddress(invalidAddrBytes)

	amount := sdk.NewInt64Coin("test", 100)
	require.NoError(t, NewMsgAirDrop(addr.String(), amount, sdk.NewInt(20)).ValidateBasic())
	require.Error(t, NewMsgAirDrop(addrInvalid.String(), amount, sdk.NewInt(20)).ValidateBasic())
}
