package runtime

import (
	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/core"
	"github.com/Venachain/Venachain/core/vm"
)

// create env for wasmvm
func NewEnv(cfg *Config) *vm.EVM {
	context := vm.Context{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     func(uint64) common.Hash { return common.Hash{} },

		Origin:      cfg.Origin,
		Coinbase:    cfg.Coinbase,
		BlockNumber: cfg.BlockNumber,
		Time:        cfg.Time,
		GasLimit:    cfg.GasLimit,
		GasPrice:    cfg.GasPrice,
	}

	return vm.NewEVM(context, cfg.State, cfg.ChainConfig, cfg.EVMConfig)
}
