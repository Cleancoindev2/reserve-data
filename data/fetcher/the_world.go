package fetcher

import (
	"github.com/KyberNetwork/reserve-data/common"
)

// TheWorld is the interface that wraps all methods to get in real life
// pricing information. For now, only gold pricing is supported.
type TheWorld interface {
	GetGoldInfo() (common.GoldData, error)
	GetBTCInfo() (common.BTCData, error)
	GetUSDCInfo() (common.USDCData, error)
	GetUSDInfo() (common.USDData, error)
}
