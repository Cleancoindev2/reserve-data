package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/data/testutil"
	commonv3 "github.com/KyberNetwork/reserve-data/v3/common"
)

func TestHasPendingDepositBoltStorage(t *testing.T) {
	boltFile := "test_bolt.db"
	tmpDir, err := ioutil.TempDir("", "pending_deposit")
	if err != nil {
		t.Fatal(err)
	}
	storage, err := NewBoltStorage(filepath.Join(tmpDir, boltFile))
	if err != nil {
		t.Fatalf("Couldn't init bolt storage %v", err)
	}
	exchange := common.TestExchange{}
	timepoint := common.GetTimepoint()
	asset := commonv3.Asset{
		ID:                 1,
		Symbol:             "OMG",
		Name:               "omise-go",
		Address:            ethereum.HexToAddress("0x1111111111111111111111111111111111111111"),
		OldAddresses:       nil,
		Decimals:           12,
		Transferable:       true,
		SetRate:            commonv3.SetRateNotSet,
		Rebalance:          false,
		IsQuote:            false,
		PWI:                nil,
		RebalanceQuadratic: nil,
		Exchanges:          nil,
		Target:             nil,
		Created:            time.Now(),
		Updated:            time.Now(),
	}
	out, err := storage.HasPendingDeposit(asset, exchange)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if out != false {
		t.Fatalf("Expected ram storage to return true false there is no pending deposit for the same currency and exchange")
	}
	err = storage.Record(
		"deposit",
		common.NewActivityID(1, "1"),
		string(exchange.ID()),
		map[string]interface{}{
			"exchange":  exchange,
			"asset":     asset,
			"amount":    "1.0",
			"timepoint": timepoint,
		},
		map[string]interface{}{
			"tx":    "",
			"error": nil,
		},
		"",
		"submitted",
		common.GetTimepoint())
	if err != nil {
		t.Fatalf("Store activity error: %s", err.Error())
	}
	b, err := storage.HasPendingDeposit(commonv3.Asset{
		ID:                 1,
		Symbol:             "OMG",
		Name:               "omise-go",
		Address:            ethereum.HexToAddress("0x1111111111111111111111111111111111111111"),
		OldAddresses:       nil,
		Decimals:           12,
		Transferable:       true,
		SetRate:            commonv3.SetRateNotSet,
		Rebalance:          false,
		IsQuote:            false,
		PWI:                nil,
		RebalanceQuadratic: nil,
		Exchanges:          nil,
		Target:             nil,
		Created:            time.Now(),
		Updated:            time.Now(),
	}, exchange)
	out, err = b, err
	if err != nil {
		t.Fatalf(err.Error())
	}
	if out != true {
		t.Fatalf("Expected ram storage to return true when there is pending deposit")
	}

	if err = os.RemoveAll(tmpDir); err != nil {
		t.Error(err)
	}
}

func TestGlobalStorageBoltImplementation(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test_bolt_storage")
	if err != nil {
		t.Fatal(err)
	}
	storage, err := NewBoltStorage(filepath.Join(tmpDir, "test_bolt.db"))
	if err != nil {
		t.Fatal(err)
	}
	testutil.NewGlobalStorageTestSuite(t, storage, storage).Run()

	if err = os.RemoveAll(tmpDir); err != nil {
		t.Error(err)
	}
}
