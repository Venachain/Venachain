// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"errors"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/consensus"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/core/vm"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"github.com/PlatONEnetwork/PlatONE-Go/rpc"
)

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	signer   types.Signer
	config   *params.ChainConfig // Chain configuration options
	bc       *BlockChain         // Canonical block chain
	engine   consensus.Engine    // Consensus engine used for block rewards
	timeout  time.Duration
	poolSize int
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *params.ChainConfig, bc *BlockChain, engine consensus.Engine) *StateProcessor {
	size := common.GetParallelPoolSize()
	if size == 0 {
		size = runtime.NumCPU() * 2
	}
	return &StateProcessor{
		signer:   types.NewEIP155Signer(config.ChainID),
		config:   config,
		bc:       bc,
		engine:   engine,
		timeout:  10 * time.Second,
		poolSize: size,
	}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to
// the processor (coinbase).
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (*types.Block, types.Receipts, []*types.Log, uint64, error) {
	header := block.Header()
	header.TxHash = common.Hash{}
	header.ReceiptHash = common.Hash{}
	if len(block.Dag()) != 0 && len(block.Transactions()) != 0 {
		receipts, logs, err := p.ParallelProcessTxsWithDag(block, statedb, false)
		if err != nil {
			return nil, nil, nil, 0, err
		}
		header.TxHash = statedb.GetTxHash()
		header.ReceiptHash = statedb.GetReceiptHash()
		// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
		cblock, err := p.engine.Finalize(p.bc, header, statedb, block.Transactions(), receipts, block.Dag())
		if err != nil {
			return nil, nil, nil, 0, err
		}
		return cblock, receipts, logs, statedb.GetGasUsed(), nil
	}

	var (
		receipts types.Receipts
		usedGas  = new(uint64)
		allLogs  []*types.Log
		gp       = new(GasPool).AddGas(block.GasLimit())
	)

	// Iterate over and process the individual transactios
	for i, tx := range block.Transactions() {
		rpc.MonitorWriteData(rpc.TransactionExecuteStartTime, tx.Hash().String(), "", p.bc.extdb)
		txHash := tx.Hash()
		statedb.Prepare(txHash, block.Hash(), i)
		log.Trace("Perform Transaction", "txHash", fmt.Sprintf("%x", txHash[:log.LogHashLen]), "blockNumber", block.Number())
		receipt, _, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		rpc.MonitorWriteData(rpc.TransactionExecuteEndTime, tx.Hash().String(), "", p.bc.extdb)
		if err != nil {
			rpc.MonitorWriteData(rpc.TransactionExecuteStatus, tx.Hash().String(), "false", p.bc.extdb)
			return nil, nil, nil, 0, err
		}
		rpc.MonitorWriteData(rpc.TransactionExecuteStatus, tx.Hash().String(), "true", p.bc.extdb)
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, receipt.Logs...)
	}
	// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
	cblock, err := p.engine.Finalize(p.bc, header, statedb, block.Transactions(), receipts, block.Dag())
	if err != nil {
		return nil, nil, nil, 0, err
	}
	return cblock, receipts, allLogs, *usedGas, nil
}

func (p *StateProcessor) CheckAndProcess(block *types.Block, statedb *state.StateDB, cfg vm.Config) (*types.Block, types.Receipts, error) {
	header := block.Header()
	header.TxHash = common.Hash{}
	header.ReceiptHash = common.Hash{}
	if len(block.Dag()) != 0 && len(block.Transactions()) != 0 {
		receipts, _, err := p.ParallelProcessTxsWithDag(block, statedb, true)
		if err != nil {
			return nil, nil, err
		}
		header.TxHash = statedb.GetTxHash()
		header.ReceiptHash = statedb.GetReceiptHash()
		// Finalize the block, applying any consensus engine specific extras (e.g. block rewards)
		cblock, err := p.engine.Finalize(p.bc, header, statedb, block.Transactions(), receipts, block.Dag())
		if err != nil {
			return nil, nil, err
		}
		return cblock, receipts, nil
	}

	var (
		receipts types.Receipts
		usedGas  = new(uint64)
		gp       = new(GasPool).AddGas(block.GasLimit())
	)

	txsMap := make(map[common.Hash]struct{})

	for i, tx := range block.Transactions() {
		statedb.Prepare(tx.Hash(), common.Hash{}, i)
		snap := statedb.Snapshot()
		if r := p.bc.GetReceiptsByHash(tx.Hash()); r != nil {
			return nil, nil, errors.New("Already executed tx")
		}
		if _, ok := txsMap[tx.Hash()]; ok {
			return nil, nil, errors.New("Repeated tx in one block")
		} else {
			txsMap[tx.Hash()] = struct{}{}
		}
		receipt, _, err := ApplyTransaction(p.config, p.bc, nil, gp, statedb, header, tx, usedGas, cfg)
		if err != nil {
			statedb.RevertToSnapshot(snap)
			return nil, nil, err
		}
		receipts = append(receipts, receipt)
	}

	cblock, err := p.engine.Finalize(p.bc, header, statedb, block.Transactions(), receipts, block.Dag())
	if err != nil {
		return nil, nil, err
	}
	return cblock, receipts, nil
}

