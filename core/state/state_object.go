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

package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"
	"sync"

	"github.com/Venachain/Venachain/log"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/crypto"
	"github.com/Venachain/Venachain/rlp"
)

var emptyCodeHash = crypto.Keccak256(nil)
var emptyFwDataHash = crypto.Keccak256(nil)

type Code []byte
type Abi []byte

// datatypes for firewall
type Action uint64

const (
	accept Action = 0
	reject Action = 1
)

var FwWildchardAddr = common.HexToAddress("0xffffffffffffffffffffffffffffffffffffffff")

var ErrInvalidFwAction = errors.New("FW: error, action is invalid")

type FwElem struct {
	Addr     common.Address
	FuncName string
}

type FwElems []FwElem

type FwStatus struct {
	ContractAddr common.Address
	Active       bool
	AcceptedList []FwElem
	RejectedList []FwElem
}

func (fw *FwStatus) canFindInList(funcName string, caller common.Address, act Action) bool {
	list := fw.RejectedList
	if act == accept {
		list = fw.AcceptedList
	}

	for _, fwElem := range list {
		if (fwElem.Addr == FwWildchardAddr || fwElem.Addr == caller) &&
			(fwElem.FuncName == "*" || fwElem.FuncName == funcName) {
			return true
		}
	}
	return false
}

func (fw *FwStatus) IsRejected(funcName string, caller common.Address) bool {
	return fw.canFindInList(funcName, caller, reject)
}

func (fw *FwStatus) IsAccepted(funcName string, caller common.Address) bool {
	return fw.canFindInList(funcName, caller, accept)
}

type FwData struct {
	AcceptedList map[string]bool
	DeniedList   map[string]bool
}

func NewAction(action string) (Action, error) {
	if strings.EqualFold(action, "ACCEPT") {
		return accept, nil
	} else if strings.EqualFold(action, "REJECT") {
		return reject, nil
	} else {
		return 0, ErrInvalidFwAction
	}
}

func (l FwElems) Len() int {
	return len(l)
}

func (l FwElems) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l FwElems) Less(i, j int) bool {
	return l[i].FuncName < l[j].FuncName
}

func NewFwData() FwData {
	return FwData{
		AcceptedList: make(map[string]bool),
		DeniedList:   make(map[string]bool),
	}
}

func FwMarshal(fw FwData) []byte {

	rawData, err := json.Marshal(fw)
	if err != nil {
		log.Warn("json Marshal failed", "err", err)
		return nil
	}
	return rawData
}

func FwUnMarshal(raw []byte, fw *FwData) *FwData {
	err := json.Unmarshal(raw, fw)
	if err != nil {
		log.Warn(err.Error())
	}
	return fw
}

func (self Code) String() string {
	return string(self) //strings.Join(Disassemble(self), " ")
}

type Storage map[string]common.Hash
type ValueStorage map[common.Hash][]byte

// Storage -> hash : hash , common.Hash ([32]byte)
//type Storage map[common.Hash]common.Hash

func (self Storage) String() (str string) {
	for key, value := range self {
		// %X -> Provide hexadecimal
		str += fmt.Sprintf("%X : %X\n", key, value)
	}

	return
}

// Copy a copy of Storage
func (self Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range self {
		cpy[key] = value
	}

	return cpy
}

func (self ValueStorage) Copy() ValueStorage {
	cpy := make(ValueStorage)
	for key, value := range self {
		cpy[key] = value
	}

	return cpy
}

// stateObject represents an Ethereum account which is being modified.
//
// The usage pattern is as follows:
// First you need to obtain a state object.
// Account values can be accessed and modified through the object.
// Finally, call CommitTrie to write the modified storage trie into a database.
type stateObject struct {
	address  common.Address
	addrHash common.Hash // hash of ethereum address of the account
	data     Account
	db       *StateDB

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr error

	// Write caches.
	trie Trie
	// storage trie, which becomes non-nil on first access
	code Code // contract bytecode, which gets set when code is loaded

	abi       Abi
	fwData    FwData // firewall data
	rawFwData []byte

	originStorage      Storage      // Storage cache of original entries to dedup rewrites
	originValueStorage ValueStorage // Storage cache of original entries to dedup rewrites

	dirtyStorage      Storage      // Storage entries that need to be flushed to disk
	dirtyValueStorage ValueStorage // Storage entries that need to be flushed to disk

	// Cache flags.
	// When an object is marked suicided it will be delete from the trie
	// during the "update" phase of the state transition.
	dirtyCode bool // true if the code was updated
	suicided  bool
	deleted   bool
	lock      sync.Mutex
}

