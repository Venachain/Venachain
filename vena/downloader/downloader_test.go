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

package downloader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Venachain/Venachain/common/hexutil"
	irisBackend "github.com/Venachain/Venachain/consensus/iris/backend"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/core"
	"github.com/Venachain/Venachain/core/types"
	"github.com/Venachain/Venachain/crypto"
	"github.com/Venachain/Venachain/event"
	"github.com/Venachain/Venachain/params"
	"github.com/Venachain/Venachain/trie"
	"github.com/Venachain/Venachain/venadb"
)

var (
	testKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	testAddress = crypto.PubkeyToAddress(testKey.PublicKey)
)

// Reduce some of the parameters to make the tester faster.
func init() {
	MaxForkAncestry = uint64(10000)
	blockCacheItems = 1024
	fsHeaderContCheck = 500 * time.Millisecond
}

// downloadTester is a test simulator for mocking out local block chain.
type downloadTester struct {
	downloader *Downloader

	genesis *types.Block    // Genesis blocks used by the tester and peers
	stateDb venadb.Database // Database used by the tester for syncing from peers
	peerDb  venadb.Database // Database of the peers containing all data

	ownHashes   []common.Hash                  // Hash chain belonging to the tester
	ownHeaders  map[common.Hash]*types.Header  // Headers belonging to the tester
	ownBlocks   map[common.Hash]*types.Block   // Blocks belonging to the tester
	ownReceipts map[common.Hash]types.Receipts // Receipts belonging to the tester
	ownChainTd  map[common.Hash]*big.Int       // Total difficulties of the blocks in the local chain

	peerHashes   map[string][]common.Hash                  // Hash chain belonging to different test peers
	peerHeaders  map[string]map[common.Hash]*types.Header  // Headers belonging to different test peers
	peerBlocks   map[string]map[common.Hash]*types.Block   // Blocks belonging to different test peers
	peerReceipts map[string]map[common.Hash]types.Receipts // Receipts belonging to different test peers
	peerChainTds map[string]map[common.Hash]*big.Int       // Total difficulties of the blocks in the peer chains

	peerMissingStates map[string]map[common.Hash]bool // State entries that fast sync should not return

	lock sync.RWMutex
}

func GenesisBlockForTesting(db venadb.Database, addr common.Address, balance *big.Int) *types.Block {
	buf, err := ioutil.ReadFile("./testdata/venachain.json")
	if err != nil {
		return nil
	}

	var gen core.Genesis
	if err := gen.UnmarshalJSON(buf); err != nil {
		return nil
	}

	gen.Alloc[addr] = core.GenesisAccount{
		Code:    nil,
		Storage: nil,
		Balance: balance,
		Nonce:   0,
	}

	block, _ := gen.Commit(db)
	return block
}

// newTester creates a new downloader test mocker.
func newTester() *downloadTester {
	testdb := venadb.NewMemDatabase()
	balanceBytes, _ := hexutil.Decode("0x2000000000000000000000000000000000000000000000000000000000000")
	balance := big.NewInt(0)
	genesis := GenesisBlockForTesting(testdb, testAddress, balance.SetBytes(balanceBytes))
	common.SysCfg.ReplayParam = &common.ReplayParam{
		Pivot:           0,
		OldSysContracts: make(map[common.Address]string),
		OldSuperAdmin:   common.NullAddress,
	}
	tester := &downloadTester{
		genesis:           genesis,
		peerDb:            testdb,
		ownHashes:         []common.Hash{genesis.Hash()},
		ownHeaders:        map[common.Hash]*types.Header{genesis.Hash(): genesis.Header()},
		ownBlocks:         map[common.Hash]*types.Block{genesis.Hash(): genesis},
		ownReceipts:       map[common.Hash]types.Receipts{genesis.Hash(): nil},
		ownChainTd:        map[common.Hash]*big.Int{genesis.Hash(): big.NewInt(0)},
		peerHashes:        make(map[string][]common.Hash),
		peerHeaders:       make(map[string]map[common.Hash]*types.Header),
		peerBlocks:        make(map[string]map[common.Hash]*types.Block),
		peerReceipts:      make(map[string]map[common.Hash]types.Receipts),
		peerChainTds:      make(map[string]map[common.Hash]*big.Int),
		peerMissingStates: make(map[string]map[common.Hash]bool),
	}
	tester.stateDb = venadb.NewMemDatabase()
	tester.stateDb.Put(genesis.Root().Bytes(), []byte{0x00})

	tester.downloader = New(FullSync, tester.stateDb, new(event.TypeMux), tester, nil, tester.dropPeer)

	return tester
}

// makeChain creates a chain of n blocks starting at and including parent.
// the returned hash chain is ordered head->parent. In addition, every 3rd block
// contains a transaction and every 5th an uncle to allow testing correct block
// reassembly.
func (dl *downloadTester) makeChain(n int, seed byte, parent *types.Block, parentReceipts types.Receipts, heavy bool) ([]common.Hash, map[common.Hash]*types.Header, map[common.Hash]*types.Block, map[common.Hash]types.Receipts) {
	// Generate the block chain
	blocks, receipts := core.GenerateChain(params.TestChainConfig, parent, irisBackend.New(nil, nil, nil), dl.peerDb, n, func(i int, block *core.BlockGen) {
		block.SetCoinbase(common.Address{seed})

		// If a heavy chain is requested, delay blocks to raise difficulty
		if heavy {
			block.OffsetTime(-1)
		}
		gas := big.NewInt(0)
		gas = gas.SetBytes(hexutil.MustDecode("0x99988888"))
		gasPrice := big.NewInt(0)
		gasPrice = gasPrice.SetBytes(hexutil.MustDecode("0x8250"))
		// If the block number is multiple of 3, send a bonus transaction to the miner
		if parent == dl.genesis && i%3 == 0 {
			signer := types.MakeSigner(params.TestChainConfig)
			tx, err := types.SignTx(types.NewTransaction(block.TxNonce(testAddress), common.HexToAddress("0x0384d39b9cbf9bab2a3b41692d426ad57e41c54c"), big.NewInt(1000), gas.Uint64(), gasPrice, hexutil.MustDecode("0xd3880000000000000002857072696e7483616263")), signer, testKey)
			if err != nil {
				panic(err)
			}
			block.AddTx(tx)
		}
	})
	// Convert the block-chain into a hash-chain and header/block maps
	hashes := make([]common.Hash, n+1)
	hashes[len(hashes)-1] = parent.Hash()

	headerm := make(map[common.Hash]*types.Header, n+1)
	headerm[parent.Hash()] = parent.Header()

	blockm := make(map[common.Hash]*types.Block, n+1)
	blockm[parent.Hash()] = parent

	receiptm := make(map[common.Hash]types.Receipts, n+1)
	receiptm[parent.Hash()] = parentReceipts

	for i, b := range blocks {
		hashes[len(hashes)-i-2] = b.Hash()
		headerm[b.Hash()] = b.Header()
		blockm[b.Hash()] = b
		receiptm[b.Hash()] = receipts[i]
	}
	return hashes, headerm, blockm, receiptm
}

// makeChainFork creates two chains of length n, such that h1[:f] and
// h2[:f] are different but have a common suffix of length n-f.
func (dl *downloadTester) makeChainFork(n, f int, parent *types.Block, parentReceipts types.Receipts, balanced bool) ([]common.Hash, []common.Hash, map[common.Hash]*types.Header, map[common.Hash]*types.Header, map[common.Hash]*types.Block, map[common.Hash]*types.Block, map[common.Hash]types.Receipts, map[common.Hash]types.Receipts) {
	// Create the common suffix
	hashes, headers, blocks, receipts := dl.makeChain(n-f, 0, parent, parentReceipts, false)

	// Create the forks, making the second heavier if non balanced forks were requested
	hashes1, headers1, blocks1, receipts1 := dl.makeChain(f, 1, blocks[hashes[0]], receipts[hashes[0]], false)
	hashes1 = append(hashes1, hashes[1:]...)

	heavy := false
	if !balanced {
		heavy = true
	}
	hashes2, headers2, blocks2, receipts2 := dl.makeChain(f, 2, blocks[hashes[0]], receipts[hashes[0]], heavy)
	hashes2 = append(hashes2, hashes[1:]...)

	for hash, header := range headers {
		headers1[hash] = header
		headers2[hash] = header
	}
	for hash, block := range blocks {
		blocks1[hash] = block
		blocks2[hash] = block
	}
	for hash, receipt := range receipts {
		receipts1[hash] = receipt
		receipts2[hash] = receipt
	}
	return hashes1, hashes2, headers1, headers2, blocks1, blocks2, receipts1, receipts2
}