func (p *StateProcessor) ParallelProcessTxs(stateDb *state.StateDB, header *types.Header, txs types.Transactions) error {
	log.Debug("Parallel Process Txs start")

	txsCount := txs.Len()
	txCh := make(chan *types.Transaction, txsCount)
	applyCh := make(chan *state.TxSimulator, txsCount)
	stopCh := make(chan bool)
	finishCh := make(chan bool)
	stopApplyCh := make(chan bool)
	timeCh := time.After(p.timeout)
	gp := new(GasPool).AddGas(header.GasLimit)
	stateDb.StartProcess()
	var errCnt int32 = 0

	goRoutinePool, err := ants.NewPool(p.poolSize, ants.WithOptions(ants.Options{
		PreAlloc: true,
		PanicHandler: func(i interface{}) {
			log.Error(fmt.Sprintf("worker exits from a panic: %v\n", i))
			atomic.AddInt32(&errCnt, 1)
		},
	}))
	if err != nil {
		return err
	}
	defer goRoutinePool.Release()

	go func() {
		for {
			select {
			case tx := <-txCh:
				err := goRoutinePool.Submit(func() {
					if !stateDb.IsProcess() {
						return
					}
					//执行模拟交易
					txSim, err := p.SimulateTx(stateDb, tx, header, gp)
					if err != nil {
						log.Warn("Transaction failed, skipped", "blockNumber", header.Number,
							"blockParentHash", header.ParentHash, "hash", tx.Hash(), "err", err)
						atomic.AddInt32(&errCnt, 1)
						return
					}

					if txSim.ReTry() {
						//需要retry的交易等待一段时间，保证前序交易已经处理完成，再重新执行
						time.Sleep(5 * time.Millisecond)
						txCh <- tx
						return
					}
					result, count := stateDb.AddTxSim(txSim, applyCh, false)
					// 严重冲突的交易需要重新执行
					if !result {
						log.Debug("get conflict retry")
						txCh <- tx
					} else {
						//log.Info("add txSim :", "hash", txSim.GetHash())
					}
					if count >= (txsCount - int(atomic.LoadInt32(&errCnt))) {
						finishCh <- true
					}
				})
				if err != nil {
					log.Warn("failed to submit tx  %s during Parallel Process Txs, %+v", tx.Hash(), err)
				}
			case <-timeCh:
				stopApplyCh <- true
				log.Debug("Parallel Process Txs reached time limit")
				return
			case <-finishCh:
				log.Debug("AddTxSim complete")
				stopApplyCh <- true
				return
			}
		}
	}()

	go func() {
		var applyCnt int
		var addCnt int
		for {
			select {
			case txSim := <-applyCh:
				stateDb.ApplyTxSim(txSim, true)
				applyCnt++
				if applyCnt >= (txsCount-int(atomic.LoadInt32(&errCnt))) || addCnt == applyCnt {
					log.Debug("applyTxSim complete")
					stopCh <- true
					return
				}
			case <-stopApplyCh:

				stateDb.StopProcess()
				addCnt = stateDb.GetTxsLen()
				if addCnt == applyCnt {
					log.Debug("stopApplyCh", "time", time.Now().Format("2006-01-02 15:04:05.000"))
					stopCh <- true
					return
				}
			}
		}
	}()

	go func() {
		for _, tx := range txs {
			txCh <- tx
		}
	}()

	// 等待交易的执行完成 或者执行超时
	<-stopCh
	stateDb.StopProcess()
	stateDb.UpdateDirtyObject()
	header.GasUsed = stateDb.GetGasUsed()
	log.Info("Parallel Process Txs stop", "txCount", stateDb.GetTxsLen(), "gasUsed", header.GasUsed)
	return nil
}

