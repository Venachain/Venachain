// Copyright 2014 The go-ethereum Authors
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

// Package state provides a caching layer atop the Ethereum state trie.
package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"

	"github.com/PlatONEnetwork/PlatONE-Go/crypto/sha3"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"github.com/PlatONEnetwork/PlatONE-Go/trie"
)

type revision struct {
	id           int
	journalIndex int
}

var (
	StoragePrefix = "storage-value-"
	// emptyState is the known hash of an empty state trie entry.
	emptyState = crypto.Keccak256Hash(nil)

	// emptyCode is the known hash of the empty EVM bytecode.
	emptyCode = crypto.Keccak256Hash(nil)

	emptyStorage = crypto.Keccak256Hash([]byte(StoragePrefix))

	cloneErr = errors.New("clone account error!")
)

// StateDBs within the ethereum protocol are used to store anything
// within the merkle trie. StateDBs take care of caching and storing
// nested states. It's the general query interface to retrieve:
// * Contracts
// * Accounts
type StateDB struct {
	db   Database
	trie Trie

	// This map holds 'live' objects, which will get modified while processing a state transition.
	objLock           sync.RWMutex
	stateObjects      map[common.Address]*stateObject
	stateObjectsDirty map[common.Address]struct{}

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// The refund counter, also used by state transitioning.
	refund uint64

	thash, bhash common.Hash
	txIndex      int
	logs         map[common.Hash][]*types.Log
	logSize      uint

	preimages map[common.Hash][]byte

	// Journal of state modifications. This is the backbone of
	// Snapshot and RevertToSnapshot.
	journal        *journal
	validRevisions []revision
	nextRevisionId int

	lock sync.Mutex

	// 用于判断是否正在执行交易
	process bool
	// 执行的过程锁
	rwLock sync.RWMutex
	// 缓存读操作 键 addr+key
	readMap map[string]*ReadOp
	// 缓存写操作 键 addr+key
	writeMap map[string]*WriteOp
	// 缓存余额变动 键 addr
	balanceMap map[common.Address]*BalanceOp
	// 缓存nonce的增加数
	nonce map[common.Address]int
	// 缓存state_object的变动
	oc map[common.Address]ObjectChange
	// 缓存已经处理的交易
	txs []*types.Transaction
	// 缓存已经处理的交易的receipt
	receipts []*types.Receipt
	// 缓存交易的依赖关系
	dag types.DAG
	// gas使用
	gasUsed uint64
	//用于生成header的txHash
	txTrie trie.Generator
	//用于生成header的receiptHash
	receiptTrie trie.Generator
}

// Create a new state from a given trie.
func New(root common.Hash, db Database) (*StateDB, error) {
	tr, err := db.OpenTrie(root)
	if err != nil {
		return nil, err
	}
	return &StateDB{
		db:                db,
		trie:              tr,
		stateObjects:      make(map[common.Address]*stateObject),
		stateObjectsDirty: make(map[common.Address]struct{}),
		logs:              make(map[common.Hash][]*types.Log),
		preimages:         make(map[common.Hash][]byte),
		readMap:           make(map[string]*ReadOp),
		writeMap:          make(map[string]*WriteOp),
		balanceMap:        make(map[common.Address]*BalanceOp),
		nonce:             make(map[common.Address]int),
		oc:                make(map[common.Address]ObjectChange),
		journal:           newJournal(),
	}, nil
}

// setError remembers the first non-nil error it is called with.
func (self *StateDB) setError(err error) {
	if self.dbErr == nil {
		self.dbErr = err
	}
}

func (self *StateDB) Error() error {
	return self.dbErr
}

// Reset clears out all ephemeral state objects from the state db, but keeps
// the underlying state trie to avoid reloading data for the next operations.
func (self *StateDB) Reset(root common.Hash) error {
	tr, err := self.db.OpenTrie(root)
	if err != nil {
		return err
	}
	self.trie = tr
	self.stateObjects = make(map[common.Address]*stateObject)
	self.stateObjectsDirty = make(map[common.Address]struct{})
	self.thash = common.Hash{}
	self.bhash = common.Hash{}
	self.txIndex = 0
	self.logs = make(map[common.Hash][]*types.Log)
	self.logSize = 0
	self.preimages = make(map[common.Hash][]byte)
	self.clearJournalAndRefund()
	return nil
}

