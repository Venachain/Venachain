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
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/consensus/istanbul"
	"github.com/Venachain/Venachain/core/state"
	"github.com/Venachain/Venachain/core/types"
	"github.com/Venachain/Venachain/crypto"
	"github.com/Venachain/Venachain/event"
	"github.com/Venachain/Venachain/params"
	"github.com/Venachain/Venachain/venadb/memorydb"
)

// testTxPoolConfig is a transaction pool configuration without stateful disk
// sideeffects used during testing.
var testTxPoolConfig TxPoolConfig

var DefaultConfig = &params.IstanbulConfig{
	RequestTimeout: 10000,
	BlockPeriod:    1,
	ProposerPolicy: istanbul.RoundRobin,
}

var TestChainConfig = params.ChainConfig{
	ChainID:       big.NewInt(1),
	Istanbul:      DefaultConfig,
	VMInterpreter: "wasm",
}

func init() {
	testTxPoolConfig = DefaultTxPoolConfig
	testTxPoolConfig.Journal = ""
}

type testBlockChain struct {
	statedb                   *state.StateDB
	gasLimit                  uint64
	chainHeadFeed             event.Feed
	blockConsensusFinishEvent event.Feed
}

func (bc *testBlockChain) CurrentBlock() *types.Block {
	return types.NewBlock(&types.Header{
		GasLimit: bc.gasLimit,
	}, nil, nil)
}

func (bc *testBlockChain) GetBlock(hash common.Hash, number uint64) *types.Block {
	return bc.CurrentBlock()
}

func (bc *testBlockChain) SubscribeChainHeadEvent(ch chan<- ChainHeadEvent) event.Subscription {
	return bc.chainHeadFeed.Subscribe(ch)
}

func (bc *testBlockChain) SubscribeBlockConsensusFinishEvent(ch chan<- BlockConsensusFinishEvent) event.Subscription {
	return bc.blockConsensusFinishEvent.Subscribe(ch)
}

func (bc *testBlockChain) GetState(header *types.Header) (*state.StateDB, error) {
	return bc.statedb, nil
}

func transaction(nonce uint64, gaslimit uint64, key *ecdsa.PrivateKey) *types.Transaction {
	return pricedTransaction(nonce, gaslimit, big.NewInt(1), key)
}

func pricedTransaction(nonce uint64, gaslimit uint64, gasprice *big.Int, key *ecdsa.PrivateKey) *types.Transaction {
	tx, _ := types.SignTx(types.NewTransaction(nonce, common.Address{}, big.NewInt(100), gaslimit, gasprice, nil), types.HomesteadSigner{}, key)
	return tx
}

func setupTxPool() (*TxPool, *ecdsa.PrivateKey) {
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(memorydb.NewMemDatabase()))
	blockchain := &testBlockChain{
		statedb:  statedb,
		gasLimit: 1000000,
	}
	db := memorydb.NewMemDatabase()

	key, _ := crypto.GenerateKey()
	pool := NewTxPool(testTxPoolConfig, &TestChainConfig, blockchain, db, nil, key)

	return pool, key
}

// validateTxPoolInternals checks various consistency invariants within the pool.
func validateTxPoolInternals(pool *TxPool) error {
	pool.mu.RLock()
	defer pool.mu.RUnlock()

	// Ensure the total transaction set is consistent with pending + queued
	pending, queued := pool.stats()
	if total := pool.all.Count(); total != pending+queued {
		return fmt.Errorf("total transaction count %d != %d pending + %d queued", total, pending, queued)
	}

	return nil
}

// validateEvents checks that the correct number of transaction addition events
// were fired on the pool's event feed.
func validateEvents(events chan NewTxsEvent, count int) error {
	var received []*types.Transaction

	for len(received) < count {
		select {
		case ev := <-events:
			received = append(received, ev.Txs...)
		case <-time.After(time.Second):
			return fmt.Errorf("event #%v not fired", received)
		}
	}
	if len(received) > count {
		return fmt.Errorf("more than %d events fired: %v", count, received[count:])
	}
	select {
	case ev := <-events:
		return fmt.Errorf("more than %d events fired: %v", count, ev.Txs)

	case <-time.After(50 * time.Millisecond):
		// This branch should be "default", but it's a data race between goroutines,
		// reading the event channel and pushing into it, so better wait a bit ensuring
		// really nothing gets injected.
	}
	return nil
}

