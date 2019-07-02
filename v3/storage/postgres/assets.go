package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/lib/pq"

	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage"
)

const addressesUniqueConstraint = "addresses_address_key"

type createAssetParams struct {
	Symbol    string `db:"symbol"`
	Name      string `db:"name"`
	Address   string `db:"address"`
	Decimals  uint64 `db:"decimals"`
	SetRate   string `db:"set_rate"`
	Rebalance bool   `db:"rebalance"`
	IsQuote   bool   `db:"is_quote"`

	AskA                   *float64 `db:"ask_a"`
	AskB                   *float64 `db:"ask_b"`
	AskC                   *float64 `db:"ask_c"`
	AskMinMinSpread        *float64 `db:"ask_min_min_spread"`
	AskPriceMultiplyFactor *float64 `db:"ask_price_multiply_factor"`
	BidA                   *float64 `db:"bid_a"`
	BidB                   *float64 `db:"bid_b"`
	BidC                   *float64 `db:"bid_c"`
	BidMinMinSpread        *float64 `db:"bid_min_min_spread"`
	BidPriceMultiplyFactor *float64 `db:"bid_price_multiply_factor"`

	RebalanceQuadraticA *float64 `db:"rebalance_quadratic_a"`
	RebalanceQuadraticB *float64 `db:"rebalance_quadratic_b"`
	RebalanceQuadraticC *float64 `db:"rebalance_quadratic_c"`

	TargetTotal              *float64 `db:"target_total"`
	TargetReserve            *float64 `db:"target_reserve"`
	TargetRebalanceThreshold *float64 `db:"target_rebalance_threshold"`
	TargetTransferThreshold  *float64 `db:"target_transfer_threshold"`
}

