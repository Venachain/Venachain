package consensus

import (
	"sync"

	"github.com/Venachain/Venachain/event"
)

type StatusInfo struct {
	sync.Mutex
	ConsensusCostTime     uint64
	CurrentRequestTimeout uint64

	Ratio               float64
	CurrentBlockTxCount uint64
	IsTimeout           bool
	Event               *event.TypeMuxSubscription
}
