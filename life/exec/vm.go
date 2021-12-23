package exec

/*
#cgo CFLAGS: -I../resolver
#cgo CXXFLAGS: -std=c++14
#include "platone_softfloat.h"
#cgo LDFLAGS: -L../resolver/softfloat/build -lsoftfloat
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"

	"github.com/Venachain/Venachain/log"

	"github.com/Venachain/Venachain/life/compiler"
	"github.com/Venachain/Venachain/life/compiler/opcodes"
	"github.com/Venachain/Venachain/life/utils"

	"github.com/go-interpreter/wagon/wasm"
)

type (
	Execute func(vm *VirtualMachine) int64
	GasCost func(vm *VirtualMachine) (uint64, error)
)

// FunctionImport represents the function import type. If len(sig.ReturnTypes) == 0, the return value will be ignored.
type FunctionImport struct {
	Execute Execute
	GasCost GasCost
}

const (
	// DefaultCallStackSize is the default call stack size.
	DefaultCallStackSize = 512

	// DefaultPageSize is the linear memory page size.  65536
	DefaultPageSize = 65536

	// JITCodeSizeThreshold is the lower-bound code size threshold for the JIT compiler.
	JITCodeSizeThreshold = 30

	DefaultMemoryPages = 16
	DynamicMemoryPages = 16

	DefaultMemPoolCount   = 5
	DefaultMemBlockSize   = 5
	DefaultMemTreeMaxPage = 8
)

// LE is a simple alias to `binary.LittleEndian`.
var LE = binary.LittleEndian
var memPool = NewMemPool(DefaultMemPoolCount, DefaultMemBlockSize)
var treePool = NewTreePool(DefaultMemPoolCount, DefaultMemBlockSize)

//var pageMemPool = NewPageMemPool()

// VirtualMachine is a WebAssembly execution environment.
type VirtualMachine struct {
	Context         *VMContext
	Module          *compiler.Module
	FunctionCode    []compiler.InterpreterCode
	FunctionImports []*FunctionImport
	JumpTable       [256]Instruction
	CallStack       []Frame
	CurrentFrame    int
	Table           []uint32
	Globals         []int64
	//Memory          *VMMemory
	Memory         *Memory
	NumValueSlots  int
	Yielded        int64
	InsideExecute  bool
	Delegate       func()
	Exited         bool
	ExitError      interface{}
	ReturnValue    int64
	Gas            uint64
	ExternalParams []int64
	InitEntryID    int
}

// VMConfig denotes a set of options passed to a single VirtualMachine insta.ce
type VMConfig struct {
	EnableJIT          bool
	DynamicMemoryPages int
	MaxMemoryPages     int
	MaxTableSize       int
	MaxValueSlots      int
	MaxCallStackDepth  int
	DefaultMemoryPages int
	DefaultTableSize   int
	GasLimit           uint64
	DisableFree        bool
}

type VMContext struct {
	Config   VMConfig
	Addr     [20]byte
	GasUsed  uint64
	GasLimit uint64

	StateDB StateDB
	Log     log.Logger
}

type VMMemory struct {
	Memory    []byte
	Start     int
	Current   int
	MemPoints map[int]int
}

// Frame represents a call frame.
type Frame struct {
	FunctionID   int
	Code         []byte
	JITInfo      interface{}
	Regs         []int64
	Locals       []int64
	IP           int
	ReturnReg    int
	Continuation int32
}

// ImportResolver is an interface for allowing one to define imports to WebAssembly modules
// ran under a single VirtualMachine instance.
type ImportResolver interface {
	ResolveFunc(module, field string) *FunctionImport
	ResolveGlobal(module, field string) int64
}

func ParseModuleAndFunc(code []byte, gasPolicy compiler.GasPolicy) (*compiler.Module, []compiler.InterpreterCode, error) {
	m, err := compiler.LoadModule(code)
	if err != nil {
		return nil, nil, err
	}

	functionCode, err := m.CompileForInterpreter(nil)
	if err != nil {
		return nil, nil, err
	}
	return m, functionCode, nil
}

func NewVirtualMachineWithModule(m *compiler.Module, functionCode []compiler.InterpreterCode, context *VMContext, impResolver ImportResolver, gasPolicy compiler.GasPolicy) (_retVM *VirtualMachine, retErr error) {
	defer utils.CatchPanic(&retErr)

	table := make([]uint32, 0)
	globals := make([]int64, 0)
	funcImports := make([]*FunctionImport, 0)

	if m.Base.Import != nil && impResolver != nil {
		for _, imp := range m.Base.Import.Entries {
			switch imp.Type.Kind() {
			case wasm.ExternalFunction:
				funcImports = append(funcImports, impResolver.ResolveFunc(imp.ModuleName, imp.FieldName))
			case wasm.ExternalGlobal:
				globals = append(globals, impResolver.ResolveGlobal(imp.ModuleName, imp.FieldName))
			case wasm.ExternalMemory:
				// TODO: Do we want a real import?
				if m.Base.Memory != nil && len(m.Base.Memory.Entries) > 0 {
					panic("cannot import another memory while we already have one")
				}
				m.Base.Memory = &wasm.SectionMemories{
					Entries: []wasm.Memory{
						wasm.Memory{
							Limits: wasm.ResizableLimits{
								Initial: uint32(context.Config.DefaultMemoryPages),
							},
						},
					},
				}
			case wasm.ExternalTable:
				// TODO: Do we want a real import?
				if m.Base.Table != nil && len(m.Base.Table.Entries) > 0 {
					panic("cannot import another table while we already have one")
				}
				m.Base.Table = &wasm.SectionTables{
					Entries: []wasm.Table{
						wasm.Table{
							Limits: wasm.ResizableLimits{
								Initial: uint32(context.Config.DefaultTableSize),
							},
						},
					},
				}
			default:
				panic(fmt.Errorf("import kind not supported: %d", imp.Type.Kind()))
			}
		}
	}

	// Load global entries.
	for _, entry := range m.Base.GlobalIndexSpace {
		globals = append(globals, execInitExpr(entry.Init, globals))
	}

	// Populate table elements.
	if m.Base.Table != nil && len(m.Base.Table.Entries) > 0 {
		t := &m.Base.Table.Entries[0]

		if context.Config.MaxTableSize != 0 && int(t.Limits.Initial) > context.Config.MaxTableSize {
			panic("max table size exceeded")
		}

		table = make([]uint32, int(t.Limits.Initial))
		for i := 0; i < int(t.Limits.Initial); i++ {
			table[i] = 0xffffffff
		}
		if m.Base.Elements != nil && len(m.Base.Elements.Entries) > 0 {
			for _, e := range m.Base.Elements.Entries {
				offset := int(execInitExpr(e.Offset, globals))
				copy(table[offset:], e.Elems)
			}
		}
	}

	// Load linear memory.
	//memory := make([]byte, 0)

	//memory := &VMMemory{
	//	Memory:    make([]byte, 0),
	//	Start:     0,
	//	Current:   0,
	//	MemPoints: make(map[int]int),
	//}
	//
	//if m.Base.Memory != nil && len(m.Base.Memory.Entries) > 0 {
	//	initialLimit := int(m.Base.Memory.Entries[0].Limits.Initial)
	//	if context.Config.MaxMemoryPages != 0 && initialLimit > context.Config.MaxMemoryPages {
	//		panic("max memory exceeded")
	//	}
	//
	//	capacity := initialLimit + context.Config.DynamicMemoryPages
	//
	//	memory.Start = initialLimit * DefaultPageSize
	//	memory.Current = initialLimit * DefaultPageSize
	//	// Initialize empty memory.
	//	//buffer := bytes.NewBuffer(make([]byte, capacity))
	//	memory.Memory = memPool.Get(capacity)
	//
	//	if m.Base.Data != nil && len(m.Base.Data.Entries) > 0 {
	//		for _, e := range m.Base.Data.Entries {
	//			offset := int(execInitExpr(e.Offset, globals))
	//			copy(memory.Memory[int(offset):], e.Data)
	//		}
	//	}
	//}

	memory := &Memory{}
	if m.Base.Memory != nil && len(m.Base.Memory.Entries) > 0 {
		initialLimit := int(m.Base.Memory.Entries[0].Limits.Initial)
		if context.Config.MaxMemoryPages != 0 && initialLimit > context.Config.MaxMemoryPages {
			panic("max memory exceeded")
		}

		capacity := initialLimit + context.Config.DynamicMemoryPages
		// Initialize empty memory.
		memory.Memory = memPool.Get(capacity)
		memory.Start = initialLimit * DefaultPageSize
		memory.tree = treePool.GetTree(capacity - initialLimit)
		memory.Size = (len(memory.tree) + 1) / 2

		if m.Base.Data != nil && len(m.Base.Data.Entries) > 0 {
			for _, e := range m.Base.Data.Entries {
				offset := int(execInitExpr(e.Offset, globals))
				copy(memory.Memory[int(offset):], e.Data)
			}
		}
	}

	return &VirtualMachine{
		Module:          m,
		Context:         context,
		FunctionCode:    functionCode,
		FunctionImports: funcImports,
		JumpTable:       GasTable,
		CallStack:       make([]Frame, DefaultCallStackSize),
		CurrentFrame:    -1,
		Table:           table,
		Globals:         globals,
		Memory:          memory,
		Exited:          true,
		ExternalParams:  make([]int64, 0),
		InitEntryID:     -1,
	}, nil
}

// NewVirtualMachine instantiates a virtual machine for a given WebAssembly module, with
// specific execution options specified under a VMConfig, and a WebAssembly module import
// resolver.
func NewVirtualMachine(code []byte, context *VMContext, impResolver ImportResolver, gasPolicy compiler.GasPolicy) (_retVM *VirtualMachine, retErr error) {
	if context.Config.EnableJIT {
		log.Warn("Warning: JIT support is incomplete and the internals are likely to change in the future.")
	}

	m, functionCode, err := ParseModuleAndFunc(code, gasPolicy)
	if err != nil {
		return nil, err
	}

	return NewVirtualMachineWithModule(m, functionCode, context, impResolver, gasPolicy)

}

func ImportGasFunc(vm *VirtualMachine, frame *Frame) (uint64, error) {
	importID := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
	gas, err := vm.FunctionImports[importID].GasCost(vm)
	return gas + 6, err

}

// Init initializes a frame. Must be called on `call` and `call_indirect`.
func (f *Frame) Init(vm *VirtualMachine, functionID int, code compiler.InterpreterCode) {
	numValueSlots := code.NumRegs + code.NumParams + code.NumLocals
	if vm.Context.Config.MaxValueSlots != 0 && vm.NumValueSlots+numValueSlots > vm.Context.Config.MaxValueSlots {
		panic("max value slot count exceeded")
	}
	vm.NumValueSlots += numValueSlots

	values := make([]int64, numValueSlots)

	f.FunctionID = functionID
	f.Regs = values[:code.NumRegs]
	f.Locals = values[code.NumRegs:]
	f.Code = code.Bytes
	f.IP = 0
	f.Continuation = 0

	if vm.Context.Config.EnableJIT {
		code := &vm.FunctionCode[functionID]
		if !code.JITDone {
			if len(code.Bytes) > JITCodeSizeThreshold {
				if !vm.GenerateCodeForFunction(functionID) {
					log.Warn("codegen for function %d failed\n", "functionID", functionID)
				} else {
					log.Debug("codegen for function %d succeeded\n", "functionID", functionID)
				}
			}
			code.JITDone = true
		}
		f.JITInfo = code.JITInfo
	}
}

// Destroy destroys a frame. Must be called on return.
func (f *Frame) Destroy(vm *VirtualMachine) {
	numValueSlots := len(f.Regs) + len(f.Locals)
	vm.NumValueSlots -= numValueSlots
}

// GetCurrentFrame returns the current frame.
func (vm *VirtualMachine) GetCurrentFrame() *Frame {
	if vm.Context.Config.MaxCallStackDepth != 0 && vm.CurrentFrame >= vm.Context.Config.MaxCallStackDepth {
		panic("max call stack depth exceeded")
	}

	if vm.CurrentFrame >= len(vm.CallStack) {
		panic("call stack overflow")
		//vm.CallStack = append(vm.CallStack, make([]Frame, DefaultCallStackSize / 2)...)
	}
	return &vm.CallStack[vm.CurrentFrame]
}

func (vm *VirtualMachine) getExport(key string, kind wasm.External) (int, bool) {
	if vm.Module.Base.Export == nil {
		return -1, false
	}

	entry, ok := vm.Module.Base.Export.Entries[key]
	if !ok {
		return -1, false
	}

	if entry.Kind != kind {
		return -1, false
	}

	return int(entry.Index), true
}

// GetGlobalExport returns the global export with the given name.
func (vm *VirtualMachine) GetGlobalExport(key string) (int, bool) {
	return vm.getExport(key, wasm.ExternalGlobal)
}

// GetFunctionExport returns the function export with the given name.
func (vm *VirtualMachine) GetFunctionExport(key string) (int, bool) {
	return vm.getExport(key, wasm.ExternalFunction)
}

// PrintStackTrace prints the entire VM stack trace for debugging.
func (vm *VirtualMachine) PrintStackTrace() {
	for i := vm.CurrentFrame; i >= 0; i-- {
		functionID := vm.CallStack[i].FunctionID
		log.Debug(fmt.Sprintf("<%d> [%d] %s\n", i, functionID, vm.Module.FunctionNames[functionID]))
	}
}

// Ignite initializes the first call frame.
func (vm *VirtualMachine) Ignite(functionID int, params ...int64) {
	if vm.ExitError != nil {
		panic("last execution exited with error; cannot ignite.")
	}

	if vm.CurrentFrame != -1 {
		panic("call stack not empty; cannot ignite.")
	}

	code := vm.FunctionCode[functionID]
	if code.NumParams != len(params) {
		panic("param count mismatch")
	}

	vm.Exited = false

	vm.CurrentFrame++
	frame := vm.GetCurrentFrame()
	frame.Init(
		vm,
		functionID,
		code,
	)
	copy(frame.Locals, params)
}

func (vm *VirtualMachine) AddAndCheckGas(delta uint64) {
	newGas := vm.Gas + delta
	if newGas < vm.Gas {
		panic("gas overflow")
	}
	if vm.Context.Config.GasLimit != 0 && newGas > vm.Context.Config.GasLimit {
		panic("gas limit exceeded")
	}
	vm.Gas = newGas
}

// Execute starts the virtual machines main instruction processing loop.
// This function may return at any point and is guaranteed to return
// at least once every 10000 instructions. Caller is responsible for
// detecting VM status in a loop.
func (vm *VirtualMachine) Execute() {
	if vm.Exited == true {
		panic("attempting to execute an exited vm")
	}

	if vm.Delegate != nil {
		panic("delegate not cleared")
	}

	if vm.InsideExecute {
		panic("vm execution is not re-entrant")
	}
	vm.InsideExecute = true

	defer func() {
		vm.InsideExecute = false
		if err := recover(); err != nil {
			vm.Exited = true
			vm.ExitError = err
		}
	}()

	frame := vm.GetCurrentFrame()

	for {
		if frame.JITInfo != nil {
			dm := frame.JITInfo.(*DynamicModule)
			var fRetVal int64
			status := dm.Run(vm, &fRetVal)
			if status < 0 {
				panic(fmt.Errorf("status = %d", status))
			}
			frame.Continuation = status
			frame.IP = int(fRetVal)
		}

		valueID := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
		ins := opcodes.Opcode(frame.Code[frame.IP+4])
		frame.IP += 5

		cost, err := vm.JumpTable[ins].GasCost(vm, frame)
		if err != nil || (cost+vm.Context.GasUsed) > vm.Context.GasLimit {
			panic(fmt.Sprintf("out of gas  cost:%d GasUsed:%d GasLimit:%d", cost, vm.Context.GasUsed, vm.Context.GasLimit))
		}
		vm.Context.GasUsed += cost

		switch ins {
		case opcodes.Nop:
		case opcodes.Unreachable:
			panic("wasm: unreachable executed")
		case opcodes.Select:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
			c := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])
			frame.IP += 12
			if c != 0 {
				frame.Regs[valueID] = a
			} else {
				frame.Regs[valueID] = b
			}
		case opcodes.I32Const:
			val := LE.Uint32(frame.Code[frame.IP : frame.IP+4])
			frame.IP += 4
			frame.Regs[valueID] = int64(val)
		case opcodes.I32Add:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			frame.Regs[valueID] = int64(a + b)
		case opcodes.I32Sub:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			frame.Regs[valueID] = int64(a - b)
		case opcodes.I32Mul:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			frame.Regs[valueID] = int64(a * b)
		case opcodes.I32DivS:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			if b == 0 {
				panic("integer division by zero")
			}

			if a == math.MinInt32 && b == -1 {
				panic("signed integer overflow")
			}

			frame.IP += 8
			frame.Regs[valueID] = int64(a / b)
		case opcodes.I32DivU:
			a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			if b == 0 {
				panic("integer division by zero")
			}

			frame.IP += 8
			frame.Regs[valueID] = int64(a / b)
		case opcodes.I32RemS:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			if b == 0 {
				panic("integer division by zero")
			}

			frame.IP += 8
			frame.Regs[valueID] = int64(a % b)
		case opcodes.I32RemU:
			a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			if b == 0 {
				panic("integer division by zero")
			}

			frame.IP += 8
			frame.Regs[valueID] = int64(a % b)
		case opcodes.I32And:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			frame.IP += 8
			frame.Regs[valueID] = int64(a & b)
		case opcodes.I32Or:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			frame.IP += 8
			frame.Regs[valueID] = int64(a | b)
		case opcodes.I32Xor:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			frame.IP += 8
			frame.Regs[valueID] = int64(a ^ b)
		case opcodes.I32Shl:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			frame.IP += 8
			frame.Regs[valueID] = int64(a << (b % 32))
		case opcodes.I32ShrS:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			frame.IP += 8
			frame.Regs[valueID] = int64(a >> (b % 32))
		case opcodes.I32ShrU:
			a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			frame.IP += 8
			frame.Regs[valueID] = int64(a >> (b % 32))
		case opcodes.I32Rotl:
			a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			frame.IP += 8
			frame.Regs[valueID] = int64(bits.RotateLeft32(a, int(b)))
		case opcodes.I32Rotr:
			a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			frame.IP += 8
			frame.Regs[valueID] = int64(bits.RotateLeft32(a, -int(b)))
		case opcodes.I32Clz:
			val := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

			frame.IP += 4
			frame.Regs[valueID] = int64(bits.LeadingZeros32(val))
		case opcodes.I32Ctz:
			val := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

			frame.IP += 4
			frame.Regs[valueID] = int64(bits.TrailingZeros32(val))
		case opcodes.I32PopCnt:
			val := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

			frame.IP += 4
			frame.Regs[valueID] = int64(bits.OnesCount32(val))
		case opcodes.I32EqZ:
			val := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

			frame.IP += 4
			if val == 0 {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I32Eq:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a == b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I32Ne:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a != b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I32LtS:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a < b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I32LtU:
			a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a < b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I32LeS:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a <= b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I32LeU:
			a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a <= b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I32GtS:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a > b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I32GtU:
			a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a > b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I32GeS:
			a := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a >= b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I32GeU:
			a := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a >= b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I64Const:
			val := LE.Uint64(frame.Code[frame.IP : frame.IP+8])
			frame.IP += 8
			frame.Regs[valueID] = int64(val)
		case opcodes.I64Add:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
			frame.IP += 8
			frame.Regs[valueID] = a + b
		case opcodes.I64Sub:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
			frame.IP += 8
			frame.Regs[valueID] = a - b
		case opcodes.I64Mul:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
			frame.IP += 8
			frame.Regs[valueID] = a * b
		case opcodes.I64DivS:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]

			if b == 0 {
				panic("integer division by zero")
			}

			if a == math.MinInt64 && b == -1 {
				panic("signed integer overflow")
			}

			frame.IP += 8
			frame.Regs[valueID] = a / b
		case opcodes.I64DivU:
			a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			if b == 0 {
				panic("integer division by zero")
			}

			frame.IP += 8
			frame.Regs[valueID] = int64(a / b)
		case opcodes.I64RemS:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]

			if b == 0 {
				panic("integer division by zero")
			}

			frame.IP += 8
			frame.Regs[valueID] = a % b
		case opcodes.I64RemU:
			a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			if b == 0 {
				panic("integer division by zero")
			}

			frame.IP += 8
			frame.Regs[valueID] = int64(a % b)
		case opcodes.I64And:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]

			frame.IP += 8
			frame.Regs[valueID] = a & b
		case opcodes.I64Or:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]

			frame.IP += 8
			frame.Regs[valueID] = a | b
		case opcodes.I64Xor:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]

			frame.IP += 8
			frame.Regs[valueID] = a ^ b
		case opcodes.I64Shl:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			frame.IP += 8
			frame.Regs[valueID] = a << (b % 64)
		case opcodes.I64ShrS:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			frame.IP += 8
			frame.Regs[valueID] = a >> (b % 64)
		case opcodes.I64ShrU:
			a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			frame.IP += 8
			frame.Regs[valueID] = int64(a >> (b % 64))
		case opcodes.I64Rotl:
			a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			frame.IP += 8
			frame.Regs[valueID] = int64(bits.RotateLeft64(a, int(b)))
		case opcodes.I64Rotr:
			a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])

			frame.IP += 8
			frame.Regs[valueID] = int64(bits.RotateLeft64(a, -int(b)))
		case opcodes.I64Clz:
			val := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

			frame.IP += 4
			frame.Regs[valueID] = int64(bits.LeadingZeros64(val))
		case opcodes.I64Ctz:
			val := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

			frame.IP += 4
			frame.Regs[valueID] = int64(bits.TrailingZeros64(val))
		case opcodes.I64PopCnt:
			val := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

			frame.IP += 4
			frame.Regs[valueID] = int64(bits.OnesCount64(val))
		case opcodes.I64EqZ:
			val := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])

			frame.IP += 4
			if val == 0 {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I64Eq:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
			frame.IP += 8
			if a == b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I64Ne:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
			frame.IP += 8
			if a != b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I64LtS:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
			frame.IP += 8
			if a < b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I64LtU:
			a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a < b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I64LeS:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
			frame.IP += 8
			if a <= b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I64LeU:
			a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a <= b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I64GtS:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
			frame.IP += 8
			if a > b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I64GtU:
			a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a > b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I64GeS:
			a := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			b := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
			frame.IP += 8
			if a >= b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.I64GeU:
			a := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			b := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))])
			frame.IP += 8
			if a >= b {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.F32Add:
			a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f32_add(C.float(a), C.float(b)))))
		case opcodes.F32Sub:
			a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f32_sub(C.float(a), C.float(b)))))
		case opcodes.F32Mul:
			a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f32_mul(C.float(a), C.float(b)))))
		case opcodes.F32Div:
			a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f32_div(C.float(a), C.float(b)))))
		case opcodes.F32Sqrt:
			val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float32bits(float32((C.platone_f32_sqrt(C.float(val))))))

		case opcodes.F32Min:
			a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f32_min(C.float(a), C.float(b)))))
		case opcodes.F32Max:
			a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f32_max(C.float(a), C.float(b)))))
		case opcodes.F32Ceil:
			val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f32_ceil(C.float(val)))))
		case opcodes.F32Floor:
			val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f32_floor(C.float(val)))))
		case opcodes.F32Trunc:
			val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			//frame.Regs[valueID] = int64(math.Float32bits(float32(math.Trunc(float64(val)))))
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f32_trunc(C.float(val)))))

		case opcodes.F32Nearest:
			val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			//frame.Regs[valueID] = int64(math.Float32bits(float32(math.RoundToEven(float64(val)))))
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f32_nearest(C.float(val)))))

		case opcodes.F32Abs:
			val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			//frame.Regs[valueID] = int64(math.Float32bits(float32(math.Abs(float64(val)))))
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f32_abs(C.float(val)))))

		case opcodes.F32Neg:
			val := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f32_neg(C.float(val)))))

		case opcodes.F32CopySign:
			a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//frame.Regs[valueID] = int64(math.Float32bits(float32(math.Copysign(float64(a), float64(b)))))
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f32_copysign(C.float(a), C.float(b)))))

			//-------------------------zbx------------------------------------------
			//-------------------------zbx------------------------------------------

		case opcodes.F32Eq:
			a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//if a == b {
			if C.platone_f32_eq(C.float(a), C.float(b)) {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.F32Ne:
			a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//if a != b {
			if C.platone_f32_ne(C.float(a), C.float(b)) {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.F32Lt:
			a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//if a < b {
			if C.platone_f32_lt(C.float(a), C.float(b)) {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.F32Le:
			a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//if a <= b {
			if C.platone_f32_le(C.float(a), C.float(b)) {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.F32Gt:
			a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//if a > b {
			if C.platone_f32_gt(C.float(a), C.float(b)) {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.F32Ge:
			a := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//if a >= b {
			if C.platone_f32_ge(C.float(a), C.float(b)) {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.F64Add:
			a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_add(C.double(a), C.double(b)))))
		case opcodes.F64Sub:
			a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_sub(C.double(a), C.double(b)))))
		case opcodes.F64Mul:
			a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_mul(C.double(a), C.double(b)))))
		case opcodes.F64Div:
			a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_div(C.double(a), C.double(b)))))
		case opcodes.F64Sqrt:
			val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_sqrt(C.double(val)))))
		case opcodes.F64Min:
			a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_min(C.double(a), C.double(b)))))
		case opcodes.F64Max:
			a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_max(C.double(a), C.double(b)))))
		case opcodes.F64Ceil:
			val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_ceil(C.double(val)))))
		case opcodes.F64Floor:
			val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_floor(C.double(val)))))
		case opcodes.F64Trunc:
			val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			//frame.Regs[valueID] = int64(math.Float64bits(math.Trunc(val)))
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_trunc(C.double(val)))))

		case opcodes.F64Nearest:
			val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			//frame.Regs[valueID] = int64(math.Float64bits(math.RoundToEven(val)))
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_nearest(C.double(val)))))

		case opcodes.F64Abs:
			val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			//frame.Regs[valueID] = int64(math.Float64bits(math.Abs(val)))
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_abs(C.double(val)))))

		case opcodes.F64Neg:
			val := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			//frame.Regs[valueID] = int64(math.Float64bits(-val))
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_neg(C.double(val)))))
		case opcodes.F64CopySign:
			a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//frame.Regs[valueID] = int64(math.Float64bits(math.Copysign(a, b)))
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f64_copysign(C.double(a), C.double(b)))))
		case opcodes.F64Eq:
			a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//if a == b {
			if C.platone_f64_eq(C.double(a), C.double(b)) {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.F64Ne:
			a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//if a != b {
			if C.platone_f64_ne(C.double(a), C.double(b)) {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.F64Lt:
			a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//if a < b {
			if C.platone_f64_lt(C.double(a), C.double(b)) {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.F64Le:
			a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//if a <= b {
			if C.platone_f64_le(C.double(a), C.double(b)) {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.F64Gt:
			a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//if a > b {
			if C.platone_f64_gt(C.double(a), C.double(b)) {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}
		case opcodes.F64Ge:
			a := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			b := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]))
			frame.IP += 8
			//if a >= b {
			if C.platone_f64_ge(C.double(a), C.double(b)) {
				frame.Regs[valueID] = 1
			} else {
				frame.Regs[valueID] = 0
			}

		case opcodes.I32WrapI64:
			v := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			frame.IP += 4
			frame.Regs[valueID] = int64(v)

		case opcodes.I32TruncSF32:
			v := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			//frame.Regs[valueID] = int64(int32(math.Trunc(float64(v))))
			frame.Regs[valueID] = int64(int32(C.platone_f32_trunc_i32s(C.float(v))))

		case opcodes.I32TruncUF32:
			v := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(int32(C.platone_f32_trunc_i32u(C.float(v))))

		case opcodes.I32TruncSF64:
			v := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			//frame.Regs[valueID] = int64(int32(math.Trunc(v)))
			frame.Regs[valueID] = int64(int32(C.platone_f64_trunc_i32s(C.double(v))))

		case opcodes.I32TruncUF64:
			v := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(int32(C.platone_f64_trunc_i32u(C.double(v))))

		case opcodes.I64TruncSF32:
			v := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			//frame.Regs[valueID] = int64(math.Trunc(float64(v)))
			frame.Regs[valueID] = int64(C.platone_f32_trunc_i64s(C.float(v)))

		case opcodes.I64TruncUF32:
			v := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(C.platone_f32_trunc_i64u(C.float(v)))

		case opcodes.I64TruncSF64:
			v := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			//frame.Regs[valueID] = int64(math.Trunc(v))
			frame.Regs[valueID] = int64(C.platone_f64_trunc_i64s(C.double(v)))

		case opcodes.I64TruncUF64:
			v := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(C.platone_f64_trunc_i64u(C.double(v)))

		case opcodes.F32DemoteF64:
			v := math.Float64frombits(uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_f64_demote(C.double(v)))))

		case opcodes.F64PromoteF32:
			v := math.Float32frombits(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_f32_promote(C.float(v)))))

		case opcodes.F32ConvertSI32:
			v := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_i32_to_f32(C.int32_t(v)))))

		case opcodes.F32ConvertUI32:
			v := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_ui32_to_f32(C.uint32_t(v)))))

		case opcodes.F32ConvertSI64:
			v := int64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_i64_to_f32(C.int64_t(v)))))

		case opcodes.F32ConvertUI64:
			v := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float32bits(float32(C.platone_ui64_to_f32(C.uint64_t(v)))))

		case opcodes.F64ConvertSI32:
			v := int32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_i32_to_f64(C.int32_t(v)))))

		case opcodes.F64ConvertUI32:
			v := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_ui32_to_f64(C.uint32_t(v)))))

		case opcodes.F64ConvertSI64:
			v := int64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_i64_to_f64(C.int64_t(v)))))

		case opcodes.F64ConvertUI64:
			v := uint64(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			frame.IP += 4
			frame.Regs[valueID] = int64(math.Float64bits(float64(C.platone_ui64_to_f64(C.uint64_t(v)))))

		case opcodes.I64ExtendUI32:
			v := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))])
			frame.IP += 4
			frame.Regs[valueID] = int64(v)

		case opcodes.I64ExtendSI32:
			v := int32(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4
			frame.Regs[valueID] = int64(v)

		case opcodes.I32Load, opcodes.I64Load32U:
			LE.Uint32(frame.Code[frame.IP : frame.IP+4])
			offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
			base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

			frame.IP += 12

			effective := int(uint64(base) + uint64(offset))
			frame.Regs[valueID] = int64(uint32(LE.Uint32(vm.Memory.Memory[effective : effective+4])))
		case opcodes.I64Load32S:
			LE.Uint32(frame.Code[frame.IP : frame.IP+4])
			offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
			base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

			frame.IP += 12

			effective := int(uint64(base) + uint64(offset))
			frame.Regs[valueID] = int64(int32(LE.Uint32(vm.Memory.Memory[effective : effective+4])))
		case opcodes.I64Load:
			LE.Uint32(frame.Code[frame.IP : frame.IP+4])
			offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
			base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

			frame.IP += 12

			effective := int(uint64(base) + uint64(offset))
			frame.Regs[valueID] = int64(LE.Uint64(vm.Memory.Memory[effective : effective+8]))
		case opcodes.I32Load8S, opcodes.I64Load8S:
			LE.Uint32(frame.Code[frame.IP : frame.IP+4])
			offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
			base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

			frame.IP += 12

			effective := int(uint64(base) + uint64(offset))
			frame.Regs[valueID] = int64(int8(vm.Memory.Memory[effective]))
		case opcodes.I32Load8U, opcodes.I64Load8U:
			LE.Uint32(frame.Code[frame.IP : frame.IP+4])
			offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
			base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

			frame.IP += 12

			effective := int(uint64(base) + uint64(offset))
			frame.Regs[valueID] = int64(uint8(vm.Memory.Memory[effective]))
		case opcodes.I32Load16S, opcodes.I64Load16S:
			LE.Uint32(frame.Code[frame.IP : frame.IP+4])
			offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
			base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

			frame.IP += 12

			effective := int(uint64(base) + uint64(offset))
			frame.Regs[valueID] = int64(int16(LE.Uint16(vm.Memory.Memory[effective : effective+2])))
		case opcodes.I32Load16U, opcodes.I64Load16U:
			LE.Uint32(frame.Code[frame.IP : frame.IP+4])
			offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
			base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

			frame.IP += 12

			effective := int(uint64(base) + uint64(offset))
			frame.Regs[valueID] = int64(uint16(LE.Uint16(vm.Memory.Memory[effective : effective+2])))
		case opcodes.I32Store, opcodes.I64Store32:
			LE.Uint32(frame.Code[frame.IP : frame.IP+4])
			offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
			base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

			value := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+12:frame.IP+16]))]

			frame.IP += 16

			effective := int(uint64(base) + uint64(offset))
			LE.PutUint32(vm.Memory.Memory[effective:effective+4], uint32(value))
		case opcodes.I64Store:
			LE.Uint32(frame.Code[frame.IP : frame.IP+4])
			offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
			base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

			value := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+12:frame.IP+16]))]

			frame.IP += 16

			effective := int(uint64(base) + uint64(offset))
			LE.PutUint64(vm.Memory.Memory[effective:effective+8], uint64(value))
		case opcodes.I32Store8, opcodes.I64Store8:
			LE.Uint32(frame.Code[frame.IP : frame.IP+4])
			offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
			base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

			value := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+12:frame.IP+16]))]

			frame.IP += 16

			effective := int(uint64(base) + uint64(offset))
			vm.Memory.Memory[effective] = byte(value)
		case opcodes.I32Store16, opcodes.I64Store16:
			LE.Uint32(frame.Code[frame.IP : frame.IP+4])
			offset := LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8])
			base := uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP+8:frame.IP+12]))])

			value := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+12:frame.IP+16]))]

			frame.IP += 16

			effective := int(uint64(base) + uint64(offset))
			LE.PutUint16(vm.Memory.Memory[effective:effective+2], uint16(value))

		case opcodes.Jmp:
			target := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			vm.Yielded = frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
			frame.IP = target
		case opcodes.JmpEither:
			targetA := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			targetB := int(LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8]))
			cond := int(LE.Uint32(frame.Code[frame.IP+8 : frame.IP+12]))
			yieldedReg := int(LE.Uint32(frame.Code[frame.IP+12 : frame.IP+16]))
			frame.IP += 16

			vm.Yielded = frame.Regs[yieldedReg]
			if frame.Regs[cond] != 0 {
				frame.IP = targetA
			} else {
				frame.IP = targetB
			}
		case opcodes.JmpIf:
			target := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			cond := int(LE.Uint32(frame.Code[frame.IP+4 : frame.IP+8]))
			yieldedReg := int(LE.Uint32(frame.Code[frame.IP+8 : frame.IP+12]))
			frame.IP += 12
			if frame.Regs[cond] != 0 {
				vm.Yielded = frame.Regs[yieldedReg]
				frame.IP = target
			}
		case opcodes.JmpTable:
			targetCount := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			frame.IP += 4

			targetsRaw := frame.Code[frame.IP : frame.IP+4*targetCount]
			frame.IP += 4 * targetCount

			defaultTarget := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			frame.IP += 4

			cond := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			frame.IP += 4

			vm.Yielded = frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			frame.IP += 4

			val := int(frame.Regs[cond])
			if val >= 0 && val < targetCount {
				frame.IP = int(LE.Uint32(targetsRaw[val*4 : val*4+4]))
			} else {
				frame.IP = defaultTarget
			}
		case opcodes.ReturnValue:
			val := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			frame.Destroy(vm)
			vm.CurrentFrame--
			if vm.CurrentFrame == -1 {
				vm.Exited = true
				vm.ReturnValue = val
				return
			} else {
				frame = vm.GetCurrentFrame()
				frame.Regs[frame.ReturnReg] = val
			}
		case opcodes.ReturnVoid:
			frame.Destroy(vm)
			vm.CurrentFrame--
			if vm.CurrentFrame == -1 {
				vm.Exited = true
				vm.ReturnValue = 0
				return
			} else {
				frame = vm.GetCurrentFrame()
			}
		case opcodes.GetLocal:
			id := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			val := frame.Locals[id]
			frame.IP += 4
			frame.Regs[valueID] = val
		case opcodes.SetLocal:
			id := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			val := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
			frame.IP += 8
			frame.Locals[id] = val
		case opcodes.GetGlobal:
			frame.Regs[valueID] = vm.Globals[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			frame.IP += 4
		case opcodes.SetGlobal:
			id := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			val := frame.Regs[int(LE.Uint32(frame.Code[frame.IP+4:frame.IP+8]))]
			frame.IP += 8

			vm.Globals[id] = val
		case opcodes.Call:
			functionID := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			frame.IP += 4
			argCount := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			frame.IP += 4
			argsRaw := frame.Code[frame.IP : frame.IP+4*argCount]
			frame.IP += 4 * argCount

			oldRegs := frame.Regs
			frame.ReturnReg = valueID

			vm.CurrentFrame++
			frame = vm.GetCurrentFrame()
			frame.Init(vm, functionID, vm.FunctionCode[functionID])
			for i := 0; i < argCount; i++ {
				frame.Locals[i] = oldRegs[int(LE.Uint32(argsRaw[i*4:i*4+4]))]
			}

		case opcodes.CallIndirect:
			typeID := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			frame.IP += 4
			argCount := int(LE.Uint32(frame.Code[frame.IP:frame.IP+4])) - 1
			frame.IP += 4
			argsRaw := frame.Code[frame.IP : frame.IP+4*argCount]
			frame.IP += 4 * argCount
			tableItemID := frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]
			frame.IP += 4

			sig := &vm.Module.Base.Types.Entries[typeID]

			functionID := int(vm.Table[tableItemID])
			code := vm.FunctionCode[functionID]

			// TODO: We are only checking CC here; Do we want strict typeck?
			if code.NumParams != len(sig.ParamTypes) || code.NumReturns != len(sig.ReturnTypes) {
				panic("type mismatch")
			}

			oldRegs := frame.Regs
			frame.ReturnReg = valueID

			vm.CurrentFrame++
			frame = vm.GetCurrentFrame()
			frame.Init(vm, functionID, code)
			for i := 0; i < argCount; i++ {
				frame.Locals[i] = oldRegs[int(LE.Uint32(argsRaw[i*4:i*4+4]))]
			}

		case opcodes.InvokeImport:
			importID := int(LE.Uint32(frame.Code[frame.IP : frame.IP+4]))
			frame.IP += 4
			vm.Delegate = func() {
				frame.Regs[valueID] = vm.FunctionImports[importID].Execute(vm)
			}
			return

		case opcodes.CurrentMemory:
			frame.Regs[valueID] = int64(len(vm.Memory.Memory) / DefaultPageSize)

		case opcodes.GrowMemory:
			n := int(uint32(frame.Regs[int(LE.Uint32(frame.Code[frame.IP:frame.IP+4]))]))
			frame.IP += 4

			current := len(vm.Memory.Memory) / DefaultPageSize
			if vm.Context.Config.MaxMemoryPages == 0 || (current+n >= current && current+n <= vm.Context.Config.MaxMemoryPages) {
				frame.Regs[valueID] = int64(current)
				vm.Memory.Memory = append(vm.Memory.Memory, make([]byte, n*DefaultPageSize)...)
			} else {
				frame.Regs[valueID] = -1
			}

		case opcodes.Phi:
			frame.Regs[valueID] = vm.Yielded

		case opcodes.AddGas:
			delta := LE.Uint64(frame.Code[frame.IP : frame.IP+8])
			frame.IP += 8
			vm.AddAndCheckGas(delta)
		default:
			panic("unknown instruction")

		}
	}
}
