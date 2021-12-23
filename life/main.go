package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/Venachain/Venachain/life/exec"
	"github.com/Venachain/Venachain/life/resolver"
	"github.com/Venachain/Venachain/log"
)

func main() {
	//entryFunctionFlag := flag.String("entry", "app_main", "entry function id")
	//dynamicPages := flag.Int("dynamicPages", 1, "dynamic memory pages")

	//jitFlag := flag.Bool("jit", false, "enable jit")
	//flag.Parse()

	// mocking test
	flag := false
	pages := 64
	functionFlag := "transfer"
	jitFlag := &flag
	dynamicPages := &pages
	entryFunctionFlag := &functionFlag

	rl := resolver.NewResolver(0x01)
	// Read WebAssembly *.wasm file.
	//input, err := ioutil.ReadFile(flag.Arg(0))
	input, err := ioutil.ReadFile("D:\\repos\\Venachain-contract\\build\\hello\\hello.wasm")
	if err != nil {
		panic(err)
	}

	rootLog := log.New()
	rootLog.SetHandler(log.StderrHandler)

	// Instantiate a new WebAssembly VM with a few resolved imports.
	vm, err := exec.NewVirtualMachine(input, &exec.VMContext{
		Config: exec.VMConfig{
			EnableJIT:          *jitFlag,
			DefaultMemoryPages: 128,
			DefaultTableSize:   65536,
			DynamicMemoryPages: *dynamicPages,
		},
		Addr:     [20]byte{},
		GasUsed:  0,
		GasLimit: 20000000,
		Log:      rootLog,
	}, rl, nil)

	if err != nil {
		panic(err)
	}

	*entryFunctionFlag = "transfer"
	// Get the function ID of the entry function to be executed.
	entryID, ok := vm.GetFunctionExport(*entryFunctionFlag)
	if !ok {
		entryID = 0
	}

	start := time.Now()

	// If any function prior to the entry function was declared to be
	// called by the module, run it first.
	if vm.Module.Base.Start != nil {
		startID := int(vm.Module.Base.Start.Index)
		_, err := vm.Run(startID)
		if err != nil {
			vm.PrintStackTrace()
			panic(err)
		}
	}

	// Run the WebAssembly module's entry function.
	ret, err := vm.Run(entryID, resolver.MallocString(vm, "hello"), resolver.MallocString(vm, "world"), 45)
	if err != nil {
		vm.PrintStackTrace()
		panic(err)
	}
	end := time.Now()

	rootLog.Info(fmt.Sprintf("return value = %d, duration = %v  gasUsed:%d \n", ret, end.Sub(start), vm.Context.GasUsed))
}