func deriveSender(tx *types.Transaction) (common.Address, error) {
	return types.Sender(types.HomesteadSigner{}, tx)
}

type testChain struct {
	*testBlockChain
	address common.Address
	trigger *bool
}

// testChain.State() is used multiple times to reset the pending state.
// when simulate is true it will create a state that indicates
// that tx0 and tx1 are included in the chain.
func (c *testChain) State() (*state.StateDB, error) {
	// delay "state change" by one. The tx pool fetches the
	// state multiple times and by delaying it a bit we simulate
	// a state change between those fetches.
	stdb := c.statedb
	if *c.trigger {
		c.statedb, _ = state.New(common.Hash{}, state.NewDatabase(memorydb.NewMemDatabase()))
		// simulate that the new head block included tx0 and tx1
		c.statedb.SetNonce(c.address, 2)
		c.statedb.SetBalance(c.address, new(big.Int).SetUint64(params.Ether))
		*c.trigger = false
	}
	return stdb, nil
}

// This test simulates a scenario where a new block is imported during a
// state reset and tests whether the pending state is in sync with the
// block head event that initiated the resetState().
func TestStateChangeDuringTransactionPoolReset(t *testing.T) {
	t.Parallel()
	var (
		key, _     = crypto.GenerateKey()
		address    = crypto.PubkeyToAddress(key.PublicKey)
		statedb, _ = state.New(common.Hash{}, state.NewDatabase(memorydb.NewMemDatabase()))
		trigger    = false
	)
	// setup pool with 2 transaction in it
	statedb.SetBalance(address, new(big.Int).SetUint64(params.Ether))
	blockchain := &testChain{
		testBlockChain: &testBlockChain{
			statedb:  statedb,
			gasLimit: 1000000000,
		},
		address: address,
		trigger: &trigger,
	}

	tx0 := transaction(1, 100000, key)
	tx1 := transaction(2, 100000, key)

	db := memorydb.NewMemDatabase()
	pool := NewTxPool(testTxPoolConfig, &TestChainConfig, blockchain, db, nil, key)
	defer pool.Stop()

	tx_numbers := pool.GetTxCount()
	if tx_numbers != 0 {
		t.Fatalf("Invalid tx size, want 0, got %d", tx_numbers)
	}
	pool.AddRemotes(types.Transactions{tx0, tx1})

	//nonce = pool.State().GetNonce(address)
	tx_numbers = pool.GetTxCount()

	if tx_numbers != 2 {
		t.Fatalf("Invalid tx pool size, want 2, got %d", tx_numbers)
	}

	//trigger state change in the background
	trigger = true
	pool.lockedReset(nil, nil)

	_, err := pool.Pending()
	if err != nil {
		t.Fatalf("Could not fetch pending transactions: %v", err)
	}
	tx_numbers = pool.GetTxCount()
	if tx_numbers != 2 {
		t.Fatalf("Invalid tx pool size, want 2, got %d", tx_numbers)
	}
}

func TestInvalidTransactions(t *testing.T) {
	t.Parallel()
	pool, key := setupTxPool()
	defer pool.Stop()

	tx := transaction(0, 100, key)

	from, _ := deriveSender(tx)
	pool.currentState.AddBalance(from, big.NewInt(1))
	if err := pool.AddRemote(tx); err != ErrInsufficientFunds {
		t.Error("expected", ErrInsufficientFunds)
	}

	balance := new(big.Int).Add(tx.Value(), new(big.Int).Mul(new(big.Int).SetUint64(tx.Gas()), tx.GasPrice()))
	pool.currentState.AddBalance(from, balance)
	if err := pool.AddLocal(tx); err != nil {
		t.Error("expected", nil, "got", err)
	}
}

func TestTransactionNegativeValue(t *testing.T) {
	t.Parallel()

	pool, key := setupTxPool()
	defer pool.Stop()

	tx, _ := types.SignTx(types.NewTransaction(0, common.Address{}, big.NewInt(-1), 100, big.NewInt(1), nil), types.HomesteadSigner{}, key)
	from, _ := deriveSender(tx)
	pool.currentState.AddBalance(from, big.NewInt(1))
	if err := pool.AddRemote(tx); err != ErrNegativeValue {
		t.Error("expected", ErrNegativeValue, "got", err)
	}
}