func (self *StateDB) AddLog(log *types.Log) {
	self.journal.append(addLogChange{txhash: self.thash})

	log.TxHash = self.thash
	log.BlockHash = self.bhash
	log.TxIndex = uint(self.txIndex)
	log.Index = self.logSize
	self.logs[self.thash] = append(self.logs[self.thash], log)
	self.logSize++
}

func (self *StateDB) GetLogs(hash common.Hash) []*types.Log {
	return self.logs[hash]
}

func (self *StateDB) Logs() []*types.Log {
	var logs []*types.Log
	for _, lgs := range self.logs {
		logs = append(logs, lgs...)
	}
	return logs
}

// AddPreimage records a SHA3 preimage seen by the VM.
func (self *StateDB) AddPreimage(hash common.Hash, preimage []byte) {
	log.Info("addPreimage", "hash", hash.String())
	if _, ok := self.preimages[hash]; !ok {
		self.journal.append(addPreimageChange{hash: hash})
		pi := make([]byte, len(preimage))
		copy(pi, preimage)
		self.preimages[hash] = pi
	}
}

// Preimages returns a list of SHA3 preimages that have been submitted.
func (self *StateDB) Preimages() map[common.Hash][]byte {
	return self.preimages
}

// AddRefund adds gas to the refund counter
func (self *StateDB) AddRefund(gas uint64) {
	self.journal.append(refundChange{prev: self.refund})
	self.refund += gas
}

// SubRefund removes gas from the refund counter.
// This method will panic if the refund counter goes below zero
func (self *StateDB) SubRefund(gas uint64) {
	self.journal.append(refundChange{prev: self.refund})
	if gas > self.refund {
		panic("Refund counter below zero")
	}
	self.refund -= gas
}

// Exist reports whether the given account address exists in the state.
// Notably this also returns true for suicided accounts.
func (self *StateDB) Exist(addr common.Address) bool {
	return self.getStateObject(addr) != nil
}

// Empty returns whether the state object is either non-existent
// or empty according to the EIP161 specification (balance = nonce = code = 0)
func (self *StateDB) Empty(addr common.Address) bool {
	so := self.getStateObject(addr)
	return so == nil || so.empty()
}

// Retrieve the balance from the given address or 0 if object not found
func (self *StateDB) GetBalance(addr common.Address) *big.Int {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}
	return common.Big0
}

func (self *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}

	return 0
}

func (self *StateDB) GetCode(addr common.Address) []byte {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Code(self.db)
	}
	return nil
}

func (self *StateDB) GetCodeSize(addr common.Address) int {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return 0
	}
	if stateObject.code != nil {
		return len(stateObject.code)
	}
	size, err := self.db.ContractCodeSize(stateObject.addrHash, common.BytesToHash(stateObject.CodeHash()))
	if err != nil {
		self.setError(err)
	}
	return size
}

func (self *StateDB) GetCodeHash(addr common.Address) common.Hash {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.CodeHash())
}

// GetState retrieves a value from the given account's storage trie.
func (self *StateDB) GetState(addr common.Address, key []byte) []byte {
	keyTrie := GetKeyTrie(addr, key)
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.GetState(self.db, keyTrie)
	}
	return []byte{}
}

// GetStateByKeyTrie retrieves a value from the given account's storage trie.
func (self *StateDB) GetStateByKeyTrie(addr common.Address, keyTrie string) []byte {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		stateObject.CreateTrie(self.db)
		return stateObject.GetCommittedStateNoCache(self.db, keyTrie)
	}
	return []byte{}
}

// GetCommittedState retrieves a value from the given account's committed storage trie.
func (self *StateDB) GetCommittedState(addr common.Address, key []byte) []byte {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		var buffer bytes.Buffer
		buffer.WriteString(addr.String())
		buffer.WriteString(string(key))
		key := buffer.String()
		value := stateObject.GetCommittedState(self.db, key)
		return value
	}
	return []byte{}
}