func (p *StateProcessor) SimulateTx(stateDb *state.StateDB, tx *types.Transaction, header *types.Header, pool *GasPool) (*state.TxSimulator, error) {
	txSim := state.NewTxSimulator(stateDb, tx)
	receipt, _, err := ApplyTransactionForSimulator(p.config, p.bc, pool, txSim, header, tx, p.bc.vmConfig)
	if err != nil {
		return nil, err
	}
	txSim.SetReceipt(receipt)
	return txSim, nil
}

//ParallelProcessTxsWithDag 根据区块中的dag信息并行执行交易
func (p *StateProcessor) ParallelProcessTxsWithDag(block *types.Block, statedb *state.StateDB, check bool) (types.Receipts, []*types.Log, error) {
	log.Debug("Parallel Process Txs with dag start")
	count := len(block.Dag())
	if count != block.Transactions().Len() {
		return nil, nil, fmt.Errorf("the length of dag does not equal the length of txs")
	}
	runIndexs, txLeft, txDependency := initDag(block.Dag())

	if len(runIndexs) == 0 {
		return nil, nil, fmt.Errorf("cycle dependency error")
	}

	header := block.Header()
	gp := new(GasPool).AddGas(header.GasLimit)
	runCh := make(chan int, count)                  //用于传递待运行的交易
	completeCh := make(chan int, count)             // 用于传递已经完成的交易
	timeoutCh := time.After(p.timeout)              // 超时
	finishCh := make(chan bool)                     // 交易全部执行完成
	stopApplyCh := make(chan bool)                  // 停止应用模拟交易
	stopCh := make(chan bool)                       // 用于等待所有交易的执行完成或者执行错误
	errCh := make(chan error)                       // 用于传递错误信息
	var ProcessErr error                            // 用于记录返回的错误信息
	applyCh := make(chan *state.TxSimulator, count) // 用于传递已经执行完成的模拟交易
	txs := block.Transactions()
	receipts := make([]*types.Receipt, count)
	var allLogs []*types.Log
	txsMap := make(map[common.Hash]struct{})

	goRoutinePool, err := ants.NewPool(p.poolSize, ants.WithOptions(ants.Options{
		PreAlloc: true,
		PanicHandler: func(i interface{}) {
			log.Error(fmt.Sprintf("worker exits from a panic: %v\n", i))
			errCh <- fmt.Errorf("worker exits from a panic: %v\n", i)
			return
		},
	}))
	if err != nil {
		return nil, nil, err
	}
	defer goRoutinePool.Release()
	statedb.StartProcess()
	go func() {
		for {
			select {
			case txIndex := <-runCh:
				tx := txs[txIndex]
				err := goRoutinePool.Submit(func() {
					if check {
						if r := p.bc.GetReceiptsByHash(tx.Hash()); r != nil {
							errCh <- errors.New("already executed transaction")
							return
						}
					}
					if !statedb.IsProcess() {
						return
					}
					//执行模拟交易
					txSim, err := p.SimulateTx(statedb, tx, header, gp)
					txSim.SetIndex(txIndex)
					if err != nil {
						log.Warn("Transaction failed, skipped", "blockNumber", header.Number,
							"blockParentHash", header.ParentHash, "hash", tx.Hash(), "err", err)
						errCh <- err
						return
					}

					if txSim.ReTry() {
						//需要retry的交易等待一段时间，保证前序交易已经处理完成，再重新执行
						time.Sleep(5 * time.Millisecond)
						runCh <- txIndex
						return
					}

					result, addCount := statedb.AddTxSim(txSim, applyCh, true)
					if !result {
						errCh <- fmt.Errorf("DAG is not correctly")
						return
					}
					completeCh <- txIndex
					if addCount >= count {
						finishCh <- true
					}
				})
				if err != nil {
					errCh <- err
					log.Warn("failed to submit tx  %s during Parallel Process Txs, %+v", tx.Hash(), err)
				}
			case txIndex := <-completeCh:
				runs := getNoneDependency(txIndex, txLeft, txDependency)
				log.Debug("add index to run ", "value ", runs)
				if len(runs) != 0 {
					for _, v := range runs {
						runCh <- v
					}
				}
			case err := <-errCh:
				log.Error("ParallelProcessTxsWithDag error", "err", err)
				ProcessErr = err
				stopApplyCh <- true
				return
			case <-finishCh:
				log.Debug("ParallelProcessTxsWithDag add Simulate finish correctly")
				stopApplyCh <- true
				return
			case <-timeoutCh:
				log.Debug("ParallelProcessTxsWithDag timeout")
				ProcessErr = fmt.Errorf("ParallelProcessTxsWithDag timeout: %s", p.timeout.String())
				stopApplyCh <- true
				return
			}
		}
	}()

	go func() {
		var applyCnt int
		for {
			select {
			case txSim := <-applyCh:
				if check {
					if _, ok := txsMap[txSim.GetHash()]; ok {
						ProcessErr = errors.New("repeated tx in one block")
						stopCh <- true
						return
					} else {
						txsMap[txSim.GetHash()] = struct{}{}
					}
				}
				statedb.ApplyTxSim(txSim, false)
				receipts[txSim.GetIndex()] = txSim.GetReceipt()
				allLogs = append(allLogs, txSim.GetReceipt().Logs...)
				applyCnt++
				if applyCnt >= count {
					stopCh <- true
					return
				}
			case <-stopApplyCh:
				if ProcessErr != nil || count == applyCnt {
					stopCh <- true
					return
				}
			}
		}
	}()

	go func() {
		for _, index := range runIndexs {
			runCh <- index
		}
	}()

	<-stopCh
	statedb.StopProcess()
	statedb.UpdateDirtyObject()
	if ProcessErr == nil {
		ProcessErr = CalculateCumulativeGasUsed(receipts)
		if ProcessErr == nil {
			statedb.UpdateTrie(receipts, txs)
		}
	}
	log.Info("Parallel Process Txs with dag stop", "err", ProcessErr, "count", len(receipts))
	return receipts, allLogs, ProcessErr
}