// terminate aborts any operations on the embedded downloader and releases all
// held resources.
func (dl *downloadTester) terminate() {
	dl.downloader.Terminate()
}

// sync starts synchronizing with a remote peer, blocking until it completes.
func (dl *downloadTester) sync(id string, td *big.Int, mode SyncMode) error {
	dl.lock.RLock()
	hash := dl.peerHashes[id][0]
	// If no particular TD was requested, load from the peer's blockchain
	if td == nil {
		td = big.NewInt(1)
		if diff, ok := dl.peerChainTds[id][hash]; ok {
			td = diff
		}
	}
	dl.lock.RUnlock()

	// Synchronise with the chosen peer and ensure proper cleanup afterwards
	err := dl.downloader.synchronise(id, hash, td, mode)
	select {
	case <-dl.downloader.cancelCh:
		// Ok, downloader fully cancelled after sync cycle
	default:
		// Downloader is still accepting packets, can block a peer up
		panic("downloader active post sync cycle") // panic will be caught by tester
	}
	return err
}

// HasHeader checks if a header is present in the testers canonical chain.
func (dl *downloadTester) HasHeader(hash common.Hash, number uint64) bool {
	return dl.GetHeaderByHash(hash) != nil
}

// HasBlock checks if a block is present in the testers canonical chain.
func (dl *downloadTester) HasBlock(hash common.Hash, number uint64) bool {
	return dl.GetBlockByHash(hash) != nil
}

// GetHeader retrieves a header from the testers canonical chain.
func (dl *downloadTester) GetHeaderByHash(hash common.Hash) *types.Header {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	return dl.ownHeaders[hash]
}

// GetBlock retrieves a block from the testers canonical chain.
func (dl *downloadTester) GetBlockByHash(hash common.Hash) *types.Block {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	return dl.ownBlocks[hash]
}

// CurrentHeader retrieves the current head header from the canonical chain.
func (dl *downloadTester) CurrentHeader() *types.Header {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	for i := len(dl.ownHashes) - 1; i >= 0; i-- {
		if header := dl.ownHeaders[dl.ownHashes[i]]; header != nil {
			return header
		}
	}
	return dl.genesis.Header()
}

// CurrentBlock retrieves the current head block from the canonical chain.
func (dl *downloadTester) CurrentBlock() *types.Block {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	for i := len(dl.ownHashes) - 1; i >= 0; i-- {
		if block := dl.ownBlocks[dl.ownHashes[i]]; block != nil {
			if _, err := dl.stateDb.Get(block.Root().Bytes()); err == nil {
				return block
			}
		}
	}
	return dl.genesis
}

// CurrentFastBlock retrieves the current head fast-sync block from the canonical chain.
func (dl *downloadTester) CurrentFastBlock() *types.Block {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	for i := len(dl.ownHashes) - 1; i >= 0; i-- {
		if block := dl.ownBlocks[dl.ownHashes[i]]; block != nil {
			return block
		}
	}
	return dl.genesis
}

// FastSyncCommitHead manually sets the head block to a given hash.
func (dl *downloadTester) FastSyncCommitHead(hash common.Hash) error {
	// For now only check that the state trie is correct
	if block := dl.GetBlockByHash(hash); block != nil {
		_, err := trie.NewSecure(block.Root(), trie.NewDatabase(dl.stateDb), 0)
		return err
	}
	return fmt.Errorf("non existent block: %x", hash[:4])
}

// GetTd retrieves the block's total difficulty from the canonical chain.
func (dl *downloadTester) GetTd(hash common.Hash, number uint64) *big.Int {
	dl.lock.RLock()
	defer dl.lock.RUnlock()

	return dl.ownChainTd[hash]
}

// InsertHeaderChain injects a new batch of headers into the simulated chain.
func (dl *downloadTester) InsertHeaderChain(headers []*types.Header, checkFreq int) (int, error) {
	dl.lock.Lock()
	defer dl.lock.Unlock()

	// Do a quick check, as the blockchain.InsertHeaderChain doesn't insert anything in case of errors
	if _, ok := dl.ownHeaders[headers[0].ParentHash]; !ok {
		return 0, errors.New("unknown parent")
	}
	for i := 1; i < len(headers); i++ {
		if headers[i].ParentHash != headers[i-1].Hash() {
			return i, errors.New("unknown parent")
		}
	}
	// Do a full insert if pre-checks passed
	for i, header := range headers {
		if _, ok := dl.ownHeaders[header.Hash()]; ok {
			continue
		}
		if _, ok := dl.ownHeaders[header.ParentHash]; !ok {
			return i, errors.New("unknown parent")
		}
		dl.ownHashes = append(dl.ownHashes, header.Hash())
		dl.ownHeaders[header.Hash()] = header
		dl.ownChainTd[header.Hash()] = new(big.Int).Add(dl.ownChainTd[header.ParentHash], big.NewInt(0))
	}
	return len(headers), nil
}

// InsertChain injects a new batch of blocks into the simulated chain.
func (dl *downloadTester) InsertChain(blocks types.Blocks) (int, error) {
	dl.lock.Lock()
	defer dl.lock.Unlock()

	for i, block := range blocks {
		if parent, ok := dl.ownBlocks[block.ParentHash()]; !ok {
			return i, errors.New("unknown parent")
		} else if _, err := dl.stateDb.Get(parent.Root().Bytes()); err != nil {
			return i, fmt.Errorf("unknown parent state %x: %v", parent.Root(), err)
		}
		if _, ok := dl.ownHeaders[block.Hash()]; !ok {
			dl.ownHashes = append(dl.ownHashes, block.Hash())
			dl.ownHeaders[block.Hash()] = block.Header()
		}
		dl.ownBlocks[block.Hash()] = block
		dl.stateDb.Put(block.Root().Bytes(), []byte{0x00})
		dl.ownChainTd[block.Hash()] = new(big.Int).Add(dl.ownChainTd[block.ParentHash()], big.NewInt(0))
	}
	return len(blocks), nil
}

// InsertReceiptChain injects a new batch of receipts into the simulated chain.
func (dl *downloadTester) InsertReceiptChain(blocks types.Blocks, receipts []types.Receipts) (int, error) {
	dl.lock.Lock()
	defer dl.lock.Unlock()

	for i := 0; i < len(blocks) && i < len(receipts); i++ {
		if _, ok := dl.ownHeaders[blocks[i].Hash()]; !ok {
			return i, errors.New("unknown owner")
		}
		if _, ok := dl.ownBlocks[blocks[i].ParentHash()]; !ok {
			return i, errors.New("unknown parent")
		}
		dl.ownBlocks[blocks[i].Hash()] = blocks[i]
		dl.ownReceipts[blocks[i].Hash()] = receipts[i]
	}
	return len(blocks), nil
}

// Rollback removes some recently added elements from the chain.
func (dl *downloadTester) Rollback(hashes []common.Hash) {
	dl.lock.Lock()
	defer dl.lock.Unlock()

	for i := len(hashes) - 1; i >= 0; i-- {
		if dl.ownHashes[len(dl.ownHashes)-1] == hashes[i] {
			dl.ownHashes = dl.ownHashes[:len(dl.ownHashes)-1]
		}
		delete(dl.ownChainTd, hashes[i])
		delete(dl.ownHeaders, hashes[i])
		delete(dl.ownReceipts, hashes[i])
		delete(dl.ownBlocks, hashes[i])
	}
}