// empty returns whether the account is considered empty.
func (s *stateObject) empty() bool {
	return s.data.Nonce == 0 && s.data.Balance.Sign() == 0 && bytes.Equal(s.data.CodeHash, emptyCodeHash)
}

// Account is the Ethereum consensus representation of accounts.
// These objects are stored in the main account trie.
type Account struct {
	Nonce      uint64
	FwActive   uint64
	Balance    *big.Int
	Root       common.Hash // merkle root of the storage trie
	CodeHash   []byte
	AbiHash    []byte
	Creator    common.Address
	FwDataHash []byte
}

// newObject creates a state object.
func newObject(db *StateDB, address common.Address, data Account) *stateObject {
	if data.Balance == nil {
		data.Balance = new(big.Int)
	}
	if data.CodeHash == nil {
		data.CodeHash = emptyCodeHash
	}
	if data.FwDataHash == nil {
		data.FwDataHash = emptyFwDataHash
	}
	return &stateObject{
		db:       db,
		address:  address,
		addrHash: crypto.Keccak256Hash(address[:]),
		data:     data,
		fwData:   NewFwData(),

		originStorage:      make(Storage),
		originValueStorage: make(map[common.Hash][]byte),

		dirtyStorage:      make(Storage),
		dirtyValueStorage: make(map[common.Hash][]byte),
	}
}

// EncodeRLP implements rlp.Encoder.
func (c *stateObject) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, c.data)
}

// setError remembers the first non-nil error it is called with.
func (self *stateObject) setError(err error) {
	if self.dbErr == nil {
		self.dbErr = err
	}
}

func (self *stateObject) markSuicided() {
	self.suicided = true
}

func (c *stateObject) touch() {
	c.db.journal.append(touchChange{
		account: &c.address,
	})
	if c.address == ripemd {
		// Explicitly put it in the dirty-cache, which is otherwise generated from
		// flattened journals.
		c.db.journal.dirty(c.address)
	}
}

func (c *stateObject) getTrie(db Database) Trie {
	if c.trie == nil {
		var err error
		c.trie, err = db.OpenStorageTrie(c.addrHash, c.data.Root)
		if err != nil {
			c.trie, _ = db.OpenStorageTrie(c.addrHash, common.Hash{})
			c.setError(fmt.Errorf("can't create storage trie: %v", err))
		}
	}
	return c.trie
}

func (c *stateObject) CreateTrie(db Database) {
	if c.trie == nil {
		c.lock.Lock()
		defer c.lock.Unlock()
		if c.trie != nil {
			return
		}
		var err error
		c.trie, err = db.OpenStorageTrie(c.addrHash, c.data.Root)
		if err != nil {
			c.trie, _ = db.OpenStorageTrie(c.addrHash, common.Hash{})
			c.setError(fmt.Errorf("can't create storage trie: %v", err))
		}
	}
}

// GetState retrieves a value from the account storage trie.
//func (self *stateObject) GetState(db Database, key common.Hash) common.Hash {
//	// If we have a dirty value for this state entry, return it
//	value, dirty := self.dirtyStorage[key]
//	if dirty {
//		return value
//	}
//	// Otherwise return the entry's original value
//	return self.GetCommittedState(db, key)
//}

// GetState retrieves a value from the account storage trie.
func (self *stateObject) GetState(db Database, keyTree string) []byte {
	// If we have a dirty value for this state entry, return it
	valueKey, dirty := self.dirtyStorage[keyTree]
	if dirty {
		value, ok := self.dirtyValueStorage[valueKey]
		if ok {
			return value
		}
	}
	// Otherwise return the entry's original value
	return self.GetCommittedState(db, keyTree)
}

