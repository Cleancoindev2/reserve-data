package postgres

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pkg/errors"

	pgutil "github.com/KyberNetwork/reserve-data/common/postgres"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// GetAssetExchange return asset exchange by its id
func (s *Storage) GetAssetExchange(id uint64) (common.AssetExchange, error) {
	var (
		result assetExchangeDB
	)
	err := s.stmts.getAssetExchange.Get(&result, assetExchangeCondition{
		ID: &id,
	})
	switch err {
	case sql.ErrNoRows:
		s.l.Infow("asset exchange not found", "id", id)
		return common.AssetExchange{}, common.ErrNotFound
	case nil:

	default:
		return common.AssetExchange{}, errors.Errorf("failed to get asset exchange from database id=%d err=%s", id, err.Error())
	}

	var tradingPairResults []tradingPairDB
	if err := s.stmts.getTradingPair.Select(&tradingPairResults, result.AssetID); err != nil {
		return common.AssetExchange{}, fmt.Errorf("failed to query for trading pairs err=%s", err.Error())
	}
	assetExchange := result.ToCommon()
	for _, tpResult := range tradingPairResults {
		if tpResult.ExchangeID == result.ExchangeID {
			assetExchange.TradingPairs = append(assetExchange.TradingPairs, tpResult.ToCommon())
		}
	}
	return assetExchange, nil
}

// GetAssetExchangeBySymbol return asset by its symbol
func (s *Storage) GetAssetExchangeBySymbol(exchangeID uint64, symbol string) (common.Asset, error) {
	var (
		result common.Asset
	)

	tx, err := s.db.Beginx()
	if err != nil {
		return result, err
	}
	defer pgutil.RollbackUnlessCommitted(tx)

	s.l.Infow("getting asset exchange", "symbol", symbol, "exchange", exchangeID)
	err = tx.Stmtx(s.stmts.getAssetExchangeBySymbol).Get(&result, exchangeID, symbol)
	switch err {
	case sql.ErrNoRows:
		s.l.Infow("asset not found", "symbol", symbol)
		return result, common.ErrNotFound
	case nil:
		return result, nil
	default:
		return result, fmt.Errorf("failed to get asset from database symbol=%s err=%s", symbol, err.Error())
	}
}

func (s *Storage) deleteAssetExchange(tx *sqlx.Tx, assetExchangeID uint64) error {
	var returnedID uint64
	err := tx.Stmtx(s.stmts.deleteAssetExchange.Stmt).Get(&returnedID, assetExchangeID)
	switch err {
	case nil:
		return nil
	case sql.ErrNoRows:
		return common.ErrNotFound
	default:
		pErr, ok := err.(*pq.Error)
		if !ok {
			return errors.Errorf("unknown returned err=%s", err.Error())
		}
		if pErr.Code == errRestrictViolation {
			return common.ErrAssetExchangeDeleteViolation
		}
		return err
	}
}