// newPeer registers a new block download source into the downloader.
func (dl *downloadTester) newPeer(id string, version int, hashes []common.Hash, headers map[common.Hash]*types.Header, blocks map[common.Hash]*types.Block, receipts map[common.Hash]types.Receipts) error {
	return dl.newSlowPeer(id, version, hashes, headers, blocks, receipts, 0)
}

// newSlowPeer registers a new block download source into the downloader, with a
// specific delay time on processing the network packets sent to it, simulating
// potentially slow network IO.
func (dl *downloadTester) newSlowPeer(id string, version int, hashes []common.Hash, headers map[common.Hash]*types.Header, blocks map[common.Hash]*types.Block, receipts map[common.Hash]types.Receipts, delay time.Duration) error {
	dl.lock.Lock()
	defer dl.lock.Unlock()

	var err = dl.downloader.RegisterPeer(id, version, &downloadTesterPeer{dl: dl, id: id, delay: delay})
	if err == nil {
		// Assign the owned hashes, headers and blocks to the peer (deep copy)
		dl.peerHashes[id] = make([]common.Hash, len(hashes))
		copy(dl.peerHashes[id], hashes)

		dl.peerHeaders[id] = make(map[common.Hash]*types.Header)
		dl.peerBlocks[id] = make(map[common.Hash]*types.Block)
		dl.peerReceipts[id] = make(map[common.Hash]types.Receipts)
		dl.peerChainTds[id] = make(map[common.Hash]*big.Int)
		dl.peerMissingStates[id] = make(map[common.Hash]bool)

		genesis := hashes[len(hashes)-1]
		if header := headers[genesis]; header != nil {
			dl.peerHeaders[id][genesis] = header
			dl.peerChainTds[id][genesis] = big.NewInt(0)
		}
		if block := blocks[genesis]; block != nil {
			dl.peerBlocks[id][genesis] = block
			dl.peerChainTds[id][genesis] = big.NewInt(0)
		}

		for i := len(hashes) - 2; i >= 0; i-- {
			hash := hashes[i]

			if header, ok := headers[hash]; ok {
				dl.peerHeaders[id][hash] = header
				if _, ok := dl.peerHeaders[id][header.ParentHash]; ok {
					dl.peerChainTds[id][hash] = new(big.Int).Add(big.NewInt(0), dl.peerChainTds[id][header.ParentHash])
				}
			}
			if block, ok := blocks[hash]; ok {
				dl.peerBlocks[id][hash] = block
				if _, ok := dl.peerBlocks[id][block.ParentHash()]; ok {
					dl.peerChainTds[id][hash] = new(big.Int).Add(big.NewInt(0), dl.peerChainTds[id][block.ParentHash()])
				}
			}
			if receipt, ok := receipts[hash]; ok {
				dl.peerReceipts[id][hash] = receipt
			}
		}
	}
	return err
}

// dropPeer simulates a hard peer removal from the connection pool.
func (dl *downloadTester) dropPeer(id string) {
	dl.lock.Lock()
	defer dl.lock.Unlock()

	delete(dl.peerHashes, id)
	delete(dl.peerHeaders, id)
	delete(dl.peerBlocks, id)
	delete(dl.peerChainTds, id)

	dl.downloader.UnregisterPeer(id)
}

type downloadTesterPeer struct {
	dl    *downloadTester
	id    string
	delay time.Duration
	lock  sync.RWMutex
}

// setDelay is a thread safe setter for the network delay value.
func (dlp *downloadTesterPeer) setDelay(delay time.Duration) {
	dlp.lock.Lock()
	defer dlp.lock.Unlock()

	dlp.delay = delay
}

// waitDelay is a thread safe way to sleep for the configured time.
func (dlp *downloadTesterPeer) waitDelay() {
	dlp.lock.RLock()
	delay := dlp.delay
	dlp.lock.RUnlock()

	time.Sleep(delay)
}

// Head constructs a function to retrieve a peer's current head hash
// and total difficulty.
func (dlp *downloadTesterPeer) Head() (common.Hash, *big.Int) {
	dlp.dl.lock.RLock()
	defer dlp.dl.lock.RUnlock()

	return dlp.dl.peerHashes[dlp.id][0], nil
}

// RequestHeadersByHash constructs a GetBlockHeaders function based on a hashed
// origin; associated with a particular peer in the download tester. The returned
// function can be used to retrieve batches of headers from the particular peer.
func (dlp *downloadTesterPeer) RequestHeadersByHash(origin common.Hash, amount int, skip int, reverse bool) error {
	// Find the canonical number of the hash
	dlp.dl.lock.RLock()
	number := uint64(0)
	for num, hash := range dlp.dl.peerHashes[dlp.id] {
		if hash == origin {
			number = uint64(len(dlp.dl.peerHashes[dlp.id]) - num - 1)
			break
		}
	}
	dlp.dl.lock.RUnlock()

	// Use the absolute header fetcher to satisfy the query
	return dlp.RequestHeadersByNumber(number, amount, skip, reverse)
}

// RequestHeadersByNumber constructs a GetBlockHeaders function based on a numbered
// origin; associated with a particular peer in the download tester. The returned
// function can be used to retrieve batches of headers from the particular peer.
func (dlp *downloadTesterPeer) RequestHeadersByNumber(origin uint64, amount int, skip int, reverse bool) error {
	dlp.waitDelay()

	dlp.dl.lock.RLock()
	defer dlp.dl.lock.RUnlock()

	// Gather the next batch of headers
	hashes := dlp.dl.peerHashes[dlp.id]
	headers := dlp.dl.peerHeaders[dlp.id]
	result := make([]*types.Header, 0, amount)
	for i := 0; i < amount && len(hashes)-int(origin)-1-i*(skip+1) >= 0; i++ {
		if header, ok := headers[hashes[len(hashes)-int(origin)-1-i*(skip+1)]]; ok {
			result = append(result, header)
		}
	}
	// Delay delivery a bit to allow attacks to unfold
	go func() {
		time.Sleep(time.Millisecond)
		dlp.dl.downloader.DeliverHeaders(dlp.id, result)
	}()
	return nil
}

// RequestBodies constructs a getBlockBodies method associated with a particular
// peer in the download tester. The returned function can be used to retrieve
// batches of block bodies from the particularly requested peer.
func (dlp *downloadTesterPeer) RequestBodies(hashes []common.Hash) error {
	dlp.waitDelay()

	dlp.dl.lock.RLock()
	defer dlp.dl.lock.RUnlock()

	blocks := dlp.dl.peerBlocks[dlp.id]

	transactions := make([][]*types.Transaction, 0, len(hashes))

	for _, hash := range hashes {
		if block, ok := blocks[hash]; ok {
			transactions = append(transactions, block.Transactions())
		}
	}
	go dlp.dl.downloader.DeliverBodies(dlp.id, transactions)

	return nil
}

// RequestReceipts constructs a getReceipts method associated with a particular
// peer in the download tester. The returned function can be used to retrieve
// batches of block receipts from the particularly requested peer.
func (dlp *downloadTesterPeer) RequestReceipts(hashes []common.Hash) error {
	dlp.waitDelay()

	dlp.dl.lock.RLock()
	defer dlp.dl.lock.RUnlock()

	receipts := dlp.dl.peerReceipts[dlp.id]

	results := make([][]*types.Receipt, 0, len(hashes))
	for _, hash := range hashes {
		if receipt, ok := receipts[hash]; ok {
			results = append(results, receipt)
		}
	}
	go dlp.dl.downloader.DeliverReceipts(dlp.id, results)

	return nil
}