// Database retrieves the low level database supporting the lower level trie ops.
func (self *StateDB) Database() Database {
	return self.db
}

// StorageTrie returns the storage trie of an account.
// The return value is a copy and is nil for non-existent accounts.
func (self *StateDB) StorageTrie(addr common.Address) Trie {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return nil
	}
	cpy := stateObject.deepCopy(self)
	return cpy.updateTrie(self.db)
}

func (self *StateDB) HasSuicided(addr common.Address) bool {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.suicided
	}
	return false
}

/*
 * SETTERS
 */

// AddBalance adds amount to the account associated with addr.
func (self *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	}
}

// SubBalance subtracts amount from the account associated with addr.
func (self *StateDB) SubBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SubBalance(amount)
	}
}

func (self *StateDB) SetBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetBalance(amount)
	}
}

func (self *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	}
}

func (self *StateDB) AddNonce(addr common.Address) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddNonce()
	}
}

func (self *StateDB) SetCode(addr common.Address, code []byte) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(crypto.Keccak256Hash(code), code)
	}
}

func (self *StateDB) SetState(address common.Address, key, value []byte) {
	stateObject := self.GetOrNewStateObject(address)
	keyTrie, valueKey, value := getKeyValue(address, key, value)
	if stateObject != nil {
		stateObject.SetState(self.db, keyTrie, valueKey, value)
	}
}

func getKeyValue(address common.Address, key []byte, value []byte) (string, common.Hash, []byte) {
	var buffer bytes.Buffer
	buffer.WriteString(address.String())
	buffer.WriteString(string(key))
	keyTrie := buffer.String()

	//if value != nil && !bytes.Equal(value,[]byte{}){
	buffer.Reset()
	buffer.WriteString(StoragePrefix)
	buffer.WriteString(string(value))

	valueKey := common.Hash{}
	keccak := sha3.NewKeccak256()
	keccak.Write(buffer.Bytes())
	keccak.Sum(valueKey[:0])

	return keyTrie, valueKey, value
	//}
	//return keyTrie, common.Hash{}, value
}

func GetKeyTrie(address common.Address, key []byte) string {
	var buffer bytes.Buffer
	buffer.WriteString(address.String())
	buffer.WriteString(string(key))
	keyTrie := buffer.String()
	return keyTrie
}

func GetKeyTrieValueKey(address common.Address, key []byte, value []byte) (string, common.Hash) {
	var buffer bytes.Buffer
	buffer.WriteString(address.String())
	buffer.WriteString(string(key))
	keyTrie := buffer.String()
	buffer.Reset()
	buffer.WriteString(StoragePrefix)
	buffer.WriteString(string(value))

	valueKey := common.Hash{}
	keccak := sha3.NewKeccak256()
	keccak.Write(buffer.Bytes())
	keccak.Sum(valueKey[:0])

	return keyTrie, valueKey
}

// Suicide marks the given account as suicided.
// This clears the account balance.
//
// The account's state object is still available until the state is committed,
// getStateObject will return a non-nil account after Suicide.
func (self *StateDB) Suicide(addr common.Address) bool {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return false
	}
	self.journal.append(suicideChange{
		account:     &addr,
		prev:        stateObject.suicided,
		prevbalance: new(big.Int).Set(stateObject.Balance()),
	})
	stateObject.markSuicided()
	stateObject.data.Balance = new(big.Int)

	return true
}

//
// Setting, updating & deleting state object methods.
//

// updateStateObject writes the given object to the trie.
func (self *StateDB) updateStateObject(stateObject *stateObject) {
	addr := stateObject.Address()
	data, err := rlp.EncodeToBytes(stateObject)
	if err != nil {
		panic(fmt.Errorf("can't encode object at %x: %v", addr[:], err))
	}
	self.setError(self.trie.TryUpdate(addr[:], data))
}

// deleteStateObject removes the given object from the state trie.
func (self *StateDB) deleteStateObject(stateObject *stateObject) {
	stateObject.deleted = true
	addr := stateObject.Address()
	self.setError(self.trie.TryDelete(addr[:]))
}

