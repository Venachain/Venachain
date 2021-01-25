package vm

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/byteutil"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"math/big"
	"reflect"
)

var (
	gasContractNameKey                 string = "GasContractName"
	isProduceEmptyBlockKey             string = "IsProduceEmptyBlock"
	txGasLimitKey                      string = "TxGasLimit"
	blockGasLimitKey                   string = "BlockGasLimit"
	isCheckContractDeployPermissionKey string = "IsCheckContractDeployPermission"
	isApproveDeployedContractKey       string = "IsApproveDeployedContract"
	isTxUseGasKey                      string = "IsTxUseGas"
	vrfParamsKey                       string = "VRFParams"
	isBlockUseTrieHashKey              string = "IsBlockUseTrieHash"
)

var preDefinedParamKeys = map[string]paramType{
	gasContractNameKey:                 &gasContractNameType{},
	isProduceEmptyBlockKey:             &IsProduceEmptyBlockType{},
	txGasLimitKey:                      &TxGasLimitType{},
	blockGasLimitKey:                   &BlockGasLimitType{},
	isCheckContractDeployPermissionKey: &CheckContractDeployPermissionType{},
	isApproveDeployedContractKey:       &IsApproveDeployedContractype{},
	isTxUseGasKey:                      &IsTxUseGastype{},
	vrfParamsKey:                       &VRFParamsType{},
	isBlockUseTrieHashKey:              &IsBlockUseTrieHashType{},
}

var (
	errDataTypeInvalid = errors.New("the data type invalid")
	errUnsupported= errors.New("the operation is unsupported")
)

const (
	paramTrue  uint32 = 1
	paramFalse uint32 = 0
)

const (
	TxGasLimitMinValue        uint64 = 12771596 * 100 // 12771596 大致相当于 0.012772s
	TxGasLimitMaxValue        uint64 = 2e9            // 相当于 2s
	txGasLimitDefaultValue    uint64 = 1.5e9          // 相当于 1.5s
	BlockGasLimitMinValue     uint64 = 12771596 * 100 // 12771596 大致相当于 0.012772s
	BlockGasLimitMaxValue     uint64 = 2e10           // 相当于 20s
	blockGasLimitDefaultValue uint64 = 1e10           // 相当于 10s
	failFlag                         = -1
	sucFlag                          = 0
)
const (
	doParamSetSuccess     CodeType = 0
	callerHasNoPermission CodeType = 1
	encodeFailure         CodeType = 2
	paramInvalid          CodeType = 3
	contractNameNotExists CodeType = 4
)

type ParamManager struct {
	stateDB      StateDB
	contractAddr *common.Address
	caller       common.Address
	blockNumber  *big.Int
}

type paramType interface {
	defalutVal() interface{}
	decodeAndCheck(ctx *ParamManager, b []byte) (interface{}, error)
}

type stringParamType struct{}

func (s *stringParamType) defalutVal() interface{} {
	return ""
}

func (s stringParamType) decodeAndCheck(ctx *ParamManager, b []byte) (interface{}, error) {
	val := byteutil.BytesToString(b)
	if b, _ := checkNameFormat(val); !b {
		ctx.emitNotifyEventInParam("StringParam", paramInvalid, fmt.Sprintf("param is invalid."))
		return val, errParamInvalid
	}
	return val, nil
}


// ======GasContractName=============================================================================
type gasContractNameType struct{}

func (c *gasContractNameType) defalutVal() interface{} {
	return ""
}

func (c *gasContractNameType) decodeAndCheck(ctx *ParamManager, b []byte) (interface{}, error) {
	contractName := byteutil.BytesToString(b)
	if b, _ := checkNameFormat(contractName); !b {
		ctx.emitNotifyEventInParam(gasContractNameKey, paramInvalid, fmt.Sprintf("param is invalid."))
		return contractName, errParamInvalid
	}
	res, err := getRegisterStatusByName(ctx.stateDB, contractName)
	if err != nil {
		return contractName, err
	}
	if !res {
		ctx.emitNotifyEventInParam(gasContractNameKey, contractNameNotExists, fmt.Sprintf("contract does not exsits."))
		return contractName, errContactNameNotExist
	}
	return contractName, nil
}

// ======IsProduceEmptyBlock=============================================================================
type IsProduceEmptyBlockType struct{}

func (c *IsProduceEmptyBlockType) defalutVal() interface{} {
	return paramFalse
}

