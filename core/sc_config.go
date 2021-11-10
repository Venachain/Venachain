package core

import (
	"encoding/json"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"
	"github.com/PlatONEnetwork/PlatONE-Go/life/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/p2p"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

func UpdateParamSysContractConfig(bc *BlockChain, sysContractConf *common.SystemConfig) {
	paramAddr := syscontracts.ParameterManagementAddress

	funcName := "getTxGasLimit"
	funcParams := []interface{}{}
	res, err := InnerCallContractReadOnly(bc, paramAddr, funcName, funcParams)
	if res != nil && nil == err {
		ret := common.CallResAsInt64(res)
		if ret > 0 {
			sysContractConf.SysParam.TxGasLimit = ret
		}
	}

	funcName = "getBlockGasLimit"
	funcParams = []interface{}{}
	res, err = InnerCallContractReadOnly(bc, paramAddr, funcName, funcParams)
	if res != nil && nil == err {
		ret := common.CallResAsInt64(res)
		if ret > 0 {
			sysContractConf.SysParam.BlockGasLimit = ret
		}
	}

	funcName = "getCheckContractDeployPermission"
	funcParams = []interface{}{}
	res, err = InnerCallContractReadOnly(bc, paramAddr, funcName, funcParams)
	if res != nil && nil == err {
		ret := common.CallResAsInt64(res)
		sysContractConf.SysParam.CheckContractDeployPermission = ret
	}

	funcName = "getIsProduceEmptyBlock"
	funcParams = []interface{}{}
	res, err = InnerCallContractReadOnly(bc, paramAddr, funcName, funcParams)
	if res != nil && nil == err {
		ret := common.CallResAsInt64(res)
		sysContractConf.SysParam.IsProduceEmptyBlock = ret == 1
	}

	funcName = "getIsTxUseGas"
	funcParams = []interface{}{}
	res, err = InnerCallContractReadOnly(bc, paramAddr, funcName, funcParams)
	if res != nil && nil == err {
		ret := common.CallResAsInt64(res)
		sysContractConf.SysParam.IsTxUseGas = ret == 1
	}

	funcName = "getIsBlockUseTrieHash"
	funcParams = []interface{}{}
	res, err = InnerCallContractReadOnly(bc, paramAddr, funcName, funcParams)
	if res != nil && nil == err {
		ret := common.CallResAsInt64(res)
		sysContractConf.SysParam.IsBlockUseTrieHash = ret == 1
	}

	funcName = "getIntParam"
	funcParams = []interface{}{vm.IsUseDAG}
	res, err = InnerCallContractReadOnly(bc, paramAddr, funcName, funcParams)
	if res != nil && nil == err {
		ret := common.CallResAsInt64(res)
		sysContractConf.SysParam.IsUseDAG = ret == 1
	}

	funcName = "getVRFParams"
	funcParams = []interface{}{}
	res, err = InnerCallContractReadOnly(bc, paramAddr, funcName, funcParams)
	if res != nil && nil == err {
		strRes := common.CallResAsString(res)
		var tmpVrfParam common.VRFParams
		if err := json.Unmarshal(utils.String2bytes(strRes), &tmpVrfParam); err != nil {
			log.Warn("unmarshal vrf params failed", "result", strRes, "err", err.Error())
		} else {
			sysContractConf.SysParam.VRF.ElectionEpoch = tmpVrfParam.ElectionEpoch
			sysContractConf.SysParam.VRF.NextElectionBlock = tmpVrfParam.NextElectionBlock
			sysContractConf.SysParam.VRF.ValidatorCount = tmpVrfParam.ValidatorCount
		}
	}

	funcName = "getGasContractName"
	funcParams = []interface{}{}
	res, err = InnerCallContractReadOnly(bc, paramAddr, funcName, funcParams)
	if res != nil && nil == err {
		sysContractConf.SysParam.GasContractName = common.CallResAsString(res)
	}

	if sysContractConf.SysParam.GasContractName != "" {
		cnsAddr := syscontracts.CnsManagementAddress
		funcName = "getContractAddress"
		funcParams = []interface{}{sysContractConf.SysParam.GasContractName, "latest"}
		res, err = InnerCallContractReadOnly(bc, cnsAddr, funcName, funcParams)
		if res != nil && nil == err {
			sysContractConf.SysParam.GasContractAddr = common.HexToAddress(common.CallResAsString(res))
		}
	}
}

func UpdateNodeSysContractConfig(bc *BlockChain, sysContractConf *common.SystemConfig) {
	funcName := "getAllNodes"
	funcParams := []interface{}{}
	res, err := InnerCallContractReadOnly(bc, syscontracts.NodeManagementAddress, funcName, funcParams)
	if nil != err {
		return
	}

	strRes := common.CallResAsString(res)

	var tmp common.CommonResult
	if err := json.Unmarshal(utils.String2bytes(strRes), &tmp); err != nil {
		log.Warn("unmarshal consensus node list failed", "result", strRes, "err", err.Error())
	} else if tmp.RetCode != 0 {
		log.Debug("contract inner error", "code", tmp.RetCode, "msg", tmp.RetMsg)
	} else {
		sysContractConf.SystemConfigMu.Lock()
		defer sysContractConf.SystemConfigMu.Unlock()
		sysContractConf.Nodes = tmp.Data
		sysContractConf.GenerateNodeData()
		p2p.UpdatePeer()
	}
}

func InitBlockReplayConfig(bc *BlockChain, sysContractConf *common.SystemConfig) {
	sysContractConf.SystemConfigMu.Lock()
	defer sysContractConf.SystemConfigMu.Unlock()

	sysContractConf.ReplayParam = &common.ReplayParam{
		Pivot:           0,
		OldSysContracts: make(map[common.Address]string),
		OldSuperAdmin:   common.NullAddress,
	}

	b, err := bc.Get(common.Sys_pivot_key)
	if err != nil {
		return
	}
	if err := rlp.DecodeBytes(b, &sysContractConf.ReplayParam.Pivot); err != nil {
		return
	}

	b, err = bc.Get(common.Sys_old_system_contract_key)
	if err != nil {
		return
	}

	var mb []byte

	if err := rlp.DecodeBytes(b, &mb); err != nil {
		return
	}
	if err = json.Unmarshal(mb, &sysContractConf.ReplayParam.OldSysContracts); err != nil {
		return
	}

	b, err = bc.Get(common.Sys_old_super_admin_key)
	if err != nil {
		return
	}
	if err := rlp.DecodeBytes(b, &sysContractConf.ReplayParam.OldSuperAdmin); err != nil {
		return
	}
}

func UpdateSysContractConfig(bc *BlockChain, sysContractConf *common.SystemConfig) {
	InitBlockReplayConfig(bc, sysContractConf)
	UpdateParamSysContractConfig(bc, sysContractConf)
	UpdateNodeSysContractConfig(bc, sysContractConf)
}