func TestDuplicateTx(t *testing.T) {
	t.Parallel()

	pool, key := setupTxPool()
	defer pool.Stop()

	addr := crypto.PubkeyToAddress(key.PublicKey)
	resetState := func() {
		statedb, _ := state.New(common.Hash{}, state.NewDatabase(memorydb.NewMemDatabase()))
		statedb.AddBalance(addr, big.NewInt(100000000000000))
		pool.chain = &testBlockChain{
			statedb:  statedb,
			gasLimit: 1000000,
		}
		pool.lockedReset(nil, nil)
	}
	resetState()

	signer := types.HomesteadSigner{}
	tx1, _ := types.SignTx(types.NewTransaction(0, common.Address{}, big.NewInt(100), 100000, big.NewInt(1), nil), signer, key)
	tx2, _ := types.SignTx(types.NewTransaction(0, common.Address{}, big.NewInt(100), 100000, big.NewInt(1), nil), signer, key)

	// Add the first two transaction, ensure higher priced stays only
	if _, err := pool.add(tx1, false); err != nil {
		t.Errorf("first transaction insert failed (%v)", err)
	}
	if _, err := pool.add(tx2, false); err != nil {
		t.Errorf("first transaction insert failed (%v)", err)
	}

	if pool.pending[addr].Len() != 1 {
		t.Error("expected 1 pending transactions, got", pool.pending[addr].Len())
	}
	// Ensure the total transaction count is correct
	if pool.all.Count() != 1 {
		t.Error("expected 1 total transactions, got", pool.all.Count())
	}
}

func TestTransactionMissingNonce(t *testing.T) {
	t.Parallel()

	pool, key := setupTxPool()
	defer pool.Stop()

	addr := crypto.PubkeyToAddress(key.PublicKey)
	pool.currentState.AddBalance(addr, big.NewInt(100000000000000))
	tx := transaction(1, 100000, key)
	if _, err := pool.add(tx, false); err != nil {
		t.Error("didn't expect error", err)
	}

	// Because there is no queue, the transaction can be directly added to pending.
	if len(pool.pending) != 1 {
		t.Error("expected 1 pending transactions, got", len(pool.pending))
	}
	if pool.all.Count() != 1 {
		t.Error("expected 1 total transactions, got", pool.all.Count())
	}
}

// Tests that if an account runs out of funds, any pending and queued transactions
// are dropped.
func TestTransactionDropping(t *testing.T) {
	t.Parallel()

	// Create a test account and fund it
	pool, key := setupTxPool()
	defer pool.Stop()

	account, _ := deriveSender(transaction(0, 0, key))
	pool.currentState.AddBalance(account, big.NewInt(1000))

	// Add some pending and some queued transactions
	var (
		tx0 = transaction(0, 100, key)
		tx1 = transaction(1, 200, key)
		tx2 = transaction(2, 300, key)
	)
	pool.promoteTx(account, tx0.Hash(), tx0)
	pool.promoteTx(account, tx1.Hash(), tx1)
	pool.promoteTx(account, tx2.Hash(), tx2)

	// Check that pre and post validations leave the pool as is
	if pool.pending[account].Len() != 3 {
		t.Errorf("pending transaction mismatch: have %d, want %d", pool.pending[account].Len(), 3)
	}

	if pool.all.Count() != 3 {
		t.Errorf("total transaction mismatch: have %d, want %d", pool.all.Count(), 3)
	}
	if pool.pending[account].Len() != 3 {
		t.Errorf("pending transaction mismatch: have %d, want %d", pool.pending[account].Len(), 3)
	}

	if pool.all.Count() != 3 {
		t.Errorf("total transaction mismatch: have %d, want %d", pool.all.Count(), 3)
	}

	// Reduce the balance of the account, and check that invalidated transactions are dropped
	pool.currentState.AddBalance(account, big.NewInt(-650))
	pool.lockedReset(nil, nil)

	if pool.all.Count() != 3 {
		t.Errorf("total transaction mismatch: have %d, want %d", pool.all.Count(), 3)
	}
}

