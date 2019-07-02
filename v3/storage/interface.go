package storage

import (
	ethereum "github.com/ethereum/go-ethereum/common"

	v3 "github.com/KyberNetwork/reserve-data/v3/common"
)

// Interface is the common persistent storage interface of V3 APIs.
type Interface interface {
	SettingReader

	GetExchanges() ([]v3.Exchange, error)
	GetExchange(id uint64) (v3.Exchange, error)
	UpdateExchange(id uint64, opts ...UpdateExchangeOption) error

	CreateAsset(
		symbol, name string,
		address ethereum.Address,
		decimals uint64,
		setRate v3.SetRate,
		rebalance bool,
		isQuote bool,
		pwi *v3.AssetPWI,
		rb *v3.RebalanceQuadratic,
		exchanges []v3.AssetExchange,
		target *v3.AssetTarget,
	) (uint64, error)
	GetAsset(id uint64) (v3.Asset, error)
	UpdateAsset(id uint64, opts ...UpdateAssetOption) error
	// ChangeAssetAddress make the current address address of asset old address and set new address as current.
	ChangeAssetAddress(id uint64, address ethereum.Address) error

	// TODO update precision pairs and live deposit addresses on startup

	// TODO method for batch update PWI
	// TODO method for batch update rebalance quadratic
	// TODO method for batch update exchange configuration
	// TODO meethod for batch update target
	// TODO method for update address
}

// SettingReader is the common interface for reading exchanges, assets configuration.
type SettingReader interface {
	GetTradingPairSymbols(exchangeID uint64) ([]v3.TradingPairSymbols, error)
	GetDepositAddresses(exchangeID uint64) (map[string]ethereum.Address, error)
	GetAssets() ([]v3.Asset, error)
	// GetSetRateAssets returns all assets that the set rate strategy is not not_set.
	// TODO: this should be call ReserveAssets or something else
	GetSetRateAssets() ([]v3.Asset, error)
	GetMinNotional(exchangeID, baseID, quoteID uint64) (float64, error)
	// TODO: this method should be removed, as we will accept asset_id instead of symbol everywhere
	GetAssetBySymbol(exchangeID uint64, symbol string) (v3.Asset, error)
}

// UpdateExchangeOpts is the options of UpdateAsset method.
type UpdateExchangeOpts struct {
	tradingFeeMaker *float64
	tradingFeeTaker *float64
	disable         *bool
}

func (u *UpdateExchangeOpts) TradingFeeMaker() *float64 {
	return u.tradingFeeMaker
}

func (u *UpdateExchangeOpts) TradingFeeTaker() *float64 {
	return u.tradingFeeTaker
}

func (u *UpdateExchangeOpts) Disable() *bool {
	return u.disable
}

type UpdateExchangeOption func(opts *UpdateExchangeOpts)

func WithTradingFeeMakerUpdateExchangeOpt(tradingFeeMaker float64) UpdateExchangeOption {
	return func(opts *UpdateExchangeOpts) {
		opts.tradingFeeMaker = &tradingFeeMaker
	}
}

func WithTradingFeeTakerUpdateExchangeOpt(tradingFeeTaker float64) UpdateExchangeOption {
	return func(opts *UpdateExchangeOpts) {
		opts.tradingFeeTaker = &tradingFeeTaker
	}
}

func WithDisableExchangeOpt(disable bool) UpdateExchangeOption {
	return func(opts *UpdateExchangeOpts) {
		opts.disable = &disable
	}
}

type UpdateAssetOpts struct {
	symbol    *string
	name      *string
	address   *ethereum.Address
	decimals  *uint64
	setRate   *v3.SetRate
	rebalance *bool
	isQuote   *bool
}

func (u *UpdateAssetOpts) Symbol() *string {
	return u.symbol
}

func (u *UpdateAssetOpts) Name() *string {
	return u.name
}

func (u *UpdateAssetOpts) Address() *ethereum.Address {
	return u.address
}

func (u *UpdateAssetOpts) Decimals() *uint64 {
	return u.decimals
}

func (u *UpdateAssetOpts) SetRate() *v3.SetRate {
	return u.setRate
}

func (u *UpdateAssetOpts) Rebalance() *bool {
	return u.rebalance
}

func (u *UpdateAssetOpts) IsQuote() *bool {
	return u.isQuote
}

type UpdateAssetOption func(opts *UpdateAssetOpts)

func WithSymbolUpdateAssetOption(symbol string) UpdateAssetOption {
	return func(opts *UpdateAssetOpts) {
		opts.symbol = &symbol
	}
}

func WithNameUpdateAssetOption(name string) UpdateAssetOption {
	return func(opts *UpdateAssetOpts) {
		opts.name = &name
	}
}

func WithAddressUpdateAssetOption(address ethereum.Address) UpdateAssetOption {
	return func(opts *UpdateAssetOpts) {
		opts.address = &address
	}
}

func WithDecimalsUpdateAssetOption(decimals uint64) UpdateAssetOption {
	return func(opts *UpdateAssetOpts) {
		opts.decimals = &decimals
	}
}

func WithSetRateUpdateAssetOption(setRate v3.SetRate) UpdateAssetOption {
	return func(opts *UpdateAssetOpts) {
		opts.setRate = &setRate
	}
}

func WithRebalanceUpdateAssetOption(rebalance bool) UpdateAssetOption {
	return func(opts *UpdateAssetOpts) {
		opts.rebalance = &rebalance
	}
}

func WithIsQuoteUpdateAssetOption(isQuote bool) UpdateAssetOption {
	return func(opts *UpdateAssetOpts) {
		opts.isQuote = &isQuote
	}
}