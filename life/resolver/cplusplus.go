package resolver
//#cgo LDFLAGS: -L ./nizkpail/ -lnizkpail -lpthread
//#include "./nizkpail/nizkpail.h"

/*
#cgo CFLAGS: -I../softfloat/source/include
#define SOFTFLOAT_FAST_INT64
#include "softfloat.h"
#cgo CXXFLAGS: -std=c++14
#include "printqf.h"
#include "print128.h"
#cgo LDFLAGS: -L ./sig/ -lsig -lcrypto -lssl -ldl -lpthread
#include "./sig/sig.h"
*/
import "C"

import "C"
import (
	cryptoZk "github.com/PlatONEnetwork/PlatONE-Go/cmd/ptransfer/client/crypto"
	"unsafe"
)
import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	inner "github.com/PlatONEnetwork/PlatONE-Go/common/math"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/life/exec"
)

var (
	cfc  = newCfcSet()
	cgbl = newGlobalSet()
)

type CResolver struct{}

func (r *CResolver) ResolveFunc(module, field string) *exec.FunctionImport {
	df := &exec.FunctionImport{
		Execute: func(vm *exec.VirtualMachine) int64 {
			panic(fmt.Sprintf("unsupport func module:%s field:%s", module, field))
		},
		GasCost: func(vm *exec.VirtualMachine) (uint64, error) {
			panic(fmt.Sprintf("unsupport gas cost module:%s field:%s", module, field))
		},
	}

	if m, exist := cfc[module]; exist == true {
		if f, exist := m[field]; exist == true {
			return f
		} else {
			return df
		}
	} else {
		return df
	}
}

func (r *CResolver) ResolveGlobal(module, field string) int64 {
	if m, exist := cgbl[module]; exist == true {
		if g, exist := m[field]; exist == true {
			return g
		} else {
			return 0
			//panic("unknown field " + field)

		}
	} else {
		return 0
		//panic("unknown module " + module)
	}
}

