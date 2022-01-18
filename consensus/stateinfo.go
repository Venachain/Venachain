package consensus

import (
	"errors"
	"sync/atomic"

	"github.com/Venachain/Venachain/event"
)

type StatusInfo struct {
	costInfo atomic.Value
	event    atomic.Value
}

// CostInfo consensus cost time
type CostInfo struct {
	BlockNum            uint64
	ConsensusCostTime   uint64
	CurrentBlockTxCount uint64
	IsTimeout           bool
	IsUsed              bool //this costInfo whether to use
}

func (s *StatusInfo) StoreCostInfo(info *CostInfo) {
	s.costInfo.Store(info)
}

func (s *StatusInfo) LoadCostInfo() (*CostInfo, error) {
	if s.costInfo.Load() == nil {
		return nil, errors.New("consensus costInfo is nil")
	}
	return s.costInfo.Load().(*CostInfo), nil
}

func (s *StatusInfo) StoreEvent(event *event.TypeMuxSubscription) {
	s.event.Store(event)
}

func (s *StatusInfo) LoadEvent() (*event.TypeMuxSubscription, error) {
	if s.event.Load() == nil {
		return nil, errors.New("consensus statusInfo event is nil")
	}
	return s.event.Load().(*event.TypeMuxSubscription), nil
}
