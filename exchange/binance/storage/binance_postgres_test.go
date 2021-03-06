package storage

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/exchange"
)

func Test_BinancePostgres(t *testing.T) {
	var storage exchange.BinanceStorage
	var err error
	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()
	storage, err = NewPostgresStorage(db)
	require.NoError(t, err)

	exchangeTradeHistory := common.ExchangeTradeHistory{
		1: []common.TradeHistory{
			{
				ID:        "12342",
				Price:     0.132131,
				Qty:       12.3123,
				Type:      "buy",
				Timestamp: 1528949872000,
			},
		},
	}

	// store trade history
	err = storage.StoreTradeHistory(exchangeTradeHistory)
	if err != nil {
		t.Fatal(err)
	}

	// get trade history
	var tradeHistory common.ExchangeTradeHistory
	fromTime := uint64(1528934400000)
	toTime := uint64(1529020800000)
	tradeHistory, err = storage.GetTradeHistory(fromTime, toTime)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tradeHistory, exchangeTradeHistory) {
		t.Fatalf("Get wrong trade history %+v", tradeHistory)
	}

	// get last trade history id
	var lastHistoryID string
	lastHistoryID, err = storage.GetLastIDTradeHistory(1)
	if err != nil {
		t.Fatalf("Get last trade history id error: %s", err.Error())
	}
	if lastHistoryID != "12342" {
		t.Fatalf("Get last trade history wrong")
	}
}
