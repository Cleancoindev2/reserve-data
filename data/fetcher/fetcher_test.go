package fetcher

import (
	"io/ioutil"
	"os"
	"path"
	"sync"
	"testing"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/data/fetcher/httprunner"
	"github.com/KyberNetwork/reserve-data/data/storage"
	"github.com/KyberNetwork/reserve-data/world"
)

func TestUnchangedFunc(t *testing.T) {
	// test different len
	a1 := map[common.ActivityID]common.ActivityStatus{
		common.NewActivityID(1, "1"): common.NewActivityStatus("done", "0x123", 0, "mined", nil),
	}
	b1 := map[common.ActivityID]common.ActivityStatus{
		common.NewActivityID(1, "1"): common.NewActivityStatus("done", "0x123", 0, "mined", nil),
		common.NewActivityID(2, "1"): common.NewActivityStatus("done", "0x123", 0, "mined", nil),
	}
	if unchanged(a1, b1) != false {
		t.Fatalf("Expected unchanged() to return false, got true")
	}
	// test different id
	a1 = map[common.ActivityID]common.ActivityStatus{
		common.NewActivityID(1, "1"): common.NewActivityStatus("done", "0x123", 0, "mined", nil),
	}
	b1 = map[common.ActivityID]common.ActivityStatus{
		common.NewActivityID(2, "1"): common.NewActivityStatus("done", "0x123", 0, "mined", nil),
	}
	if unchanged(a1, b1) != false {
		t.Fatalf("Expected unchanged() to return false, got true")
	}
	// test different exchange status
	a1 = map[common.ActivityID]common.ActivityStatus{
		common.NewActivityID(1, "1"): common.NewActivityStatus("", "0x123", 0, "mined", nil),
	}
	b1 = map[common.ActivityID]common.ActivityStatus{
		common.NewActivityID(1, "1"): common.NewActivityStatus("done", "0x123", 0, "mined", nil),
	}
	if unchanged(a1, b1) != false {
		t.Fatalf("Expected unchanged() to return false, got true")
	}
	// test different mining status
	a1 = map[common.ActivityID]common.ActivityStatus{
		common.NewActivityID(1, "1"): common.NewActivityStatus("done", "0x123", 0, "mined", nil),
	}
	b1 = map[common.ActivityID]common.ActivityStatus{
		common.NewActivityID(1, "1"): common.NewActivityStatus("done", "0x123", 0, "", nil),
	}
	if unchanged(a1, b1) != false {
		t.Fatalf("Expected unchanged() to return false, got true")
	}
	// test different tx
	a1 = map[common.ActivityID]common.ActivityStatus{
		common.NewActivityID(1, "1"): common.NewActivityStatus("done", "0x123", 0, "mined", nil),
	}
	b1 = map[common.ActivityID]common.ActivityStatus{
		common.NewActivityID(1, "1"): common.NewActivityStatus("done", "0x124", 0, "mined", nil),
	}
	if unchanged(a1, b1) != false {
		t.Fatalf("Expected unchanged() to return false, got true")
	}
	// test identical statuses
	a1 = map[common.ActivityID]common.ActivityStatus{
		common.NewActivityID(1, "1"): common.NewActivityStatus("done", "0x123", 0, "mined", nil),
	}
	b1 = map[common.ActivityID]common.ActivityStatus{
		common.NewActivityID(1, "1"): common.NewActivityStatus("done", "0x123", 0, "mined", nil),
	}
	if unchanged(a1, b1) != true {
		t.Fatalf("Expected unchanged() to return true, got false")
	}
}

func TestExchangeDown(t *testing.T) {
	// mock fetcher
	tmpDir, err := ioutil.TempDir("", "test_fetcher")
	if err != nil {
		t.Fatal(err)
	}
	testFetcherStoragePath := path.Join(tmpDir, "test_fetcher.db")

	fstorage, err := storage.NewBoltStorage(testFetcherStoragePath)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer func() {
		if rErr := os.RemoveAll(tmpDir); rErr != nil {
			t.Error(rErr)
		}
	}()
	runner, err := httprunner.NewHTTPRunner(httprunner.WithPort(9000))
	if err != nil {
		t.Fatal(err)
	}

	addressConf := &common.ContractAddressConfiguration{}

	fetcher := NewFetcher(fstorage, fstorage, &world.TheWorld{}, runner, true, addressConf)
	var KNC = common.AssetID(1)
	// mock normal data
	var estatuses, bstatuses sync.Map
	ebalanceValue := common.EBalanceEntry{
		Valid:      true,
		Error:      "",
		Timestamp:  common.GetTimestamp(),
		ReturnTime: common.GetTimestamp(),
		AvailableBalance: map[common.AssetID]float64{
			KNC: 500,
		},
		LockedBalance:  map[common.AssetID]float64{},
		DepositBalance: map[common.AssetID]float64{},
		Status:         true,
	}
	ebalance := sync.Map{}
	ebalance.Store(common.Binance, ebalanceValue)

	rawBalance := common.RawBalance{}
	tokenBalance := common.BalanceEntry{
		Valid:      true,
		Error:      "",
		Timestamp:  common.GetTimestamp(),
		ReturnTime: common.GetTimestamp(),
		Balance:    rawBalance,
	}

	bbalance := map[common.AssetID]common.BalanceEntry{
		KNC: tokenBalance,
	}

	// empty pending activities
	pendings := []common.ActivityRecord{}

	var snapshot common.AuthDataSnapshot
	timepoint := common.NowInMillis()

	// Persist normal auth snapshot
	err = fetcher.PersistSnapshot(&ebalance, bbalance, &estatuses, &bstatuses, pendings, &snapshot, timepoint)
	if err != nil {
		t.Fatalf("Cannot persist snapshot: %s", err.Error())
	}

	// mock empty data as exchange down
	ebalanceValue = common.EBalanceEntry{
		Valid:            false,
		Error:            "Connection time out",
		Timestamp:        common.GetTimestamp(),
		ReturnTime:       common.GetTimestamp(),
		AvailableBalance: map[common.AssetID]float64{},
		LockedBalance:    map[common.AssetID]float64{},
		DepositBalance:   map[common.AssetID]float64{},
		Status:           false, // exchange status false - down, true - up
	}
	ebalance.Store(common.Binance, ebalanceValue)
	err = fetcher.PersistSnapshot(&ebalance, bbalance, &estatuses, &bstatuses, pendings, &snapshot, timepoint)
	if err != nil {
		t.Fatalf("Cannot persist snapshot: %s", err.Error())
	}
	// check if snapshot store latest data instead of empty
	version, err := fetcher.storage.CurrentAuthDataVersion(common.NowInMillis())
	if err != nil {
		t.Fatalf("Snapshot did not saved: %s", err.Error())
	}
	authData, err := fetcher.storage.GetAuthData(version)
	if err != nil {
		t.Fatalf("Cannot get snapshot: %s", err.Error())
	}
	exchangeBalance := authData.ExchangeBalances[common.Binance]
	if exchangeBalance.AvailableBalance[KNC] != 500 {
		t.Fatalf("Snapshot did not get the latest auth data instead")
	}

	if exchangeBalance.Error != "Connection time out" {
		t.Fatalf("Snapshot did not save exchange error")
	}
}