// GetCommittedState retrieves a value from the committed account storage trie.
//func (self *stateObject) GetCommittedState(db Database, key common.Hash) common.Hash {
//	// If we have the original value cached, return that
//	value, cached := self.originStorage[key]
//	if cached {
//		return value
//	}
//	// Otherwise load the value from the database
//	enc, err := self.getTrie(db).TryGet(key[:])
//	if err != nil {
//		self.setError(err)
//		return common.Hash{}
//	}
//	if len(enc) > 0 {
//		_, content, _, err := rlp.Split(enc)
//		if err != nil {
//			self.setError(err)
//		}
//		value.SetBytes(content)
//	}
//	self.originStorage[key] = value
//	return value
//}

// GetCommittedState retrieves a value from the committed account storage trie.
func (self *stateObject) GetCommittedState(db Database, key string) []byte {
	var value []byte
	// If we have the original value cached, return that
	valueKey, cached := self.originStorage[key]
	if cached {
		value, cached2 := self.originValueStorage[valueKey]
		if cached2 {
			return value
		}
	}

	// Otherwise load the valueKey from trie
	enc, err := self.getTrie(db).TryGet([]byte(key))
	if err != nil {
		self.setError(err)
		return []byte{}
	}
	if len(enc) > 0 {
		_, content, _, err := rlp.Split(enc)
		if err != nil {
			self.setError(err)
		}
		valueKey.SetBytes(content)

		//load value from db
		value = self.db.trie.GetKey(valueKey.Bytes())
		if err != nil {
			self.setError(err)
		}
	}

	self.originStorage[key] = valueKey
	self.originValueStorage[valueKey] = value
	return value
}

func (self *stateObject) GetCommittedStateNoCache(db Database, key string) []byte {
	var value []byte
	var valueKey common.Hash

	// Otherwise load the valueKey from trie
	self.lock.Lock()
	enc, err := self.getTrie(db).TryGet([]byte(key))
	self.lock.Unlock()
	if err != nil {
		self.setError(err)
		return []byte{}
	}
	if len(enc) > 0 {
		_, content, _, err := rlp.Split(enc)
		if err != nil {
			self.setError(err)
		}
		valueKey.SetBytes(content)

		//load value from db
		value = self.db.trie.GetKey(valueKey.Bytes())
		if err != nil {
			self.setError(err)
		}
	}

	return value
}

// SetState updates a value in account storage.
// set [keyTrie,valueKey] to storage
// set [valueKey,value] to db
func (self *stateObject) SetState(db Database, keyTrie string, valueKey common.Hash, value []byte) {

	//if the new value is the same as old,don't set
	preValue := self.GetState(db, keyTrie) // get value key
	if bytes.Equal(preValue, value) {
		return
	}

	//New value is different, update and journal the change
	self.db.journal.append(storageChange{
		account:  &self.address,
		key:      keyTrie,
		valueKey: self.originStorage[keyTrie],
		preValue: preValue,
	})

	self.setState(keyTrie, valueKey, value)
}

func (self *stateObject) setState(key string, valueKey common.Hash, value []byte) {
	self.dirtyStorage[key] = valueKey
	self.dirtyValueStorage[valueKey] = value
}

// updateTrie writes cached storage modifications into the object's storage trie.
func (self *stateObject) updateTrie(db Database) Trie {
	tr := self.getTrie(db)
	for key, valueKey := range self.dirtyStorage {
		delete(self.dirtyStorage, key)

		if valueKey == self.originStorage[key] {
			continue
		}

		self.originStorage[key] = valueKey

		if valueKey == emptyStorage {
			self.setError(self.trie.TryDelete([]byte(key)))
			continue
		}

		v, _ := rlp.EncodeToBytes(bytes.TrimLeft(valueKey[:], "\x00"))
		self.setError(self.trie.TryUpdate([]byte(key), v))

		//flush dirty value
		if value, ok := self.dirtyValueStorage[valueKey]; ok {
			delete(self.originValueStorage, valueKey)
			self.originValueStorage[valueKey] = value
			self.setError(self.trie.TryUpdateValue(valueKey.Bytes(), value))
		}
	}

	return tr
}

func (self *stateObject) UpdateKey(key, value []byte) {
	self.lock.Lock()
	defer self.lock.Unlock()
	err := self.trie.TryUpdate(key, value)
	if err != nil {
		self.setError(err)
	}
}

func (self *stateObject) DeleteKey(key []byte) {
	self.lock.Lock()
	defer self.lock.Unlock()
	err := self.trie.TryDelete(key)
	if err != nil {
		self.setError(err)
	}
}

