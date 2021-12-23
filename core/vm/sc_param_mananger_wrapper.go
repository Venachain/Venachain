package vm

import (
	"encoding/json"
	"reflect"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/params"
)

type scParamManagerWrapper struct {
	base *ParamManager
}

func newSCParamManagerWrapper(db StateDB) *scParamManagerWrapper {
	return &scParamManagerWrapper{
		base: &ParamManager{
			stateDB: db,
		},
	}
}

func (u *scParamManagerWrapper) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.ParamManagerGas
}

func (u *scParamManagerWrapper) Run(input []byte) ([]byte, error) {
	fnName, ret, err := execSC(input, u.AllExportFns())
	if err != nil {
		if fnName == "" {
			fnName = "Notify"
		}
		u.base.emitNotifyEventInParam(fnName, operateFail, err.Error())
	}
	return ret, err
}

// Deprecated: Use setParam() instead
func (u *scParamManagerWrapper) setGasContractName(contractName string) (int32, error) {
	return u.base.setParam(GasContractNameKey, []byte(contractName))
}

// Deprecated: Use setParam() instead
func (u *scParamManagerWrapper) setIsProduceEmptyBlock(isProduceEmptyBlock uint32) (int32, error) {
	return u.base.setParam(IsProduceEmptyBlockKey, common.Uint32ToBytes(isProduceEmptyBlock))
}

// Deprecated: Use setParam() instead
func (u *scParamManagerWrapper) setTxGasLimit(txGasLimit uint64) (int32, error) {
	return u.base.setParam(TxGasLimitKey, common.Uint64ToBytes(txGasLimit))
}

// Deprecated: Use setParam() instead
func (u *scParamManagerWrapper) setBlockGasLimit(blockGasLimit uint64) (int32, error) {
	return u.base.setParam(BlockGasLimitKey, common.Uint64ToBytes(blockGasLimit))
}

// Deprecated: Use setParam() instead
// 设置是否检查合约部署权限
// 0: 不检查合约部署权限，允许任意用户部署合约  1: 检查合约部署权限，用户具有相应权限才可以部署合约
// 默认为0，不检查合约部署权限，即允许任意用户部署合约
func (u *scParamManagerWrapper) setCheckContractDeployPermission(permission uint32) (int32, error) {
	return u.base.setParam(IsCheckContractDeployPermissionKey, common.Uint32ToBytes(permission))
}

// Deprecated: Use setParam() instead
// 设置是否审核已部署的合约
// @isApproveDeployedContract:
// 1: 审核已部署的合约  0: 不审核已部署的合约
func (u *scParamManagerWrapper) setIsApproveDeployedContract(isApproveDeployedContract uint32) (int32, error) {
	return u.base.setParam(IsApproveDeployedContractKey, common.Uint32ToBytes(isApproveDeployedContract))
}

// Deprecated: Use setParam() instead
// 本参数根据最新的讨论（2019.03.06之前）不再需要，即交易需要消耗gas。但是计费相关如消耗特定合约代币的参数由 setGasContractName 进行设置
// 设置交易是否消耗 gas
// @isTxUseGas:
// 1:  交易消耗 gas  0: 交易不消耗 gas
func (u *scParamManagerWrapper) setIsTxUseGas(isTxUseGas uint32) (int32, error) {
	return u.base.setParam(IsTxUseGasKey, common.Uint32ToBytes(isTxUseGas))
}

// Deprecated: Use setParam() instead
func (u *scParamManagerWrapper) setVRFParams(params string) (int32, error) {
	return u.base.setParam(VrfParamsKey, []byte(params))
}

// Deprecated: Use setParam() instead
// 1:  header 使用trie hash  // 0:
func (u *scParamManagerWrapper) setIsBlockUseTrieHash(isBlockUseTrieHash uint32) (int32, error) {
	return u.base.setParam(IsBlockUseTrieHashKey, common.Uint32ToBytes(isBlockUseTrieHash))
}

func (u *scParamManagerWrapper) setIntParam(key string, value uint64) (int32, error) {
	if _, ok := preDefinedParamKeys[key]; ok {
		return u.setParam(key, common.Uint64ToBytes(value))
	}
	return failFlag, errUnsupported
}

func (u *scParamManagerWrapper) setStrParam(key string, value string) (int32, error) {
	return u.base.setParam(key, []byte(value))
}

func (u *scParamManagerWrapper) setParam(key string, b []byte) (int32, error) {
	return u.base.setParam(key, b)
}

//===================================================================================
// Deprecated: Use getParam() instead
func (u *scParamManagerWrapper) getGasContractName() (string, error) {
	data, err := u.base.getParam(GasContractNameKey)
	return data.(string), err
}