// Tests that if the transaction pool has both executable and non-executable
// transactions from an origin account, filling the nonce gap moves all queued
// ones into the pending pool.
func TestTransactionGapFilling(t *testing.T) {
	t.Parallel()

	// Create a test account and fund it
	pool, key := setupTxPool()
	defer pool.Stop()

	account, _ := deriveSender(transaction(0, 0, key))
	pool.currentState.AddBalance(account, big.NewInt(1000000))

	// Keep track of transaction events to ensure all executables get announced
	events := make(chan NewTxsEvent, 69)
	sub := pool.txFeed.Subscribe(events)
	defer sub.Unsubscribe()

	// Create a pending and a queued transaction with a nonce-gap in between
	if err := pool.AddRemote(transaction(0, 100000, key)); err != nil {
		t.Fatalf("failed to add pending transaction: %v", err)
	}
	if err := pool.AddRemote(transaction(2, 100000, key)); err != nil {
		t.Fatalf("failed to add queued transaction: %v", err)
	}
	pending, _ := pool.Stats()
	if pending != 2 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 2)
	}

	if err := validateEvents(events, 2); err != nil {
		t.Fatalf("original event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// Fill the nonce gap and ensure all transactions become pending
	if err := pool.AddRemote(transaction(1, 100000, key)); err != nil {
		t.Fatalf("failed to add gapped transaction: %v", err)
	}
	pending, _ = pool.Stats()
	if pending != 3 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 3)
	}

	if err := validateEvents(events, 1); err != nil {
		t.Fatalf("gap-filling event firing failed: %v", err)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
}

// Tests that the transaction limits are enforced the same way irrelevant whether
// the transactions are added one by one or in batches.
func TestTransactionQueueLimitingEquivalency(t *testing.T) { testTransactionLimitingEquivalency(t, 1) }
func TestTransactionPendingLimitingEquivalency(t *testing.T) {
	testTransactionLimitingEquivalency(t, 0)
}

func testTransactionLimitingEquivalency(t *testing.T, origin uint64) {
	t.Parallel()

	// Add a batch of transactions to a pool one by one
	pool1, key1 := setupTxPool()
	defer pool1.Stop()

	account1, _ := deriveSender(transaction(0, 0, key1))
	pool1.currentState.AddBalance(account1, big.NewInt(1000000))

	for i := uint64(0); i < 5; i++ {
		if err := pool1.AddRemote(transaction(origin+i, 100000, key1)); err != nil {
			t.Fatalf("tx %d: failed to add transaction: %v", i, err)
		}
	}
	// Add a batch of transactions to a pool in one big batch
	pool2, key2 := setupTxPool()
	defer pool2.Stop()

	account2, _ := deriveSender(transaction(0, 0, key2))
	pool2.currentState.AddBalance(account2, big.NewInt(1000000))

	txs := []*types.Transaction{}
	for i := uint64(0); i < 5; i++ {
		txs = append(txs, transaction(origin+i, 100000, key2))
	}
	pool2.AddRemotes(txs)

	// Ensure the batch optimization honors the same pool mechanics
	if len(pool1.pending) != len(pool2.pending) {
		t.Errorf("pending transaction count mismatch: one-by-one algo: %d, batch algo: %d", len(pool1.pending), len(pool2.pending))
	}
	if pool1.all.Count() != pool2.all.Count() {
		t.Errorf("total transaction count mismatch: one-by-one algo %d, batch algo %d", pool1.all.Count(), pool2.all.Count())
	}
	if err := validateTxPoolInternals(pool1); err != nil {
		t.Errorf("pool 1 internal state corrupted: %v", err)
	}
	if err := validateTxPoolInternals(pool2); err != nil {
		t.Errorf("pool 2 internal state corrupted: %v", err)
	}
}

