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
        "address": "fetch19xav7r0kja7ldzf3jypvr7fhdd92zk6h8xtse5",
        "coins": [
          {
            "denom": "stake",
            "amount": "400000"
          }
        ],
        "public_key": "fetchpub1addwnpepqgrq8p4mt8n3xmvpshlyx89xr75wln0jj997auhw00cmd6emumd7wjhu8gx",
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
        "address": "fetch13k6l84d7ceu744p660zy3zgtsz93v976zfuqml",
        "coins": [
          {
            "denom": "stake",
            "amount": "10000205"
          }
        ],
        "public_key": "fetchpub1addwnpepqgah4lqpza3ye0e8npm6f6cvut77heanlswg637uxs7ancttel4nunwcsfg",
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
        "address": "fetch1ckvh6fp75eenpkpnj88e2fv4hsdqnhc9lmd60r",
        "coins": [
          {
            "denom": "stake",
            "amount": "10000205"
          }
        ],
        "public_key": "fetchpub1addwnpepqg5vsltaz0x0awrhp99mvx32xuhcndtwa5yvcn9n34gh79dly6tss0pr0l0",
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

var expectedGenAuthState = []byte(`{"params":{"max_memo_characters":"10","tx_sig_limit":"10","tx_size_cost_per_byte":"10","sig_verify_cost_ed25519":"10","sig_verify_cost_secp256k1":"10"},"accounts":[{"type":"cosmos-sdk/Account","value":{"address":"fetch19xav7r0kja7ldzf3jypvr7fhdd92zk6h8xtse5","coins":[{"denom":"stake","amount":"400000"}],"public_key":{"type":"tendermint/PubKeySecp256k1","value":"AgYDhrtZ5xNtgYX+QxymH6jvzfKRS+7y7nvxtus75tvn"},"account_number":"1","sequence":"1"}},{"type":"cosmos-sdk/ModuleAccount","value":{"address":"fetch1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3xxqtmq","coins":[{"denom":"stake","amount":"400000000"}],"public_key":"","account_number":"2","sequence":"4","name":"bonded_tokens_pool","permissions":["burner","staking"]}},{"type":"cosmos-sdk/ContinuousVestingAccount","value":{"address":"fetch13k6l84d7ceu744p660zy3zgtsz93v976zfuqml","coins":[{"denom":"stake","amount":"10000205"}],"public_key":{"type":"tendermint/PubKeySecp256k1","value":"Ajt6/AEXYky/J5h3pOsM4v3r57P8HI1H3DQ92eFrz+s+"},"account_number":"3","sequence":"5","original_vesting":[{"denom":"stake","amount":"10000205"}],"delegated_free":[],"delegated_vesting":[],"end_time":"1596125048","start_time":"1595952248"}},{"type":"cosmos-sdk/DelayedVestingAccount","value":{"address":"fetch1ckvh6fp75eenpkpnj88e2fv4hsdqnhc9lmd60r","coins":[{"denom":"stake","amount":"10000205"}],"public_key":{"type":"tendermint/PubKeySecp256k1","value":"AijIfX0TzP64dwlLthoqNy+JtW7tCMxMs41RfxW/JpcI"},"account_number":"4","sequence":"15","original_vesting":[{"denom":"stake","amount":"10000205"}],"delegated_free":[],"delegated_vesting":[],"end_time":"1596125048"}}]}`)

func TestMigrate(t *testing.T) {
	genesis := types.AppMap{
		v038auth.ModuleName: genAuthState,
	}

	var migrated types.AppMap
	require.NotPanics(t, func() { migrated = v039.Migrate(genesis) })
	require.Equal(t, string(expectedGenAuthState), string(migrated[v039auth.ModuleName]))
}