// Deprecated: Use getParam() instead
func (u *scParamManagerWrapper) getIsProduceEmptyBlock() (uint32, error) {
	data, err := u.base.getParam(IsProduceEmptyBlockKey)
	return data.(uint32), err
}

// Deprecated: Use getParam() instead
func (u *scParamManagerWrapper) getTxGasLimit() (uint64, error) {
	data, err := u.base.getParam(TxGasLimitKey)
	return data.(uint64), err
}

// Deprecated: Use getParam() instead
func (u *scParamManagerWrapper) getBlockGasLimit() (uint64, error) {
	data, err := u.base.getParam(BlockGasLimitKey)
	return data.(uint64), err
}

// Deprecated: Use getParam() instead
// 获取是否是否检查合约部署权限
// 0: 不检查合约部署权限，允许任意用户部署合约  1: 检查合约部署权限，用户具有相应权限才可以部署合约
// 默认为0，不检查合约部署权限，即允许任意用户部署合约
func (u *scParamManagerWrapper) getCheckContractDeployPermission() (uint32, error) {
	data, err := u.base.getParam(IsCheckContractDeployPermissionKey)
	return data.(uint32), err
}

// Deprecated: Use getParam() instead
// 获取是否审核已部署的合约的标志
func (u *scParamManagerWrapper) getIsApproveDeployedContract() (uint32, error) {
	data, err := u.base.getParam(IsApproveDeployedContractKey)
	return data.(uint32), err
}

// Deprecated: Use getParam() instead
// 获取交易是否消耗 gas
func (u *scParamManagerWrapper) getIsTxUseGas() (uint32, error) {
	data, err := u.base.getParam(IsTxUseGasKey)
	return data.(uint32), err
}

// Deprecated: Use getParam() instead
func (u *scParamManagerWrapper) getVRFParams() (common.VRFParams, error) {
	data, err := u.base.getParam(VrfParamsKey)
	return data.(common.VRFParams), err
}

// Deprecated: Use getParam() instead
// 获取header是否使用trie hash
func (u *scParamManagerWrapper) getIsBlockUseTrieHash() (uint32, error) {
	data, err := u.base.getParam(IsBlockUseTrieHashKey)
	return data.(uint32), err
}

func (u *scParamManagerWrapper) getIntParam(key string) (uint64, error) {
	data, err := u.base.getParam(key)
	if err != nil {
		return 0, err
	}
	if reflect.TypeOf(data).Kind() == reflect.Uint64 {
		return data.(uint64), nil
	}
	if reflect.TypeOf(data).Kind() == reflect.Uint32 {
		return uint64(data.(uint32)), nil
	}

	return 0, errDataTypeInvalid
}

func (u *scParamManagerWrapper) getStrParam(key string) (string, error) {
	data, err := u.base.getParam(key)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(data)
	return string(b), err
}

func (u *scParamManagerWrapper) getParam(key string) (interface{}, error) {
	return u.base.getParam(key)
}

//for access control
func (u *scParamManagerWrapper) AllExportFns() SCExportFns {
	return SCExportFns{
		// Deprecated: Use getParam()/setParam() instead
		"setGasContractName":               u.setGasContractName,
		"getGasContractName":               u.getGasContractName,
		"setIsProduceEmptyBlock":           u.setIsProduceEmptyBlock,
		"getIsProduceEmptyBlock":           u.getIsProduceEmptyBlock,
		"setTxGasLimit":                    u.setTxGasLimit,
		"getTxGasLimit":                    u.getTxGasLimit,
		"setBlockGasLimit":                 u.setBlockGasLimit,
		"getBlockGasLimit":                 u.getBlockGasLimit,
		"setCheckContractDeployPermission": u.setCheckContractDeployPermission,
		"getCheckContractDeployPermission": u.getCheckContractDeployPermission,
		"setIsApproveDeployedContract":     u.setIsApproveDeployedContract,
		"getIsApproveDeployedContract":     u.getIsApproveDeployedContract,
		"setIsTxUseGas":                    u.setIsTxUseGas,
		"getIsTxUseGas":                    u.getIsTxUseGas,
		"setVRFParams":                     u.setVRFParams,
		"getVRFParams":                     u.getVRFParams,
		"setIsBlockUseTrieHash":            u.setIsBlockUseTrieHash,
		"getIsBlockUseTrieHash":            u.getIsBlockUseTrieHash,
		"getIntParam":                      u.getIntParam,
		"setIntParam":                      u.setIntParam,
		"getStrParam":                      u.getStrParam,
		"setStrParam":                      u.setStrParam,
	}
}