// Retrieve a state object given by the address. Returns nil if not found.
func (self *StateDB) getStateObjectold(addr common.Address) (stateObject *stateObject) {
	// Prefer 'live' objects.
	if obj := self.stateObjects[addr]; obj != nil {
		if obj.deleted {
			return nil
		}
		return obj
	}

	// Load the object from the database.
	enc, err := self.trie.TryGet(addr[:])
	if len(enc) == 0 {
		self.setError(err)
		return nil
	}
	var data Account
	if err := rlp.DecodeBytes(enc, &data); err != nil {
		log.Error("Failed to decode state object", "addr", addr, "err", err)
		return nil
	}
	// Insert into the live set.
	obj := newObject(self, addr, data)
	self.setStateObject(obj)
	return obj
}

func (self *StateDB) getStateObject(addr common.Address) (stateObject *stateObject) {
	// Prefer 'live' objects.
	if obj := self.getCacheObject(addr); obj != nil {
		if obj.deleted {
			return nil
		}
		return obj
	}

	self.objLock.Lock()
	defer self.objLock.Unlock()

	if obj, ok := self.stateObjects[addr]; ok {
		return obj
	}
	// Load the object from the database.
	enc, err := self.trie.TryGet(addr[:])
	if len(enc) == 0 {
		self.setError(err)
		return nil
	}
	var data Account
	if err := rlp.DecodeBytes(enc, &data); err != nil {
		log.Error("Failed to decode state object", "addr", addr, "err", err)
		return nil
	}
	// Insert into the live set.
	obj := newObject(self, addr, data)
	self.setStateObject(obj)
	return obj
}

func (self *StateDB) setStateObject(object *stateObject) {
	self.stateObjects[object.Address()] = object
}

func (self *StateDB) setStateObjectSafe(object *stateObject) {
	self.objLock.Lock()
	defer self.objLock.Unlock()
	self.stateObjects[object.Address()] = object
}

func (self *StateDB) getCacheObject(addr common.Address) *stateObject {
	self.objLock.RLock()
	defer self.objLock.RUnlock()
	return self.stateObjects[addr]
}

// Retrieve a state object or create a new state object if nil.
func (self *StateDB) GetOrNewStateObject(addr common.Address) *stateObject {
	stateObject := self.getStateObject(addr)
	if stateObject == nil || stateObject.deleted {
		stateObject, _ = self.createObject(addr)
	}
	return stateObject
}

func (self *StateDB) GetOrNewStateObjectSafe(addr common.Address) *stateObject {
	stateObject := self.getStateObject(addr)
	if stateObject == nil || stateObject.deleted {
		stateObject, _ = self.createObjectSafe(addr)
	}
	return stateObject
}

// createObject creates a new state object. If there is an existing account with
// the given address, it is overwritten and returned as the second return value.
func (self *StateDB) createObject(addr common.Address) (newobj, prev *stateObject) {
	prev = self.getStateObject(addr)
	newobj = newObject(self, addr, Account{})
	newobj.setNonce(0) // sets the object to dirty
	if prev == nil {
		self.journal.append(createObjectChange{account: &addr})
	} else {
		self.journal.append(resetObjectChange{prev: prev})
	}
	self.setStateObject(newobj)
	return newobj, prev
}

func (self *StateDB) createObjectSafe(addr common.Address) (newobj, prev *stateObject) {
	prev = self.getStateObject(addr)
	self.objLock.Lock()
	defer self.objLock.Unlock()

	if obj, ok := self.stateObjects[addr]; ok {
		return obj, prev
	}
	newobj = newObject(self, addr, Account{})
	newobj.setNonce(0)
	self.setStateObject(newobj)
	return newobj, prev
}

// CreateAccount explicitly creates a state object. If a state object with the address
// already exists the balance is carried over to the new account.
//
// CreateAccount is called during the EVM CREATE operation. The situation might arise that
// a contract does the following:
//
//   1. sends funds to sha(account ++ (nonce + 1))
//   2. tx_create(sha(account ++ nonce)) (note that this gets the address of 1)
//
// Carrying over the balance ensures that Ether doesn't disappear.
func (self *StateDB) CreateAccount(addr common.Address) {
	new, prev := self.createObject(addr)
	if prev != nil {
		new.setBalance(prev.data.Balance)
	}
}

