package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	addr = sdk.AccAddress([]byte("addr________________"))
	addrInvalid = sdk.AccAddress([]byte(""))
)

type AirdropTestSuite struct {
	suite.Suite
}

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

func TestMsgAirdropValidateBasic (t *testing.T) {
	amount := sdk.NewInt64Coin("test", 100)
	require.NoError(t, NewMsgAirDrop(addr, amount, sdk.NewInt(20)).ValidateBasic())
	require.Error(t, NewMsgAirDrop(addrInvalid, amount, sdk.NewInt(20)).ValidateBasic())
}

func TestAirdropTestSuite(t *testing.T) {
	suite.Run(t, new(AirdropTestSuite))
}