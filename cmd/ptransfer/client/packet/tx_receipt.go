package packet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/core/types"

	precompile "github.com/PlatONEnetwork/PlatONE-Go/cmd/ptransfer/client/precompiled"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ptransfer/client/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

var (
	txReceiptSuccessCode = hexutil.EncodeUint64(types.ReceiptStatusSuccessful)
	txReceiptFailureCode = hexutil.EncodeUint64(types.ReceiptStatusFailed)
)

const (
	TxReceiptSuccessMsg = "Operation Succeeded"
	TxReceiptFailureMsg = "Operation Failed"
)

func receiptStatusReturn(status string) (result string) {

	switch status {
	case txReceiptSuccessCode:
		result = TxReceiptSuccessMsg
	case txReceiptFailureCode:
		result = TxReceiptFailureMsg
	default:
		result = "undefined status. Something wrong"
	}

	return
}

// Receipt, eth_getTransactionReceipt return data struct
type Receipt struct {
	BlockHash         string    `json:"blockHash"`         // hash of the block
	BlockNumber       string    `json:"blockNumber"`       // height of the block
	ContractAddress   string    `json:"contractAddress"`   // contract address of the contract deployment. otherwise null
	CumulativeGasUsed string    `json:"cumulativeGasUsed"` //
	From              string    `json:"from"`              // the account address used to send the transaction
	GasUsed           string    `json:"gasUsed"`           // gas used by executing the transaction
	Root              string    `json:"root"`
	To                string    `json:"to"`               // the address the transaction is sent to
	TransactionHash   string    `json:"transactionHash"`  // the hash of the transaction
	TransactionIndex  string    `json:"transactionIndex"` // the index of the transaction
	Logs              RecptLogs `json:"logs"`
	Status            string    `json:"status"` // the execution status of the transaction, "0x1" for success
}

type Log struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
}

type RecptLogs []*Log

// ParseSysContractResult parsed the rpc response to Receipt object
func ParseTxReceipt(response interface{}) (*Receipt, error) {
	var receipt = &Receipt{}

	temp, _ := json.Marshal(response)
	err := json.Unmarshal(temp, receipt)
	if err != nil {
		// LogErr.Printf(ErrUnmarshalBytesFormat, "transaction receipt", err.Error())
		errStr := fmt.Sprintf(utils.ErrUnmarshalBytesFormat, "transaction receipt", err.Error())
		return nil, errors.New(errStr)
	}

	return receipt, nil
}

func getSysEventAbis(SysEventList []string) (abiBytesArr [][]byte) {
	for _, data := range SysEventList {
		p := precompile.List[data]
		abiBytes, _ := precompile.Asset(p)
		abiBytesArr = append(abiBytesArr, abiBytes)
	}

	return
}

type eventParsingFunc func(eLog *Log, abiBytes []byte) string

func WasmEventParsingPerLog(eLog *Log, abiBytes []byte) string {
	var rlpList []interface{}

	eventName, topicTypes := findWasmLogTopic(eLog.Topics[0], abiBytes)

	if len(topicTypes) == 0 {
		return ""
	}

	dataBytes, _ := hexutil.Decode(eLog.Data)
	err := rlp.DecodeBytes(dataBytes, &rlpList)
	if err != nil {
		// todo: error handle
		fmt.Printf("the error is %v\n", err)
	}

	result := fmt.Sprintf("Event %s: ", eventName)
	result += parseReceiptLogData(rlpList, topicTypes)

	return result
}

func EventParsing(logs RecptLogs, abiBytesArr [][]byte, fn eventParsingFunc) []string {
	var res []string

	for _, logData := range logs {
		for _, data := range abiBytesArr {
			result := fn(logData, data)
			if result != "" {
				res = append(res, result)
				break
			}
		}
	}

	return res
}

func findWasmLogTopic(topic string, abiBytes []byte) (string, []string) {
	abiFunc, err := ParseAbiFromJson(abiBytes)
	if err != nil {
		return "", nil
	}

	for _, data := range abiFunc {
		if data.Type != "event" {
			continue
		}

		if strings.EqualFold(wasmLogTopicEncode(data.Name), topic) {
			topicTypes := make([]string, 0)
			name := data.Name
			for _, v := range data.Inputs {
				topicTypes = append(topicTypes, v.Type)
			}
			return name, topicTypes
		}
	}

	return "", nil
}

func parseReceiptLogData(data []interface{}, types []string) string {
	var str string

	for i, v := range data {
		result := ConvertRlpBytesTo(v.([]uint8), types[i])
		str += fmt.Sprintf("%v ", result)
	}

	return str
}

func wasmLogTopicEncode(name string) string {
	return common.BytesToHash(crypto.Keccak256([]byte(name))).String()
}

func ConvertRlpBytesTo(input []byte, targetType string) interface{} {
	v, ok := Bytes2X_CMD[targetType]
	if !ok {
		panic("unsupported type")
	}

	return reflect.ValueOf(v).Call([]reflect.Value{reflect.ValueOf(input)})[0].Interface()
}

var Bytes2X_CMD = map[string]interface{}{
	"string": BytesToString,

	// "uint8":  RlpBytesToUint,
	"uint16": RlpBytesToUint16,
	"uint32": RlpBytesToUint32,
	"uint64": RlpBytesToUint64,

	// "uint8":  RlpBytesToUint,
	"int16": RlpBytesToUint16,
	"int32": RlpBytesToUint32,
	"int64": RlpBytesToUint64,

	"bool": RlpBytesToBool,
}

func BytesToString(b []byte) string {
	return string(b)
}

func RlpBytesToUint16(b []byte) uint16 {
	b = common.LeftPadBytes(b, 32)
	result := common.CallResAsUint32(b)
	return uint16(result)
}

func RlpBytesToUint32(b []byte) uint32 {
	b = common.LeftPadBytes(b, 32)
	return common.CallResAsUint32(b)
}

func RlpBytesToUint64(b []byte) uint64 {
	b = common.LeftPadBytes(b, 32)
	return common.CallResAsUint64(b)
}

func RlpBytesToBool(b []byte) bool {
	if bytes.Compare(b, []byte{1}) == 0 {
		return true
	}
	return false
}

/*
func RlpBytesToUintV2(b []byte) interface{} {
	var val interface{}

	for _, v := range b {
		val = val << 8
		val |= uint(v)
	}

	return val
}*/