// Tests that if the transaction count belonging to multiple accounts go above
// some hard threshold, the higher transactions are dropped to prevent DOS
// attacks.
func TestTransactionPendingGlobalLimiting(t *testing.T) {
	t.Parallel()

	pool, _ := setupTxPool()
	pool.config.GlobalSlots = 1
	defer pool.Stop()

	// Create a number of test accounts and fund them
	keys := make([]*ecdsa.PrivateKey, 1)
	for i := 0; i < len(keys); i++ {
		keys[i], _ = crypto.GenerateKey()
		pool.currentState.AddBalance(crypto.PubkeyToAddress(keys[i].PublicKey), big.NewInt(1000000))
	}
	// Generate and queue a batch of transactions
	nonces := make(map[common.Address]uint64)

	txs := types.Transactions{}
	for _, key := range keys {
		addr := crypto.PubkeyToAddress(key.PublicKey)
		for j := 0; j < 2; j++ {
			txs = append(txs, transaction(nonces[addr], 100000, key))
			nonces[addr]++
		}
	}
	// Import the batch and verify that limits have been enforced
	result := 0
	for _, j := range txs {
		if err := pool.AddRemote(j); err == ErrTxpoolIsFull {
			result = 1
		}
	}
	if result != 1 {
		t.Error("expected", ErrTxpoolIsFull)
	}
}

// Tests that if transactions start being capped, transactions are also removed from 'all'
func TestTransactionCapClearsFromAll(t *testing.T) {
	t.Parallel()

	config := testTxPoolConfig
	config.GlobalSlots = 8
	pool, _ := setupTxPool()
	defer pool.Stop()

	// Create a number of test accounts and fund them
	key, _ := crypto.GenerateKey()
	addr := crypto.PubkeyToAddress(key.PublicKey)
	pool.currentState.AddBalance(addr, big.NewInt(1000000))

	txs := types.Transactions{}
	for j := 0; j < int(config.GlobalSlots)*2; j++ {
		txs = append(txs, transaction(uint64(j), 100000, key))
	}
	// Import the batch and verify that limits have been enforced
	pool.AddRemotes(txs)
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
}

// Tests that local transactions are journaled to disk, but remote transactions
// get discarded between restarts.
func TestTransactionJournaling(t *testing.T)         { testTransactionJournaling(t, false) }
func TestTransactionJournalingNoLocals(t *testing.T) { testTransactionJournaling(t, true) }

func testTransactionJournaling(t *testing.T, nolocals bool) {
	t.Parallel()

	// Create a temporary file for the journal
	file, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("failed to create temporary journal: %v", err)
	}
	journal := file.Name()
	defer os.Remove(journal)

	// Clean up the temporary file, we only need the path for now
	file.Close()
	os.Remove(journal)

	config := testTxPoolConfig
	config.NoLocals = nolocals
	config.Journal = journal
	config.Rejournal = time.Second

	pool, _ := setupTxPool()

	// Create two test accounts to ensure remotes expire but locals do not
	local, _ := crypto.GenerateKey()
	remote, _ := crypto.GenerateKey()

	pool.currentState.AddBalance(crypto.PubkeyToAddress(local.PublicKey), big.NewInt(1000000000))
	pool.currentState.AddBalance(crypto.PubkeyToAddress(remote.PublicKey), big.NewInt(1000000000))

	// Add three local and a remote transactions and ensure they are queued up
	if err := pool.AddLocal(pricedTransaction(0, 100000, big.NewInt(1), local)); err != nil {
		t.Fatalf("failed to add local transaction: %v", err)
	}
	if err := pool.AddLocal(pricedTransaction(1, 100000, big.NewInt(1), local)); err != nil {
		t.Fatalf("failed to add local transaction: %v", err)
	}
	if err := pool.AddLocal(pricedTransaction(2, 100000, big.NewInt(1), local)); err != nil {
		t.Fatalf("failed to add local transaction: %v", err)
	}
	if err := pool.AddRemote(pricedTransaction(0, 100000, big.NewInt(1), remote)); err != nil {
		t.Fatalf("failed to add remote transaction: %v", err)
	}
	pending, queued := pool.Stats()
	fmt.Println("pending", pending)
	if pending != 4 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 4)
	}
	if queued != 0 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 0)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}

	pool.Stop()
	statedb, _ := state.New(common.Hash{}, state.NewDatabase(memorydb.NewMemDatabase()))
	blockchain := &testBlockChain{
		statedb:  statedb,
		gasLimit: 1000000,
	}
	db := memorydb.NewMemDatabase()

	key, _ := crypto.GenerateKey()
	pool = NewTxPool(testTxPoolConfig, &TestChainConfig, blockchain, db, nil, key)
	statedb.SetNonce(crypto.PubkeyToAddress(local.PublicKey), 1)
	pending, queued = pool.Stats()
	if queued != 0 {
		t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 0)
	}
	if nolocals {
		if pending != 0 {
			t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 0)
		}
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}

	// Bump the nonce temporarily and ensure the newly invalidated transaction is removed
	//statedb.SetNonce(crypto.PubkeyToAddress(local.PublicKey), 2)
	pool.lockedReset(nil, nil)
	time.Sleep(2 * config.Rejournal)
	pool.Stop()

	statedb, _ = state.New(common.Hash{}, state.NewDatabase(memorydb.NewMemDatabase()))
	blockchain = &testBlockChain{
		statedb:  statedb,
		gasLimit: 1000000,
	}
	db = memorydb.NewMemDatabase()

	key, _ = crypto.GenerateKey()
	pool = NewTxPool(testTxPoolConfig, &TestChainConfig, blockchain, db, nil, key)

	statedb.SetNonce(crypto.PubkeyToAddress(local.PublicKey), 1)
	pending, queued = pool.Stats()
	if pending != 0 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 0)
	}
	if nolocals {
		if queued != 0 {
			t.Fatalf("queued transactions mismatched: have %d, want %d", queued, 0)
		}
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	pool.Stop()
}