func newCfcSet() map[string]map[string]*exec.FunctionImport {
	return map[string]map[string]*exec.FunctionImport{
		"env": {
			"malloc":       &exec.FunctionImport{Execute: envMalloc, GasCost: envMallocGasCost},
			"free":         &exec.FunctionImport{Execute: envFree, GasCost: envFreeGasCost},
			"calloc":       &exec.FunctionImport{Execute: envCalloc, GasCost: envCallocGasCost},
			"realloc":      &exec.FunctionImport{Execute: envRealloc, GasCost: envReallocGasCost},
			"disable_free": &exec.FunctionImport{Execute: envDisableFree, GasCost: envDisableFreeGasCost},

			"memcpy":  &exec.FunctionImport{Execute: envMemcpy, GasCost: envMemcpyGasCost},
			"memmove": &exec.FunctionImport{Execute: envMemmove, GasCost: envMemmoveGasCost},
			"memcmp":  &exec.FunctionImport{Execute: envMemcmp, GasCost: envMemcmpGasCost},
			"memset":  &exec.FunctionImport{Execute: envMemset, GasCost: envMemsetGasCost},

			"prints":     &exec.FunctionImport{Execute: envPrints, GasCost: envPrintsGasCost},
			"prints_l":   &exec.FunctionImport{Execute: envPrintsl, GasCost: envPrintslGasCost},
			"printi":     &exec.FunctionImport{Execute: envPrinti, GasCost: envPrintiGasCost},
			"printui":    &exec.FunctionImport{Execute: envPrintui, GasCost: envPrintuiGasCost},
			"printi128":  &exec.FunctionImport{Execute: envPrinti128, GasCost: envPrinti128GasCost},
			"printui128": &exec.FunctionImport{Execute: envPrintui128, GasCost: envPrintui128GasCost},
			"printsf":    &exec.FunctionImport{Execute: envPrintsf, GasCost: envPrintsfGasCost},
			"printdf":    &exec.FunctionImport{Execute: envPrintdf, GasCost: envPrintdfGasCost},
			"printqf":    &exec.FunctionImport{Execute: envPrintqf, GasCost: envPrintqfGasCost},
			"printn":     &exec.FunctionImport{Execute: envPrintn, GasCost: envPrintnGasCost},
			"printhex":   &exec.FunctionImport{Execute: envPrinthex, GasCost: envPrinthexGasCost},

			"abort": &exec.FunctionImport{Execute: envAbort, GasCost: envAbortGasCost},

			// compiler builtins
			// arithmetic long double
			"__ashlti3": &exec.FunctionImport{Execute: env__ashlti3, GasCost: env__ashlti3GasCost},
			"__ashrti3": &exec.FunctionImport{Execute: env__ashrti3, GasCost: env__ashrti3GasCost},
			"__lshlti3": &exec.FunctionImport{Execute: env__lshlti3, GasCost: env__lshlti3GasCost},
			"__lshrti3": &exec.FunctionImport{Execute: env__lshrti3, GasCost: env__lshrti3GasCost},
			"__divti3":  &exec.FunctionImport{Execute: env__divti3, GasCost: env__divti3GasCost},
			"__udivti3": &exec.FunctionImport{Execute: env__udivti3, GasCost: env__udivti3GasCost},
			"__modti3":  &exec.FunctionImport{Execute: env__modti3, GasCost: env__modti3GasCost},
			"__umodti3": &exec.FunctionImport{Execute: env__umodti3, GasCost: env__umodti3GasCost},
			"__multi3":  &exec.FunctionImport{Execute: env__multi3, GasCost: env__multi3GasCost},
			"__addtf3":  &exec.FunctionImport{Execute: env__addtf3, GasCost: env__addtf3GasCost},
			"__subtf3":  &exec.FunctionImport{Execute: env__subtf3, GasCost: env__subtf3GasCost},
			"__multf3":  &exec.FunctionImport{Execute: env__multf3, GasCost: env__multf3GasCost},
			"__divtf3":  &exec.FunctionImport{Execute: env__divtf3, GasCost: env__divtf3GasCost},

			// conversion long double
			"__floatsitf":   &exec.FunctionImport{Execute: env__floatsitf, GasCost: env__floatsitfGasCost},
			"__floatunsitf": &exec.FunctionImport{Execute: env__floatunsitf, GasCost: env__floatunsitfGasCost},
			"__floatditf":   &exec.FunctionImport{Execute: env__floatditf, GasCost: env__floatditfGasCost},
			"__floatunditf": &exec.FunctionImport{Execute: env__floatunditf, GasCost: env__floatunditfGasCost},
			"__floattidf":   &exec.FunctionImport{Execute: env__floattidf, GasCost: env__floattidfGasCost},
			"__floatuntidf": &exec.FunctionImport{Execute: env__floatuntidf, GasCost: env__floatuntidfGasCost},
			"__floatsidf":   &exec.FunctionImport{Execute: env__floatsidf, GasCost: env__floatsidfGasCost},
			"__extendsftf2": &exec.FunctionImport{Execute: env__extendsftf2, GasCost: env__extendsftf2GasCost},
			"__extenddftf2": &exec.FunctionImport{Execute: env__extenddftf2, GasCost: env__extenddftf2GasCost},
			"__fixtfti":     &exec.FunctionImport{Execute: env__fixtfti, GasCost: env__fixtftiGasCost},
			"__fixtfdi":     &exec.FunctionImport{Execute: env__fixtfdi, GasCost: env__fixtfdiGasCost},
			"__fixtfsi":     &exec.FunctionImport{Execute: env__fixtfsi, GasCost: env__fixtfsiGasCost},
			"__fixunstfti":  &exec.FunctionImport{Execute: env__fixunstfti, GasCost: env__fixunstftiGasCost},
			"__fixunstfdi":  &exec.FunctionImport{Execute: env__fixunstfdi, GasCost: env__fixunstfdiGasCost},
			"__fixunstfsi":  &exec.FunctionImport{Execute: env__fixunstfsi, GasCost: env__fixunstfsiGasCost},
			"__fixsfti":     &exec.FunctionImport{Execute: env__fixsfti, GasCost: env__fixsftiGasCost},
			"__fixdfti":     &exec.FunctionImport{Execute: env__fixdfti, GasCost: env__fixdftiGasCost},
			"__trunctfdf2":  &exec.FunctionImport{Execute: env__trunctfdf2, GasCost: env__trunctfdf2GasCost},
			"__trunctfsf2":  &exec.FunctionImport{Execute: env__trunctfsf2, GasCost: env__trunctfsf2GasCost},

			"__eqtf2":    &exec.FunctionImport{Execute: env__eqtf2, GasCost: env__eqtf2GasCost},
			"__netf2":    &exec.FunctionImport{Execute: env__netf2, GasCost: env__netf2GasCost},
			"__getf2":    &exec.FunctionImport{Execute: env__getf2, GasCost: env__getf2GasCost},
			"__gttf2":    &exec.FunctionImport{Execute: env__gttf2, GasCost: env__gttf2GasCost},
			"__lttf2":    &exec.FunctionImport{Execute: env__lttf2, GasCost: env__lttf2GasCost},
			"__letf2":    &exec.FunctionImport{Execute: env__letf2, GasCost: env__letf2GasCost},
			"__cmptf2":   &exec.FunctionImport{Execute: env__cmptf2, GasCost: env__cmptf2GasCost},
			"__unordtf2": &exec.FunctionImport{Execute: env__unordtf2, GasCost: env__unordtf2GasCost},
			"__negtf2":   &exec.FunctionImport{Execute: env__negtf2, GasCost: env__negtf2GasCost},

			// for blockchain function
			//"gasPrice":     &exec.FunctionImport{Execute: envGasPrice, GasCost: constGasFunc(compiler.GasQuickStep)},
			//"blockHash":    &exec.FunctionImport{Execute: envBlockHash, GasCost: constGasFunc(compiler.GasQuickStep)},
			//"number":       &exec.FunctionImport{Execute: envNumber, GasCost: constGasFunc(compiler.GasQuickStep)},
			//"gasLimit":     &exec.FunctionImport{Execute: envGasLimit, GasCost: constGasFunc(compiler.GasQuickStep)},
			//"timestamp":    &exec.FunctionImport{Execute: envTimestamp, GasCost: constGasFunc(compiler.GasQuickStep)},
			//"coinbase":     &exec.FunctionImport{Execute: envCoinbase, GasCost: constGasFunc(compiler.GasQuickStep)},
			//"balance":      &exec.FunctionImport{Execute: envBalance, GasCost: constGasFunc(compiler.GasQuickStep)},
			//"origin":       &exec.FunctionImport{Execute: envOrigin, GasCost: constGasFunc(compiler.GasQuickStep)},
			//"caller":       &exec.FunctionImport{Execute: envCaller, GasCost: constGasFunc(compiler.GasQuickStep)},
			//"isOwner":      &exec.FunctionImport{Execute: envIsOwner, GasCost: constGasFunc(compiler.GasQuickStep)},
			//"isFromInit":   &exec.FunctionImport{Execute: envIsFromInit, GasCost: constGasFunc(compiler.GasQuickStep)},
			//"callValue":    &exec.FunctionImport{Execute: envCallValue, GasCost: constGasFunc(compiler.GasQuickStep)},
			//"address":      &exec.FunctionImport{Execute: envAddress, GasCost: constGasFunc(compiler.GasQuickStep)},
			"gasPrice":   &exec.FunctionImport{Execute: envGasPrice, GasCost: envGasPriceGasCost},
			"blockHash":  &exec.FunctionImport{Execute: envBlockHash, GasCost: envBlockHashGasCost},
			"number":     &exec.FunctionImport{Execute: envNumber, GasCost: envNumberGasCost},
			"gasLimit":   &exec.FunctionImport{Execute: envGasLimit, GasCost: envGasLimitGasCost},
			"timestamp":  &exec.FunctionImport{Execute: envTimestamp, GasCost: envTimestampGasCost},
			"coinbase":   &exec.FunctionImport{Execute: envCoinbase, GasCost: envCoinbaseGasCost},
			"balance":    &exec.FunctionImport{Execute: envBalance, GasCost: envBalanceGasCost},
			"origin":     &exec.FunctionImport{Execute: envOrigin, GasCost: envOriginGasCost},
			"caller":     &exec.FunctionImport{Execute: envCaller, GasCost: envCallerGasCost},
			"isOwner":    &exec.FunctionImport{Execute: envIsOwner, GasCost: envIsOwnerGasCost},
			"isFromInit": &exec.FunctionImport{Execute: envIsFromInit, GasCost: envIsFromInitGasCost},
			"callValue":  &exec.FunctionImport{Execute: envCallValue, GasCost: envCallValueGasCost},
			"address":    &exec.FunctionImport{Execute: envAddress, GasCost: envAddressGasCost},

			"sha3":         &exec.FunctionImport{Execute: envSha3, GasCost: envSha3GasCost},
			"emitEvent":    &exec.FunctionImport{Execute: envEmitEvent, GasCost: envEmitEventGasCost},
			"setState":     &exec.FunctionImport{Execute: envSetState, GasCost: envSetStateGasCost},
			"getState":     &exec.FunctionImport{Execute: envGetState, GasCost: envGetStateGasCost},
			"getStateSize": &exec.FunctionImport{Execute: envGetStateSize, GasCost: envGetStateSizeGasCost},
			"ecrecover":    &exec.FunctionImport{Execute: envEcrecover, GasCost: envEcrecoverGasCost},

			// support for vc
			//Temporarily comment the following code to prepare for cross platform
			//"vc_InitGadgetEnv":          &exec.FunctionImport{Execute: envInitGadgetEnv, GasCost: envInitGadgetEnvGasCost},
			//"vc_UninitGadgetEnv":        &exec.FunctionImport{Execute: envUninitGadgetEnv, GasCost: envUninitGadgetEnvGasCost},
			//"vc_CreatePBVar":            &exec.FunctionImport{Execute: envCreatePBVarEnv, GasCost: envCreatePBVarGasCost},
			//"vc_CreateGadget":           &exec.FunctionImport{Execute: envCreateGadgetEnv, GasCost: envCreateGadgetGasCost},
			//"vc_SetVar":                 &exec.FunctionImport{Execute: envSetVarEnv, GasCost: envSetVarGasCost},
			//"vc_SetRetIndex":            &exec.FunctionImport{Execute: envSetRetIndexEnv, GasCost: envSetRetIndexGasCost},
			//"vc_GenerateWitness":        &exec.FunctionImport{Execute: envGenWitnessEnv, GasCost: envGenWitnessGasCost},
			//"vc_GenerateProofAndResult": &exec.FunctionImport{Execute: envGenProofAndResultEnv, GasCost: envGenProofAndResultGasCost},
			//"vc_Verify":                 &exec.FunctionImport{Execute: envVerifyEnv, GasCost: envVerifyGasCost},

			// supplement
			"getCallerNonce": &exec.FunctionImport{Execute: envGetCallerNonce, GasCost: envGetCallerNonceGasCost},
			"callTransfer":   &exec.FunctionImport{Execute: envCallTransfer, GasCost: envCallTransferGasCost},

			"bcwasmCall":               &exec.FunctionImport{Execute: envBCWasmCall, GasCost: envBCWasmCallGasCost},
			"bcwasmCallInt64":          &exec.FunctionImport{Execute: envBCWasmCallInt64, GasCost: envBCWasmCallInt64GasCost},
			"bcwasmCallString":         &exec.FunctionImport{Execute: envBCWasmCallString, GasCost: envBCWasmCallStringGasCost},
			"bcwasmDelegateCall":       &exec.FunctionImport{Execute: envBCWasmDelegateCall, GasCost: envBCWasmDelegateCallGasCost},
			"bcwasmDelegateCallInt64":  &exec.FunctionImport{Execute: envBCWasmDelegateCallInt64, GasCost: envBCWasmDelegateCallInt64GasCost},
			"bcwasmDelegateCallString": &exec.FunctionImport{Execute: envBCWasmDelegateCallString, GasCost: envBCWasmDelegateCallStringGasCost},

			//nizkpail
			//"pailEncrypt":     &exec.FunctionImport{Execute: envPailEncrypt, GasCost: envPailEncryptGasCost},

			//Temporarily comment the following code to prepare for cross platform
			//"pailHomAdd":      &exec.FunctionImport{Execute: envPailHomAdd, GasCost: envPailHomAddGasCost},
			//"pailHomSub":      &exec.FunctionImport{Execute: envPailHomSub, GasCost: envPailHomSubGasCost},
			//"nizkVerifyProof": &exec.FunctionImport{Execute: envNizkVerifyProof, GasCost: envNizkVerifyProofGasCost},

			//sm2
			"smSigVerify": &exec.FunctionImport{Execute: envSmSigVerify, GasCost: envSMVerifyGasCost},
			//sm2sec
			"sm2secSigVerify": &exec.FunctionImport{Execute: envSmSecSigVerify, GasCost: envSMVerifyGasCost},
			//secp256r1
			"secp256r1SigVerify": &exec.FunctionImport{Execute: envP256r1SigVerify, GasCost: envSMVerifyGasCost},
			//secp256k1
			"secp256k1SigVerify": &exec.FunctionImport{Execute: envP256k1SigVerify, GasCost: envSMVerifyGasCost},
			//sm3
			"sm3":         &exec.FunctionImport{Execute: envSm3, GasCost: envSha3GasCost},
			//bulletproof
			"bulletProofVerify": &exec.FunctionImport{Execute: envBulletProofVerify, GasCost: envSMVerifyGasCost},
		},
	}
}

