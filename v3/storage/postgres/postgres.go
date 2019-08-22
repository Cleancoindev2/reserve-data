package postgres

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/KyberNetwork/reserve-data/common"
	v3 "github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage"
)

// Storage is an implementation of storage.Interface that use PostgreSQL as database system.
type Storage struct {
	db    *sqlx.DB
	stmts *preparedStmts
}

func (s *Storage) initExchanges() error {
	const query = `INSERT INTO "exchanges" (id, name)
VALUES (unnest($1::INT[]),
        unnest($2::TEXT[]))`

	var (
		idParams   []int
		nameParams []string
	)
	for name, ex := range common.ValidExchangeNames {
		nameParams = append(nameParams, name)
		idParams = append(idParams, int(ex))
	}

	_, err := s.db.Exec(query, pq.Array(idParams), pq.Array(nameParams))
	if err != nil {
		return err
	}

	maker := 1.0
	taker := 1.0
	disable := false

	// stable exchange is not a real exchange, we will just enable it by default with fake fee configuration
	err = s.UpdateExchange(uint64(common.StableExchange), storage.UpdateExchangeOpts{
		TradingFeeMaker: &maker,
		TradingFeeTaker: &taker,
		Disable:         &disable,
	})
	return err
}

func (s *Storage) initAssets() error {
	ethAddr := "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"
	_, err := s.stmts.newAsset.Exec(&createAssetParams{
		Symbol:       "ETH",
		Name:         "Ethereum",
		Address:      &ethAddr,
		Decimals:     18,
		Transferable: true,
		SetRate:      v3.SetRateNotSet.String(),
		Rebalance:    false,
		IsQuote:      true,
	})
	if err != nil {
		return err
	}
	dgxAddr := "0x4f3AfEC4E5a3F2A6a1A411DEF7D7dFe50eE057bF"
	_, err = s.stmts.newAsset.Exec(&createAssetParams{
		Symbol:       "DGX",
		Name:         "Digix Gold Token",
		Address:      &dgxAddr,
		Decimals:     18,
		Transferable: true,
		SetRate:      v3.SetRateNotSet.String(),
		Rebalance:    false,
		IsQuote:      true,
	})
	return err
}

// NewStorage creates a new Storage instance from given configuration.
func NewStorage(db *sqlx.DB) (*Storage, error) {
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to intialize database schema err=%s", err.Error())
	}

	stmts, err := newPreparedStmts(db)
	if err != nil {
		return nil, fmt.Errorf("failed to preprare statements err=%s", err.Error())
	}

	s := &Storage{db: db, stmts: stmts}

	exchanges, err := s.GetExchanges()
	if err != nil {
		return nil, fmt.Errorf("failed to get existing exchanges")
	}

	if len(exchanges) == 0 {
		log.Printf("database is empty, initializing exchanges and assets")

		if err = s.initExchanges(); err != nil {
			return nil, fmt.Errorf("failed to initialize exchanges err=%s", err.Error())
		}

		if err = s.initAssets(); err != nil {
			return nil, fmt.Errorf("failed to initialize assets err=%s", err.Error())
		}
	}
	return s, nil
}