// RequestNodeData constructs a getNodeData method associated with a particular
// peer in the download tester. The returned function can be used to retrieve
// batches of node state data from the particularly requested peer.
func (dlp *downloadTesterPeer) RequestNodeData(hashes []common.Hash) error {
	dlp.waitDelay()

	dlp.dl.lock.RLock()
	defer dlp.dl.lock.RUnlock()

	results := make([][]byte, 0, len(hashes))
	for _, hash := range hashes {
		if data, err := dlp.dl.peerDb.Get(hash.Bytes()); err == nil {
			if !dlp.dl.peerMissingStates[dlp.id][hash] {
				results = append(results, data)
			}
		} else {
			seckeybuf := [43]byte{}
			secureKeyPrefix := []byte("secure-key-")
			buf := append(seckeybuf[:0], secureKeyPrefix...)
			buf = append(buf, hash[:]...)
			if data2, err := dlp.dl.peerDb.Get(buf); err == nil {
				if !dlp.dl.peerMissingStates[dlp.id][hash] {
					results = append(results, data2)
				}
			}
		}
	}
	go dlp.dl.downloader.DeliverNodeData(dlp.id, results)

	return nil
}

// assertOwnChain checks if the local chain contains the correct number of items
// of the various chain components.
func assertOwnChain(t *testing.T, tester *downloadTester, length int) {
	assertOwnForkedChain(t, tester, 1, []int{length})
}

// assertOwnForkedChain checks if the local forked chain contains the correct
// number of items of the various chain components.
func assertOwnForkedChain(t *testing.T, tester *downloadTester, common int, lengths []int) {
	// Initialize the counters for the first fork
	headers, blocks, receipts := lengths[0], lengths[0], lengths[0]-fsMinFullBlocks

	if receipts < 0 {
		receipts = 1
	}
	// Update the counters for each subsequent fork
	for _, length := range lengths[1:] {
		headers += length - common
		blocks += length - common
		receipts += length - common - fsMinFullBlocks
	}
	switch tester.downloader.mode {
	case FullSync:
		receipts = 1
	case LightSync:
		blocks, receipts = 1, 1
	}
	if hs := len(tester.ownHeaders); hs != headers {
		t.Fatalf("synchronised headers mismatch: have %v, want %v", hs, headers)
	}
	if bs := len(tester.ownBlocks); bs != blocks {
		t.Fatalf("synchronised blocks mismatch: have %v, want %v", bs, blocks)
	}
	if rs := len(tester.ownReceipts); rs != receipts {
		t.Fatalf("synchronised receipts mismatch: have %v, want %v", rs, receipts)
	}
	// Verify the state trie too for fast syncs
	/*if tester.downloader.mode == FastSync {
		pivot := uint64(0)
		var index int
		if pivot := int(tester.downloader.queue.fastSyncPivot); pivot < common {
			index = pivot
		} else {
			index = len(tester.ownHashes) - lengths[len(lengths)-1] + int(tester.downloader.queue.fastSyncPivot)
		}
		if index > 0 {
			if statedb, err := state.New(tester.ownHeaders[tester.ownHashes[index]].Root, state.NewDatabase(trie.NewDatabase(tester.stateDb))); statedb == nil || err != nil {
				t.Fatalf("state reconstruction failed: %v", err)
			}
		}
	}*/
}

// Tests that simple synchronization against a canonical chain works correctly.
// In this test common ancestor lookup should be short circuited and not require
// binary searching.
func TestCanonicalSynchronisationFull(t *testing.T)  { testCanonicalSynchronisation(t, 1, FullSync) }
func TestCanonicalSynchronisationFast(t *testing.T)  { testCanonicalSynchronisation(t, 1, FastSync) }
func TestCanonicalSynchronisationLight(t *testing.T) { testCanonicalSynchronisation(t, 1, LightSync) }

func testCanonicalSynchronisation(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()

	tester := newTester()
	defer tester.terminate()

	// Create a small enough block chain to download
	targetBlocks := blockCacheItems - 15
	hashes, headers, blocks, receipts := tester.makeChain(targetBlocks, 0, tester.genesis, nil, false)

	tester.newPeer("peer", protocol, hashes, headers, blocks, receipts)

	// Synchronise with the peer and make sure all relevant data was retrieved
	if err := tester.sync("peer", nil, mode); err != nil {
		t.Fatalf("failed to synchronise blocks: %v", err)
	}
	assertOwnChain(t, tester, targetBlocks+1)
}

// Tests that if a large batch of blocks are being downloaded, it is throttled
// until the cached blocks are retrieved.
func TestThrottlingFull(t *testing.T) { testThrottling(t, 1, FullSync) }
func TestThrottlingFast(t *testing.T) { testThrottling(t, 1, FastSync) }

func testThrottling(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()
	tester := newTester()
	defer tester.terminate()

	// Create a long block chain to download and the tester
	targetBlocks := 8 * blockCacheItems
	hashes, headers, blocks, receipts := tester.makeChain(targetBlocks, 0, tester.genesis, nil, false)

	tester.newPeer("peer", protocol, hashes, headers, blocks, receipts)

	// Wrap the importer to allow stepping
	blocked, proceed := uint32(0), make(chan struct{})
	tester.downloader.chainInsertHook = func(results []*fetchResult) {
		atomic.StoreUint32(&blocked, uint32(len(results)))
		<-proceed
	}
	// Start a synchronisation concurrently
	errc := make(chan error)
	go func() {
		errc <- tester.sync("peer", nil, mode)
	}()
	// Iteratively take some blocks, always checking the retrieval count
	for {
		// Check the retrieval count synchronously (! reason for this ugly block)
		tester.lock.RLock()
		retrieved := len(tester.ownBlocks)
		tester.lock.RUnlock()
		if retrieved >= targetBlocks+1 {
			break
		}
		// Wait a bit for sync to throttle itself
		var cached, frozen int
		for start := time.Now(); time.Since(start) < 3*time.Second; {
			time.Sleep(25 * time.Millisecond)

			tester.lock.Lock()
			tester.downloader.queue.lock.Lock()
			cached = len(tester.downloader.queue.blockDonePool)
			if mode == FastSync {
				if receipts := len(tester.downloader.queue.receiptDonePool); receipts < cached {
					//if tester.downloader.queue.resultCache[receipts].Header.Number.Uint64() < tester.downloader.queue.fastSyncPivot {
					cached = receipts
					//}
				}
			}
			frozen = int(atomic.LoadUint32(&blocked))
			retrieved = len(tester.ownBlocks)
			tester.downloader.queue.lock.Unlock()
			tester.lock.Unlock()

			if cached == blockCacheItems || retrieved+cached+frozen == targetBlocks+1 {
				break
			}
		}
		// Make sure we filled up the cache, then exhaust it
		time.Sleep(25 * time.Millisecond) // give it a chance to screw up

		tester.lock.RLock()
		retrieved = len(tester.ownBlocks)
		tester.lock.RUnlock()
		if cached != blockCacheItems && retrieved+cached+frozen != targetBlocks+1 {
			t.Fatalf("block count mismatch: have %v, want %v (owned %v, blocked %v, target %v)", cached, blockCacheItems, retrieved, frozen, targetBlocks+1)
		}
		// Permit the blocked blocks to import
		if atomic.LoadUint32(&blocked) > 0 {
			atomic.StoreUint32(&blocked, uint32(0))
			proceed <- struct{}{}
		}
	}
	// Check that we haven't pulled more blocks than available
	assertOwnChain(t, tester, targetBlocks+1)
	if err := <-errc; err != nil {
		t.Fatalf("block synchronization failed: %v", err)
	}
}