func newGlobalSet() map[string]map[string]int64 {
	return map[string]map[string]int64{
		"env": {
			"stderr": 0,
			"stdin":  0,
			"stdout": 0,
		},
	}
}

//void * memcpy ( void * destination, const void * source, size_t num );
func envMemcpy(vm *exec.VirtualMachine) int64 {
	dest := int(uint32(vm.GetCurrentFrame().Locals[0]))
	src := int(uint32(vm.GetCurrentFrame().Locals[1]))
	len := int(uint32(vm.GetCurrentFrame().Locals[2]))

	copy(vm.Memory.Memory[dest:dest+len], vm.Memory.Memory[src:src+len])
	return int64(dest)
}

func envMemcpyGasCost(vm *exec.VirtualMachine) (uint64, error) {
	//len := int(uint32(vm.GetCurrentFrame().Locals[2]))
	//return uint64(len), nil
	return 23, nil
}

//void * memmove ( void * destination, const void * source, size_t num );
func envMemmove(vm *exec.VirtualMachine) int64 {
	dest := int(uint32(vm.GetCurrentFrame().Locals[0]))
	src := int(uint32(vm.GetCurrentFrame().Locals[1]))
	len := int(uint32(vm.GetCurrentFrame().Locals[2]))

	copy(vm.Memory.Memory[dest:dest+len], vm.Memory.Memory[src:src+len])
	return int64(dest)
}

func envMemmoveGasCost(vm *exec.VirtualMachine) (uint64, error) {
	//len := int(uint32(vm.GetCurrentFrame().Locals[2]))
	//return uint64(len), nil
	return 22, nil

}

//int memcmp ( const void * ptr1, const void * ptr2, size_t num );
func envMemcmp(vm *exec.VirtualMachine) int64 {
	ptr1 := int(uint32(vm.GetCurrentFrame().Locals[0]))
	ptr2 := int(uint32(vm.GetCurrentFrame().Locals[1]))
	num := int(uint32(vm.GetCurrentFrame().Locals[2]))

	return int64(bytes.Compare(vm.Memory.Memory[ptr1:ptr1+num], vm.Memory.Memory[ptr2:ptr2+num]))
}

func envMemcmpGasCost(vm *exec.VirtualMachine) (uint64, error) {
	//len := int(uint32(vm.GetCurrentFrame().Locals[2]))
	//return uint64(len), nil
	return 25, nil
}

//void * memset ( void * ptr, int value, size_t num );
func envMemset(vm *exec.VirtualMachine) int64 {
	ptr := int(uint32(vm.GetCurrentFrame().Locals[0]))
	value := int(uint32(vm.GetCurrentFrame().Locals[1]))
	num := int(uint32(vm.GetCurrentFrame().Locals[2]))

	pos := 0
	for pos < num {
		vm.Memory.Memory[ptr+pos] = byte(value)
		pos++
	}
	return int64(ptr)
}

func envMemsetGasCost(vm *exec.VirtualMachine) (uint64, error) {
	//len := int(uint32(vm.GetCurrentFrame().Locals[2]))
	//return uint64(len), nil
	return 125, nil
}

//libc prints()
func envPrints(vm *exec.VirtualMachine) int64 {
	start := int(uint32(vm.GetCurrentFrame().Locals[0]))
	end := 0
	for end = start; end < len(vm.Memory.Memory); end++ {
		if vm.Memory.Memory[end] == 0 {
			break
		}
	}
	vm.Context.Log.Debug(string(vm.Memory.Memory[start:end]))

	//fmt.Printf("%s", string(vm.Memory.Memory[start:end]))
	return 0
}

func envPrintsGasCost(vm *exec.VirtualMachine) (uint64, error) {
	//start := int(uint32(vm.GetCurrentFrame().Locals[0]))
	//end := 0
	//for end = start; end < len(vm.Memory.Memory); end++ {
	//	if vm.Memory.Memory[end] == 0 {
	//		break
	//	}
	//}
	//return uint64(end - start), nil
	return 319400, nil
}

//libc prints_l
func envPrintsl(vm *exec.VirtualMachine) int64 {
	ptr := int(uint32(vm.GetCurrentFrame().Locals[0]))
	msgLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
	msg := vm.Memory.Memory[ptr : ptr+msgLen]
	vm.Context.Log.Debug(string(msg))
	return 0
}

func envPrintslGasCost(vm *exec.VirtualMachine) (uint64, error) {
	//msgLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
	//return uint64(msgLen), nil
	return 358400, nil

}

//libc printi()
func envPrinti(vm *exec.VirtualMachine) int64 {
	vm.Context.Log.Debug(fmt.Sprintf("%d", vm.GetCurrentFrame().Locals[0]))
	return 0
}

func envPrintiGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 334600, nil
}

func envPrintui(vm *exec.VirtualMachine) int64 {
	vm.Context.Log.Debug(fmt.Sprintf("%d", vm.GetCurrentFrame().Locals[0]))
	return 0
}

func envPrintuiGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 293600, nil
}

