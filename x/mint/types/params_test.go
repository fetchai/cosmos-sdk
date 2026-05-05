package types

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
)

func TestDefaultParams_ValidateOK(t *testing.T) {
	t.Parallel()
	require.NoError(t, DefaultParams().Validate())
}

func TestNewParams_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		params      Params
		wantErr     bool
		errContains string
	}{
		{
			name: "valid typical config",
			params: NewParams(
				"uatom",
				sdkmath.LegacyMustNewDecFromStr("0.13"),
				uint64(60*60*8766/5),
			),
			wantErr:     false,
			errContains: "",
		},
		{
			name: "invalid denom",
			params: NewParams(
				"",
				sdkmath.LegacyMustNewDecFromStr("0.10"),
				1,
			),
			wantErr:     true,
			errContains: "",
		},
		{
			name: "blocks per year zero",
			params: NewParams(
				"uatom",
				sdkmath.LegacyMustNewDecFromStr("0.10"),
				0,
			),
			wantErr:     true,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMintDenom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		denom       string
		wantErr     bool
		errContains string
	}{
		{name: "blank", denom: "", wantErr: true, errContains: "cannot be blank"},
		{name: "spaces", denom: "   ", wantErr: true, errContains: "cannot be blank"},
		{name: "valid lower", denom: "uatom", wantErr: false, errContains: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateMintDenom(tt.denom)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateBlocksPerYear(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		val         uint64
		wantErr     bool
		errContains string
	}{
		{name: "zero", val: 0, wantErr: true, errContains: "must be positive"},
		{name: "one ok", val: 1, wantErr: false, errContains: ""},
		{name: "maxInt64 ok", val: uint64(math.MaxInt64), wantErr: false, errContains: ""},
		{name: "maxInt64+1 too large", val: uint64(math.MaxInt64) + 1, wantErr: true, errContains: "too large"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateBlocksPerYear(tt.val)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.ErrorContains(t, err, tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
