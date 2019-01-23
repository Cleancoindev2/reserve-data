package data

import (
	"github.com/KyberNetwork/reserve-data/common"
)

// GlobalStorage is the interfaces that wraps database operations of real world
// pricing information of ReserveData.
type GlobalStorage interface {
	GetGoldInfo(version common.Version) (common.GoldData, error)
	CurrentGoldInfoVersion(timepoint uint64) (common.Version, error)

	GetBTCInfo(version common.Version) (common.BTCData, error)
	CurrentBTCInfoVersion(timepoint uint64) (common.Version, error)
	UpdateFeedConfiguration(string, bool) error
	GetFeedConfiguration() ([]common.FeedConfiguration, error)

	UpdateFetcherConfiguration(common.FetcherConfigurationRequest) error
	GetAllFetcherConfiguration() (common.FetcherConfigurationRequest, error)
	GetFetcherConfiguration(token string) (bool, error)
}