func envPrinti128(vm *exec.VirtualMachine) int64 {
	pos := vm.GetCurrentFrame().Locals[0]
	buf := vm.Memory.Memory[pos : pos+16]
	lo := uint64(binary.LittleEndian.Uint64(buf[:8]))
	ho := uint64(binary.LittleEndian.Uint64(buf[8:]))
	ret := C.printi128(C.uint64_t(lo), C.uint64_t(ho))
	vm.Context.Log.Debug(fmt.Sprintf("%s", C.GoString(ret)))
	return 0
}

func envPrinti128GasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 300000, nil
}

func envPrintui128(vm *exec.VirtualMachine) int64 {
	pos := vm.GetCurrentFrame().Locals[0]
	buf := vm.Memory.Memory[pos : pos+16]
	lo := uint64(binary.LittleEndian.Uint64(buf[:8]))
	ho := uint64(binary.LittleEndian.Uint64(buf[8:]))
	ret := C.printui128(C.uint64_t(lo), C.uint64_t(ho))
	vm.Context.Log.Debug(fmt.Sprintf("%s", C.GoString(ret)))
	return 0
}

func envPrintui128GasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 300000, nil
}

func envPrintsf(vm *exec.VirtualMachine) int64 {
	pos := vm.GetCurrentFrame().Locals[0]
	float := math.Float32frombits(uint32(pos))
	vm.Context.Log.Debug(fmt.Sprintf("%g", float))
	return 0
}

func envPrintsfGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 320800, nil
}

func envPrintdf(vm *exec.VirtualMachine) int64 {
	pos := vm.GetCurrentFrame().Locals[0]
	double := math.Float64frombits(uint64(pos))
	vm.Context.Log.Debug(fmt.Sprintf("%g", double))
	return 0
}

func envPrintdfGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 323600, nil
}

func envPrintqf(vm *exec.VirtualMachine) int64 {
	frame := vm.GetCurrentFrame()
	pos := frame.Locals[0]

	low := C.uint64_t(binary.LittleEndian.Uint64(vm.Memory.Memory[pos : pos+8]))
	high := C.uint64_t(binary.LittleEndian.Uint64(vm.Memory.Memory[pos+8 : pos+16]))

	buf := C.GoString(C.__printqf(low, high))
	vm.Context.Log.Debug(fmt.Sprintf("%s", buf))
	return 0
}

func envPrintqfGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 300000, nil
}

func envPrintn(vm *exec.VirtualMachine) int64 {
	vm.Context.Log.Debug(fmt.Sprintf("%d", int(uint32(vm.GetCurrentFrame().Locals[0]))))
	return 0
}

func envPrintnGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 301400, nil
}

func envPrinthex(vm *exec.VirtualMachine) int64 {
	data := int(uint32(vm.GetCurrentFrame().Locals[0]))
	dataLen := int(uint32(vm.GetCurrentFrame().Locals[1]))
	vm.Context.Log.Debug(fmt.Sprintf("%x", vm.Memory.Memory[data:data+dataLen]))
	return 0
}

func envPrinthexGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 387400, nil
}

//libc malloc()
func envMalloc(vm *exec.VirtualMachine) int64 {
	//mem := vm.Memory
	size := int(uint32(vm.GetCurrentFrame().Locals[0]))

	pos := vm.Memory.Malloc(size)
	if pos == -1 {
		panic("melloc error...")
	}

	return int64(pos)
}

func envMallocGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 40, nil
}

//libc free()
func envFree(vm *exec.VirtualMachine) int64 {
	if vm.Context.Config.DisableFree {
		return 0
	}

	mem := vm.Memory
	offset := int(uint32(vm.GetCurrentFrame().Locals[0]))

	err := mem.Free(offset)
	if err != nil {
		panic("free error...")
	}

	return 0
}

func envFreeGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 1424, nil
}

//libc calloc()
func envCalloc(vm *exec.VirtualMachine) int64 {
	mem := vm.Memory
	num := int(int32(vm.GetCurrentFrame().Locals[0]))
	size := int(int32(vm.GetCurrentFrame().Locals[1]))
	total := num * size

	pos := mem.Malloc(total)

	return int64(pos)
}

func envCallocGasCost(vm *exec.VirtualMachine) (uint64, error) {
	//num := int(int32(vm.GetCurrentFrame().Locals[0]))
	//size := int(int32(vm.GetCurrentFrame().Locals[1]))
	//total := num * size
	//return uint64(total), nil
	return 251, nil
}

func envRealloc(vm *exec.VirtualMachine) int64 {
	mem := vm.Memory
	ptr := int(int32(vm.GetCurrentFrame().Locals[0]))
	size := int(int32(vm.GetCurrentFrame().Locals[1]))

	if size == 0 {
		return 0
	}

	pos := mem.Realloc(ptr, size)

	return int64(pos)
}

func envReallocGasCost(vm *exec.VirtualMachine) (uint64, error) {
	//size := int(int32(vm.GetCurrentFrame().Locals[1]))
	//return uint64(size), nil
	return 161, nil
}

func envDisableFree(vm *exec.VirtualMachine) int64 {
	vm.Context.Config.DisableFree = true
	return 0
}

func envDisableFreeGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 1, nil
}

func envAbort(vm *exec.VirtualMachine) int64 {
	panic("abort")
}

func envAbortGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 1, nil
}

// define: int64_t gasPrice();
func envGasPrice(vm *exec.VirtualMachine) int64 {
	gasPrice := vm.Context.StateDB.GasPrice()
	return gasPrice
}

// define: void blockHash(int num, char hash[20]);
func envBlockHash(vm *exec.VirtualMachine) int64 {
	num := int(int32(vm.GetCurrentFrame().Locals[0]))
	offset := int(int32(vm.GetCurrentFrame().Locals[1]))
	blockHash := vm.Context.StateDB.BlockHash(uint64(num))
	copy(vm.Memory.Memory[offset:], blockHash.Bytes())
	return 0
}

// define: int64_t number();
func envNumber(vm *exec.VirtualMachine) int64 {
	return int64(vm.Context.StateDB.BlockNumber().Uint64())
}

// define: int64_t gasLimit();
func envGasLimit(vm *exec.VirtualMachine) int64 {
	return int64(vm.Context.StateDB.GasLimimt())
}

// define: int64_t timestamp();
func envTimestamp(vm *exec.VirtualMachine) int64 {
	return vm.Context.StateDB.Time().Int64()
}

// define: void coinbase(char addr[20]);
func envCoinbase(vm *exec.VirtualMachine) int64 {
	offset := int(int32(vm.GetCurrentFrame().Locals[0]))
	coinBase := vm.Context.StateDB.Coinbase()
	copy(vm.Memory.Memory[offset:], coinBase.Bytes())
	return 0
}

// define: u256 balance();
func envBalance(vm *exec.VirtualMachine) int64 {
	balance := vm.Context.StateDB.GetBalance(vm.Context.StateDB.Address())
	ptr := int(int32(vm.GetCurrentFrame().Locals[0]))
	// 256 bits
	if len(balance.Bytes()) > 32 {
		panic(fmt.Sprintf("balance overflow(%d>32)", len(balance.Bytes())))
	}
	// bigendian
	offset := 32 - len(balance.Bytes())
	if offset > 0 {
		empty := make([]byte, offset)
		copy(vm.Memory.Memory[ptr:ptr+offset], empty)
	}
	copy(vm.Memory.Memory[ptr+offset:], balance.Bytes())
	return 0
}

// define: void origin(char addr[20]);
func envOrigin(vm *exec.VirtualMachine) int64 {
	offset := int(int32(vm.GetCurrentFrame().Locals[0]))
	address := vm.Context.StateDB.Origin()
	copy(vm.Memory.Memory[offset:], address.Bytes())
	return 0
}

