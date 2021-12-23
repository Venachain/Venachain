package core

import (
	"errors"
	"sync"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/core/state"
	"github.com/Venachain/Venachain/core/types"
	"github.com/Venachain/Venachain/log"
)

var (
	errMakeStateDB = errors.New("make StateDB error")
)

type BlockChainCache struct {
	*BlockChain
	stateDBCache  map[common.Hash]*stateDBCache  // key is header SealHash
	receiptsCache map[common.Hash]*receiptsCache // key is header SealHash
	stateDBMu     sync.RWMutex
	receiptsMu    sync.RWMutex
}

type stateDBCache struct {
	stateDB  *state.StateDB
	blockNum uint64
}

type receiptsCache struct {
	receipts []*types.Receipt
	blockNum uint64
}

func (pbc *BlockChainCache) CurrentBlock() *types.Block {
	log.Debug("get CurrentBlock() in chain")
	return pbc.currentBlock.Load().(*types.Block)
}

func (pbc *BlockChainCache) GetBlock(hash common.Hash, number uint64) *types.Block {
	block := pbc.getBlock(hash, number)
	if block == nil {
		log.Trace("cannot find block in chain", "RoutineID", common.CurrentGoRoutineID(), "hash", hash, "number", number)
	}
	return block
}

func NewBlockChainCache(blockChain *BlockChain) *BlockChainCache {
	pbc := &BlockChainCache{}
	pbc.BlockChain = blockChain
	pbc.stateDBCache = make(map[common.Hash]*stateDBCache)
	pbc.receiptsCache = make(map[common.Hash]*receiptsCache)

	return pbc
}

// Read the Receipt collection from the cache map.
func (bcc *BlockChainCache) ReadReceipts(sealHash common.Hash) []*types.Receipt {
	bcc.receiptsMu.RLock()
	defer bcc.receiptsMu.RUnlock()
	if obj, exist := bcc.receiptsCache[sealHash]; exist {
		return obj.receipts
	}
	return nil
}

// GetState returns a new mutable state based on a particular point in time.
func (bcc *BlockChainCache) GetState(header *types.Header) (*state.StateDB, error) {
	state := bcc.ReadStateDB(header.SealHash())
	if state != nil {
		return state, nil
	} else {
		return bcc.StateAt(header.Root)
	}
}

// Read the StateDB instance from the cache map
func (pbc *BlockChainCache) ReadStateDB(sealHash common.Hash) *state.StateDB {
	pbc.stateDBMu.RLock()
	defer pbc.stateDBMu.RUnlock()
	if obj, exist := pbc.stateDBCache[sealHash]; exist {
		log.Debug("Read the StateDB instance from the cache map", "sealHash", sealHash)
		return obj.stateDB.Copy()
	}
	return nil
}

// Write Receipt to the cache
func (pbc *BlockChainCache) WriteReceipts(sealHash common.Hash, receipts []*types.Receipt, blockNum uint64) {
	pbc.receiptsMu.Lock()
	defer pbc.receiptsMu.Unlock()
	obj, exist := pbc.receiptsCache[sealHash]
	if exist && obj.blockNum == blockNum {
		obj.receipts = append(obj.receipts, receipts...)
	} else if !exist {
		pbc.receiptsCache[sealHash] = &receiptsCache{receipts: receipts, blockNum: blockNum}
	}
}

// Write a StateDB instance to the cache
func (bcc *BlockChainCache) WriteStateDB(sealHash common.Hash, stateDB *state.StateDB, blockNum uint64) {
	bcc.stateDBMu.Lock()
	defer bcc.stateDBMu.Unlock()
	log.Info("Write a StateDB instance to the cache", "sealHash", sealHash, "blockNum", blockNum)
	if _, exist := bcc.stateDBCache[sealHash]; !exist {
		bcc.stateDBCache[sealHash] = &stateDBCache{stateDB: stateDB, blockNum: blockNum}
	}
}

// Read the Receipt collection from the cache map
func (bcc *BlockChainCache) clearReceipts(sealHash common.Hash) {
	bcc.receiptsMu.Lock()
	defer bcc.receiptsMu.Unlock()

	var blockNum uint64
	if obj, exist := bcc.receiptsCache[sealHash]; exist {
		blockNum = obj.blockNum
		//delete(pbc.receiptsCache, sealHash)
	}
	for hash, obj := range bcc.receiptsCache {
		if obj.blockNum <= blockNum {
			delete(bcc.receiptsCache, hash)
		}
	}
}

// Read the StateDB instance from the cache map
func (bcc *BlockChainCache) clearStateDB(sealHash common.Hash) {
	bcc.stateDBMu.Lock()
	defer bcc.stateDBMu.Unlock()

	var blockNum uint64
	if obj, exist := bcc.stateDBCache[sealHash]; exist {
		blockNum = obj.blockNum
		//delete(pbc.stateDBCache, sealHash)
	}
	for hash, obj := range bcc.stateDBCache {
		if obj.blockNum <= blockNum {
			delete(bcc.stateDBCache, hash)
		}
	}
}

// Get the StateDB instance of the corresponding block
func (bcc *BlockChainCache) MakeStateDB(block *types.Block) (*state.StateDB, error) {
	// Create a StateDB instance from the blockchain based on stateRoot
	if state, err := bcc.StateAt(block.Root()); err == nil && state != nil {
		return state, nil
	}
	// Read and copy the stateDB instance in the cache
	sealHash := bcc.Engine().SealHash(block.Header())
	log.Info("Read and copy the stateDB instance in the cache", "sealHash", sealHash, "blockHash", block.Hash(), "blockNum", block.NumberU64(), "stateRoot", block.Root())
	if state := bcc.ReadStateDB(sealHash); state != nil {
		//return state.Copy(), nil
		return state, nil
	} else {
		return nil, errMakeStateDB
	}
}

// Get the StateDB instance of the corresponding block
func (bcc *BlockChainCache) ClearCache(block *types.Block) {
	sealHash := bcc.Engine().SealHash(block.Header())
	bcc.clearReceipts(sealHash)
	bcc.clearStateDB(sealHash)
}
