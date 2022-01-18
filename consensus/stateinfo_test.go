package consensus

import (
	"testing"

	"github.com/Venachain/Venachain/event"
)

func TestStoreCostInfo(t *testing.T) {
	costInfo := &CostInfo{
		BlockNum:            5,
		ConsensusCostTime:   6000,
		CurrentBlockTxCount: 5000,
		IsTimeout:           false,
		IsUsed:              false,
	}
	statusInfo := StatusInfo{}
	statusInfo.StoreCostInfo(costInfo)
	res, _ := statusInfo.LoadCostInfo()
	if res != costInfo {
		t.Errorf("res should equal costinfo")
	}
}

func TestStoreEvent(t *testing.T) {
	event := &event.TypeMuxSubscription{}
	statusInfo := StatusInfo{}
	statusInfo.StoreEvent(event)
	res, _ := statusInfo.LoadEvent()
	if res != event {
		t.Errorf("res should equal event")
	}
}
