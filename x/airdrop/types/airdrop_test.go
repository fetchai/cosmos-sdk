package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

func TestAirdropTestSuite(t *testing.T) {
	suite.Run(t, new(AirdropTestSuite))
}