func (self *StateDB) CloneAccount(src common.Address, dest common.Address) error {

	srcObject := self.getStateObject(src)
	if srcObject == nil {
		return cloneErr
	}
	it := trie.NewIterator(srcObject.getTrie(self.db).NodeIterator(nil))
	for it.Next() {

		var value []byte

		keyTrieCode := self.trie.GetKey(it.Key)
		if keyTrieCode == nil {
			log.Warn("CloneAccount Iterator: unable to get keyTrie from hashKey.")
			continue
		}
		keyTrie := string(keyTrieCode)

		if len(keyTrie) <= 42 {
			log.Warn("Invalid keyTrie length.")
			continue
		}
		key := []byte(keyTrie[42:])
		value = self.trie.GetKey(it.Value)
		self.SetState(dest, key, value)
	}
	return nil
}

func (self *StateDB) ForEachStorage(addr common.Address, cb func(key, value common.Hash) bool) {
	so := self.getStateObject(addr)
	if so == nil {
		return
	}
	it := trie.NewIterator(so.getTrie(self.db).NodeIterator(nil))
	for it.Next() {
		key := common.BytesToHash(self.trie.GetKey(it.Key))
		if value, dirty := so.dirtyValueStorage[key]; dirty {
			cb(key, common.BytesToHash(value))
			continue
		}
		cb(key, common.BytesToHash(it.Value))
	}
}