// Tests that an inactive downloader will not accept incoming block headers,
// bodies and receipts.
func TestInactiveDownloader(t *testing.T) {
	t.Parallel()

	tester := newTester()
	defer tester.terminate()

	// Check that neither block headers nor bodies are accepted
	if err := tester.downloader.DeliverHeaders("bad peer", []*types.Header{}); err != errNoSyncActive {
		t.Errorf("error mismatch: have %v, want %v", err, errNoSyncActive)
	}
	if err := tester.downloader.DeliverBodies("bad peer", [][]*types.Transaction{}); err != errNoSyncActive {
		t.Errorf("error mismatch: have %v, want %v", err, errNoSyncActive)
	}
	if err := tester.downloader.DeliverReceipts("bad peer", [][]*types.Receipt{}); err != errNoSyncActive {
		t.Errorf("error mismatch: have %v, want %v", err, errNoSyncActive)
	}
}

// Tests that a canceled download wipes all previously accumulated state.
func TestCancelFull(t *testing.T)  { testCancel(t, 1, FullSync) }
func TestCancelFast(t *testing.T)  { testCancel(t, 1, FastSync) }
func TestCancelLight(t *testing.T) { testCancel(t, 1, LightSync) }

func testCancel(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()

	tester := newTester()
	defer tester.terminate()

	// Create a small enough block chain to download and the tester
	targetBlocks := blockCacheItems - 15
	if targetBlocks >= MaxHashFetch {
		targetBlocks = MaxHashFetch - 15
	}
	if targetBlocks >= MaxHeaderFetch {
		targetBlocks = MaxHeaderFetch - 15
	}
	hashes, headers, blocks, receipts := tester.makeChain(targetBlocks, 0, tester.genesis, nil, false)

	tester.newPeer("peer", protocol, hashes, headers, blocks, receipts)

	// Make sure canceling works with a pristine downloader
	tester.downloader.Cancel()
	if !tester.downloader.queue.Idle() {
		t.Errorf("download queue not idle")
	}
	// Synchronise with the peer, but cancel afterwards
	if err := tester.sync("peer", nil, mode); err != nil {
		t.Fatalf("failed to synchronise blocks: %v", err)
	}
	tester.downloader.Cancel()
	if !tester.downloader.queue.Idle() {
		t.Errorf("download queue not idle")
	}
}

// Tests that synchronisation from multiple peers works as intended (multi thread sanity test).
func TestMultiSynchronisationFull(t *testing.T)  { testMultiSynchronisation(t, 1, FullSync) }
func TestMultiSynchronisationFast(t *testing.T)  { testMultiSynchronisation(t, 1, FastSync) }
func TestMultiSynchronisationLight(t *testing.T) { testMultiSynchronisation(t, 1, LightSync) }

func testMultiSynchronisation(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()

	tester := newTester()
	defer tester.terminate()

	// Create various peers with various parts of the chain
	targetPeers := 8
	targetBlocks := targetPeers*blockCacheItems - 15
	hashes, headers, blocks, receipts := tester.makeChain(targetBlocks, 0, tester.genesis, nil, false)

	for i := 0; i < targetPeers; i++ {
		id := fmt.Sprintf("peer #%d", i)
		tester.newPeer(id, protocol, hashes[i*blockCacheItems:], headers, blocks, receipts)
	}
	if err := tester.sync("peer #0", nil, mode); err != nil {
		t.Fatalf("failed to synchronise blocks: %v", err)
	}
	assertOwnChain(t, tester, targetBlocks+1)
}

// Tests that synchronisations behave well in multi-version protocol environments
// and not wreak havoc on other nodes in the network.
func TestMultiProtoSynchronisationFull(t *testing.T)  { testMultiProtoSync(t, 1, FullSync) }
func TestMultiProtoSynchronisationFast(t *testing.T)  { testMultiProtoSync(t, 1, FastSync) }
func TestMultiProtoSynchronisationLight(t *testing.T) { testMultiProtoSync(t, 1, LightSync) }

func testMultiProtoSync(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()

	tester := newTester()
	defer tester.terminate()

	// Create a small enough block chain to download
	targetBlocks := blockCacheItems - 15
	hashes, headers, blocks, receipts := tester.makeChain(targetBlocks, 0, tester.genesis, nil, false)

	// Create peers of every type
	tester.newPeer("peer 1", 1, hashes, headers, blocks, receipts)

	// Synchronise with the requested peer and make sure all blocks were retrieved
	if err := tester.sync(fmt.Sprintf("peer %d", protocol), nil, mode); err != nil {
		t.Fatalf("failed to synchronise blocks: %v", err)
	}
	assertOwnChain(t, tester, targetBlocks+1)

	// Check that no peers have been dropped off
	for _, version := range []int{1} {
		peer := fmt.Sprintf("peer %d", version)
		if _, ok := tester.peerHashes[peer]; !ok {
			t.Errorf("%s dropped", peer)
		}
	}
}

// Tests that if a block is empty (e.g. header only), no body request should be
// made, and instead the header should be assembled into a whole block in itself.
func TestEmptyShortCircuitFull(t *testing.T)  { testEmptyShortCircuit(t, 1, FullSync) }
func TestEmptyShortCircuitFast(t *testing.T)  { testEmptyShortCircuit(t, 1, FastSync) }
func TestEmptyShortCircuitLight(t *testing.T) { testEmptyShortCircuit(t, 1, LightSync) }

func testEmptyShortCircuit(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()

	tester := newTester()
	defer tester.terminate()

	// Create a block chain to download
	targetBlocks := 2*blockCacheItems - 15
	hashes, headers, blocks, receipts := tester.makeChain(targetBlocks, 0, tester.genesis, nil, false)

	tester.newPeer("peer", protocol, hashes, headers, blocks, receipts)

	// Instrument the downloader to signal body requests
	bodiesHave, receiptsHave := int32(0), int32(0)
	tester.downloader.bodyFetchHook = func(headers []*types.Header) {
		atomic.AddInt32(&bodiesHave, int32(len(headers)))
	}
	tester.downloader.receiptFetchHook = func(headers []*types.Header) {
		atomic.AddInt32(&receiptsHave, int32(len(headers)))
	}
	// Synchronise with the peer and make sure all blocks were retrieved
	if err := tester.sync("peer", nil, mode); err != nil {
		t.Fatalf("failed to synchronise blocks: %v", err)
	}
	assertOwnChain(t, tester, targetBlocks+1)

	// Validate the number of block bodies that should have been requested
	bodiesNeeded, receiptsNeeded := 0, 0
	for _, block := range blocks {
		if mode != LightSync && block != tester.genesis && (len(block.Transactions()) > 0) {
			bodiesNeeded++
		}
	}
	for _, receipt := range receipts {
		if mode == FastSync && len(receipt) > 0 {
			receiptsNeeded++
		}
	}
	if int(bodiesHave) != bodiesNeeded {
		t.Errorf("body retrieval count mismatch: have %v, want %v", bodiesHave, bodiesNeeded)
	}
	if int(receiptsHave) != receiptsNeeded {
		t.Errorf("receipt retrieval count mismatch: have %v, want %v", receiptsHave, receiptsNeeded)
	}
}

// Tests that headers are enqueued continuously, preventing malicious nodes from
// stalling the downloader by feeding gapped header chains.
func TestMissingHeaderAttackFull(t *testing.T)  { testMissingHeaderAttack(t, 1, FullSync) }
func TestMissingHeaderAttackFast(t *testing.T)  { testMissingHeaderAttack(t, 1, FastSync) }
func TestMissingHeaderAttackLight(t *testing.T) { testMissingHeaderAttack(t, 1, LightSync) }

func testMissingHeaderAttack(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()

	tester := newTester()
	defer tester.terminate()

	// Create a small enough block chain to download
	targetBlocks := blockCacheItems - 15
	hashes, headers, blocks, receipts := tester.makeChain(targetBlocks, 0, tester.genesis, nil, false)

	// Attempt a full sync with an attacker feeding gapped headers
	tester.newPeer("attack", protocol, hashes, headers, blocks, receipts)
	missing := targetBlocks / 2
	delete(tester.peerHeaders["attack"], hashes[missing])

	if err := tester.sync("attack", nil, mode); err == nil {
		t.Fatalf("succeeded attacker synchronisation")
	}
	// Synchronise with the valid peer and make sure sync succeeds
	tester.newPeer("valid", protocol, hashes, headers, blocks, receipts)
	if err := tester.sync("valid", nil, mode); err != nil {
		t.Fatalf("failed to synchronise blocks: %v", err)
	}
	assertOwnChain(t, tester, targetBlocks+1)
}