func (c *IsProduceEmptyBlockType) decodeAndCheck(ctx *ParamManager, b []byte) (interface{}, error)  {
	isProduceEmptyBlock := byteutil.BytesToUint32(b)
	if isProduceEmptyBlock/2 != 0 {
		ctx.emitNotifyEventInParam(isProduceEmptyBlockKey, paramInvalid, fmt.Sprintf("param is invalid."))
		return isProduceEmptyBlock,errParamInvalid
	}
	return isProduceEmptyBlock,nil
}

// ======TxGasLimit=============================================================================
type TxGasLimitType struct{}

func (c *TxGasLimitType) defalutVal() interface{} {
	return txGasLimitDefaultValue
}

func (c *TxGasLimitType) decodeAndCheck(ctx *ParamManager, b []byte) (interface{}, error) {
	txGasLimit := byteutil.BytesToUint64(b)
	if txGasLimit < TxGasLimitMinValue || txGasLimit > TxGasLimitMaxValue {
		ctx.emitNotifyEventInParam(txGasLimitKey, paramInvalid, fmt.Sprintf("param is invalid."))
		return txGasLimit,errParamInvalid
	}
	// 获取区块 gas limit，其值应大于或等于每笔交易 gas limit 参数的值
	blockGasLimit, err := (&scParamManagerWrapper{ctx}).getBlockGasLimit()
	if err != nil && err != errEmptyValue {
		return txGasLimit,err
	}
	if txGasLimit > blockGasLimit {
		ctx.emitNotifyEventInParam(txGasLimitKey, paramInvalid, fmt.Sprintf("setting value is larger than block gas limit"))
		return txGasLimit,errParamInvalid
	}
	return txGasLimit,nil
}

// ======BlockGasLimit=============================================================================
type BlockGasLimitType struct{}

func (c *BlockGasLimitType) defalutVal() interface{} {
	return blockGasLimitDefaultValue
}

func (c *BlockGasLimitType) decodeAndCheck(ctx *ParamManager, b []byte) (interface{}, error) {
	blockGasLimit := byteutil.BytesToUint64(b)
	if blockGasLimit < BlockGasLimitMinValue || blockGasLimit > BlockGasLimitMaxValue {
		ctx.emitNotifyEventInParam(blockGasLimitKey, paramInvalid, fmt.Sprintf("param is invalid."))
		return blockGasLimit,errParamInvalid
	}

	txGasLimit, err := (&scParamManagerWrapper{ctx}).getTxGasLimit()
	if err != nil && err != errEmptyValue {
		return blockGasLimit,err
	}
	if txGasLimit > blockGasLimit {
		ctx.emitNotifyEventInParam(blockGasLimitKey, paramInvalid, fmt.Sprintf("setting value is smaller than tx gas limit"))
		return blockGasLimit,nil
	}
	return blockGasLimit,nil
}

// ======CheckContractDeployPermission=============================================================================
type CheckContractDeployPermissionType struct{}

func (c *CheckContractDeployPermissionType) defalutVal() interface{} {
	return paramFalse
}

func (c *CheckContractDeployPermissionType) decodeAndCheck(ctx *ParamManager, b []byte) (interface{}, error)  {
	permission := byteutil.BytesToUint32(b)
	if permission/2 != 0 {
		ctx.emitNotifyEventInParam(isCheckContractDeployPermissionKey, paramInvalid, fmt.Sprintf("param is invalid."))
		return permission,errParamInvalid
	}
	return permission,nil
}

// ======IsApproveDeployedContract=============================================================================
type IsApproveDeployedContractype struct{}

func (c *IsApproveDeployedContractype) defalutVal() interface{} {
	return paramFalse
}

func (c *IsApproveDeployedContractype) decodeAndCheck(ctx *ParamManager, b []byte) (interface{}, error) {
	isApproveDeployedContract := byteutil.BytesToUint32(b)
	if isApproveDeployedContract/2 != 0 {
		ctx.emitNotifyEventInParam(isApproveDeployedContractKey, paramInvalid, fmt.Sprintf("param is invalid."))
		return isApproveDeployedContract,errParamInvalid
	}
	return isApproveDeployedContract,nil
}

// ======IsTxUseGas=============================================================================
type IsTxUseGastype struct{}

func (c *IsTxUseGastype) defalutVal() interface{} {
	return paramFalse
}

func (c *IsTxUseGastype) decodeAndCheck(ctx *ParamManager, b []byte) (interface{}, error) {
	isTxUseGas := byteutil.BytesToUint32(b)
	if isTxUseGas/2 != 0 {
		ctx.emitNotifyEventInParam(isTxUseGasKey, paramInvalid, fmt.Sprintf("param is invalid."))
		return isTxUseGas,errParamInvalid
	}
	return isTxUseGas,nil
}