// Copy creates a deep, independent copy of the state.
// Snapshots of the copied state cannot be applied to the copy.
func (self *StateDB) Copy() *StateDB {
	self.lock.Lock()
	defer self.lock.Unlock()

	// Copy all the basic fields, initialize the memory ones
	state := &StateDB{
		db:                self.db,
		trie:              self.db.CopyTrie(self.trie),
		stateObjects:      make(map[common.Address]*stateObject, len(self.journal.dirties)),
		stateObjectsDirty: make(map[common.Address]struct{}, len(self.journal.dirties)),
		refund:            self.refund,
		logs:              make(map[common.Hash][]*types.Log, len(self.logs)),
		logSize:           self.logSize,
		preimages:         make(map[common.Hash][]byte),
		journal:           newJournal(),
	}
	// Copy the dirty states, logs, and preimages
	for addr := range self.journal.dirties {
		// As documented [here](https://github.com/ethereum/go-ethereum/pull/16485#issuecomment-380438527),
		// and in the Finalise-method, there is a case where an object is in the journal but not
		// in the stateObjects: OOG after touch on ripeMD prior to Byzantium. Thus, we need to check for
		// nil
		if object, exist := self.stateObjects[addr]; exist {
			state.stateObjects[addr] = object.deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}
	// Above, we don't copy the actual journal. This means that if the copy is copied, the
	// loop above will be a no-op, since the copy's journal is empty.
	// Thus, here we iterate over stateObjects, to enable copies of copies
	for addr := range self.stateObjectsDirty {
		if _, exist := state.stateObjects[addr]; !exist {
			state.stateObjects[addr] = self.stateObjects[addr].deepCopy(state)
			state.stateObjectsDirty[addr] = struct{}{}
		}
	}
	for hash, logs := range self.logs {
		cpy := make([]*types.Log, len(logs))
		for i, l := range logs {
			cpy[i] = new(types.Log)
			*cpy[i] = *l
		}
		state.logs[hash] = cpy
	}
	for hash, preimage := range self.preimages {
		state.preimages[hash] = preimage
	}
	return state
}

// Snapshot returns an identifier for the current revision of the state.
func (self *StateDB) Snapshot() int {
	id := self.nextRevisionId
	self.nextRevisionId++
	self.validRevisions = append(self.validRevisions, revision{id, self.journal.length()})
	return id
}

// RevertToSnapshot reverts all state changes made since the given revision.
func (self *StateDB) RevertToSnapshot(revid int) {
	// Find the snapshot in the stack of valid snapshots.
	idx := sort.Search(len(self.validRevisions), func(i int) bool {
		return self.validRevisions[i].id >= revid
	})
	if idx == len(self.validRevisions) || self.validRevisions[idx].id != revid {
		panic(fmt.Errorf("revision id %v cannot be reverted", revid))
	}
	snapshot := self.validRevisions[idx].journalIndex

	// Replay the journal to undo changes and remove invalidated snapshots
	self.journal.revert(self, snapshot)
	self.validRevisions = self.validRevisions[:idx]
}

// GetRefund returns the current value of the refund counter.
func (self *StateDB) GetRefund() uint64 {
	return self.refund
}

// Finalise finalises the state by removing the self destructed objects
// and clears the journal as well as the refunds.
func (self *StateDB) Finalise(deleteEmptyObjects bool) {
	for addr := range self.journal.dirties {
		stateObject, exist := self.stateObjects[addr]
		if !exist {
			// ripeMD is 'touched' at block 1714175, in tx 0x1237f737031e40bcde4a8b7e717b2d15e3ecadfe49bb1bbc71ee9deb09c6fcf2
			// That tx goes out of gas, and although the notion of 'touched' does not exist there, the
			// touch-event will still be recorded in the journal. Since ripeMD is a special snowflake,
			// it will persist in the journal even though the journal is reverted. In this special circumstance,
			// it may exist in `s.journal.dirties` but not in `s.stateObjects`.
			// Thus, we can safely ignore it here
			continue
		}

		if stateObject.suicided || (deleteEmptyObjects && stateObject.empty()) {
			self.deleteStateObject(stateObject)
		} else {
			stateObject.updateRoot(self.db)
			self.updateStateObject(stateObject)
		}
		self.stateObjectsDirty[addr] = struct{}{}
	}
	// Invalidate journal because reverting across transactions is not allowed.
	self.clearJournalAndRefund()
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (self *StateDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash {
	self.Finalise(deleteEmptyObjects)
	return self.trie.Hash()
}

// Prepare sets the current transaction hash and index and block hash which is
// used when the EVM emits new state logs.
func (self *StateDB) Prepare(thash, bhash common.Hash, ti int) {
	self.thash = thash
	self.bhash = bhash
	self.txIndex = ti
}

func (self *StateDB) clearJournalAndRefund() {
	self.journal = newJournal()
	self.validRevisions = self.validRevisions[:0]
	self.refund = 0
}

// Commit writes the state to the underlying in-memory trie database.
func (self *StateDB) Commit(deleteEmptyObjects bool) (root common.Hash, err error) {
	self.lock.Lock()
	defer self.lock.Unlock()

	defer self.clearJournalAndRefund()

	for addr := range self.journal.dirties {
		self.stateObjectsDirty[addr] = struct{}{}
	}
	// Commit objects to the trie.
	for addr, stateObject := range self.stateObjects {
		_, isDirty := self.stateObjectsDirty[addr]
		switch {
		case stateObject.suicided || (isDirty && deleteEmptyObjects && stateObject.empty()):
			// If the object has been removed, don't bother syncing it
			// and just mark it for deletion in the trie.
			self.deleteStateObject(stateObject)
		case isDirty:
			// Write any contract code associated with the state object
			if stateObject.code != nil && stateObject.dirtyCode {
				self.db.TrieDB().InsertBlob(common.BytesToHash(stateObject.CodeHash()), stateObject.code)
				stateObject.dirtyCode = false
			}
			// Write any storage changes in the state object to its storage trie.
			if err := stateObject.CommitTrie(self.db); err != nil {
				return common.Hash{}, err
			}
			// Update the object in the main account trie.
			self.updateStateObject(stateObject)
		}
		delete(self.stateObjectsDirty, addr)
	}
	// Write trie changes.
	root, err = self.trie.Commit(func(leaf []byte, parent common.Hash) error {
		var account Account
		if err := rlp.DecodeBytes(leaf, &account); err != nil {
			return nil
		}
		if account.Root != emptyState {
			self.db.TrieDB().Reference(account.Root, parent)
		}
		code := common.BytesToHash(account.CodeHash)
		if code != emptyCode {
			self.db.TrieDB().Reference(code, parent)
		}
		return nil
	})

	log.Debug("Trie cache stats after commit", "misses", trie.CacheMisses(), "unloads", trie.CacheUnloads())
	return root, err
}

func (self *StateDB) SetInt32(addr common.Address, key []byte, value int32) {
	self.SetState(addr, key, common.Int32ToBytes(value))
}
func (self *StateDB) SetInt64(addr common.Address, key []byte, value int64) {
	self.SetState(addr, key, common.Int64ToBytes(value))
}
func (self *StateDB) SetFloat32(addr common.Address, key []byte, value float32) {
	self.SetState(addr, key, common.Float32ToBytes(value))
}
func (self *StateDB) SetFloat64(addr common.Address, key []byte, value float64) {
	self.SetState(addr, key, common.Float64ToBytes(value))
}
func (self *StateDB) SetString(addr common.Address, key []byte, value string) {
	self.SetState(addr, key, []byte(value))
}
func (self *StateDB) SetByte(addr common.Address, key []byte, value byte) {
	self.SetState(addr, key, []byte{value})
}

func (self *StateDB) GetInt32(addr common.Address, key []byte) int32 {
	return common.BytesToInt32(self.GetState(addr, key))
}
func (self *StateDB) GetInt64(addr common.Address, key []byte) int64 {
	return common.BytesToInt64(self.GetState(addr, key))
}
func (self *StateDB) GetFloat32(addr common.Address, key []byte) float32 {
	return common.BytesToFloat32(self.GetState(addr, key))
}
func (self *StateDB) GetFloat64(addr common.Address, key []byte) float64 {
	return common.BytesToFloat64(self.GetState(addr, key))
}
func (self *StateDB) GetString(addr common.Address, key []byte) string {
	return string(self.GetState(addr, key))
}
func (self *StateDB) GetByte(addr common.Address, key []byte) byte {
	ret := self.GetState(addr, key)
	//if len(ret) != 1{
	//	return byte('')
	//}
	return ret[0]
}

// todo: new method -> GetAbiHash
func (self *StateDB) GetAbiHash(addr common.Address) common.Hash {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return common.Hash{}
	}
	return common.BytesToHash(stateObject.AbiHash())
}

// todo: new method -> GetAbi
func (self *StateDB) GetAbi(addr common.Address) []byte {
	stateObject := self.getStateObject(addr)
	if stateObject != nil {
		return stateObject.Abi(self.db)
	}
	return nil
}

// todo: new method -> SetAbi
func (self *StateDB) SetAbi(addr common.Address, abi []byte) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetAbi(crypto.Keccak256Hash(abi), abi)
	}
}