// define: void caller(char addr[20]);
func envCaller(vm *exec.VirtualMachine) int64 {
	offset := int(int32(vm.GetCurrentFrame().Locals[0]))
	caller := vm.Context.StateDB.Caller()
	copy(vm.Memory.Memory[offset:], caller.Bytes())
	return 0
}

// define: void owner(char addr[20]);
func envIsOwner(vm *exec.VirtualMachine) int64 {
	contract := int(int32(vm.GetCurrentFrame().Locals[0]))
	contractLen := int(int32(vm.GetCurrentFrame().Locals[1]))
	account := int(int32(vm.GetCurrentFrame().Locals[2]))
	accountLen := int(int32(vm.GetCurrentFrame().Locals[3]))

	contractAddress := common.BytesToAddress(vm.Memory.Memory[contract : contract+contractLen])
	accountAddress := common.BytesToAddress(vm.Memory.Memory[account : account+accountLen])

	return vm.Context.StateDB.IsOwner(contractAddress, accountAddress)
}

// define: int64_t isFromInit();
func envIsFromInit(vm *exec.VirtualMachine) int64 {
	if vm.InitEntryID != -1 {
		return 0
	} else {
		return 1
	}
}

// define: int64_t callValue();
func envCallValue(vm *exec.VirtualMachine) int64 {
	value := vm.Context.StateDB.CallValue()
	ptr := int(int32(vm.GetCurrentFrame().Locals[0]))
	if len(value.Bytes()) > 32 {
		panic(fmt.Sprintf("balance overflow(%d > 32)", len(value.Bytes())))
	}
	// bigendian
	offset := 32 - len(value.Bytes())
	if offset > 0 {
		empty := make([]byte, offset)
		copy(vm.Memory.Memory[ptr:ptr+offset], empty)
	}
	copy(vm.Memory.Memory[ptr+offset:], value.Bytes())
	return 0
}

// define: void address(char hash[20]);
func envAddress(vm *exec.VirtualMachine) int64 {
	offset := int(int32(vm.GetCurrentFrame().Locals[0]))
	address := vm.Context.StateDB.Address()
	copy(vm.Memory.Memory[offset:], address.Bytes())
	return 0
}

// define: void sha3(char *src, size_t srcLen, char *dest, size_t destLen);
func envSha3(vm *exec.VirtualMachine) int64 {
	offset := int(int32(vm.GetCurrentFrame().Locals[0]))
	size := int(int32(vm.GetCurrentFrame().Locals[1]))
	destOffset := int(int32(vm.GetCurrentFrame().Locals[2]))
	destSize := int(int32(vm.GetCurrentFrame().Locals[3]))
	data := vm.Memory.Memory[offset : offset+size]
	hash := crypto.Keccak256(data)
	if destSize < len(hash) {
		return 0
	}
	copy(vm.Memory.Memory[destOffset:], hash)
	return 0
}

func envSm3(vm *exec.VirtualMachine) int64 {
	offset := int(int32(vm.GetCurrentFrame().Locals[0]))
	size := int(int32(vm.GetCurrentFrame().Locals[1]))
	destOffset := int(int32(vm.GetCurrentFrame().Locals[2]))
	destSize := int(int32(vm.GetCurrentFrame().Locals[3]))
	data := vm.Memory.Memory[offset : offset+size]
	data = append(data, 0)

	dest := vm.Memory.Memory[destOffset : destOffset + destSize]
	dest = append(dest, 0)
	dataPtr := (*C.char)(unsafe.Pointer(&data[0]))
	destPtr := (*C.char)(unsafe.Pointer(&dest[0]))

	hash := C.sm3_compute(dataPtr, destPtr)
	if hash == (C.int)(0){
		return 0
	}
	return 0
}


func envEcrecover(vm *exec.VirtualMachine) int64 {
	hashOffset := int(int32(vm.GetCurrentFrame().Locals[0]))
	rsOffset := int(int32(vm.GetCurrentFrame().Locals[1]))
	addrOffset := int(int32(vm.GetCurrentFrame().Locals[2]))

	h := vm.Memory.Memory[hashOffset : hashOffset+32]
	rs := vm.Memory.Memory[rsOffset : rsOffset+65]

	pubK, err := crypto.SigToPub(h, rs)
	if err != nil {
		return 0
	}
	addr := crypto.PubkeyToAddress(*pubK)
	copy(vm.Memory.Memory[addrOffset:], addr.Bytes())

	return 0
}

func envEcrecoverGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 280738, nil
}

func envSmSigVerify(vm *exec.VirtualMachine) int64 {
	msgOffset := int(int32(vm.GetCurrentFrame().Locals[0]))
	msgSize := int(int32(vm.GetCurrentFrame().Locals[1]))
	useridOffset := int(int32(vm.GetCurrentFrame().Locals[2]))
	useridSize := int(int32(vm.GetCurrentFrame().Locals[3]))
	pubkeyOffset := int(int32(vm.GetCurrentFrame().Locals[4]))
	pubkeySize := int(int32(vm.GetCurrentFrame().Locals[5]))
	sigOffset := int(int32(vm.GetCurrentFrame().Locals[6]))
	sigSize := int(int32(vm.GetCurrentFrame().Locals[7]))
	resultOffset := int(int32(vm.GetCurrentFrame().Locals[8]))
	resultSize := int(int32(vm.GetCurrentFrame().Locals[9]))

	msg := vm.Memory.Memory[msgOffset : msgOffset+msgSize]
	userid := vm.Memory.Memory[useridOffset : useridOffset+useridSize]
	pubkey := vm.Memory.Memory[pubkeyOffset : pubkeyOffset+pubkeySize]
	sig := vm.Memory.Memory[sigOffset : sigOffset+sigSize]
	msg = append(msg, 0)
	userid = append(userid, 0)
	pubkey = append(pubkey, 0)
	sig = append(sig, 0)

	msgPtr := (*C.char)(unsafe.Pointer(&msg[0]))
	//useridPtr := (*C.char)(unsafe.Pointer(&userid[0]))
	pubkeyPtr := (*C.char)(unsafe.Pointer(&pubkey[0]))
	sigPtr := (*C.char)(unsafe.Pointer(&sig[0]))

	//result := C.smSigVerify(msgPtr, useridPtr, pubkeyPtr, sigPtr)
	result := C.sm_verify_sig(msgPtr, pubkeyPtr, sigPtr)
	ret := "0"
	if result == 1 {
		ret = "1"
	}
	resultBts := []byte(ret)
	resultBts = append(resultBts, 0)

	if resultSize < len(resultBts) {
		return 0
	}

	copy(vm.Memory.Memory[resultOffset:], resultBts)

	return 0
}


func envSmSecSigVerify(vm *exec.VirtualMachine) int64 {
	msgOffset := int(int32(vm.GetCurrentFrame().Locals[0]))
	msgSize := int(int32(vm.GetCurrentFrame().Locals[1]))
	pubkeyOffset := int(int32(vm.GetCurrentFrame().Locals[2]))
	pubkeySize := int(int32(vm.GetCurrentFrame().Locals[3]))
	sigOffset := int(int32(vm.GetCurrentFrame().Locals[4]))
	sigSize := int(int32(vm.GetCurrentFrame().Locals[5]))
	resultOffset := int(int32(vm.GetCurrentFrame().Locals[6]))
	resultSize := int(int32(vm.GetCurrentFrame().Locals[7]))

	msg := vm.Memory.Memory[msgOffset : msgOffset+msgSize]
	pubkey := vm.Memory.Memory[pubkeyOffset : pubkeyOffset+pubkeySize]
	sig := vm.Memory.Memory[sigOffset : sigOffset+sigSize]
	msg = append(msg, 0)
	pubkey = append(pubkey, 0)
	sig = append(sig, 0)

	msgPtr := (*C.char)(unsafe.Pointer(&msg[0]))
	pubkeyPtr := (*C.char)(unsafe.Pointer(&pubkey[0]))
	sigPtr := (*C.char)(unsafe.Pointer(&sig[0]))
	result := C.sm_verify_sig(msgPtr, pubkeyPtr, sigPtr)
	ret := "0"
	if result == 1 {
		ret = "1"
	}
	resultBts := []byte(ret)
	resultBts = append(resultBts, 0)

	if resultSize < len(resultBts) {
		return 0
	}

	copy(vm.Memory.Memory[resultOffset:], resultBts)

	return 0
}