type txNeed map[int]struct{}

//initDag 初始化，根据dag信息获得三种数据：返回1：无依赖的交易index集合 2：剩余有依赖的交易index集合 3：所有被依赖的关系
func initDag(dag types.DAG) ([]int, map[int]txNeed, map[int]txNeed) {
	var runIndexs []int                  //待执行的无依赖的数据
	txLeft := make(map[int]txNeed)       // 剩余有依赖的数据 value是依赖的前序交易的集合
	txDependency := make(map[int]txNeed) //所有依赖的关系  value是被依赖的后续交易的集合
	for index, dependency := range dag {
		if len(dependency) == 0 {
			runIndexs = append(runIndexs, index)
		} else {
			need := make(txNeed)
			for _, v := range dependency {
				vint := int(v)
				//根据dependency中的数据设置依赖的前序交易
				need[vint] = struct{}{}
				//优先检查前序交易vint 是否有其他的后续交易集合，如果有的话直接加入该集合，否则创建集合再加入
				if data, ok := txDependency[vint]; ok {
					data[index] = struct{}{}
				} else {
					data = make(txNeed)
					data[index] = struct{}{}
					txDependency[vint] = data
				}
			}
			//设置待执行的且存在前序交易的数据
			txLeft[index] = need
		}
	}
	return runIndexs, txLeft, txDependency
}

