//go:build cgo && ledger && !test_ledger_mock
// +build cgo,ledger,!test_ledger_mock

package ledger

import ledger "github.com/cosmos/ledger-cosmos-go"
import "errors"

type MyLedgerDevice struct {
	ledger *ledger.LedgerCosmos
}

func (d *MyLedgerDevice) Close() error {
	return d.ledger.Close()
}

func (d *MyLedgerDevice) GetPublicKeySECP256K1(bip32Path []uint32) ([]byte, error) {
	return d.ledger.GetPublicKeySECP256K1(bip32Path)
}

func (d *MyLedgerDevice) GetAddressPubKeySECP256K1(bip32Path []uint32, hrp string) ([]byte, string, error) {
	return d.ledger.GetAddressPubKeySECP256K1(bip32Path, hrp)
}

//func (d *MyLedgerDevice) SignSECP256K1(bip32Path []uint32, transaction []byte, p2 byte) ([]byte, error) {
//	return d.ledger.SignSECP256K1(bip32Path, transaction, p2)
//}

func (d *MyLedgerDevice) SignSECP256K1(bip32Path []uint32, transaction []byte) ([]byte, error) {
	ver, err := d.ledger.GetVersion()
	if err != nil {
		return nil, err
	}

	if ver.Major != 1 {
		return nil, errors.New("App version is not supported")
	}
	return d.ledger.SignSECP256K1(bip32Path, transaction, 0)
}

// If ledger support (build tag) has been enabled, which implies a CGO dependency,
// set the discoverLedger function which is responsible for loading the Ledger
// device at runtime or returning an error.
func init() {
	discoverLedger = func() (SECP256K1, error) {
		device, err := ledger.FindLedgerCosmosUserApp()
		if err != nil {
			return nil, err
		}

		return &MyLedgerDevice{device}, nil
	}
}
