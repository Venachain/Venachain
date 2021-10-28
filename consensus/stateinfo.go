package consensus

import (
	"sync"

	"github.com/PlatONEnetwork/PlatONE-Go/event"
)

type StatusInfo struct {
	sync.Mutex
	ConsensusCostTime     uint64
	CurrentRequestTimeout uint64

	Ratio     float64
	TxCount   uint64
	IsTimeout bool
	Event     *event.TypeMuxSubscription
}
