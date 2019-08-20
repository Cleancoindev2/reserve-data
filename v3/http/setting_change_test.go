package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/common/testutil"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage/postgres"
)

// TODO write more test cases
func TestServer_SettingChangeBasic(t *testing.T) {
	const (
		settingChangePath = "/v3/setting-change"
	)

	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := postgres.NewStorage(db)
	require.NoError(t, err)

	assetID, err := createSampleAsset(s)
	require.NoError(t, err)

	server := NewServer(s, "", nil)

	var tests = []testCase{
		{
			msg:      "create asset exchange",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x007"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
						},
					},
					{
						Type: common.ChangeTypeUpdateAssetExchange,
						Data: common.UpdateAssetExchangeEntry{
							ID:             2,
							DepositAddress: common.AddressPointer(eth.HexToAddress("0x0000000000001")),
							MinDeposit:     common.FloatPointer(6.0),
							WithdrawFee:    common.FloatPointer(9.0),
						},
					},
				},
			},
			assert: httputil.ExpectSuccess,
		},
		{
			msg:      "confirm setting change",
			endpoint: fmt.Sprint(settingChangePath, "/", 1),
			method:   http.MethodPut,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				httputil.ExpectSuccess(t, resp)
				//check update asset exchange
				assetExchange, err := s.GetAssetExchange(2)
				require.NoError(t, err)
				require.Equal(t, 6.0, assetExchange.MinDeposit)
				require.Equal(t, 9.0, assetExchange.WithdrawFee)
				//check create asset exchange
				assetExchange, err = s.GetAssetExchange(4)
				require.NoError(t, err)
				require.Equal(t, "ETH", assetExchange.Symbol)
				require.Equal(t, huobi, assetExchange.ExchangeID)
			},
		},
		{
			msg:      "create asset",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAsset,
						Data: common.CreateAssetEntry{
							Symbol:    "KNC",
							Name:      "Kyber",
							Address:   eth.HexToAddress("0x3f105f78359ad80562b4c34296a87b8e66c584c5"),
							Decimals:  18,
							SetRate:   common.SetRateNotSet,
							IsQuote:   true,
							Rebalance: true,
							PWI: &common.AssetPWI{
								Ask: common.PWIEquation{
									A:                   234,
									B:                   23,
									C:                   12,
									MinMinSpread:        234,
									PriceMultiplyFactor: 123,
								},
								Bid: common.PWIEquation{
									A:                   23,
									B:                   234,
									C:                   234,
									MinMinSpread:        234,
									PriceMultiplyFactor: 234,
								},
							},
							Target: &common.AssetTarget{
								Total:              12,
								Reserve:            24,
								TransferThreshold:  34,
								RebalanceThreshold: 1,
							},
							Exchanges: []common.AssetExchange{
								{
									Symbol:      "KNC",
									ExchangeID:  binance,
									MinDeposit:  34,
									WithdrawFee: 34,
									TargetRatio: 34,
									TradingPairs: []common.TradingPair{
										{
											Quote: 1,
										},
									},
									DepositAddress:    eth.HexToAddress("0x3f105f78359ad80562b4c34296a87b8e66c584c5"),
									TargetRecommended: 234,
								},
							},
							RebalanceQuadratic: &common.RebalanceQuadratic{
								A: 12,
								B: 34,
								C: 24,
							},
						},
					},
				},
			},
			assert: httputil.ExpectSuccess,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) {
			testHTTPRequest(t, tc, server.r)
		})
	}

}

func TestHTTPServerAssetExchangeWithOptionalTradingPair(t *testing.T) {
	const (
		settingChangePath = "/v3/setting-change"
	)

	db, tearDown := testutil.MustNewDevelopmentDB()
	defer func() {
		assert.NoError(t, tearDown())
	}()

	s, err := postgres.NewStorage(db)
	require.NoError(t, err)

	assetID, err := createSampleAsset(s)
	require.NoError(t, err)

	_, err = s.CreateAssetExchange(2, 1, "ETH",
		eth.HexToAddress("0x8cacc5bf46ea076a09febb2ae213dc4da6384e74"),
		0.001,
		0.001,
		0.001,
		0.001,
	)
	require.NoError(t, err)

	asset, err := s.GetAsset(1)
	require.NoError(t, err)
	t.Log(asset)

	server := NewServer(s, "", nil)

	var tests = []testCase{
		{
			msg:      "create asset exchange with invalid trading pair",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailure,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x001"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
							TradingPairs: []common.TradingPair{
								{
									Base:            1,
									Quote:           1,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
							},
						},
					},
				},
			},
		},
		{
			msg:      "create asset exchange with invalid trading pair - both base and quote are zero",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailure,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x001"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
							TradingPairs: []common.TradingPair{
								{
									Base:            0,
									Quote:           0,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
							},
						},
					},
				},
			},
		},
		{
			msg:      "create asset exchange with invalid trading pair - invalid quote token",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailure,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x001"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
							TradingPairs: []common.TradingPair{
								{
									Base:            0,
									Quote:           1234,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
							},
						},
					},
				},
			},
		},
		{
			msg:      "create asset exchange with invalid trading pair - invalid base token",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailure,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x001"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
							TradingPairs: []common.TradingPair{
								{
									Base:            1234,
									Quote:           0,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
							},
						},
					},
				},
			},
		},
		{
			msg:      "create asset exchange with duplicate trading pair",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			assert:   httputil.ExpectFailureWithReason(`failed to create trading pair base=2 quote=1 exchange_id=2 err=duplicate key value violates unique constraint "trading_pairs_exchange_id_base_id_quote_id_key"`),
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x001"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
							TradingPairs: []common.TradingPair{
								{
									Base:            0,
									Quote:           1,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
								{
									Base:            0,
									Quote:           1,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
							},
						},
					},
				},
			},
		},
		{
			msg:      "confirm asset exchange with duplicate trading pair",
			endpoint: settingChangePath + "/1",
			method:   http.MethodPut,
			assert:   httputil.ExpectFailureWithReason(`failed to create trading pair base=2 quote=1 exchange_id=2 err=duplicate key value violates unique constraint "trading_pairs_exchange_id_base_id_quote_id_key"`),
			data:     nil,
		},
		{
			msg:      "create asset exchange with trading pair successfully",
			endpoint: settingChangePath,
			method:   http.MethodPost,
			assert:   httputil.ExpectSuccess,
			data: &common.SettingChange{
				ChangeList: []common.SettingChangeEntry{
					{
						Type: common.ChangeTypeCreateAssetExchange,
						Data: common.CreateAssetExchangeEntry{
							AssetID:           assetID,
							ExchangeID:        huobi,
							Symbol:            "ETH",
							DepositAddress:    eth.HexToAddress("0x001"),
							MinDeposit:        10.0,
							WithdrawFee:       11.0,
							TargetRecommended: 12.0,
							TargetRatio:       13.0,
							TradingPairs: []common.TradingPair{
								{
									Base:            0,
									Quote:           1,
									AmountPrecision: 10,
									AmountLimitMin:  1000,
									AmountLimitMax:  10000,
									PriceLimitMin:   0.1,
									PriceLimitMax:   10.10,
									MinNotional:     0.001,
								},
							},
						},
					},
				},
			},
		},
		{
			msg:      "confirm asset exchange with trading pair successfully",
			endpoint: settingChangePath + "/2",
			method:   http.MethodPut,
			assert:   httputil.ExpectSuccess,
			data:     nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Log(tc.msg)
		t.Run(tc.msg, func(t *testing.T) { testHTTPRequest(t, tc, server.r) })
	}
}
