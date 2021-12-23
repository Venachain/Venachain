package vm

import (
	"fmt"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/core/types"
	"github.com/Venachain/Venachain/crypto"
	"github.com/Venachain/Venachain/rlp"
)

func emitEvent(address common.Address, stateDB StateDB, bn uint64, topic string, params ...interface{}) {
	eLog := types.Log{}
	eLog.Address = address
	eLog.Topics = []common.Hash{common.BytesToHash(crypto.Keccak256([]byte(topic)))}
	eLog.BlockNumber = bn

	bin, err := rlp.EncodeToBytes(params)
	if nil != err {
		panic(fmt.Sprintf("failed to emit event,address:%s,bn:%d,topic:%s,params:%#v", address, bn, topic, params))
	}

	eLog.Data = bin

	stateDB.AddLog(&eLog)
}