func envP256k1SigVerify(vm *exec.VirtualMachine) int64 {
	msgOffset := int(int32(vm.GetCurrentFrame().Locals[0]))
	msgSize := int(int32(vm.GetCurrentFrame().Locals[1]))
	pubkeyOffset := int(int32(vm.GetCurrentFrame().Locals[2]))
	pubkeySize := int(int32(vm.GetCurrentFrame().Locals[3]))
	sigOffset := int(int32(vm.GetCurrentFrame().Locals[4]))
	sigSize := int(int32(vm.GetCurrentFrame().Locals[5]))
	resultOffset := int(int32(vm.GetCurrentFrame().Locals[6]))
	resultSize := int(int32(vm.GetCurrentFrame().Locals[7]))

	msg := vm.Memory.Memory[msgOffset : msgOffset+msgSize]
	pubkey := vm.Memory.Memory[pubkeyOffset : pubkeyOffset+pubkeySize]
	sig := vm.Memory.Memory[sigOffset : sigOffset+sigSize]
	fmt.Println("base64:", msg)
	md := crypto.Keccak256(msg)
	md = append(md, 0)
	pubkey = append(pubkey, 0)
	sig = append(sig, 0)

	mdPtr := (*C.char)(unsafe.Pointer(&md[0]))
	pubkeyPtr := (*C.char)(unsafe.Pointer(&pubkey[0]))
	sigPtr := (*C.char)(unsafe.Pointer(&sig[0]))
	result := C.p256k1_verify_with_base64(mdPtr, pubkeyPtr, sigPtr)
	fmt.Println("result:", result)
	ret := "0"
	if result == 1 {
		ret = "1"
	}
	resultBts := []byte(ret)
	resultBts = append(resultBts, 0)

	if resultSize < len(resultBts) {
		return 0
	}

	copy(vm.Memory.Memory[resultOffset:], resultBts)

	return 0
}
func envP256r1SigVerify(vm *exec.VirtualMachine) int64 {
	msgOffset := int(int32(vm.GetCurrentFrame().Locals[0]))
	msgSize := int(int32(vm.GetCurrentFrame().Locals[1]))
	pubkeyOffset := int(int32(vm.GetCurrentFrame().Locals[2]))
	pubkeySize := int(int32(vm.GetCurrentFrame().Locals[3]))
	sigOffset := int(int32(vm.GetCurrentFrame().Locals[4]))
	sigSize := int(int32(vm.GetCurrentFrame().Locals[5]))
	resultOffset := int(int32(vm.GetCurrentFrame().Locals[6]))
	resultSize := int(int32(vm.GetCurrentFrame().Locals[7]))

	msg := vm.Memory.Memory[msgOffset : msgOffset+msgSize]
	pubkey := vm.Memory.Memory[pubkeyOffset : pubkeyOffset+pubkeySize]
	sig := vm.Memory.Memory[sigOffset : sigOffset+sigSize]
	msg = append(msg, 0)
	pubkey = append(pubkey, 0)
	sig = append(sig, 0)

	msgPtr := (*C.char)(unsafe.Pointer(&msg[0]))
	pubkeyPtr := (*C.char)(unsafe.Pointer(&pubkey[0]))
	sigPtr := (*C.char)(unsafe.Pointer(&sig[0]))
	result := C.p256r1_verify_with_base64(msgPtr, pubkeyPtr, sigPtr)
	ret := "0"
	if result == 1 {
		ret = "1"
	}
	resultBts := []byte(ret)
	resultBts = append(resultBts, 0)

	if resultSize < len(resultBts) {
		return 0
	}

	copy(vm.Memory.Memory[resultOffset:], resultBts)

	return 0
}

func envBulletProofVerify(vm *exec.VirtualMachine) int64 {
	proofOffset := int(int32(vm.GetCurrentFrame().Locals[0]))
	proofSize := int(int32(vm.GetCurrentFrame().Locals[1]))
	resultOffset := int(int32(vm.GetCurrentFrame().Locals[2]))
	resultSize := int(int32(vm.GetCurrentFrame().Locals[3]))

	proof := vm.Memory.Memory[proofOffset : proofSize + proofOffset]

	//hexproof := hexutil.Encode(proof)
	statement := cryptoZk.GenerateAggBpStatement(2, 16)
	result, err := cryptoZk.AggBpVerify_s(string(proof), statement)

	//fmt.Println("result:", result)
	ret := "1"
	if err !=nil || !result {
		ret = "0"
	}
	resultBts := []byte(ret)
	resultBts = append(resultBts, 0)

	if resultSize < len(resultBts) {
		return 0
	}

	copy(vm.Memory.Memory[resultOffset:], resultBts)

	return 0
}