// ======VRFParams=============================================================================
type VRFParamsType struct{}

func (c *VRFParamsType) defalutVal() interface{} {
	return common.VRFParams{
		ElectionEpoch:     0,
		NextElectionBlock: 0,
		ValidatorCount:    0,
	}
}

func (c *VRFParamsType) decodeAndCheck(ctx *ParamManager, b []byte) (interface{}, error) {
	var params common.VRFParams
	if err := json.Unmarshal(b, &params); nil != err {
		return params,err
	}

	if params.ValidatorCount < 1 {
		ctx.emitNotifyEventInParam(vrfParamsKey, paramInvalid, errValidatorCountInvalid.Error())
		return params,errValidatorCountInvalid
	}
	return params,nil
}

// ======IsBlockUseTrieHash=============================================================================
type IsBlockUseTrieHashType struct{}

func (c *IsBlockUseTrieHashType) defalutVal() interface{} {
	return paramTrue
}

func (c *IsBlockUseTrieHashType) decodeAndCheck(ctx *ParamManager, b []byte) (interface{}, error){
	isBlockUseTrieHash := byteutil.BytesToUint32(b)
	if isBlockUseTrieHash/2 != 0 {
		ctx.emitNotifyEventInParam(isBlockUseTrieHashKey, paramInvalid, fmt.Sprintf("param is invalid."))
		return isBlockUseTrieHash,errParamInvalid
	}
	return isBlockUseTrieHash,nil
}

//===========================================================================
func (u *ParamManager) setParam(key string, dataInBytes []byte) (int32, error) {
	var paramType paramType
	var ok bool
	if paramType, ok = preDefinedParamKeys[key]; !ok {
		paramType = &stringParamType{}
	}

	data,err := paramType.decodeAndCheck(u,dataInBytes)
	if err != nil {
		return failFlag, err
	}

	ret, err := u.doParamSet(key, data)
	return ret, err
}
func (u *ParamManager) getParam(key string) (interface{}, error) {
	var paramType paramType
	var ok bool
	if paramType, ok = preDefinedParamKeys[key]; !ok {
		paramType = &stringParamType{}
	}

	defaultVal := paramType.defalutVal()
	defaultValPtr := structToPtr(defaultVal)

	value := u.getState(generateStateKey(key))
	if len(value) == 0 {
		return defaultVal, nil
	}
	if err := rlp.DecodeBytes(value, defaultValPtr); err != nil {
		return defaultVal, err
	}

	defaultVal = ptrToStruct(defaultValPtr)
	return defaultVal, nil
}

func (u *ParamManager) doParamSet(key string, value interface{}) (int32, error) {
	if !hasParamOpPermission(u.stateDB, u.caller) {
		u.emitNotifyEventInParam(key, callerHasNoPermission, fmt.Sprintf("%s has no permission to adjust param.", u.caller.String()))
		return failFlag, errNoPermission
	}

	keyInBytes := generateStateKey(key)
	valueInBytes, err := rlp.EncodeToBytes(value)
	if err != nil {
		u.emitNotifyEventInParam(key, encodeFailure, fmt.Sprintf("%v failed to encode.", keyInBytes))
		return failFlag, errEncodeFailure
	}

	u.setState(keyInBytes, valueInBytes)
	u.emitNotifyEventInParam(key, doParamSetSuccess, fmt.Sprintf("param set successful."))
	return sucFlag, nil
}

func (u *ParamManager) emitNotifyEventInParam(topic string, code CodeType, msg string) {
	emitEvent(*u.contractAddr, u.stateDB, u.blockNumber.Uint64(), topic, code, msg)
}

func (u *ParamManager) setState(key, value []byte) {
	u.stateDB.SetState(*u.contractAddr, key, value)
}

func (u *ParamManager) getState(key []byte) []byte {
	return u.stateDB.GetState(*u.contractAddr, key)
}

func encode(i interface{}) ([]byte, error) {
	return rlp.EncodeToBytes(i)
}

func structToPtr(obj interface{}) interface{} {
	vp := reflect.New(reflect.TypeOf(obj))
	vp.Elem().Set(reflect.ValueOf(obj))
	return vp.Interface()
}

func ptrToStruct(obj interface{}) interface{} {
	return reflect.Indirect(reflect.ValueOf(obj)).Interface()
}