func (self *StateDB) FwAdd(addr common.Address, action Action, list []FwElem) {
	stateObject := self.GetOrNewStateObject(addr)
	fwData := stateObject.FwData()
	switch action {
	case reject:
		for _, addr := range list {
			fwData.DeniedList[addr.FuncName+":"+addr.Addr.String()] = true
		}
	case accept:
		for _, addr := range list {
			fwData.AcceptedList[addr.FuncName+":"+addr.Addr.String()] = true
		}
	}
	stateObject.SetFwData(fwData)
}
func (self *StateDB) FwClear(addr common.Address, action Action) {
	stateObject := self.GetOrNewStateObject(addr)

	fwData := stateObject.FwData()
	switch action {
	case reject:
		fwData.DeniedList = make(map[string]bool)
	case accept:
		fwData.AcceptedList = make(map[string]bool)
	}
	stateObject.SetFwData(fwData)
}
func (self *StateDB) FwDel(addr common.Address, action Action, list []FwElem) {
	stateObject := self.GetOrNewStateObject(addr)

	fwData := stateObject.FwData()
	switch action {
	case reject:
		for _, addr := range list {
			fwData.DeniedList[addr.FuncName+":"+addr.Addr.String()] = false
			delete(fwData.DeniedList, (addr.FuncName + ":" + addr.Addr.String()))
		}
	case accept:
		for _, addr := range list {
			fwData.AcceptedList[addr.FuncName+":"+addr.Addr.String()] = false
			delete(fwData.AcceptedList, (addr.FuncName + ":" + addr.Addr.String()))
		}
	}
	stateObject.SetFwData(fwData)
}
func (self *StateDB) FwSet(addr common.Address, action Action, list []FwElem) {
	stateObject := self.GetOrNewStateObject(addr)

	fwData := NewFwData()
	switch action {
	case reject:
		for _, addr := range list {
			fwData.DeniedList[addr.FuncName+":"+addr.Addr.String()] = true
		}
		fwData.AcceptedList = stateObject.FwData().AcceptedList
	case accept:
		for _, addr := range list {
			fwData.AcceptedList[addr.FuncName+":"+addr.Addr.String()] = true
		}
		fwData.DeniedList = stateObject.FwData().DeniedList
	}
	stateObject.SetFwData(fwData)
}
func (self *StateDB) SetFwStatus(addr common.Address, status FwStatus) {
	stateObject := self.GetOrNewStateObject(addr)
	fwActive := status.Active
	stateObject.SetFwActive(fwActive)

	acc := status.AcceptedList
	self.FwSet(addr, accept, acc)

	denied := status.RejectedList
	self.FwSet(addr, reject, denied)
}