//
//func envPailHomAdd(vm *exec.VirtualMachine) int64 {
//	cipher1Offset := int(int32(vm.GetCurrentFrame().Locals[0]))
//	cipher1Size := int(int32(vm.GetCurrentFrame().Locals[1]))
//	cipher2Offset := int(int32(vm.GetCurrentFrame().Locals[2]))
//	cipher2Size := int(int32(vm.GetCurrentFrame().Locals[3]))
//	pubkeyOffset := int(int32(vm.GetCurrentFrame().Locals[4]))
//	pubkeySize := int(int32(vm.GetCurrentFrame().Locals[5]))
//	resultOffset := int(int32(vm.GetCurrentFrame().Locals[6]))
//	resultSize := int(int32(vm.GetCurrentFrame().Locals[7]))
//
//	cipher1 := vm.Memory.Memory[cipher1Offset : cipher1Offset+cipher1Size]
//	cipher2 := vm.Memory.Memory[cipher2Offset : cipher2Offset+cipher2Size]
//	pubkey := vm.Memory.Memory[pubkeyOffset : pubkeyOffset+pubkeySize]
//	cipher1 = append(cipher1, 0)
//	cipher2 = append(cipher2, 0)
//	pubkey = append(pubkey, 0)
//
//	cipher1Ptr := (*C.char)(unsafe.Pointer(&cipher1[0]))
//	cipher2Ptr := (*C.char)(unsafe.Pointer(&cipher2[0]))
//	pubkeyPtr := (*C.char)(unsafe.Pointer(&pubkey[0]))
//
//	resultPtr := C.pailHomAdd(cipher1Ptr, cipher2Ptr, pubkeyPtr)
//	resultStr := C.GoString(resultPtr)
//	C.free(unsafe.Pointer(resultPtr))
//
//	resultBts := []byte(resultStr)
//	resultBts = append(resultBts, 0)
//
//	if resultSize < len(resultBts) {
//		return 0
//	}
//
//	copy(vm.Memory.Memory[resultOffset:], resultBts)
//
//	return 0
//}
//
//func envPailHomSub(vm *exec.VirtualMachine) int64 {
//	cipher1Offset := int(int32(vm.GetCurrentFrame().Locals[0]))
//	cipher1Size := int(int32(vm.GetCurrentFrame().Locals[1]))
//	cipher2Offset := int(int32(vm.GetCurrentFrame().Locals[2]))
//	cipher2Size := int(int32(vm.GetCurrentFrame().Locals[3]))
//	pubkeyOffset := int(int32(vm.GetCurrentFrame().Locals[4]))
//	pubkeySize := int(int32(vm.GetCurrentFrame().Locals[5]))
//	resultOffset := int(int32(vm.GetCurrentFrame().Locals[6]))
//	resultSize := int(int32(vm.GetCurrentFrame().Locals[7]))
//
//	cipher1 := vm.Memory.Memory[cipher1Offset : cipher1Offset+cipher1Size]
//	cipher2 := vm.Memory.Memory[cipher2Offset : cipher2Offset+cipher2Size]
//	pubkey := vm.Memory.Memory[pubkeyOffset : pubkeyOffset+pubkeySize]
//	cipher1 = append(cipher1, 0)
//	cipher2 = append(cipher2, 0)
//	pubkey = append(pubkey, 0)
//
//	cipher1Ptr := (*C.char)(unsafe.Pointer(&cipher1[0]))
//	cipher2Ptr := (*C.char)(unsafe.Pointer(&cipher2[0]))
//	pubkeyPtr := (*C.char)(unsafe.Pointer(&pubkey[0]))
//
//	resultPtr := C.pailHomSub(cipher1Ptr, cipher2Ptr, pubkeyPtr)
//	resultStr := C.GoString(resultPtr)
//	C.free(unsafe.Pointer(resultPtr))
//
//	resultBts := []byte(resultStr)
//	resultBts = append(resultBts, 0)
//
//	if resultSize < len(resultBts) {
//		return 0
//	}
//
//	copy(vm.Memory.Memory[resultOffset:], resultBts)
//
//	return 0
//}
//
//func envNizkVerifyProof(vm *exec.VirtualMachine) int64 {
//	paiOffset := int(int32(vm.GetCurrentFrame().Locals[0]))
//	paiSize := int(int32(vm.GetCurrentFrame().Locals[1]))
//	fromBalCipherOffset := int(int32(vm.GetCurrentFrame().Locals[2]))
//	fromBalCipherSize := int(int32(vm.GetCurrentFrame().Locals[3]))
//	fromAmountCipherOffset := int(int32(vm.GetCurrentFrame().Locals[4]))
//	fromAmountCipherSize := int(int32(vm.GetCurrentFrame().Locals[5]))
//	toAmountCipherOffset := int(int32(vm.GetCurrentFrame().Locals[6]))
//	toAmountCipherSize := int(int32(vm.GetCurrentFrame().Locals[7]))
//	fromPubkeyOffset := int(int32(vm.GetCurrentFrame().Locals[8]))
//	fromPubkeySize := int(int32(vm.GetCurrentFrame().Locals[9]))
//	toPubkeyOffset := int(int32(vm.GetCurrentFrame().Locals[10]))
//	toPubkeySize := int(int32(vm.GetCurrentFrame().Locals[11]))
//	resultOffset := int(int32(vm.GetCurrentFrame().Locals[12]))
//	resultSize := int(int32(vm.GetCurrentFrame().Locals[13]))
//
//	pai := vm.Memory.Memory[paiOffset : paiOffset+paiSize]
//	fromBalCipher := vm.Memory.Memory[fromBalCipherOffset : fromBalCipherOffset+fromBalCipherSize]
//	fromAmountCipher := vm.Memory.Memory[fromAmountCipherOffset : fromAmountCipherOffset+fromAmountCipherSize]
//	toAmountCipher := vm.Memory.Memory[toAmountCipherOffset : toAmountCipherOffset+toAmountCipherSize]
//	fromPubkey := vm.Memory.Memory[fromPubkeyOffset : fromPubkeyOffset+fromPubkeySize]
//	toPubkey := vm.Memory.Memory[toPubkeyOffset : toPubkeyOffset+toPubkeySize]
//	pai = append(pai, 0)
//	fromBalCipher = append(fromBalCipher, 0)
//	fromAmountCipher = append(fromAmountCipher, 0)
//	toAmountCipher = append(toAmountCipher, 0)
//	fromPubkey = append(fromPubkey, 0)
//	toPubkey = append(toPubkey, 0)
//
//	paiPtr := (*C.char)(unsafe.Pointer(&pai[0]))
//	fromBalCipherPtr := (*C.char)(unsafe.Pointer(&fromBalCipher[0]))
//	fromAmountCipherPtr := (*C.char)(unsafe.Pointer(&fromAmountCipher[0]))
//	toAmountCipherPtr := (*C.char)(unsafe.Pointer(&toAmountCipher[0]))
//	fromPubkeyPtr := (*C.char)(unsafe.Pointer(&fromPubkey[0]))
//	toPubkeyPtr := (*C.char)(unsafe.Pointer(&toPubkey[0]))
//
//	resultPtr := C.nizkVerifyProof(paiPtr, fromBalCipherPtr, fromAmountCipherPtr, toAmountCipherPtr, fromPubkeyPtr, toPubkeyPtr)
//	resultStr := C.GoString(resultPtr)
//	C.free(unsafe.Pointer(resultPtr))
//
//	resultBts := []byte(resultStr)
//	resultBts = append(resultBts, 0)
//
//	if resultSize < len(resultBts) {
//		return 0
//	}
//
//	copy(vm.Memory.Memory[resultOffset:], resultBts)
//
//	return 0
//}

func envSha3GasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 1310, nil
}

func envGasPriceGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 9, nil
}
func envBlockHashGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 25, nil
}
func envNumberGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 9, nil
}
func envGasLimitGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 7, nil
}

func envTimestampGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 9, nil
}

func envCoinbaseGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 27, nil
}

func envBalanceGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 1609, nil
}

func envOriginGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 28, nil
}

func envCallerGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 35, nil
}

func envIsOwnerGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 5077, nil
}

func envIsFromInitGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 4, nil
}
func envCallValueGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 215, nil
}
func envAddressGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 79, nil
}

func constGasFunc(gas uint64) exec.GasCost {
	return func(vm *exec.VirtualMachine) (uint64, error) {
		return gas, nil
	}
}

func envPailHomAddGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 2936, nil
}
func envPailHomSubGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 2848, nil
}
func envNizkVerifyProofGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 23188, nil
}
func envSMVerifyGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 2936, nil
}
func envGetCallerNonceGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 1358, nil
}

func envCallTransferGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 1349, nil
}

//void emitEvent(const char *topic, size_t topicLen, const uint8_t *data, size_t dataLen);
func envEmitEvent(vm *exec.VirtualMachine) int64 {
	topic := int(int32(vm.GetCurrentFrame().Locals[0]))
	topicLen := int(int32(vm.GetCurrentFrame().Locals[1]))
	dataSrc := int(int32(vm.GetCurrentFrame().Locals[2]))
	dataLen := int(int32(vm.GetCurrentFrame().Locals[3]))

	t := make([]byte, topicLen)
	d := make([]byte, dataLen)
	copy(t, vm.Memory.Memory[topic:topic+topicLen])
	copy(d, vm.Memory.Memory[dataSrc:dataSrc+dataLen])
	address := vm.Context.StateDB.Address()
	topics := []common.Hash{common.BytesToHash(crypto.Keccak256(t))}
	bn := vm.Context.StateDB.BlockNumber().Uint64()

	vm.Context.StateDB.AddLog(address, topics, d, bn)
	return 0
}

func envEmitEventGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 300000, nil
}

