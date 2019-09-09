package fetcher

import (
	"github.com/KyberNetwork/reserve-data/common"
)

// GlobalStorage is the storage to store real world data pricing information.
type GlobalStorage interface {
	StoreGoldInfo(data common.GoldData) error
	StoreBTCInfo(data common.BTCData) error
	StoreUSDCInfo(data common.USDCData) error
	StoreUSDInfo(data common.USDData) error

	GetAllFetcherConfiguration() (common.FetcherConfiguration, error)
}