func (self *stateObject) UpdateValue(key, value []byte) {
	err := self.trie.TryUpdateValue(key, value)
	if err != nil {
		self.setError(err)
	}
}

// UpdateRoot sets the trie root to the current root hash of
func (self *stateObject) updateRoot(db Database) {
	self.updateTrie(db)
	self.data.Root = self.trie.Hash()
}

// CommitTrie the storage trie of the object to db.
// This updates the trie root.
func (self *stateObject) CommitTrie(db Database) error {
	self.updateTrie(db)
	if self.dbErr != nil {
		return self.dbErr
	}

	root, err := self.trie.Commit(nil)
	if err == nil {
		self.data.Root = root
	}
	return err
}

// AddBalance removes amount from c's balance.
// It is used to add funds to the destination account of a transfer.
func (c *stateObject) AddBalance(amount *big.Int) {
	// EIP158: We must check emptiness for the objects such that the account
	// clearing (0,0,0 objects) can take effect.
	if amount.Sign() == 0 {
		if c.empty() {
			c.touch()
		}

		return
	}
	c.SetBalance(new(big.Int).Add(c.Balance(), amount))
}

// SubBalance removes amount from c's balance.
// It is used to remove funds from the origin account of a transfer.
func (c *stateObject) SubBalance(amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	c.SetBalance(new(big.Int).Sub(c.Balance(), amount))
}

func (self *stateObject) SetBalance(amount *big.Int) {
	self.db.journal.append(balanceChange{
		account: &self.address,
		prev:    new(big.Int).Set(self.data.Balance),
	})
	self.setBalance(amount)
}

func (self *stateObject) setBalance(amount *big.Int) {
	self.data.Balance = amount
}

// Return the gas back to the origin. Used by the Virtual machine or Closures
func (c *stateObject) ReturnGas(gas *big.Int) {}

func (self *stateObject) deepCopy(db *StateDB) *stateObject {
	stateObject := newObject(db, self.address, self.data)
	if self.trie != nil {
		stateObject.trie = db.db.CopyTrie(self.trie)
	}
	stateObject.rawFwData = self.rawFwData
	stateObject.fwData = self.fwData
	stateObject.code = self.code
	stateObject.dirtyStorage = self.dirtyStorage.Copy()
	stateObject.dirtyValueStorage = self.dirtyValueStorage.Copy()
	stateObject.originStorage = self.originStorage.Copy()
	stateObject.originValueStorage = self.originValueStorage.Copy()
	stateObject.suicided = self.suicided
	stateObject.dirtyCode = self.dirtyCode
	stateObject.deleted = self.deleted
	return stateObject
}

//
// Attribute accessors
//

// Returns the address of the contract/account
func (c *stateObject) Address() common.Address {
	return c.address
}

// Code returns the contract code associated with this object, if any.
func (self *stateObject) Code(db Database) []byte {
	if self.code != nil {
		return self.code
	}
	if bytes.Equal(self.CodeHash(), emptyCodeHash) {
		return nil
	}
	code, err := db.ContractCode(self.addrHash, common.BytesToHash(self.CodeHash()))
	if err != nil {
		self.setError(fmt.Errorf("can't load code hash %x: %v", self.CodeHash(), err))
	}
	self.code = code
	return code
}

func (self *stateObject) SetCode(codeHash common.Hash, code []byte) {
	prevcode := self.Code(self.db.db)
	self.db.journal.append(codeChange{
		account:  &self.address,
		prevhash: self.CodeHash(),
		prevcode: prevcode,
	})
	self.setCode(codeHash, code)
}

func (self *stateObject) setCode(codeHash common.Hash, code []byte) {
	self.code = code
	self.data.CodeHash = codeHash[:]
	self.dirtyCode = true
}

func (self *stateObject) SetNonce(nonce uint64) {
	self.db.journal.append(nonceChange{
		account: &self.address,
		prev:    self.data.Nonce,
	})
	self.setNonce(nonce)
}

func (self *stateObject) AddNonce() {
	self.db.journal.append(nonceChange{
		account: &self.address,
		prev:    self.data.Nonce,
	})
	self.setNonce(self.data.Nonce + 1)
}

func (self *stateObject) setNonce(nonce uint64) {
	self.data.Nonce = nonce
}