func (s *Storage) CreateAsset(
	symbol, name string,
	address ethereum.Address,
	decimals uint64,
	setRate common.SetRate,
	rebalance, isQuote bool,
	pwi *common.AssetPWI,
	rb *common.RebalanceQuadratic,
	exchanges []common.AssetExchange,
	target *common.AssetTarget,
) (uint64, error) {
	var assetID uint64
	tx, err := s.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer rollbackUnlessCommitted(tx)

	if common.IsZeroAddress(address) {
		if isQuote {
			log.Printf("creating quote asset without address symbol=%s is_quote=%t", symbol, isQuote)
		} else {
			log.Printf("refusing to create non-quote asset without address")
			return 0, common.ErrAddressMissing
		}
	} else {
		log.Printf("creating new asset symbol=%s adress=%s", symbol, address.String())
	}

	arg := createAssetParams{
		Symbol:    symbol,
		Name:      name,
		Address:   address.String(),
		Decimals:  decimals,
		SetRate:   setRate.String(),
		Rebalance: rebalance,
		IsQuote:   isQuote,
	}

	if pwi != nil {
		arg.AskA = &pwi.Ask.A
		arg.AskB = &pwi.Ask.B
		arg.AskC = &pwi.Ask.C
		arg.AskMinMinSpread = &pwi.Ask.MinMinSpread
		arg.AskPriceMultiplyFactor = &pwi.Ask.PriceMultiplyFactor
		arg.BidA = &pwi.Bid.A
		arg.BidB = &pwi.Bid.B
		arg.BidC = &pwi.Bid.C
		arg.BidMinMinSpread = &pwi.Bid.MinMinSpread
		arg.BidPriceMultiplyFactor = &pwi.Bid.PriceMultiplyFactor
	}

	if rebalance {
		if rb == nil {
			log.Printf("rebalance is enabled but rebalance quadratic is invalid symbol=%s", symbol)
			return 0, common.ErrRebalanceQuadraticMissing
		}

		if len(exchanges) == 0 {
			log.Printf("rebalance is enabled but no exchange configuration is provided symbol=%s", symbol)
			return 0, common.ErrAssetExchangeMissing
		}

		if target == nil {
			log.Printf("rebalance is enabled but target configuration is invalid symbol=%s", symbol)
			return 0, common.ErrAssetTargetMissing
		}
	}

	if rb != nil {
		arg.RebalanceQuadraticA = &rb.A
		arg.RebalanceQuadraticB = &rb.B
		arg.RebalanceQuadraticC = &rb.C
	}

	if target != nil {
		arg.TargetTotal = &target.Total
		arg.TargetReserve = &target.Reserve
		arg.TargetRebalanceThreshold = &target.RebalanceThreshold
		arg.TargetTransferThreshold = &target.TransferThreshold
	}

	if err := tx.NamedStmt(s.stmts.newAsset).Get(&assetID, arg); err != nil {
		pErr, ok := err.(*pq.Error)
		if !ok {
			return 0, fmt.Errorf("unknown returned err=%s", err.Error())
		}

		log.Printf("failed to create new asset err=%s", pErr.Message)
		switch pErr.Code {
		case errCodeUniqueViolation:
			switch pErr.Constraint {
			case addressesUniqueConstraint:
				return 0, common.ErrAddressExists
			case "assets_symbol_key":
				return 0, common.ErrSymbolExists
			}
		case errCodeCheckViolation:
			if pErr.Constraint == "pwi_check" {
				return 0, common.ErrPWIMissing
			}
		}
		return 0, pErr
	}

	for _, exchange := range exchanges {
		var assetExchangeID uint64
		err := tx.NamedStmt(s.stmts.newAssetExchange).Get(&assetExchangeID, struct {
			ExchangeID        uint64  `db:"exchange_id"`
			AssetID           uint64  `db:"asset_id"`
			Symbol            string  `db:"symbol"`
			DepositAddress    string  `db:"deposit_address"`
			MinDeposit        float64 `db:"min_deposit"`
			WithdrawFee       float64 `db:"withdraw_fee"`
			PricePrecision    int64   `db:"price_precision"`
			AmountPrecision   int64   `db:"amount_precision"`
			AmountLimitMin    float64 `db:"amount_limit_min"`
			AmountLimitMax    float64 `db:"amount_limit_max"`
			PriceLimitMin     float64 `db:"price_limit_min"`
			PriceLimitMax     float64 `db:"price_limit_max"`
			TargetRecommended float64 `db:"target_recommended"`
			TargetRatio       float64 `db:"target_ratio"`
		}{
			ExchangeID:        exchange.ExchangeID,
			AssetID:           assetID,
			Symbol:            exchange.Symbol,
			DepositAddress:    exchange.DepositAddress.String(),
			MinDeposit:        exchange.MinDeposit,
			WithdrawFee:       exchange.WithdrawFee,
			PricePrecision:    exchange.PricePrecision,
			AmountPrecision:   exchange.AmountPrecision,
			AmountLimitMin:    exchange.AmountLimitMin,
			AmountLimitMax:    exchange.AmountLimitMax,
			PriceLimitMin:     exchange.PriceLimitMin,
			PriceLimitMax:     exchange.PriceLimitMax,
			TargetRecommended: exchange.TargetRecommended,
			TargetRatio:       exchange.TargetRatio,
		})

		if err != nil {
			return 0, err
		}

		log.Printf("asset exchange is created id=%d", assetExchangeID)

		for _, pair := range exchange.TradingPairs {
			var (
				tradingPairID uint64
				baseID        = pair.Base
				quoteID       = pair.Quote
			)

			if baseID != 0 && quoteID != 0 {
				log.Printf(
					"both base and quote are provided asset_symbol=%s exchange_id=%d",
					symbol,
					exchange.ExchangeID)
				return 0, common.ErrBadTradingPairConfiguration
			}

			if baseID == 0 {
				baseID = assetID
			}
			if quoteID == 0 {
				quoteID = assetID
			}

			err = tx.Stmtx(s.stmts.newTradingPair).Get(
				&tradingPairID,
				exchange.ExchangeID,
				baseID,
				quoteID,
				pair.MinNotional,
			)
			if err != nil {
				pErr, ok := err.(*pq.Error)
				if !ok {
					return 0, fmt.Errorf("unknown error returned err=%s", err.Error())
				}

				switch pErr.Code {
				case errAssertFailure, errForeignKeyViolation:
					log.Printf("failed to create trading pair as assertion failed symbol=%s exchange_id=%d err=%s",
						symbol,
						exchange.ExchangeID,
						pErr.Message)
					return 0, common.ErrBadTradingPairConfiguration
				}

				return 0, fmt.Errorf("failed to create trading pair symbol=%s exchange_id=%d err=%s",
					symbol,
					exchange.ExchangeID,
					pErr.Message,
				)
			}
			log.Printf("trading pair created id=%d", tradingPairID)
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return assetID, nil
}

type assetExchangeDB struct {
	ID                uint64          `db:"id"`
	ExchangeID        uint64          `db:"exchange_id"`
	AssetID           uint64          `db:"asset_id"`
	Symbol            string          `db:"symbol"`
	DepositAddress    string          `db:"deposit_address"`
	MinDeposit        float64         `db:"min_deposit"`
	WithdrawFee       float64         `db:"withdraw_fee"`
	PricePrecision    int64           `db:"price_precision"`
	AmountPrecision   int64           `db:"amount_precision"`
	AmountLimitMin    float64         `db:"amount_limit_min"`
	AmountLimitMax    float64         `db:"amount_limit_max"`
	PriceLimitMin     float64         `db:"price_limit_min"`
	PriceLimitMax     float64         `db:"price_limit_max"`
	TargetRecommended sql.NullFloat64 `db:"target_recommended"`
	TargetRatio       sql.NullFloat64 `db:"target_ratio"`
}

func (aedb *assetExchangeDB) ToCommon() common.AssetExchange {
	result := common.AssetExchange{
		ID:              aedb.ID,
		ExchangeID:      aedb.ExchangeID,
		Symbol:          aedb.Symbol,
		DepositAddress:  ethereum.HexToAddress(aedb.DepositAddress),
		MinDeposit:      aedb.MinDeposit,
		WithdrawFee:     aedb.WithdrawFee,
		PricePrecision:  aedb.PricePrecision,
		AmountPrecision: aedb.AmountPrecision,
		AmountLimitMin:  aedb.AmountLimitMin,
		AmountLimitMax:  aedb.AmountLimitMax,
		PriceLimitMin:   aedb.PriceLimitMin,
		PriceLimitMax:   aedb.PriceLimitMax,
		TradingPairs:    nil,
	}
	if aedb.TargetRecommended.Valid {
		result.TargetRecommended = aedb.TargetRecommended.Float64
	}
	if aedb.TargetRatio.Valid {
		result.TargetRatio = aedb.TargetRatio.Float64
	}
	return result
}

type tradingPairDB struct {
	ID          uint64  `db:"id"`
	ExchangeID  uint64  `db:"exchange_id"`
	BaseID      uint64  `db:"base_id"`
	QuoteID     uint64  `db:"quote_id"`
	MinNotional float64 `db:"min_notional"`
}

func (tpd *tradingPairDB) ToCommon() common.TradingPair {
	return common.TradingPair{
		ID:          tpd.ID,
		Base:        tpd.BaseID,
		Quote:       tpd.QuoteID,
		MinNotional: tpd.MinNotional,
	}
}

type assetDB struct {
	ID           uint64         `db:"id"`
	Symbol       string         `db:"symbol"`
	Name         string         `db:"name"`
	Address      sql.NullString `db:"address"`
	OldAddresses pq.StringArray `db:"old_addresses"`
	Decimals     uint64         `db:"decimals"`
	SetRate      string         `db:"set_rate"`
	Rebalance    bool           `db:"rebalance"`
	IsQuote      bool           `db:"is_quote"`

	PWIAskA                   *float64 `db:"pwi_ask_a"`
	PWIAskB                   *float64 `db:"pwi_ask_b"`
	PWIAskC                   *float64 `db:"pwi_ask_c"`
	PWIAskMinMinSpread        *float64 `db:"pwi_ask_min_min_spread"`
	PWIAskPriceMultiplyFactor *float64 `db:"pwi_ask_price_multiply_factor"`
	PWIBidA                   *float64 `db:"pwi_bid_a"`
	PWIBidB                   *float64 `db:"pwi_bid_b"`
	PWIBidC                   *float64 `db:"pwi_bid_c"`
	PWIBidMinMinSpread        *float64 `db:"pwi_bid_min_min_spread"`
	PWIBidPriceMultiplyFactor *float64 `db:"pwi_bid_price_multiply_factor"`

	RebalanceQuadraticA *float64 `db:"rebalance_quadratic_a"`
	RebalanceQuadraticB *float64 `db:"rebalance_quadratic_b"`
	RebalanceQuadraticC *float64 `db:"rebalance_quadratic_c"`

	TargetTotal              *float64 `db:"target_total"`
	TargetReserve            *float64 `db:"target_reserve"`
	TargetRebalanceThreshold *float64 `db:"target_rebalance_threshold"`
	TargetTransferThreshold  *float64 `db:"target_transfer_threshold"`

	Created time.Time `db:"created"`
	Updated time.Time `db:"updated"`
}

func (adb *assetDB) ToCommon() (common.Asset, error) {
	result := common.Asset{
		ID:        adb.ID,
		Symbol:    adb.Symbol,
		Name:      adb.Name,
		Address:   ethereum.Address{},
		Decimals:  adb.Decimals,
		Rebalance: adb.Rebalance,
		IsQuote:   adb.IsQuote,
		Created:   adb.Created,
		Updated:   adb.Updated,
	}

	if adb.Address.Valid {
		result.Address = ethereum.HexToAddress(adb.Address.String)
	}

	for _, oldAddress := range adb.OldAddresses {
		result.OldAddresses = append(result.OldAddresses, ethereum.HexToAddress(oldAddress))
	}

	setRate, ok := common.SetRateFromString(adb.SetRate)
	if !ok {
		return common.Asset{}, fmt.Errorf("invalid set rate value %s", adb.SetRate)
	}

	result.SetRate = setRate

	if adb.PWIAskA != nil &&
		adb.PWIAskB != nil &&
		adb.PWIAskC != nil &&
		adb.PWIAskMinMinSpread != nil &&
		adb.PWIAskPriceMultiplyFactor != nil &&
		adb.PWIBidA != nil &&
		adb.PWIBidB != nil &&
		adb.PWIBidC != nil &&
		adb.PWIBidMinMinSpread != nil &&
		adb.PWIBidPriceMultiplyFactor != nil {
		result.PWI = &common.AssetPWI{
			Ask: common.PWIEquation{
				A:                   *adb.PWIAskA,
				B:                   *adb.PWIAskB,
				C:                   *adb.PWIAskC,
				MinMinSpread:        *adb.PWIAskMinMinSpread,
				PriceMultiplyFactor: *adb.PWIAskPriceMultiplyFactor,
			},
			Bid: common.PWIEquation{
				A:                   *adb.PWIBidA,
				B:                   *adb.PWIBidB,
				C:                   *adb.PWIBidC,
				MinMinSpread:        *adb.PWIBidMinMinSpread,
				PriceMultiplyFactor: *adb.PWIBidPriceMultiplyFactor,
			},
		}
	}
	if adb.RebalanceQuadraticA != nil && adb.RebalanceQuadraticB != nil && adb.RebalanceQuadraticC != nil {
		result.RebalanceQuadratic = &common.RebalanceQuadratic{
			A: *adb.RebalanceQuadraticA,
			B: *adb.RebalanceQuadraticB,
			C: *adb.RebalanceQuadraticC,
		}
	}

	if adb.TargetTotal != nil &&
		adb.TargetReserve != nil &&
		adb.TargetRebalanceThreshold != nil &&
		adb.TargetTransferThreshold != nil {
		result.Target = &common.AssetTarget{
			Total:              *adb.TargetTotal,
			Reserve:            *adb.TargetReserve,
			RebalanceThreshold: *adb.TargetRebalanceThreshold,
			TransferThreshold:  *adb.TargetTransferThreshold,
		}
	}

	return result, nil
}

func (s *Storage) GetAssets() ([]common.Asset, error) {
	var (
		allAssetDBs       []assetDB
		allAssetExchanges []assetExchangeDB
		allOrderBooks     []tradingPairDB
		results           []common.Asset
	)

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer rollbackUnlessCommitted(tx)

	if err := tx.Stmtx(s.stmts.getAsset).Select(&allAssetDBs, nil); err != nil {
		return nil, err
	}

	if err := tx.Stmtx(s.stmts.getAssetExchange).Select(&allAssetExchanges, nil); err != nil {
		return nil, err
	}

	if err := tx.Stmtx(s.stmts.getTradingPair).Select(&allOrderBooks, nil); err != nil {
		return nil, err
	}

	for _, assetDBResult := range allAssetDBs {
		result, err := assetDBResult.ToCommon()
		if err != nil {
			return nil, fmt.Errorf("invalid database record for asset id=%d err=%s", assetDBResult.ID, err.Error())
		}

		for _, assetExchangeResult := range allAssetExchanges {
			if assetExchangeResult.AssetID == assetDBResult.ID {
				exchange := assetExchangeResult.ToCommon()
				for _, tradingPairResult := range allOrderBooks {
					if assetExchangeResult.ExchangeID == tradingPairResult.ExchangeID {
						exchange.TradingPairs = append(exchange.TradingPairs, tradingPairResult.ToCommon())
					}
				}
				result.Exchanges = append(result.Exchanges, exchange)
			}
		}

		results = append(results, result)
	}

	return results, nil
}

func (s *Storage) GetAsset(id uint64) (common.Asset, error) {
	var (
		assetDBResult        assetDB
		assetExchangeResults []assetExchangeDB
		tradingPairResults   []tradingPairDB
		exchanges            []common.AssetExchange
	)

	tx, err := s.db.Beginx()
	if err != nil {
		return common.Asset{}, err
	}
	defer rollbackUnlessCommitted(tx)

	if err := tx.Stmtx(s.stmts.getAssetExchange).Select(&assetExchangeResults, id); err != nil {
		return common.Asset{}, fmt.Errorf("failed to query asset exchanges err=%s", err.Error())
	}

	for _, ae := range assetExchangeResults {
		exchanges = append(exchanges, ae.ToCommon())
	}

	if err := tx.Stmtx(s.stmts.getTradingPair).Select(&tradingPairResults, id); err != nil {
		return common.Asset{}, fmt.Errorf("failed to query for trading pairs err=%s", err.Error())
	}

	for _, pair := range tradingPairResults {
		for i := range exchanges {
			if exchanges[i].ExchangeID == pair.ExchangeID {
				exchanges[i].TradingPairs = append(exchanges[i].TradingPairs, pair.ToCommon())
			}
		}
	}

	log.Printf("getting asset id=%d", id)
	err = tx.Stmtx(s.stmts.getAsset).Get(&assetDBResult, id)
	switch err {
	case sql.ErrNoRows:
		log.Printf("asset not found id=%d", id)
		return common.Asset{}, common.ErrNotFound
	case nil:
		result, err := assetDBResult.ToCommon()
		if err != nil {
			return common.Asset{}, fmt.Errorf("invalid database record for asset id=%d err=%s", assetDBResult.ID, err.Error())
		}
		result.Exchanges = exchanges
		return result, nil
	default:
		return common.Asset{}, fmt.Errorf("failed to get asset from database id=%d err=%s", id, err.Error())
	}
}

func (s *Storage) UpdateAsset(id uint64, opts ...storage.UpdateAssetOption) error {
	var (
		updateOpts   = &storage.UpdateAssetOpts{}
		addressParam *string
		setRatePram  *string
	)

	if len(opts) == 0 {
		log.Printf("no update option is provided, doing nothing")
		return nil
	}

	for _, opt := range opts {
		opt(updateOpts)
	}

	var updateMsgs []string
	if updateOpts.Symbol() != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("symbol=%s", *updateOpts.Symbol()))
	}
	if updateOpts.Name() != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("name=%s", *updateOpts.Name()))
	}
	if updateOpts.Address() != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("address=%s", updateOpts.Address().String()))
		addressStr := updateOpts.Address().String()
		addressParam = &addressStr
	}
	if updateOpts.Decimals() != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("decimals=%d", *updateOpts.Decimals()))
	}
	if updateOpts.SetRate() != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("set_rate=%s", updateOpts.SetRate().String()))
		setRateStr := updateOpts.SetRate().String()
		setRatePram = &setRateStr
	}
	if updateOpts.Rebalance() != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("rebalance=%t", *updateOpts.Rebalance()))
	}
	if updateOpts.IsQuote() != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("is_quote=%t", *updateOpts.IsQuote()))
	}

	log.Printf("updating asset %d %s", id, strings.Join(updateMsgs, " "))
	var updatedID uint64
	err := s.stmts.updateAsset.Get(&updatedID,
		struct {
			ID        uint64  `db:"id"`
			Symbol    *string `db:"symbol"`
			Name      *string `db:"name"`
			Address   *string `db:"address"`
			Decimals  *uint64 `db:"decimals"`
			SetRate   *string `db:"set_rate"`
			Rebalance *bool   `db:"rebalance"`
			IsQuote   *bool   `db:"is_quote"`
		}{
			ID:        id,
			Symbol:    updateOpts.Symbol(),
			Name:      updateOpts.Name(),
			Address:   addressParam,
			Decimals:  updateOpts.Decimals(),
			SetRate:   setRatePram,
			Rebalance: updateOpts.Rebalance(),
			IsQuote:   updateOpts.IsQuote(),
		},
	)
	if err == sql.ErrNoRows {
		log.Printf("asset not found in database id=%d", id)
		return common.ErrNotFound
	} else if err != nil {
		pErr, ok := err.(*pq.Error)
		if !ok {
			return fmt.Errorf("unknown error returned err=%s", err.Error())
		}

		if pErr.Code == errCodeUniqueViolation {
			switch pErr.Constraint {
			case "assets_symbol_key":
				log.Printf("conflict symbol when updating asset id=%d err=%s", id, pErr.Message)
				return common.ErrSymbolExists
			case addressesUniqueConstraint:
				log.Printf("conflict address when updating asset id=%d err=%s", id, pErr.Message)
				return common.ErrAddressExists
			}
		}

		return fmt.Errorf("failed to update asset err=%s", pErr)
	}
	return nil
}

func (s *Storage) ChangeAssetAddress(id uint64, address ethereum.Address) error {
	log.Printf("changing address of asset id=%d new_address=%s", id, address.String())

	_, err := s.stmts.changeAssetAddress.Exec(id, address.String())
	if err != nil {
		pErr, ok := err.(*pq.Error)
		if !ok {
			return fmt.Errorf("unknown error returned err=%s", err.Error())
		}

		switch pErr.Code {
		case errCodeUniqueViolation:
			if pErr.Constraint == addressesUniqueConstraint {
				log.Printf("conflict address when changing asset address id=%d err=%s", id, pErr.Message)
				return common.ErrAddressExists
			}
		case errAssertFailure:
			log.Printf("asset not found in database id=%d err=%s", id, pErr.Message)
			return common.ErrNotFound
		}
		return fmt.Errorf("failed to update asset err=%s", pErr)
	}
	return nil
}