// TestTransactionStatusCheck tests that the pool can correctly retrieve the
// pending status of individual transactions.
func TestTransactionStatusCheck(t *testing.T) {
	t.Parallel()

	pool, _ := setupTxPool()
	defer pool.Stop()

	// Create the test accounts to check various transaction statuses with
	keys := make([]*ecdsa.PrivateKey, 3)
	for i := 0; i < len(keys); i++ {
		keys[i], _ = crypto.GenerateKey()
		pool.currentState.AddBalance(crypto.PubkeyToAddress(keys[i].PublicKey), big.NewInt(1000000))
	}
	// Generate and queue a batch of transactions, both pending and queued
	txs := types.Transactions{}

	txs = append(txs, pricedTransaction(0, 100000, big.NewInt(1), keys[0])) // Pending only
	txs = append(txs, pricedTransaction(2, 100000, big.NewInt(1), keys[1]))

	// Import the transaction and ensure they are correctly added
	pool.AddRemotes(txs)

	pending, _ := pool.Stats()
	if pending != 2 {
		t.Fatalf("pending transactions mismatched: have %d, want %d", pending, 2)
	}
	if err := validateTxPoolInternals(pool); err != nil {
		t.Fatalf("pool internal state corrupted: %v", err)
	}
	// Retrieve the status of each transaction and validate them
	hashes := make([]common.Hash, len(txs))
	for i, tx := range txs {
		hashes[i] = tx.Hash()
	}
	hashes = append(hashes, common.Hash{})

	statuses := pool.Status(hashes)
	expect := []TxStatus{TxStatusPending, TxStatusPending, TxStatusUnknown}

	for i := 0; i < len(statuses); i++ {
		if statuses[i] != expect[i] {
			t.Errorf("transaction %d: status mismatch: have %v, want %v", i, statuses[i], expect[i])
		}
	}
}

// Benchmarks the speed of validating the contents of the pending queue of the
// transaction pool.
func BenchmarkPendingDemotion100(b *testing.B)   { benchmarkPendingDemotion(b, 100) }
func BenchmarkPendingDemotion1000(b *testing.B)  { benchmarkPendingDemotion(b, 1000) }
func BenchmarkPendingDemotion10000(b *testing.B) { benchmarkPendingDemotion(b, 10000) }

func benchmarkPendingDemotion(b *testing.B, size int) {
	// Add a batch of transactions to a pool one by one
	pool, key := setupTxPool()
	defer pool.Stop()

	blockHash := common.Hash{}
	account, _ := deriveSender(transaction(0, 0, key))
	pool.currentState.AddBalance(account, big.NewInt(1000000))
	pool.recentRemovedPending.Add(blockHash, struct{}{})

	var txs types.Transactions
	for i := 0; i < size; i++ {
		tx := transaction(uint64(i), 100000, key)
		txs = append(txs, tx)
		pool.promoteTx(account, tx.Hash(), tx)
	}
	// Benchmark the speed of pool validation
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.demoteUnexecutables(txs, blockHash)
	}
}

