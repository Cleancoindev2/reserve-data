package world

import (
	"encoding/json"
	"io/ioutil"
)

var (
	allFeeds = []string{
		"DGX",
		"OneForgeETH",
		"OneForgeUSD",
		"GDAX",
		"Kraken",
		"Gemini",

		"CoinbaseBTC",
		"GeminiBTC",

		"CoinbaseUSD",
		"BinanceUSD",

		"CoinbaseUSDC",
		"BinanceUSDC",

		"CoinbaseDAI",
		"HitDAI",

		"BitFinexUSD",
		"BinanceUSDT",
		"BinancePAX",
		"BinanceTUSD",
	}
	// remove unused feeds
)

// AllFeeds returns all configured feed sources.
func AllFeeds() []string {
	return allFeeds
}

// Endpoint returns all API endpoints to use in TheWorld struct.
type Endpoint interface {
	GoldDataEndpoint() string
	OneForgeGoldETHDataEndpoint() string
	OneForgeGoldUSDDataEndpoint() string
	GDAXDataEndpoint() string
	KrakenDataEndpoint() string
	GeminiDataEndpoint() string

	CoinbaseBTCEndpoint() string
	GeminiBTCEndpoint() string

	CoinbaseUSDCEndpoint() string
	BinanceUSDCEndpoint() string
	CoinbaseUSDEndpoint() string
	CoinbaseDAIEndpoint() string
	HitDaiEndpoint() string

	BitFinexUSDTEndpoint() string
	BinanceUSDTEndpoint() string
	BinancePAXEndpoint() string
	BinanceTUSDEndpoint() string
}

type RealEndpoint struct {
	OneForgeKey string `json:"oneforge"`
}

func (re RealEndpoint) BitFinexUSDTEndpoint() string {
	return "https://api-pub.bitfinex.com/v2/ticker/tETHUSD"
}

func (re RealEndpoint) BinanceUSDTEndpoint() string {
	return "https://api.binance.com/api/v3/ticker/bookTicker?symbol=ETHUSDT"
}

func (re RealEndpoint) BinancePAXEndpoint() string {
	return "https://api.binance.com/api/v3/ticker/bookTicker?symbol=ETHPAX"
}

func (re RealEndpoint) BinanceTUSDEndpoint() string {
	return "https://api.binance.com/api/v3/ticker/bookTicker?symbol=ETHTUSD"
}

func (re RealEndpoint) CoinbaseDAIEndpoint() string {
	return "https://api.pro.coinbase.com/products/eth-dai/ticker"
}

func (re RealEndpoint) HitDaiEndpoint() string {
	return "https://api.hitbtc.com/api/2/public/ticker/ETHDAI"
}

func (re RealEndpoint) CoinbaseUSDEndpoint() string {
	return "https://api.pro.coinbase.com/products/eth-usd/ticker"
}

func (re RealEndpoint) CoinbaseUSDCEndpoint() string {
	return "https://api.pro.coinbase.com/products/eth-usdc/ticker"
}

func (re RealEndpoint) BinanceUSDCEndpoint() string {
	return "https://api.binance.com/api/v3/ticker/bookTicker?symbol=ETHUSDC"
}

func (re RealEndpoint) GoldDataEndpoint() string {
	return "https://datafeed.digix.global/tick/"
}

func (re RealEndpoint) OneForgeGoldETHDataEndpoint() string {
	return "https://forex.1forge.com/1.0.3/convert?from=XAU&to=ETH&quantity=1&api_key=" + re.OneForgeKey
}

func (re RealEndpoint) OneForgeGoldUSDDataEndpoint() string {
	return "https://forex.1forge.com/1.0.3/convert?from=XAU&to=USD&quantity=1&api_key=" + re.OneForgeKey
}

func (re RealEndpoint) GDAXDataEndpoint() string {
	return "https://api.pro.coinbase.com/products/eth-usd/ticker"
}

func (re RealEndpoint) KrakenDataEndpoint() string {
	return "https://api.kraken.com/0/public/Ticker?pair=ETHUSD"
}

func (re RealEndpoint) GeminiDataEndpoint() string {
	return "https://api.gemini.com/v1/pubticker/ethusd"
}

func (re RealEndpoint) CoinbaseBTCEndpoint() string {
	return "https://api.pro.coinbase.com/products/eth-btc/ticker"
}

func (re RealEndpoint) GeminiBTCEndpoint() string {
	return "https://api.gemini.com/v1/pubticker/ethbtc"
}

func NewRealEndpointFromFile(path string) (*RealEndpoint, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := RealEndpoint{}
	err = json.Unmarshal(data, &result)
	return &result, err
}

type SimulatedEndpoint struct {
}

func (se SimulatedEndpoint) BitFinexUSDTEndpoint() string {
	panic("implement me")
}

func (se SimulatedEndpoint) BinanceUSDTEndpoint() string {
	panic("implement me")
}

func (se SimulatedEndpoint) BinancePAXEndpoint() string {
	panic("implement me")
}

func (se SimulatedEndpoint) BinanceTUSDEndpoint() string {
	panic("implement me")
}

func (se SimulatedEndpoint) CoinbaseDAIEndpoint() string {
	panic("implement me")
}

func (se SimulatedEndpoint) HitDaiEndpoint() string {
	panic("implement me")
}

func (se SimulatedEndpoint) CoinbaseUSDEndpoint() string {
	panic("implement me")
}

// TODO: support simulation
func (se SimulatedEndpoint) CoinbaseUSDCEndpoint() string {
	panic("implement me")
}

func (se SimulatedEndpoint) BinanceUSDCEndpoint() string {
	panic("implement me")
}

func (se SimulatedEndpoint) GoldDataEndpoint() string {
	return "http://simulator:5400/tick"
}

func (se SimulatedEndpoint) OneForgeGoldETHDataEndpoint() string {
	return "http://simulator:5500/1.0.3/convert?from=XAU&to=ETH&quantity=1&api_key="
}

func (se SimulatedEndpoint) OneForgeGoldUSDDataEndpoint() string {
	return "http://simulator:5500/1.0.3/convert?from=XAU&to=USD&quantity=1&api_key="
}

func (se SimulatedEndpoint) GDAXDataEndpoint() string {
	return "http://simulator:5600/products/eth-usd/ticker"
}

func (se SimulatedEndpoint) KrakenDataEndpoint() string {
	return "http://simulator:5700/0/public/Ticker?pair=ETHUSD"
}

func (se SimulatedEndpoint) GeminiDataEndpoint() string {
	return "http://simulator:5800/v1/pubticker/ethusd"
}

func (se SimulatedEndpoint) CoinbaseBTCEndpoint() string {
	return "http://simulator:5600/products/eth-btc/ticker"
}

func (se SimulatedEndpoint) GeminiBTCEndpoint() string {
	return "http://simulator:5800/v1/pubticker/ethbtc"
}
