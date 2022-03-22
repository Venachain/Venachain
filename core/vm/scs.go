package vm

import (
	"fmt"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/common/syscontracts"
	"github.com/Venachain/Venachain/log"
)

//system contract export functions
type (
	SCExportFn  interface{}
	SCExportFns map[string]SCExportFn //map[function name]function pointer
)

var VenachainPrecompiledContracts = map[common.Address]PrecompiledContract{
	syscontracts.UserManagementAddress:        &UserManagement{},
	syscontracts.NodeManagementAddress:        &scNodeWrapper{},
	syscontracts.CnsManagementAddress:         &CnsWrapper{},
	syscontracts.ParameterManagementAddress:   &scParamManagerWrapper{},
	syscontracts.FirewallManagementAddress:    &FwWrapper{},
	syscontracts.GroupManagementAddress:       &GroupManagement{},
	syscontracts.ContractDataProcessorAddress: &ContractDataProcessor{},
	syscontracts.GroupManagementAddress:       &GroupManagement{},
	syscontracts.CnsInvokeAddress:             &CnsInvoke{},
	syscontracts.EvidenceManagementAddress:    &SCEvidenceWrapper{},
	syscontracts.BulletProofAddress:           &SCBulletProofWrapper{},
	syscontracts.PaillierAddress:              &SCPaillierWrapper{},
}

func RunVenachainPrecompiledSC(p PrecompiledContract, input []byte, contract *Contract, evm *EVM) (ret []byte, err error) {
	defer func() {
		if err := recover(); nil != err {
			log.Error("failed to run precompiled system contract", "err", err)
			ret, err = nil, fmt.Errorf("failed to run precompiled system contract,err:%v", err)
		}
	}()

	gas := p.RequiredGas(input)

	if contract.UseGas(gas) {
		switch p.(type) {
		case *UserManagement:
			um := &UserManagement{
				stateDB:      evm.StateDB,
				caller:       contract.Caller(),
				contractAddr: syscontracts.UserManagementAddress,
				blockNumber:  evm.BlockNumber,
			}
			return um.Run(input)
		case *scNodeWrapper:
			node := newSCNodeWrapper(evm.StateDB)
			node.base.caller = evm.Origin
			node.base.blockNumber = evm.BlockNumber
			node.base.contractAddr = *contract.CodeAddr

			return node.Run(input)
		case *CnsWrapper:
			cns := newCnsManager(evm.StateDB)
			cns.caller = contract.CallerAddress
			cns.origin = evm.Origin
			cns.isInit = evm.InitEntryID
			cns.blockNumber = evm.BlockNumber

			cnsWrap := new(CnsWrapper)
			cnsWrap.base = cns

			return cnsWrap.Run(input)
		case *scParamManagerWrapper:
			p := newSCParamManagerWrapper(evm.StateDB)
			p.base.contractAddr = contract.CodeAddr
			p.base.caller = evm.Context.Origin
			p.base.blockNumber = evm.BlockNumber
			return p.Run(input)
		case *FwWrapper:
			fw := new(FwWrapper)
			fw.base = NewFireWall(evm, contract)

			return fw.Run(input)
		case *GroupManagement:
			gm := &GroupManagement{
				stateDB:      evm.StateDB,
				contractAddr: contract.self.Address(),
				caller:       contract.caller.Address(),
				blockNumber:  evm.BlockNumber,
			}
			return gm.Run(input)
		case *ContractDataProcessor:
			dp := &ContractDataProcessor{
				stateDB:      evm.StateDB,
				contractAddr: contract.self.Address(),
				caller:       contract.caller.Address(),
				blockNumber:  evm.BlockNumber,
			}
			return dp.Run(input)
		case *CnsInvoke:
			ci := &CnsInvoke{
				evm:         evm,
				caller:      evm.Context.Origin,
				contract:    contract,
				blockNumber: evm.BlockNumber,
			}
			return ci.Run(input)
		case *SCEvidenceWrapper:
			ew := NewEvidenceWrapper(evm.StateDB)
			ew.base.caller = evm.Context.Origin
			ew.base.blockNumber = evm.BlockNumber
			ew.base.contractAddr = *contract.CodeAddr
			return ew.Run(input)
		case *SCBulletProofWrapper:
			bpw := NewBPWrapper(evm.StateDB)
			bpw.base.caller = evm.Context.Origin
			bpw.base.blockNumber = evm.BlockNumber
			bpw.base.contractAddr = *contract.CodeAddr
			return bpw.Run(input)
		case *SCPaillierWrapper:
			plw := NewPLWrapper(evm.StateDB)
			plw.base.caller = evm.Context.Origin
			plw.base.blockNumber = evm.BlockNumber
			plw.base.contractAddr = *contract.CodeAddr
			return plw.Run(input)
		default:
			panic("system contract handler not found")
		}
	}

	return nil, ErrOutOfGas
}
