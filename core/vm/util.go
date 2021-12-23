package vm

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/common/math"
	"github.com/Venachain/Venachain/life/utils"
	"github.com/Venachain/Venachain/log"
	"github.com/Venachain/Venachain/rlp"
)

var PermissionErr = errors.New("Permission Denied!")

func toContractReturnValueIntType(txType int, res int64) []byte {
	if txType == common.CallContractFlag {
		return utils.Int64ToBytes(res)
	}

	bigRes := new(big.Int)
	bigRes.SetInt64(res)
	finalRes := utils.Align32Bytes(math.U256(bigRes).Bytes())
	return finalRes
}

func toContractReturnValueUintType(txType int, res uint64) []byte {
	if txType == common.CallContractFlag {
		return utils.Uint64ToBytes(res)
	}

	finalRes := utils.Align32Bytes(utils.Uint64ToBytes((res)))
	return finalRes
}

func toContractReturnValueStringType(txType int, res []byte) []byte {
	if txType == common.CallContractFlag || txType == common.TxTypeCallSollCompatibleWasm {
		return res
	}

	return MakeReturnBytes(res)
}

func toContractReturnValueStructType(txType int, res interface{}) []byte {
	b, err := json.Marshal(res)
	if err != nil {
		b = []byte{}
	}
	if txType == common.CallContractFlag || txType == common.TxTypeCallSollCompatibleWasm {
		return b
	}
	return MakeReturnBytes(b)
}

func MakeReturnBytes(ret []byte) []byte {
	var dataRealSize = len(ret)
	if (dataRealSize % 32) != 0 {
		dataRealSize = dataRealSize + (32 - (dataRealSize % 32))
	}
	dataByt := make([]byte, dataRealSize)
	copy(dataByt[0:], ret)

	strHash := common.BytesToHash(common.Int32ToBytes(32))
	sizeHash := common.BytesToHash(common.Int64ToBytes(int64(len(ret))))

	finalData := make([]byte, 0)
	finalData = append(finalData, strHash.Bytes()...)
	finalData = append(finalData, sizeHash.Bytes()...)
	finalData = append(finalData, dataByt...)

	return finalData
}

func FwCheck(stateDb StateDB, contractAddr common.Address, caller common.Address, input []byte) ([]byte, bool) {
	return fwCheck(stateDb, contractAddr, caller, input)
}

// 合约防火墙的检查：
//  1. 如果账户结构体code字段为空，pass
//  2. 如果账户data字段为空，pass
// 	3. 黑名单优先于白名单，后续只有不在黑名单列表，同时在白名单列表里的账户才能pass
func fwCheck(stateDb StateDB, contractAddr common.Address, caller common.Address, input []byte) ([]byte, bool) {
	if stateDb.IsFwOpened(contractAddr) == false {
		return nil, true
	}

	// 如果账户结构体code字段为空或tx.data为空，pass
	if len(stateDb.GetCode(contractAddr)) == 0 || len(input) == 0 {
		return nil, true
	}

	var data [][]byte
	if err := rlp.DecodeBytes(input, &data); err != nil {
		log.Debug("FW : Input decode error")
		return MakeReturnBytes([]byte("FW : Input decode error")), false
	}
	if len(data) < 2 {
		log.Debug("FW : Missing function name")
		return MakeReturnBytes([]byte("FW : Missing function name")), false
	}
	funcName := string(data[1])

	if stateDb.GetContractCreator(contractAddr) == caller {
		return nil, true
	}

	fwStatus := stateDb.GetFwStatus(contractAddr)

	fwLog := "FW : Access to contract:" + contractAddr.String() + " by " + funcName + "is refused by firewall."

	if fwStatus.IsRejected(funcName, caller) {
		return MakeReturnBytes([]byte(fwLog)), false
	}

	if fwStatus.IsAccepted(funcName, caller) {
		return nil, true
	}

	return MakeReturnBytes([]byte(fwLog)), false
}