func envSetState(vm *exec.VirtualMachine) int64 {
	key := int(int32(vm.GetCurrentFrame().Locals[0]))
	keyLen := int(int32(vm.GetCurrentFrame().Locals[1]))
	value := int(int32(vm.GetCurrentFrame().Locals[2]))
	valueLen := int(int32(vm.GetCurrentFrame().Locals[3]))

	copyKey := make([]byte, keyLen)
	copyValue := make([]byte, valueLen)
	copy(copyKey, vm.Memory.Memory[key:key+keyLen])
	copy(copyValue, vm.Memory.Memory[value:value+valueLen])
	vm.Context.StateDB.SetState(copyKey, copyValue)
	return 0
}

func envSetStateGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 5132, nil
}

func envGetState(vm *exec.VirtualMachine) int64 {
	key := int(int32(vm.GetCurrentFrame().Locals[0]))
	keyLen := int(int32(vm.GetCurrentFrame().Locals[1]))
	value := int(int32(vm.GetCurrentFrame().Locals[2]))
	valueLen := int(int32(vm.GetCurrentFrame().Locals[3]))

	val := vm.Context.StateDB.GetState(vm.Memory.Memory[key : key+keyLen])

	if len(val) > valueLen {
		return 0
	}

	copy(vm.Memory.Memory[value:value+valueLen], val)
	return 0
}

func envGetStateGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 4503, nil
}

func envGetStateSize(vm *exec.VirtualMachine) int64 {
	key := int(int32(vm.GetCurrentFrame().Locals[0]))
	keyLen := int(int32(vm.GetCurrentFrame().Locals[1]))
	val := vm.Context.StateDB.GetState(vm.Memory.Memory[key : key+keyLen])

	return int64(len(val))
}

func envGetStateSizeGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 4573, nil
}

// define: int64_t getNonce();
func envGetCallerNonce(vm *exec.VirtualMachine) int64 {
	return vm.Context.StateDB.GetCallerNonce()
}

func envCallTransfer(vm *exec.VirtualMachine) int64 {
	key := int(int32(vm.GetCurrentFrame().Locals[0]))
	keyLen := int(int32(vm.GetCurrentFrame().Locals[1]))
	value := int(vm.GetCurrentFrame().Locals[2])
	bValue := new(big.Int)
	// 256 bits
	bValue.SetBytes(vm.Memory.Memory[value : value+32])
	value256 := inner.U256(bValue)
	addr := common.BytesToAddress(vm.Memory.Memory[key : key+keyLen])

	_, returnGas, err := vm.Context.StateDB.Transfer(addr, value256)

	vm.Context.GasUsed -= returnGas
	if err != nil {

		return 1
	} else {
		return 0
	}
}

func envBCWasmCall(vm *exec.VirtualMachine) int64 {
	addr := int(int32(vm.GetCurrentFrame().Locals[0]))
	params := int(int32(vm.GetCurrentFrame().Locals[1]))
	paramsLen := int(int32(vm.GetCurrentFrame().Locals[2]))

	_, err := vm.Context.StateDB.Call(vm.Memory.Memory[addr : addr+20], vm.Memory.Memory[params:params+paramsLen])
	if err != nil {
		common.ErrPrintln("call error: ", err.Error())
		return 0
	}
	return 0
}

func envBCWasmDelegateCall(vm *exec.VirtualMachine) int64 {
	addr := int(int32(vm.GetCurrentFrame().Locals[0]))
	params := int(int32(vm.GetCurrentFrame().Locals[1]))
	paramsLen := int(int32(vm.GetCurrentFrame().Locals[2]))

	address := vm.Memory.Memory[addr : addr+20]
	input, err := parseWasmCallSolInput(vm, address, vm.Memory.Memory[params:params+paramsLen])
	if err != nil {
		common.ErrPrintln("call parseWasmInput err: ", err.Error())
		return 0
	}
	_, err = vm.Context.StateDB.DelegateCall(address, input)
	if err != nil {
		common.ErrPrintln("call error: ", err.Error())
		return 0
	}
	return 0
}

func envBCWasmCallInt64(vm *exec.VirtualMachine) int64 {
	addr := int(int32(vm.GetCurrentFrame().Locals[0]))
	params := int(int32(vm.GetCurrentFrame().Locals[1]))
	paramsLen := int(int32(vm.GetCurrentFrame().Locals[2]))

	ret, err := vm.Context.StateDB.Call(vm.Memory.Memory[addr : addr+20], vm.Memory.Memory[params:params+paramsLen])
	if err != nil {
		common.ErrPrintln("call error: ", err.Error())
		return 0
	}
	ret = common.WasmCallResultCompatibleSolInt64(ret)
	return common.BytesToInt64(ret)
}

func envBCWasmDelegateCallInt64(vm *exec.VirtualMachine) int64 {
	addr := int(int32(vm.GetCurrentFrame().Locals[0]))
	params := int(int32(vm.GetCurrentFrame().Locals[1]))
	paramsLen := int(int32(vm.GetCurrentFrame().Locals[2]))

	address := vm.Memory.Memory[addr : addr+20]
	// Resolve the delegate call solidity contract input problem
	input, err := parseWasmCallSolInput(vm, address, vm.Memory.Memory[params:params+paramsLen])
	if err != nil {
		common.ErrPrintln("call parseWasmInput err: ", err.Error())
		return 0
	}

	ret, err := vm.Context.StateDB.DelegateCall(address, input)
	if err != nil {
		common.ErrPrintln("call error: ", err.Error())
		return 0
	}
	ret = common.WasmCallResultCompatibleSolInt64(ret)
	return common.BytesToInt64(ret)
}

func envBCWasmCallString(vm *exec.VirtualMachine) int64 {
	addr := int(int32(vm.GetCurrentFrame().Locals[0]))
	params := int(int32(vm.GetCurrentFrame().Locals[1]))
	paramsLen := int(int32(vm.GetCurrentFrame().Locals[2]))

	ret, err := vm.Context.StateDB.Call(vm.Memory.Memory[addr : addr+20], vm.Memory.Memory[params:params+paramsLen])
	if err != nil {
		common.ErrPrintln("call error: ", err.Error())
		return 0
	}
	ret = common.WasmCallResultCompatibleSolString(ret)
	return MallocString(vm, string(ret))
}

func envBCWasmDelegateCallString(vm *exec.VirtualMachine) int64 {
	addr := int(int32(vm.GetCurrentFrame().Locals[0]))
	params := int(int32(vm.GetCurrentFrame().Locals[1]))
	paramsLen := int(int32(vm.GetCurrentFrame().Locals[2]))

	address := vm.Memory.Memory[addr : addr+20]
	// Resolve the delegate call solidity contract input problem
	input, err := parseWasmCallSolInput(vm, address, vm.Memory.Memory[params:params+paramsLen])
	if err != nil {
		common.ErrPrintln("call parseWasmInput err: ", err.Error())
		return 0
	}

	ret, err := vm.Context.StateDB.DelegateCall(address, input)
	if err != nil {
		common.ErrPrintln("call error: ", err.Error())
		return 0
	}
	ret = common.WasmCallResultCompatibleSolString(ret)
	return MallocString(vm, string(ret))
}

func envBCWasmCallGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 1231, nil
}

func envBCWasmCallInt64GasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 1282, nil
}

func envBCWasmCallStringGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 1283, nil
}
func envBCWasmDelegateCallStringGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 3826, nil
}

func envBCWasmDelegateCallInt64GasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 3807, nil
}

func envBCWasmDelegateCallGasCost(vm *exec.VirtualMachine) (uint64, error) {
	return 3469, nil
}