// Tests that if requested headers are shifted (i.e. first is missing), the queue
// detects the invalid numbering.
func TestShiftedHeaderAttackFull(t *testing.T)  { testShiftedHeaderAttack(t, 1, FullSync) }
func TestShiftedHeaderAttackFast(t *testing.T)  { testShiftedHeaderAttack(t, 1, FastSync) }
func TestShiftedHeaderAttackLight(t *testing.T) { testShiftedHeaderAttack(t, 1, LightSync) }

func testShiftedHeaderAttack(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()

	tester := newTester()
	defer tester.terminate()

	// Create a small enough block chain to download
	targetBlocks := blockCacheItems - 15
	hashes, headers, blocks, receipts := tester.makeChain(targetBlocks, 0, tester.genesis, nil, false)

	// Attempt a full sync with an attacker feeding shifted headers
	tester.newPeer("attack", protocol, hashes, headers, blocks, receipts)
	delete(tester.peerHeaders["attack"], hashes[len(hashes)-2])
	delete(tester.peerBlocks["attack"], hashes[len(hashes)-2])
	delete(tester.peerReceipts["attack"], hashes[len(hashes)-2])

	if err := tester.sync("attack", nil, mode); err == nil {
		t.Fatalf("succeeded attacker synchronisation")
	}
	// Synchronise with the valid peer and make sure sync succeeds
	tester.newPeer("valid", protocol, hashes, headers, blocks, receipts)
	if err := tester.sync("valid", nil, mode); err != nil {
		t.Fatalf("failed to synchronise blocks: %v", err)
	}
	assertOwnChain(t, tester, targetBlocks+1)
}

// Tests that a peer advertising an high TD doesn't get to stall the downloader
// afterwards by not sending any useful hashes.
func TestHighTDStarvationAttackFull(t *testing.T)  { testHighTDStarvationAttack(t, 1, FullSync) }
func TestHighTDStarvationAttackFast(t *testing.T)  { testHighTDStarvationAttack(t, 1, FastSync) }
func TestHighTDStarvationAttackLight(t *testing.T) { testHighTDStarvationAttack(t, 1, LightSync) }

func testHighTDStarvationAttack(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()

	tester := newTester()
	defer tester.terminate()

	hashes, headers, blocks, receipts := tester.makeChain(0, 0, tester.genesis, nil, false)
	tester.newPeer("attack", protocol, []common.Hash{hashes[0]}, headers, blocks, receipts)

	if err := tester.sync("attack", big.NewInt(1000000), mode); err != errStallingPeer {
		t.Fatalf("synchronisation error mismatch: have %v, want %v", err, errStallingPeer)
	}
}

// Tests that misbehaving peers are disconnected, whilst behaving ones are not.
func TestBlockHeaderAttackerDropping(t *testing.T) { testBlockHeaderAttackerDropping(t, 1) }

func testBlockHeaderAttackerDropping(t *testing.T, protocol int) {
	t.Parallel()

	// Define the disconnection requirement for individual hash fetch errors
	tests := []struct {
		result error
		drop   bool
	}{
		{nil, false},                        // Sync succeeded, all is well
		{errBusy, false},                    // Sync is already in progress, no problem
		{errUnknownPeer, false},             // Peer is unknown, was already dropped, don't double drop
		{errBadPeer, true},                  // Peer was deemed bad for some reason, drop it
		{errStallingPeer, true},             // Peer was detected to be stalling, drop it
		{errNoPeers, false},                 // No peers to download from, soft race, no issue
		{errTimeout, true},                  // No hashes received in due time, drop the peer
		{errEmptyHeaderSet, true},           // No headers were returned as a response, drop as it's a dead end
		{errPeersUnavailable, true},         // Nobody had the advertised blocks, drop the advertiser
		{errInvalidAncestor, true},          // Agreed upon ancestor is not acceptable, drop the chain rewriter
		{errInvalidChain, true},             // Hash chain was detected as invalid, definitely drop
		{errInvalidBlock, false},            // A bad peer was detected, but not the sync origin
		{errInvalidBody, false},             // A bad peer was detected, but not the sync origin
		{errInvalidReceipt, false},          // A bad peer was detected, but not the sync origin
		{errCancelBlockFetch, false},        // Synchronisation was canceled, origin may be innocent, don't drop
		{errCancelHeaderFetch, false},       // Synchronisation was canceled, origin may be innocent, don't drop
		{errCancelBodyFetch, false},         // Synchronisation was canceled, origin may be innocent, don't drop
		{errCancelReceiptFetch, false},      // Synchronisation was canceled, origin may be innocent, don't drop
		{errCancelHeaderProcessing, false},  // Synchronisation was canceled, origin may be innocent, don't drop
		{errCancelContentProcessing, false}, // Synchronisation was canceled, origin may be innocent, don't drop
	}
	// Run the tests and check disconnection status
	tester := newTester()
	defer tester.terminate()

	for i, tt := range tests {
		// Register a new peer and ensure it's presence
		id := fmt.Sprintf("test %d", i)
		if err := tester.newPeer(id, protocol, []common.Hash{tester.genesis.Hash()}, nil, nil, nil); err != nil {
			t.Fatalf("test %d: failed to register new peer: %v", i, err)
		}
		if _, ok := tester.peerHashes[id]; !ok {
			t.Fatalf("test %d: registered peer not found", i)
		}
		// Simulate a synchronisation and check the required result
		tester.downloader.synchroniseMock = func(string, common.Hash) error { return tt.result }

		tester.downloader.Synchronise(id, tester.genesis.Hash(), big.NewInt(1000), FullSync)
		if _, ok := tester.peerHashes[id]; !ok != tt.drop {
			t.Errorf("test %d: peer drop mismatch for %v: have %v, want %v", i, tt.result, !ok, tt.drop)
		}
	}
}

// Tests that synchronisation progress (origin block number, current block number
// and highest block number) is tracked and updated correctly.
func TestSyncProgressFull(t *testing.T)  { testSyncProgress(t, 1, FullSync) }
func TestSyncProgressFast(t *testing.T)  { testSyncProgress(t, 1, FastSync) }
func TestSyncProgressLight(t *testing.T) { testSyncProgress(t, 1, LightSync) }