// Benchmarks the speed of iterative transaction insertion.
func BenchmarkPoolInsert(b *testing.B) {
	// Generate a batch of transactions to enqueue into the pool
	pool, key := setupTxPool()
	defer pool.Stop()

	account, _ := deriveSender(transaction(0, 0, key))
	pool.currentState.AddBalance(account, big.NewInt(1000000))

	txs := make(types.Transactions, 10000)
	for i := 0; i < 10000; i++ {
		txs[i] = transaction(uint64(i), 100000, key)
	}
	// Benchmark importing the transactions into the queue
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.AddRemotes(txs)
	}
}

// Benchmarks the speed of batched transaction insertion.
func BenchmarkPoolBatchInsert100(b *testing.B)   { benchmarkPoolBatchInsert(b, 100) }
func BenchmarkPoolBatchInsert1000(b *testing.B)  { benchmarkPoolBatchInsert(b, 1000) }
func BenchmarkPoolBatchInsert10000(b *testing.B) { benchmarkPoolBatchInsert(b, 10000) }

func benchmarkPoolBatchInsert(b *testing.B, size int) {
	// Generate a batch of transactions to enqueue into the pool
	pool, key := setupTxPool()
	defer pool.Stop()

	account, _ := deriveSender(transaction(0, 0, key))
	pool.currentState.AddBalance(account, big.NewInt(1000000))

	batches := make([]types.Transactions, b.N)
	for i := 0; i < b.N; i++ {
		batches[i] = make(types.Transactions, size)
		for j := 0; j < size; j++ {
			batches[i][j] = transaction(uint64(size*i+j), 100000, key)
		}
	}
	// Benchmark importing the transactions into the queue
	b.ResetTimer()
	for _, batch := range batches {
		pool.AddRemotes(batch)
	}
}

func TestTxNumber(t *testing.T) {
	// Generate a batch of transactions to enqueue into the pool
	pool, key := setupTxPool()
	defer pool.Stop()
	//The number of transactions should be greater than config.GlobalSlots. Default GlobalSlots 40960.
	txsize := 40960
	account, _ := deriveSender(transaction(0, 0, key))
	pool.currentState.AddBalance(account, big.NewInt(1000000))

	txs := make(types.Transactions, txsize)
	for i := 0; i < txsize; i++ {
		txs[i] = transaction(uint64(i), 100000, key)
	}
	pool.AddRemotes(txs)
	if len(txs) != pool.GetTxCount() {
		t.Fatalf("tx number is wrong, want %d,got %d", len(txs), pool.GetTxCount())
	}
}

func BenchmarkTxPool_AddRemotes(b *testing.B) {
	pool, key := setupTxPool()
	defer pool.Stop()

	account, _ := deriveSender(transaction(0, 0, key))
	pool.currentState.AddBalance(account, big.NewInt(1000000))
	b.ResetTimer()
	batches := make([]types.Transactions, b.N)
	for i := 0; i < b.N; i++ {
		batches[i] = make(types.Transactions, 40960)
		for j := 0; j < 40960; j++ {
			batches[i][j] = transaction(uint64(40960*i+j), uint64(j), key)
		}
		pool.AddRemotes(batches[i])
	}
}

func BenchmarkTxPool_AddRemote(b *testing.B) {
	pool, key := setupTxPool()
	defer pool.Stop()

	account, _ := deriveSender(transaction(0, 0, key))
	pool.currentState.AddBalance(account, big.NewInt(1000000))
	b.ResetTimer()
	batches := make([]types.Transactions, b.N)
	for i := 0; i < b.N; i++ {
		batches[i] = make(types.Transactions, 40960)
		for j := 0; j < 40960; j++ {
			batches[i][j] = transaction(uint64(40960*i+j), uint64(j), key)
		}
		for _, j := range batches[i] {
			pool.AddRemote(j)
		}
	}
}