func getNoneDependency(index int, left map[int]txNeed, all map[int]txNeed) []int {
	var result []int
	//先通过all所有被依赖数据中，获取对应的被依赖的后续交易数据，如果不存在则表示index不影响后续的依赖 直接返回空数组
	if need, ok := all[index]; ok {
		//存在后续被依赖的交易,进行遍历
		for k, _ := range need {
			// 获取被依赖的交易 所依赖的集合
			dep := left[k]
			// 删除集合里的index
			delete(dep, index)
			// 如果集合为空，表示交易k所依赖的前序交易都已经执行完成
			if len(dep) == 0 {
				//将交易k放入result中作为返回内容
				result = append(result, k)
				//从left中删除k的依赖集合
				delete(left, k)
			}
		}
	}
	return result
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransaction(config *params.ChainConfig, bc ChainContext, author *common.Address, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *uint64, cfg vm.Config) (*types.Receipt, uint64, error) {
	var from common.Address
	var gas uint64
	var gasPrice int64
	var failed bool
	var err error
	signer := types.MakeSigner(config)
	to := common.Address{}
	if tx.To() != nil {
		to = *tx.To()
	}
	if (tx.Data() == nil || len(tx.Data()) == 0) && statedb.GetCode(to) == nil {
		value := tx.Value()
		from, _ = types.Sender(signer, tx)
		if statedb.GetBalance(from).Cmp(value) < 0 {
			failed = true
			err = vm.ErrInsufficientBalance
		} else {
			statedb.AddNonce(from)
			statedb.SubBalance(from, value)
			statedb.AddBalance(to, value)
			failed = false
			err = nil
		}
		gp.AddGas(params.TxGas)
		gas = params.TxGas
		gasPrice = 0
	} else {
		var msg *types.Message
		if header.Number.Uint64() <= common.SysCfg.ReplayParam.Pivot {
			msg, err = tx.OldAsMessage()
			if err == types.ErrInvalidOldTrx {
				msg, err = tx.AsMessage(signer)
			}
		} else {
			msg, err = tx.AsMessage(signer)
		}

		// Replay situation,reflect address
		if header.Number.Uint64() < common.SysCfg.ReplayParam.Pivot && msg.To() != nil {
			if n := common.SysCfg.ReplayParam.OldSysContracts[*msg.To()]; n != "" {
				msg.SetTo(vm.CnsSysContractsMap[n])
			}
			//else if msg.TxType() == types.CnsTxType {
			//	msg.SetTo(syscontracts.CnsInvokeAddress)
			//} else if msg.TxType() == types.FwTxType {
			//	msg.SetTo(syscontracts.FirewallManagementAddress)
			//}
		}
		from = msg.From()
		if err != nil {
			return nil, 0, err
		}

		// Create a new context to be used in the EVM environment
		context := NewEVMContext(msg, header, bc, author)
		// Create a new environment which holds all relevant information
		// about the transaction and calling mechanisms.
		vmenv := vm.NewEVM(context, statedb, config, cfg)
		// Apply the transaction to the current state (included in the env)
		_, gas, gasPrice, failed, err = ApplyMessage(vmenv, msg, gp)
	}

	if err != nil {
		switch err {
		case vm.PermissionErr:
			data := [][]byte{}
			data = append(data, []byte(err.Error()))
			encodeData, _ := rlp.EncodeToBytes(data)
			topics := []common.Hash{common.BytesToHash(crypto.Keccak256([]byte("contract permission")))}
			log := &types.Log{
				Address:     from,
				Topics:      topics,
				Data:        encodeData,
				BlockNumber: header.Number.Uint64(),
			}
			statedb.AddLog(log)
		default:
			return nil, 0, err
		}
	}

	if common.SysCfg.GetIsTxUseGas() {
		data := [][]byte{}
		data = append(data, []byte(common.Int64ToBytes(gasPrice)))
		encodeData, _ := rlp.EncodeToBytes(data)
		topics := []common.Hash{common.BytesToHash(crypto.Keccak256([]byte("GasPrice")))}
		log := &types.Log{
			Address:     from,
			Topics:      topics,
			Data:        encodeData,
			BlockNumber: header.Number.Uint64(),
		}
		statedb.AddLog(log)
	}
	// Update the state with pending changes
	var root []byte
	statedb.Finalise(true)
	*usedGas += gas

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing whether the root touch-delete accounts.
	receipt := types.NewReceipt(root, failed, *usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if tx.To() == nil && err == nil {
		receipt.ContractAddress = crypto.CreateAddress(from, statedb.GetNonce(from)-1)
	}
	// Set the receipt logs and create a bloom for filtering

	receipt.Logs = statedb.GetLogs(tx.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	return receipt, gas, nil
}

// ApplyTransactionForSimulator attempts to apply a transaction to the given state database
// and uses the input parameters for its environment. It returns the receipt
// for the transaction, gas used and an error if the transaction failed,
// indicating the block was invalid.
func ApplyTransactionForSimulator(config *params.ChainConfig, bc ChainContext, gp *GasPool, txSim *state.TxSimulator, header *types.Header, tx *types.Transaction, cfg vm.Config) (*types.Receipt, uint64, error) {
	var from common.Address
	var gas uint64
	var gasPrice int64
	var failed bool
	var err error
	signer := types.MakeSigner(config)
	to := common.Address{}
	if tx.To() != nil {
		to = *tx.To()
	}
	if (tx.Data() == nil || len(tx.Data()) == 0) && txSim.GetCode(to) == nil {
		value := tx.Value()
		from, _ = types.Sender(signer, tx)
		if txSim.GetBalance(from).Cmp(value) < 0 {
			failed = true
			err = vm.ErrInsufficientBalance
			return nil, 0, err
		} else {
			txSim.AddNonce(from)
			txSim.SubBalance(from, value)
			txSim.AddBalance(to, value)
			failed = false
			err = nil
		}
		gp.AddGas(params.TxGas)
		gas = params.TxGas
		gasPrice = 0
	} else {
		var msg *types.Message
		if header.Number.Uint64() <= common.SysCfg.ReplayParam.Pivot {
			msg, err = tx.OldAsMessage()
			if err == types.ErrInvalidOldTrx {
				msg, err = tx.AsMessage(signer)
			}
		} else {
			msg, err = tx.AsMessage(signer)
		}

		// Replay situation,reflect address
		if header.Number.Uint64() < common.SysCfg.ReplayParam.Pivot && msg.To() != nil {
			if n := common.SysCfg.ReplayParam.OldSysContracts[*msg.To()]; n != "" {
				msg.SetTo(vm.CnsSysContractsMap[n])
			}

		}
		from = msg.From()
		if err != nil {
			return nil, 0, err
		}

		// Create a new context to be used in the EVM environment
		context := NewEVMContext(msg, header, bc, nil)
		// Create a new environment which holds all relevant information
		// about the transaction and calling mechanisms.
		vmenv := vm.NewEVM(context, txSim, config, cfg)
		// Apply the transaction to the current state (included in the env)
		_, gas, gasPrice, failed, err = ApplyMessage(vmenv, msg, gp)
	}

	if err != nil {
		switch err {
		case vm.PermissionErr:
			data := [][]byte{}
			data = append(data, []byte(err.Error()))
			encodeData, _ := rlp.EncodeToBytes(data)
			topics := []common.Hash{common.BytesToHash(crypto.Keccak256([]byte("contract permission")))}
			log := &types.Log{
				Address:     from,
				Topics:      topics,
				Data:        encodeData,
				BlockNumber: header.Number.Uint64(),
			}
			txSim.AddLog(log)
		default:
			return nil, 0, err
		}
	}

	if common.SysCfg.GetIsTxUseGas() {
		data := [][]byte{}
		data = append(data, []byte(common.Int64ToBytes(gasPrice)))
		encodeData, _ := rlp.EncodeToBytes(data)
		topics := []common.Hash{common.BytesToHash(crypto.Keccak256([]byte("GasPrice")))}
		log := &types.Log{
			Address:     from,
			Topics:      topics,
			Data:        encodeData,
			BlockNumber: header.Number.Uint64(),
		}
		txSim.AddLog(log)
	}
	// Update the state with pending changes
	var root []byte

	// Create a new receipt for the transaction, storing the intermediate root and gas used by the tx
	// based on the eip phase, we're passing whether the root touch-delete accounts.
	receipt := types.NewReceipt(root, failed, gas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = gas
	// if the transaction created a contract, store the creation address in the receipt.
	if tx.To() == nil && err == nil {
		receipt.ContractAddress = crypto.CreateAddress(from, txSim.GetNonce(from))
	}
	// Set the receipt logs and create a bloom for filtering

	receipt.Logs = txSim.GetLogs(tx.Hash())
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	return receipt, gas, nil
}

func CalculateCumulativeGasUsed(receipts []*types.Receipt) error {
	var gasUsed uint64
	for _, v := range receipts {
		if v == nil {
			return fmt.Errorf("receipt missing err")
		}
		gasUsed += v.GasUsed
		v.CumulativeGasUsed = gasUsed
	}
	return nil
}
