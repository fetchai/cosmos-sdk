package v039_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_38"
	v039auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_39"
	v039 "github.com/cosmos/cosmos-sdk/x/genutil/legacy/v0_39"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

var genAuthState = []byte(`{
  "params": {
    "max_memo_characters": "10",
    "tx_sig_limit": "10",
    "tx_size_cost_per_byte": "10",
    "sig_verify_cost_ed25519": "10",
    "sig_verify_cost_secp256k1": "10"
  },
  "accounts": [
    {
      "type": "cosmos-sdk/Account",
      "value": {
        "address": "fetch1pzfhj8mnnhslexuulhjr69vq8ary3x50366juv",
        "coins": [
          {
            "denom": "stake",
            "amount": "400000"
          }
        ],
        "public_key": "fetchpub1addwnpepqdlmzynlevpq9swdvzl567mqmgecsur8m5hvgktz6ry6th58yq345vvrxlc",
        "account_number": 1,
        "sequence": 1
      }
    },
    {
      "type": "cosmos-sdk/ModuleAccount",
      "value": {
        "address": "fetch1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3xxqtmq",
        "coins": [
          {
            "denom": "stake",
            "amount": "400000000"
          }
        ],
        "public_key": "",
        "account_number": 2,
        "sequence": 4,
        "name": "bonded_tokens_pool",
        "permissions": [
          "burner",
          "staking"
        ]
      }
    },
    {
      "type": "cosmos-sdk/ContinuousVestingAccount",
      "value": {
        "address": "fetch1k3pgrf7p9a7pqxunpu7n5uz8ta7adgxky25xhv",
        "coins": [
          {
            "denom": "stake",
            "amount": "10000205"
          }
        ],
        "public_key": "fetchpub1addwnpepqgvg2su40xqx2arhkwul008vs8jafu5gj68f0fkgvvqmt4egnaejxrjdmf3",
        "account_number": 3,
        "sequence": 5,
        "original_vesting": [
          {
            "denom": "stake",
            "amount": "10000205"
          }
        ],
        "delegated_free": [],
        "delegated_vesting": [],
        "end_time": 1596125048,
        "start_time": 1595952248
      }
    },
    {
      "type": "cosmos-sdk/DelayedVestingAccount",
      "value": {
        "address": "fetch1g7afpkq7kjp64gpzrqjl2g3957acs5xzhp5mf7",
        "coins": [
          {
            "denom": "stake",
            "amount": "10000205"
          }
        ],
        "public_key": "fetchpub1addwnpepq2mu4qj3gvav48x2jd3l24c5se06w7u8wsckd3w0x4hsnf8aqmvrgn8k0tv",
        "account_number": 4,
        "sequence": 15,
        "original_vesting": [
          {
            "denom": "stake",
            "amount": "10000205"
          }
        ],
        "delegated_free": [],
        "delegated_vesting": [],
        "end_time": 1596125048
      }
    }
  ]
}`)

var expectedGenAuthState = []byte(`{"params":{"max_memo_characters":"10","tx_sig_limit":"10","tx_size_cost_per_byte":"10","sig_verify_cost_ed25519":"10","sig_verify_cost_secp256k1":"10"},"accounts":[{"type":"cosmos-sdk/Account","value":{"address":"fetch1pzfhj8mnnhslexuulhjr69vq8ary3x50366juv","coins":[{"denom":"stake","amount":"400000"}],"public_key":{"type":"tendermint/PubKeySecp256k1","value":"A3+xEn/LAgLBzWC/TXtg2jOIcGfdLsRZYtDJpd6HICNa"},"account_number":"1","sequence":"1"}},{"type":"cosmos-sdk/ModuleAccount","value":{"address":"fetch1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3xxqtmq","coins":[{"denom":"stake","amount":"400000000"}],"public_key":"","account_number":"2","sequence":"4","name":"bonded_tokens_pool","permissions":["burner","staking"]}},{"type":"cosmos-sdk/ContinuousVestingAccount","value":{"address":"fetch1k3pgrf7p9a7pqxunpu7n5uz8ta7adgxky25xhv","coins":[{"denom":"stake","amount":"10000205"}],"public_key":{"type":"tendermint/PubKeySecp256k1","value":"AhiFQ5V5gGV0d7O597zsgeXU8oiWjpemyGMBtdcon3Mj"},"account_number":"3","sequence":"5","original_vesting":[{"denom":"stake","amount":"10000205"}],"delegated_free":[],"delegated_vesting":[],"end_time":"1596125048","start_time":"1595952248"}},{"type":"cosmos-sdk/DelayedVestingAccount","value":{"address":"fetch1g7afpkq7kjp64gpzrqjl2g3957acs5xzhp5mf7","coins":[{"denom":"stake","amount":"10000205"}],"public_key":{"type":"tendermint/PubKeySecp256k1","value":"ArfKglFDOsqcypNj9VcUhl+ne4d0MWbFzzVvCaT9Btg0"},"account_number":"4","sequence":"15","original_vesting":[{"denom":"stake","amount":"10000205"}],"delegated_free":[],"delegated_vesting":[],"end_time":"1596125048"}}]}`)

func TestMigrate(t *testing.T) {
	genesis := types.AppMap{
		v038auth.ModuleName: genAuthState,
	}

	var migrated types.AppMap
	require.NotPanics(t, func() { migrated = v039.Migrate(genesis) })
	require.Equal(t, string(expectedGenAuthState), string(migrated[v039auth.ModuleName]))
}
