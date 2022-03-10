package backend

import (
	"encoding/json"
	"math/big"
	"reflect"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/common/syscontracts"
	"github.com/Venachain/Venachain/consensus"
	"github.com/Venachain/Venachain/core"
	"github.com/Venachain/Venachain/core/state"
	"github.com/Venachain/Venachain/core/types"
	"github.com/Venachain/Venachain/core/vm"
	"github.com/Venachain/Venachain/life/utils"
	"github.com/Venachain/Venachain/log"
	"github.com/Venachain/Venachain/p2p/discover"
)

type ChainContext struct {
	// Engine retrieves the chain's consensus engine.
	chain *consensus.ChainReader

	engine consensus.Engine
}

func (cc *ChainContext) GetHeader(hash common.Hash, number uint64) *types.Header {
	return cc.GetHeader(hash, number)
}

func (cc *ChainContext) Engine() consensus.Engine {
	return cc.engine
}

// getInitialNodesList catch initial nodes List from paramManager contract when
// new a dpos and miner a new block
func getConsensusNodesList(chain consensus.ChainReader, sb *backend, number uint64) ([]discover.NodeID, error) {
	var tmp []common.NodeInfo
	tmpVrfParam := getVRFParamsAtNumber(chain, sb, number)
	if tmpVrfParam.ElectionEpoch != 0 {
		// vrf feature is active
		tmp = getVrfConsensusNodesAtNumber(chain, sb, number)
	} else {
		tmp = getCandidateNodesAtNumber(chain, sb, number)
	}

	nodeIDs := make([]discover.NodeID, 0, len(tmp))
	for _, dataObj := range tmp {
		if pubKey := dataObj.PublicKey; len(pubKey) > 0 {
			log.Debug("Consensus node", "PublicKey", pubKey)
			if nodeID, err := discover.HexID(pubKey); err == nil {
				nodeIDs = append(nodeIDs, nodeID)
			}
		}
	}
	return nodeIDs, nil
}

func getVRFParamsAtNumber(chain consensus.ChainReader, sb *backend, number uint64) *common.VRFParams {
	isOldBlock := number < chain.CurrentHeader().Number.Uint64()
	if !isOldBlock {
		return &common.SysCfg.SysParam.VRF
	}

	resVRF := CallSystemContractAtBlockNumber(chain, sb, number, syscontracts.ParameterManagementAddress, "getVRFParams", []interface{}{})
	vrf := ParseResultToExtractType(resVRF, common.VRFParams{})
	if vrf != nil {
		return vrf.(*common.VRFParams)
	}
	return &common.VRFParams{}
}

func getVrfConsensusNodesAtNumber(chain consensus.ChainReader, sb *backend, number uint64) []common.NodeInfo {
	resVrfConsensusNodes := CallSystemContractAtBlockNumber(chain, sb, number, syscontracts.NodeManagementAddress, "getVrfConsensusNodes", []interface{}{})
	nodes := ParseResultToExtractType(resVrfConsensusNodes, common.CommonResult{})
	if nodes != nil {
		return nodes.(*common.CommonResult).Data
	}
	return []common.NodeInfo{}
}

func getCandidateNodesAtNumber(chain consensus.ChainReader, sb *backend, number uint64) []common.NodeInfo {
	isOldBlock := number < chain.CurrentHeader().Number.Uint64()
	nodes := make([]common.NodeInfo, 0)
	if isOldBlock {
		resNodes := CallSystemContractAtBlockNumber(chain, sb, number, syscontracts.NodeManagementAddress, "getAllNodes", []interface{}{})
		tmp := ParseResultToExtractType(resNodes, common.CommonResult{})
		if tmp != nil {
			nodes = tmp.(*common.CommonResult).Data
		}
	}

	return common.SysCfg.GetConsensusNodesFilterDelay(number, nodes, isOldBlock)
}

func ParseResultToExtractType(res []byte, v interface{}) interface{} {
	t := reflect.TypeOf(v)
	d := reflect.New(t).Interface()
	strRes := common.CallResAsString(res)
	if err := json.Unmarshal(utils.String2bytes(strRes), &d); err != nil {
		log.Warn("ParseResultToNodeInfos: unmarshal consensus node list failed", "result", strRes, "err", err.Error())
		return nil
	}
	return d
}

func CallSystemContractAtBlockNumber(
	chain consensus.ChainReader,
	sb *backend,
	number uint64,
	sysContractAddr common.Address,
	sysFuncName string,
	sysFuncParams []interface{},
) []byte {
	_state, _ := state.New(chain.GetHeaderByNumber(number).Root, state.NewDatabase(sb.db))
	if _state == nil {
		log.Warn("load state fail at block number", "number", number)
		return nil
	}
	msg := types.NewMessage(common.Address{}, nil, 1, big.NewInt(1), 0x1, big.NewInt(1), nil, false)
	cc := ChainContext{&chain, sb}
	context := core.NewEVMContext(msg, chain.CurrentHeader(), &cc, nil)
	evm := vm.NewEVM(context, _state, chain.Config(), vm.Config{})
	callData := common.GenCallData(sysFuncName, sysFuncParams)
	res, _, err := evm.Call(vm.AccountRef(common.Address{}), sysContractAddr, callData, uint64(0xffffffffff), big.NewInt(0))
	if err != nil {
		return nil
	}
	return res
}