func (self *StateDB) GetFwStatus(addr common.Address) FwStatus {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return FwStatus{
			ContractAddr: addr,
			Active:       false,
			RejectedList: nil,
			AcceptedList: nil,
		}
	}
	fwData := stateObject.FwData()
	fwActive := stateObject.FwActive()

	var deniedList, acceptedList []FwElem

	for elem, b := range fwData.DeniedList {
		tmp := strings.Split(elem, ":")
		if len(tmp) < 2 {
			return FwStatus{
				ContractAddr: addr,
				Active:       false,
				RejectedList: nil,
				AcceptedList: nil,
			}
		}
		api := tmp[0]
		addr := tmp[1]
		if b {
			deniedList = append(deniedList, FwElem{Addr: common.HexToAddress(addr), FuncName: api})
		}
	}
	for elem, b := range fwData.AcceptedList {
		tmp := strings.Split(elem, ":")
		if len(tmp) != 2 {
			return FwStatus{
				ContractAddr: addr,
				Active:       false,
				RejectedList: nil,
				AcceptedList: nil,
			}
		}
		api := tmp[0]
		addr := tmp[1]
		if b {
			acceptedList = append(acceptedList, FwElem{Addr: common.HexToAddress(addr), FuncName: api})
		}
	}

	sort.Sort(FwElems(deniedList))
	sort.Sort(FwElems(acceptedList))

	return FwStatus{
		ContractAddr: addr,
		Active:       fwActive,
		RejectedList: deniedList,
		AcceptedList: acceptedList,
	}
}

func (self *StateDB) FwImport(addr common.Address, data []byte) error {
	status := FwStatus{}
	err := json.Unmarshal(data, &status)
	if err != nil {
		return errors.New("Firewall import failed")
	}
	self.FwAdd(addr, reject, status.RejectedList)
	self.FwAdd(addr, accept, status.AcceptedList)
	//s.SetFwStatus(addr, status)
	return nil
}

func (self *StateDB) SetContractCreator(addr, creator common.Address) {
	stateObject := self.GetOrNewStateObject(addr)
	stateObject.SetContractCreator(creator)
}
func (self *StateDB) GetContractCreator(addr common.Address) common.Address {
	stateObject := self.getStateObject(addr)
	if stateObject == nil {
		return common.Address{}
	}
	creator := stateObject.ContractCreator()
	return creator
}

func (self *StateDB) OpenFirewall(addr common.Address) {
	stateObject := self.GetOrNewStateObject(addr)
	stateObject.SetFwActive(true)
}
func (self *StateDB) CloseFirewall(addr common.Address) {
	stateObject := self.GetOrNewStateObject(addr)
	stateObject.SetFwActive(false)
}
func (self *StateDB) IsFwOpened(addr common.Address) bool {
	stateObject := self.GetOrNewStateObject(addr)
	return stateObject.FwActive()
}
