package http

import (
	"fmt"
	"log"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"

	v1common "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func (s *Server) validateChangeEntry(e common.SettingChangeType, changeType common.ChangeType) error {
	var (
		err error
	)

	switch changeType {
	case common.ChangeTypeCreateAsset:
		err = s.checkCreateAssetParams(*(e.(*common.CreateAssetEntry)))
	case common.ChangeTypeUpdateAsset:
		err = s.checkUpdateAssetParams(*(e.(*common.UpdateAssetEntry)))
	case common.ChangeTypeCreateAssetExchange:
		err = s.checkCreateAssetExchangeParams(*(e.(*common.CreateAssetExchangeEntry)))
	case common.ChangeTypeUpdateAssetExchange:
		err = s.checkUpdateAssetExchangeParams(*(e.(*common.UpdateAssetExchangeEntry)))
	case common.ChangeTypeCreateTradingPair:
		_, _, err = s.checkCreateTradingPairParams(*(e.(*common.CreateTradingPairEntry)))
	case common.ChangeTypeCreateTradingBy:
		err = s.checkCreateTradingByParams(*e.(*common.CreateTradingByEntry))
	case common.ChangeTypeChangeAssetAddr:
		err = s.checkChangeAssetAddressParams(*e.(*common.ChangeAssetAddressEntry))
	case common.ChangeTypeUpdateExchange:
		err = s.checkUpdateExchangeParams(*e.(*common.UpdateExchangeEntry))
	case common.ChangeTypeDeleteTradingPair:
		err = s.checkDeleteTradingPairParams(*e.(*common.DeleteTradingPairEntry))
	case common.ChangeTypeDeleteAssetExchange:
		err = s.checkDeleteAssetExchangeParams(*e.(*common.DeleteAssetExchangeEntry))
	case common.ChangeTypeUpdateStableTokenParams:
		return nil
	default:
		return errors.Errorf("unknown type of setting change: %v", reflect.TypeOf(e))
	}
	return err
}

func (s *Server) fillLiveInfoSettingChange(settingChange *common.SettingChange) error {
	assets, err := s.storage.GetAssets()
	if err != nil {
		return err
	}

	for i, o := range settingChange.ChangeList {
		switch o.Type {
		case common.ChangeTypeCreateAsset:
			asset := o.Data.(*common.CreateAssetEntry)
			for _, assetExchange := range asset.Exchanges {
				err = s.fillLiveInfoAssetExchange(assets, assetExchange.ExchangeID, assetExchange.TradingPairs, assetExchange.Symbol, assetExchange.AssetID)
				if err != nil {
					return fmt.Errorf("position %d, error: %v", i, err)
				}
			}
		case common.ChangeTypeCreateTradingPair:
			entry := o.Data.(*common.CreateTradingPairEntry)
			baseSymbol, quoteSymbol, err := s.checkCreateTradingPairParams(*entry)
			if err != nil {
				return err
			}
			tradingPairSymbol := common.TradingPairSymbols{TradingPair: entry.TradingPair}
			tradingPairSymbol.BaseSymbol = baseSymbol
			tradingPairSymbol.QuoteSymbol = quoteSymbol
			tradingPairSymbol.ID = uint64(1) // mock one
			exhID := v1common.ExchangeID(entry.ExchangeID)
			centralExh, ok := s.supportedExchanges[exhID]
			if !ok {
				return errors.Errorf("position %d, exchange %s not supported", i, exhID)
			}
			exchangeInfo, err := centralExh.GetLiveExchangeInfos([]common.TradingPairSymbols{tradingPairSymbol})
			if err != nil {
				return fmt.Errorf("position %d, error: %v", i, err)
			}
			info := exchangeInfo[tradingPairSymbol.ID]
			entry.MinNotional = info.MinNotional
			entry.AmountLimitMax = info.AmountLimit.Max
			entry.AmountLimitMin = info.AmountLimit.Min
			entry.AmountPrecision = uint64(info.Precision.Amount)
			entry.PricePrecision = uint64(info.Precision.Price)
			entry.PriceLimitMax = info.PriceLimit.Max
			entry.PriceLimitMin = info.PriceLimit.Min
		case common.ChangeTypeCreateAssetExchange:
			assetExchange := o.Data.(*common.CreateAssetExchangeEntry)
			err = s.fillLiveInfoAssetExchange(assets, assetExchange.ExchangeID, assetExchange.TradingPairs, assetExchange.Symbol, assetExchange.AssetID)
			if err != nil {
				return fmt.Errorf("position %d, error: %v", i, err)
			}
		}
	}
	return nil
}