func testSyncProgress(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()

	tester := newTester()
	defer tester.terminate()

	// Create a small enough block chain to download
	targetBlocks := blockCacheItems - 15
	hashes, headers, blocks, receipts := tester.makeChain(targetBlocks, 0, tester.genesis, nil, false)

	// Set a sync init hook to catch progress changes
	starting := make(chan struct{})
	progress := make(chan struct{})

	tester.downloader.syncInitHook = func(origin, latest uint64) {
		starting <- struct{}{}
		<-progress
	}
	// Retrieve the sync progress and ensure they are zero (pristine sync)
	if progress := tester.downloader.Progress(); progress.StartingBlock != 0 || progress.CurrentBlock != 0 || progress.HighestBlock != 0 {
		t.Fatalf("Pristine progress mismatch: have %v/%v/%v, want %v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, 0, 0, 0)
	}
	// Synchronise half the blocks and check initial progress
	tester.newPeer("peer-half", protocol, hashes[targetBlocks/2:], headers, blocks, receipts)
	pending := new(sync.WaitGroup)
	pending.Add(1)

	go func() {
		defer pending.Done()
		if err := tester.sync("peer-half", nil, mode); err != nil {
			panic(fmt.Sprintf("failed to synchronise blocks: %v", err))
		}
	}()
	<-starting
	if progress := tester.downloader.Progress(); progress.StartingBlock != 0 || progress.CurrentBlock != 0 || progress.HighestBlock != uint64(targetBlocks/2+1) {
		t.Fatalf("Initial progress mismatch: have %v/%v/%v, want %v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, 0, 0, targetBlocks/2+1)
	}
	progress <- struct{}{}
	pending.Wait()

	// Synchronise all the blocks and check continuation progress
	tester.newPeer("peer-full", protocol, hashes, headers, blocks, receipts)
	pending.Add(1)

	go func() {
		defer pending.Done()
		if err := tester.sync("peer-full", nil, mode); err != nil {
			panic(fmt.Sprintf("failed to synchronise blocks: %v", err))
		}
	}()
	<-starting
	if progress := tester.downloader.Progress(); progress.StartingBlock != uint64(targetBlocks/2+1) || progress.CurrentBlock != uint64(targetBlocks/2+1) || progress.HighestBlock != uint64(targetBlocks) {
		t.Fatalf("Completing progress mismatch: have %v/%v/%v, want %v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, targetBlocks/2+1, targetBlocks/2+1, targetBlocks)
	}
	progress <- struct{}{}
	pending.Wait()

	// Check final progress after successful sync
	if progress := tester.downloader.Progress(); progress.StartingBlock != uint64(targetBlocks/2+1) || progress.CurrentBlock != uint64(targetBlocks) || progress.HighestBlock != uint64(targetBlocks) {
		t.Fatalf("Final progress mismatch: have %v/%v/%v, want %v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, targetBlocks/2+1, targetBlocks, targetBlocks)
	}
}

// Tests that synchronisation progress (origin block number and highest block
// number) is tracked and updated correctly in case of a fork (or manual head
// revertal).
func TestForkedSyncProgressFull(t *testing.T)  { testForkedSyncProgress(t, 1, FullSync) }
func TestForkedSyncProgressFast(t *testing.T)  { testForkedSyncProgress(t, 1, FastSync) }
func TestForkedSyncProgressLight(t *testing.T) { testForkedSyncProgress(t, 1, LightSync) }

func testForkedSyncProgress(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()

	tester := newTester()
	defer tester.terminate()

	// Create a forked chain to simulate origin revertal
	common, fork := MaxHashFetch, 2*MaxHashFetch
	hashesA, hashesB, headersA, headersB, blocksA, blocksB, receiptsA, receiptsB := tester.makeChainFork(common+fork, fork, tester.genesis, nil, true)

	// Set a sync init hook to catch progress changes
	starting := make(chan struct{})
	progress := make(chan struct{})

	tester.downloader.syncInitHook = func(origin, latest uint64) {
		starting <- struct{}{}
		<-progress
	}
	// Retrieve the sync progress and ensure they are zero (pristine sync)
	if progress := tester.downloader.Progress(); progress.StartingBlock != 0 || progress.CurrentBlock != 0 || progress.HighestBlock != 0 {
		t.Fatalf("Pristine progress mismatch: have %v/%v/%v, want %v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, 0, 0, 0)
	}
	// Synchronise with one of the forks and check progress
	tester.newPeer("fork A", protocol, hashesA, headersA, blocksA, receiptsA)
	pending := new(sync.WaitGroup)
	pending.Add(1)

	go func() {
		defer pending.Done()
		if err := tester.sync("fork A", nil, mode); err != nil {
			panic(fmt.Sprintf("failed to synchronise blocks: %v", err))
		}
	}()
	<-starting
	if progress := tester.downloader.Progress(); progress.StartingBlock != 0 || progress.CurrentBlock != 0 || progress.HighestBlock != uint64(len(hashesA)-1) {
		t.Fatalf("Initial progress mismatch: have %v/%v/%v, want %v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, 0, 0, len(hashesA)-1)
	}
	progress <- struct{}{}
	pending.Wait()

	// Simulate a successful sync above the fork
	tester.downloader.syncStatsChainOrigin = tester.downloader.syncStatsChainHeight

	// Synchronise with the second fork and check progress resets
	tester.newPeer("fork B", protocol, hashesB, headersB, blocksB, receiptsB)
	pending.Add(1)

	go func() {
		defer pending.Done()
		if err := tester.sync("fork B", nil, mode); err != nil {
			panic(fmt.Sprintf("failed to synchronise blocks: %v", err))
		}
	}()
	<-starting
	if progress := tester.downloader.Progress(); progress.StartingBlock != uint64(common) || progress.CurrentBlock != uint64(len(hashesA)-1) || progress.HighestBlock != uint64(len(hashesB)-1) {
		t.Fatalf("Forking progress mismatch: have %v/%v/%v, want %v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, common, len(hashesA)-1, len(hashesB)-1)
	}
	progress <- struct{}{}
	pending.Wait()

	// Check final progress after successful sync
	if progress := tester.downloader.Progress(); progress.StartingBlock != uint64(common) || progress.CurrentBlock != uint64(len(hashesB)-1) || progress.HighestBlock != uint64(len(hashesB)-1) {
		t.Fatalf("Final progress mismatch: have %v/%v/%v, want %v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, common, len(hashesB)-1, len(hashesB)-1)
	}
}

// Tests that if synchronisation is aborted due to some failure, then the progress
// origin is not updated in the next sync cycle, as it should be considered the
// continuation of the previous sync and not a new instance.
func TestFailedSyncProgressFull(t *testing.T)  { testFailedSyncProgress(t, 1, FullSync) }
func TestFailedSyncProgressFast(t *testing.T)  { testFailedSyncProgress(t, 1, FastSync) }
func TestFailedSyncProgressLight(t *testing.T) { testFailedSyncProgress(t, 1, LightSync) }

func testFailedSyncProgress(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()

	tester := newTester()
	defer tester.terminate()

	// Create a small enough block chain to download
	targetBlocks := blockCacheItems - 15
	hashes, headers, blocks, receipts := tester.makeChain(targetBlocks, 0, tester.genesis, nil, false)

	// Set a sync init hook to catch progress changes
	starting := make(chan struct{})
	progress := make(chan struct{})

	tester.downloader.syncInitHook = func(origin, latest uint64) {
		starting <- struct{}{}
		<-progress
	}
	// Retrieve the sync progress and ensure they are zero (pristine sync)
	if progress := tester.downloader.Progress(); progress.StartingBlock != 0 || progress.CurrentBlock != 0 || progress.HighestBlock != 0 {
		t.Fatalf("Pristine progress mismatch: have %v/%v/%v, want %v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, 0, 0, 0)
	}
	// Attempt a full sync with a faulty peer
	tester.newPeer("faulty", protocol, hashes, headers, blocks, receipts)
	missing := targetBlocks / 2
	delete(tester.peerHeaders["faulty"], hashes[missing])
	delete(tester.peerBlocks["faulty"], hashes[missing])
	delete(tester.peerReceipts["faulty"], hashes[missing])

	pending := new(sync.WaitGroup)
	pending.Add(1)

	go func() {
		defer pending.Done()
		if err := tester.sync("faulty", nil, mode); err == nil {
			panic("succeeded faulty synchronisation")
		}
	}()
	<-starting
	if progress := tester.downloader.Progress(); progress.StartingBlock != 0 || progress.CurrentBlock != 0 || progress.HighestBlock != uint64(targetBlocks) {
		t.Fatalf("Initial progress mismatch: have %v/%v/%v, want %v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, 0, 0, targetBlocks)
	}
	progress <- struct{}{}
	pending.Wait()

	// Synchronise with a good peer and check that the progress origin remind the same after a failure
	tester.newPeer("valid", protocol, hashes, headers, blocks, receipts)
	pending.Add(1)

	go func() {
		defer pending.Done()
		if err := tester.sync("valid", nil, mode); err != nil {
			panic(fmt.Sprintf("failed to synchronise blocks: %v", err))
		}
	}()
	<-starting
	if progress := tester.downloader.Progress(); progress.StartingBlock != 0 || progress.CurrentBlock > uint64(targetBlocks/2) || progress.HighestBlock != uint64(targetBlocks) {
		t.Fatalf("Completing progress mismatch: have %v/%v/%v, want %v/0-%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, 0, targetBlocks/2, targetBlocks)
	}
	progress <- struct{}{}
	pending.Wait()

	// Check final progress after successful sync
	if progress := tester.downloader.Progress(); progress.StartingBlock > uint64(targetBlocks/2) || progress.CurrentBlock != uint64(targetBlocks) || progress.HighestBlock != uint64(targetBlocks) {
		t.Fatalf("Final progress mismatch: have %v/%v/%v, want 0-%v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, targetBlocks/2, targetBlocks, targetBlocks)
	}
}

// Tests that if an attacker fakes a chain height, after the attack is detected,
// the progress height is successfully reduced at the next sync invocation.
func TestFakedSyncProgressFull(t *testing.T)  { testFakedSyncProgress(t, 1, FullSync) }
func TestFakedSyncProgressFast(t *testing.T)  { testFakedSyncProgress(t, 1, FastSync) }
func TestFakedSyncProgressLight(t *testing.T) { testFakedSyncProgress(t, 1, LightSync) }

func testFakedSyncProgress(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()

	tester := newTester()
	defer tester.terminate()

	// Create a small block chain
	targetBlocks := blockCacheItems - 15
	hashes, headers, blocks, receipts := tester.makeChain(targetBlocks+3, 0, tester.genesis, nil, false)

	// Set a sync init hook to catch progress changes
	starting := make(chan struct{})
	progress := make(chan struct{})

	tester.downloader.syncInitHook = func(origin, latest uint64) {
		starting <- struct{}{}
		<-progress
	}
	// Retrieve the sync progress and ensure they are zero (pristine sync)
	if progress := tester.downloader.Progress(); progress.StartingBlock != 0 || progress.CurrentBlock != 0 || progress.HighestBlock != 0 {
		t.Fatalf("Pristine progress mismatch: have %v/%v/%v, want %v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, 0, 0, 0)
	}
	//  Create and sync with an attacker that promises a higher chain than available
	tester.newPeer("attack", protocol, hashes, headers, blocks, receipts)
	for i := 1; i < 3; i++ {
		delete(tester.peerHeaders["attack"], hashes[i])
		delete(tester.peerBlocks["attack"], hashes[i])
		delete(tester.peerReceipts["attack"], hashes[i])
	}

	pending := new(sync.WaitGroup)
	pending.Add(1)

	go func() {
		defer pending.Done()
		if err := tester.sync("attack", nil, mode); err == nil {
			panic("succeeded attacker synchronisation")
		}
	}()
	<-starting
	if progress := tester.downloader.Progress(); progress.StartingBlock != 0 || progress.CurrentBlock != 0 || progress.HighestBlock != uint64(targetBlocks+3) {
		t.Fatalf("Initial progress mismatch: have %v/%v/%v, want %v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, 0, 0, targetBlocks+3)
	}
	progress <- struct{}{}
	pending.Wait()

	// Synchronise with a good peer and check that the progress height has been reduced to the true value
	tester.newPeer("valid", protocol, hashes[3:], headers, blocks, receipts)
	pending.Add(1)

	go func() {
		defer pending.Done()
		if err := tester.sync("valid", nil, mode); err != nil {
			panic(fmt.Sprintf("failed to synchronise blocks: %v", err))
		}
	}()
	<-starting
	if progress := tester.downloader.Progress(); progress.StartingBlock != 0 || progress.CurrentBlock > uint64(targetBlocks) || progress.HighestBlock != uint64(targetBlocks) {
		t.Fatalf("Completing progress mismatch: have %v/%v/%v, want %v/0-%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, 0, targetBlocks, targetBlocks)
	}
	progress <- struct{}{}
	pending.Wait()

	// Check final progress after successful sync
	if progress := tester.downloader.Progress(); progress.StartingBlock > uint64(targetBlocks) || progress.CurrentBlock != uint64(targetBlocks) || progress.HighestBlock != uint64(targetBlocks) {
		t.Fatalf("Final progress mismatch: have %v/%v/%v, want 0-%v/%v/%v", progress.StartingBlock, progress.CurrentBlock, progress.HighestBlock, targetBlocks, targetBlocks, targetBlocks)
	}
}

// This test reproduces an issue where unexpected deliveries would
// block indefinitely if they arrived at the right time.
// We use data driven subtests to manage this so that it will be parallel on its own
// and not with the other tests, avoiding intermittent failures.
func TestDeliverHeadersHang(t *testing.T) {
	testCases := []struct {
		protocol int
		syncMode SyncMode
	}{
		{1, FullSync},
		{1, FastSync},
		{1, LightSync},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("protocol %d mode %v", tc.protocol, tc.syncMode), func(t *testing.T) {
			testDeliverHeadersHang(t, tc.protocol, tc.syncMode)
		})
	}
}

type floodingTestPeer struct {
	peer   Peer
	tester *downloadTester
	pend   sync.WaitGroup
}

func (ftp *floodingTestPeer) Head() (common.Hash, *big.Int) { return ftp.peer.Head() }
func (ftp *floodingTestPeer) RequestHeadersByHash(hash common.Hash, count int, skip int, reverse bool) error {
	return ftp.peer.RequestHeadersByHash(hash, count, skip, reverse)
}
func (ftp *floodingTestPeer) RequestBodies(hashes []common.Hash) error {
	return ftp.peer.RequestBodies(hashes)
}
func (ftp *floodingTestPeer) RequestReceipts(hashes []common.Hash) error {
	return ftp.peer.RequestReceipts(hashes)
}
func (ftp *floodingTestPeer) RequestNodeData(hashes []common.Hash) error {
	return ftp.peer.RequestNodeData(hashes)
}

func (ftp *floodingTestPeer) RequestHeadersByNumber(from uint64, count, skip int, reverse bool) error {
	deliveriesDone := make(chan struct{}, 500)
	for i := 0; i < cap(deliveriesDone); i++ {
		peer := fmt.Sprintf("fake-peer%d", i)
		ftp.pend.Add(1)

		go func() {
			ftp.tester.downloader.DeliverHeaders(peer, []*types.Header{{}, {}, {}, {}})
			deliveriesDone <- struct{}{}
			ftp.pend.Done()
		}()
	}
	// Deliver the actual requested headers.
	go ftp.peer.RequestHeadersByNumber(from, count, skip, reverse)
	// None of the extra deliveries should block.
	timeout := time.After(60 * time.Second)
	for i := 0; i < cap(deliveriesDone); i++ {
		select {
		case <-deliveriesDone:
		case <-timeout:
			panic("blocked")
		}
	}
	return nil
}

func testDeliverHeadersHang(t *testing.T, protocol int, mode SyncMode) {
	t.Parallel()

	master := newTester()
	defer master.terminate()

	hashes, headers, blocks, receipts := master.makeChain(5, 0, master.genesis, nil, false)
	for i := 0; i < 10; i++ {
		tester := newTester()
		tester.peerDb = master.peerDb

		tester.newPeer("peer", protocol, hashes, headers, blocks, receipts)
		// Whenever the downloader requests headers, flood it with
		// a lot of unrequested header deliveries.
		tester.downloader.peers.peers["peer"].peer = &floodingTestPeer{
			peer:   tester.downloader.peers.peers["peer"].peer,
			tester: tester,
		}
		if err := tester.sync("peer", nil, mode); err != nil {
			t.Errorf("test %d: sync failed: %v", i, err)
		}
		tester.terminate()

		// Flush all goroutines to prevent messing with subsequent tests
		tester.downloader.peers.peers["peer"].peer.(*floodingTestPeer).pend.Wait()
	}
}