func (self *stateObject) CodeHash() []byte {
	return self.data.CodeHash
}

func (self *stateObject) Balance() *big.Int {
	return self.data.Balance
}

func (self *stateObject) Nonce() uint64 {
	return self.data.Nonce
}

// Never called, but must be present to allow stateObject to be used
// as a vm.Account interface that also satisfies the vm.ContractRef
// interface. Interfaces are awesome.
func (self *stateObject) Value() *big.Int {
	panic("Value on stateObject should never be called")
}

// todo: New method
// ======================================= New method ===============================

// todo: new method -> AbiHash
func (self *stateObject) AbiHash() []byte {
	return self.data.AbiHash
}

// ABI returns the contract abi associated with this object, if any.
func (self *stateObject) Abi(db Database) []byte {
	//if self.Abi != nil {
	//	return self.abi
	//}
	if bytes.Equal(self.AbiHash(), emptyCodeHash) {
		return nil
	}
	// Extract the code from the tree, enter the parameters: address and hash, here you need to find the acquisition rules in depth
	abi, err := db.ContractAbi(self.addrHash, common.BytesToHash(self.AbiHash()))
	if err != nil {
		self.setError(fmt.Errorf("can't load abi hash %x: %v", self.AbiHash(), err))
	}
	self.abi = abi
	return abi
}

// todo: new method -> SetAbi.
func (self *stateObject) SetAbi(abiHash common.Hash, abi []byte) {
	prevabi := self.Abi(self.db.db)
	self.db.journal.append(abiChange{
		account:  &self.address,
		prevhash: self.AbiHash(),
		prevabi:  prevabi,
	})
	self.setAbi(abiHash, abi)
}

// todo: new method -> setAbi
func (self *stateObject) setAbi(abiHash common.Hash, abi []byte) {
	self.abi = abi
	self.data.AbiHash = abiHash[:]
}

// todo: setter and getter for contractCreator
func (self *stateObject) ContractCreator() common.Address {
	return self.data.Creator
}

func (self *stateObject) SetContractCreator(addr common.Address) error {
	if bytes.Equal(self.CodeHash(), emptyCodeHash) {
		return errors.New("not a contract account")
	}
	prevCreator := self.ContractCreator()

	self.db.journal.append(creatorChange{
		account:     &self.address,
		prevCreator: prevCreator,
	})

	self.setContractCreator(addr)

	return nil
}
func (self *stateObject) setContractCreator(addr common.Address) {
	self.data.Creator = addr
}

// todo: setters and getters for firewall data

func (self *stateObject) FwData() FwData {
	if self.rawFwData != nil {
		return self.fwData
	}
	fwData := NewFwData()
	if bytes.Equal(self.FwDataHash(), emptyFwDataHash) {
		return fwData
	}
	rawData := self.db.GetState(self.Address(), self.FwDataHash())
	FwUnMarshal(rawData, &fwData)
	self.fwData = fwData
	self.rawFwData = rawData

	return self.fwData
}
func (self *stateObject) FwDataHash() []byte {
	return self.data.FwDataHash
}

func (self *stateObject) FwActive() bool {
	if self.data.FwActive != uint64(0) {
		return true
	} else {
		return false
	}
}

func (self *stateObject) SetFwData(data FwData) {
	prevFwData := self.FwData()
	self.db.journal.append(fwDataChange{
		account:    &self.address,
		prevFwData: prevFwData,
	})

	self.setFwData(data)
}

func (self *stateObject) setFwData(data FwData) {
	rawData := FwMarshal(data)

	var fwHash []byte
	if bytes.Equal(self.FwDataHash(), emptyFwDataHash) {
		fwHash = crypto.Keccak256(rawData)
		self.data.FwDataHash = fwHash
	} else {
		fwHash = self.FwDataHash()
	}

	self.fwData = data
	self.rawFwData = rawData

	self.db.SetState(self.Address(), fwHash, rawData)
}

func (self *stateObject) SetFwActive(active bool) {

	self.db.journal.append(fwActiveChange{
		account:    &self.address,
		prevActive: self.data.FwActive,
	})
	if active {
		self.setFwActive(uint64(1))
	} else {
		self.setFwActive(uint64(0))
	}
}

func (self *stateObject) setFwActive(active uint64) {
	self.data.FwActive = active
}