func (s *Server) fillLiveInfoAssetExchange(assets []common.Asset, exchangeID uint64, tradingPairs []common.TradingPair, assetSymbol string, assetID uint64) error {
	exhID := v1common.ExchangeID(exchangeID)
	centralExh, ok := s.supportedExchanges[exhID]
	if !ok {
		return errors.Errorf("exchange %s not supported", exhID)
	}
	var tps []common.TradingPairSymbols
	index := uint64(1)
	for idx, tradingPair := range tradingPairs {
		tradingPairSymbol := common.TradingPairSymbols{TradingPair: tradingPair}
		tradingPairSymbol.ID = index
		if tradingPair.Quote == 0 {
			tradingPairSymbol.QuoteSymbol = assetSymbol
			base, err := getAssetExchange(assets, tradingPair.Base, exchangeID)
			if err != nil {
				return err
			}
			tradingPairSymbol.BaseSymbol = base.Symbol
			if assetID != 0 {
				tradingPairs[idx].Quote = assetID
			}
		}
		if tradingPair.Base == 0 {
			tradingPairSymbol.BaseSymbol = assetSymbol
			quote, err := getAssetExchange(assets, tradingPair.Quote, exchangeID)
			if err != nil {
				return err
			}
			tradingPairSymbol.QuoteSymbol = quote.Symbol
			if assetID != 0 {
				tradingPairs[idx].Base = assetID
			}
		}
		tps = append(tps, tradingPairSymbol)
		index++
	}
	exchangeInfo, err := centralExh.GetLiveExchangeInfos(tps)
	if err != nil {
		return err
	}
	tradingPairID := uint64(1)
	for idx := range tradingPairs {
		if info, ok := exchangeInfo[tradingPairID]; ok {
			tradingPairs[idx].MinNotional = info.MinNotional
			tradingPairs[idx].AmountLimitMax = info.AmountLimit.Max
			tradingPairs[idx].AmountLimitMin = info.AmountLimit.Min
			tradingPairs[idx].AmountPrecision = uint64(info.Precision.Amount)
			tradingPairs[idx].PricePrecision = uint64(info.Precision.Price)
			tradingPairs[idx].PriceLimitMax = info.PriceLimit.Max
			tradingPairs[idx].PriceLimitMin = info.PriceLimit.Min
			tradingPairID++
		}
	}
	return nil
}
func (s *Server) createSettingChangeWithType(t common.ChangeCatalog) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		s.createSettingChange(ctx, t)
	}
}
func (s *Server) createSettingChange(c *gin.Context, t common.ChangeCatalog) {
	var settingChange common.SettingChange
	if err := c.ShouldBindJSON(&settingChange); err != nil {
		log.Printf("cannot bind data to create setting_change from request err=%s", err.Error())
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if len(settingChange.ChangeList) == 0 {
		httputil.ResponseFailure(c, httputil.WithReason("change_list must not empty"))
		return
	}
	for i, o := range settingChange.ChangeList {
		if err := binding.Validator.ValidateStruct(o.Data); err != nil {
			msg := fmt.Sprintf("verify change list failed, position %d, err=%s", i, err)
			httputil.ResponseFailure(c, httputil.WithReason(msg), httputil.WithField("failed-at", o))
			return
		}

		if err := s.validateChangeEntry(o.Data, o.Type); err != nil {
			msg := fmt.Sprintf("verify change list failed, postision %d, err=%s", i, err)
			httputil.ResponseFailure(c, httputil.WithReason(msg), httputil.WithField("failed-at", o))
			return
		}
	}
	if err := s.fillLiveInfoSettingChange(&settingChange); err != nil {
		msg := fmt.Sprintf("validate trading pair info failed, %s", err)
		log.Println(msg)
		httputil.ResponseFailure(c, httputil.WithReason(msg))
		return
	}

	id, err := s.storage.CreateSettingChange(t, settingChange)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(makeFriendlyMessage(err)))
		return
	}

	// test confirm
	err = s.storage.ConfirmSettingChange(id, false)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(makeFriendlyMessage(err)))
		// clean up
		if err = s.storage.RejectSettingChange(id); err != nil {
			log.Printf("failed to clean up, error")
		}
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) getSettingChange(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	result, err := s.storage.GetSettingChange(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}
func (s *Server) getSettingChangeWithType(t common.ChangeCatalog) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		s.getSettingChanges(ctx, t)
	}
}
func (s *Server) getSettingChanges(c *gin.Context, t common.ChangeCatalog) {
	result, err := s.storage.GetSettingChanges(t)
	if err != nil {
		log.Printf("failed to get setting changes %v\n", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) rejectSettingChange(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.RejectSettingChange(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) confirmSettingChange(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.ConfirmSettingChange(input.ID, true)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) checkCreateTradingPairParams(createEntry common.CreateTradingPairEntry) (string, string, error) {
	var (
		ok           bool
		quoteAssetEx common.AssetExchange
		baseAssetEx  common.AssetExchange
	)

	base, err := s.storage.GetAsset(createEntry.Base)
	if err != nil {
		return "", "", errors.Wrapf(common.ErrBaseAssetInvalid, "base id: %v", createEntry.Base)
	}
	quote, err := s.storage.GetAsset(createEntry.Quote)
	if err != nil {
		return "", "", errors.Wrapf(common.ErrBaseAssetInvalid, "quote id: %v", createEntry.Quote)
	}

	if !quote.IsQuote {
		return "", "", errors.Wrap(common.ErrQuoteAssetInvalid, "quote asset should have is_quote=true")
	}

	if baseAssetEx, ok = getAssetExchangeByExchangeID(base, createEntry.ExchangeID); !ok {
		return "", "", errors.Wrap(common.ErrBaseAssetInvalid, "base asset not config on exchange")
	}

	if quoteAssetEx, ok = getAssetExchangeByExchangeID(quote, createEntry.ExchangeID); !ok {
		return "", "", errors.Wrap(common.ErrQuoteAssetInvalid, "quote asset not config on exchange")
	}
	return baseAssetEx.Symbol, quoteAssetEx.Symbol, nil
}

func getAssetExchangeByExchangeID(asset common.Asset, exchangeID uint64) (common.AssetExchange, bool) {
	for _, exchange := range asset.Exchanges {
		if exchange.ExchangeID == exchangeID {
			return exchange,
				true
		}
	}
	return common.AssetExchange{}, false
}

func (s *Server) checkCreateTradingByParams(createEntry common.CreateTradingByEntry) error {
	tpSymBol, err := s.storage.GetTradingPair(createEntry.TradingPairID)
	if err != nil {
		return err
	}
	if tpSymBol.Base != createEntry.AssetID && tpSymBol.Quote != createEntry.AssetID {
		return common.ErrTradingByAssetIDInvalid
	}
	return nil
}

func (s *Server) checkUpdateAssetParams(updateEntry common.UpdateAssetEntry) error {
	asset, err := s.storage.GetAsset(updateEntry.AssetID)
	if err != nil {
		return errors.Wrapf(err, "failed to get asset id: %v from db ", updateEntry.AssetID)
	}

	if updateEntry.Rebalance != nil && *updateEntry.Rebalance {
		if asset.RebalanceQuadratic == nil && updateEntry.RebalanceQuadratic == nil {
			return errors.Errorf("%v at asset id: %v", common.ErrRebalanceQuadraticMissing.Error(), updateEntry.AssetID)
		}

		if asset.Target == nil && updateEntry.Target == nil {
			return errors.Errorf("%v at asset id: %v", common.ErrAssetTargetMissing.Error(), updateEntry.AssetID)
		}
	}

	if updateEntry.SetRate != nil && *updateEntry.SetRate != common.SetRateNotSet && asset.PWI == nil && updateEntry.PWI == nil {
		return errors.Errorf("%v at asset id: %v", common.ErrPWIMissing.Error(), updateEntry.AssetID)
	}

	if updateEntry.Transferable != nil && *updateEntry.Transferable {
		for _, exchange := range asset.Exchanges {
			if common.IsZeroAddress(exchange.DepositAddress) {
				return errors.Errorf("%v at asset id: %v and asset_exchange: %v", common.ErrDepositAddressMissing, updateEntry.AssetID, exchange.ID)
			}
		}
	}
	return nil
}

func (s *Server) checkUpdateAssetExchangeParams(updateEntry common.UpdateAssetExchangeEntry) error {
	assetExchange, err := s.storage.GetAssetExchange(updateEntry.ID)
	if err != nil {
		return errors.Wrap(err, "asset exchange not found")
	}

	asset, err := s.storage.GetAsset(assetExchange.AssetID)
	if err != nil {
		return errors.Wrap(err, "asset not found")
	}

	if asset.Transferable && updateEntry.DepositAddress != nil && common.IsZeroAddress(*updateEntry.DepositAddress) {
		return common.ErrDepositAddressMissing
	}
	return nil
}

func (s *Server) checkCreateAssetExchangeParams(createEntry common.CreateAssetExchangeEntry) error {
	asset, err := s.storage.GetAsset(createEntry.AssetID)
	if err != nil {
		return errors.Wrap(err, "asset not found")
	}

	_, err = s.storage.GetExchange(createEntry.ExchangeID)
	if err != nil {
		return errors.Wrap(err, "exchange not found")
	}

	for _, exchange := range asset.Exchanges {
		if exchange.ExchangeID == createEntry.ExchangeID {
			return common.ErrAssetExchangeAlreadyExist
		}
	}
	if asset.Transferable && common.IsZeroAddress(createEntry.DepositAddress) {
		return common.ErrDepositAddressMissing
	}
	for _, tradingPair := range createEntry.TradingPairs {
		if tradingPair.Base != 0 && tradingPair.Quote != 0 {
			return errors.Wrapf(common.ErrBadTradingPairConfiguration, "base id:%v quote id:%v", tradingPair.Base, tradingPair.Quote)
		}

		if tradingPair.Base == 0 && tradingPair.Quote == 0 {
			return errors.Wrapf(common.ErrBadTradingPairConfiguration, "base id:%v quote id:%v", tradingPair.Base, tradingPair.Quote)
		}

		if tradingPair.Base == 0 {
			quoteAsset, err := s.storage.GetAsset(tradingPair.Quote)
			if err != nil {
				return errors.Wrapf(common.ErrQuoteAssetInvalid, "quote id: %v", tradingPair.Quote)
			}
			if !quoteAsset.IsQuote {
				return errors.Wrapf(common.ErrQuoteAssetInvalid, "quote id: %v", tradingPair.Quote)
			}
		}

		if tradingPair.Quote == 0 {
			_, err := s.storage.GetAsset(tradingPair.Base)
			if err != nil {
				return errors.Wrapf(common.ErrBaseAssetInvalid, "base id: %v", tradingPair.Base)
			}

			if !asset.IsQuote {
				return errors.Wrapf(common.ErrQuoteAssetInvalid, "quote id: %v", tradingPair.Quote)
			}
		}
	}
	return nil
}

func getAssetExchange(assets []common.Asset, assetID, exchangeID uint64) (common.AssetExchange, error) {
	for _, asset := range assets {
		if asset.ID == assetID {
			for _, assetExchange := range asset.Exchanges {
				if assetExchange.ExchangeID == exchangeID {
					return assetExchange, nil
				}
			}
		}
	}
	return common.AssetExchange{}, fmt.Errorf("AssetExchange not found, asset=%d exchange=%d", assetID, exchangeID)
}

func (s *Server) checkCreateAssetParams(createEntry common.CreateAssetEntry) error {
	if createEntry.Transferable {
		if s.blockchain == nil {
			return common.ErrBlockchainHaveNotInitiated
		}
		if err := s.blockchain.CheckTokenIndices(createEntry.Address); err != nil {
			return common.ErrAssetAddressIsNotIndexInContract
		}
	}
	if createEntry.Rebalance && createEntry.RebalanceQuadratic == nil {
		return common.ErrRebalanceQuadraticMissing
	}

	if createEntry.Rebalance && createEntry.Target == nil {
		return common.ErrAssetTargetMissing
	}

	if createEntry.SetRate != common.SetRateNotSet && createEntry.PWI == nil {
		return common.ErrPWIMissing
	}

	for _, exchange := range createEntry.Exchanges {
		if common.IsZeroAddress(exchange.DepositAddress) && createEntry.Transferable {
			return errors.Wrapf(common.ErrDepositAddressMissing, "exchange %v", exchange.Symbol)
		}

		for _, tradingPair := range exchange.TradingPairs {

			if tradingPair.Base != 0 && tradingPair.Quote != 0 {
				return errors.Wrapf(common.ErrBadTradingPairConfiguration, "base id:%v quote id:%v", tradingPair.Base, tradingPair.Quote)
			}

			if tradingPair.Base == 0 && tradingPair.Quote == 0 {
				return errors.Wrapf(common.ErrBadTradingPairConfiguration, "base id:%v quote id:%v", tradingPair.Base, tradingPair.Quote)
			}

			if tradingPair.Base == 0 {
				quoteAsset, err := s.storage.GetAsset(tradingPair.Quote)
				if err != nil {
					return errors.Wrapf(common.ErrQuoteAssetInvalid, "quote id: %v", tradingPair.Quote)
				}
				if !quoteAsset.IsQuote {
					return errors.Wrapf(common.ErrQuoteAssetInvalid, "quote id: %v", tradingPair.Quote)
				}
			}

			if tradingPair.Quote == 0 {
				_, err := s.storage.GetAsset(tradingPair.Base)
				if err != nil {
					return errors.Wrapf(common.ErrBaseAssetInvalid, "base id: %v", tradingPair.Base)
				}

				if !createEntry.IsQuote {
					return errors.Wrapf(common.ErrQuoteAssetInvalid, "quote id: %v", tradingPair.Quote)
				}
			}
		}
	}

	return nil
}

func (s *Server) checkChangeAssetAddressParams(changeAssetAddressEntry common.ChangeAssetAddressEntry) error {
	asset, err := s.storage.GetAsset(changeAssetAddressEntry.ID)
	if err != nil {
		return err
	}
	if asset.Address == changeAssetAddressEntry.Address {
		return common.ErrAddressExists
	}
	return nil
}

func (s *Server) checkUpdateExchangeParams(updateExchangeEntry common.UpdateExchangeEntry) error {
	//check if exchange exist
	_, err := s.storage.GetExchange(updateExchangeEntry.ExchangeID)
	return err
}
