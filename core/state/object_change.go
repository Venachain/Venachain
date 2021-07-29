package state

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
)

//ObjectChange 用于记录state_object的变更内容
type ObjectChange interface {
	change(*StateDB) //实际的变更内容

	getAddr() common.Address //获取变更的地址

	getObject() *stateObject //获取变更的stateobject

	setVersion(version int)

	getVersion() int
}

//CreateAccount 创建账户的变更操作
type CreateAccount struct {
	account common.Address
	prev    *stateObject
	newObj  *stateObject
	version int
}

func NewCreateAccount(po *stateObject, no *stateObject) *CreateAccount {
	return &CreateAccount{
		account: no.address,
		prev:    po,
		newObj:  no,
	}
}

func (ca CreateAccount) change(s *StateDB) {
	if ca.prev != nil {
		ca.newObj.setBalance(ca.prev.Balance())
	}
	s.setStateObject(ca.newObj)
}

func (ca CreateAccount) getAddr() common.Address {
	return ca.account
}

func (ca CreateAccount) getObject() *stateObject {
	return ca.newObj
}

func (ca CreateAccount) getVersion() int {
	return ca.version
}

func (ca CreateAccount) setVersion(version int) {
	ca.version = version
}

type SetCode struct {
	address common.Address
	obj     *stateObject
	code    []byte
	hash    common.Hash
	version int
}

func NewSetCode(stateObject *stateObject, addr common.Address, code []byte) *SetCode {
	hashCode := crypto.Keccak256Hash(code)

	return &SetCode{
		address: addr,
		obj:     stateObject,
		code:    code,
		hash:    hashCode,
	}
}

func (ca SetCode) change(s *StateDB) {
	ca.obj.setCode(ca.hash, ca.code)
}

func (ca SetCode) getAddr() common.Address {
	return ca.address
}

func (ca SetCode) getObject() *stateObject {
	return ca.obj
}

func (ca SetCode) getVersion() int {
	return ca.version
}

func (ca SetCode) setVersion(version int) {
	ca.version = version
}

type SetAbi struct {
	address common.Address
	obj     *stateObject
	code    []byte
	hash    common.Hash
	version int
}

func NewSetAbi(stateObject *stateObject, addr common.Address, code []byte) *SetAbi {
	hashCode := crypto.Keccak256Hash(code)
	return &SetAbi{
		address: addr,
		obj:     stateObject,
		code:    code,
		hash:    hashCode,
	}
}

func (ca SetAbi) change(s *StateDB) {
	ca.obj.setAbi(ca.hash, ca.code)
}

func (ca SetAbi) getAddr() common.Address {
	return ca.address
}

func (ca SetAbi) getObject() *stateObject {
	return ca.obj
}

func (ca SetAbi) getVersion() int {
	return ca.version
}

func (ca SetAbi) setVersion(version int) {
	ca.version = version
}

type Suicide struct {
	address common.Address
	obj     *stateObject
	version int
}

func NewSuicide(stateObject *stateObject, addr common.Address) *Suicide {
	return &Suicide{
		address: addr,
		obj:     stateObject,
	}
}

func (ca Suicide) change(s *StateDB) {
	ca.obj.markSuicided()
	s.deleteStateObject(ca.obj)
}

func (ca Suicide) getAddr() common.Address {
	return ca.address
}

func (ca Suicide) getObject() *stateObject {
	return ca.obj
}

func (ca Suicide) getVersion() int {
	return ca.version
}

func (ca Suicide) setVersion(version int) {
	ca.version = version
}

type SetFwData struct {
	address common.Address
	obj     *stateObject
	data    FwData
	version int
}

func NewSetFwData(object *stateObject, addr common.Address, data FwData) *SetFwData {
	return &SetFwData{
		address: addr,
		obj:     object,
		data:    data,
	}
}

func (ca SetFwData) change(s *StateDB) {
	ca.obj.setFwData(ca.data)
}

func (ca SetFwData) getAddr() common.Address {
	return ca.address
}

func (ca SetFwData) getObject() *stateObject {
	return ca.obj
}

func (ca SetFwData) getVersion() int {
	return ca.version
}

func (ca SetFwData) setVersion(version int) {
	ca.version = version
}

type SetFwActive struct {
	address common.Address
	obj     *stateObject
	active  uint64
	version int
}

func NewSetFwActive(object *stateObject, addr common.Address, active uint64) *SetFwActive {
	return &SetFwActive{
		address: addr,
		obj:     object,
		active:  active,
	}
}

func (ca SetFwActive) change(s *StateDB) {
	pre := ca.obj.data.FwActive
	result := ca.active ^ pre
	if result == 1 {
		ca.obj.setFwActive(ca.active)
	}
}

func (ca SetFwActive) getAddr() common.Address {
	return ca.address
}

func (ca SetFwActive) getObject() *stateObject {
	return ca.obj
}

func (ca SetFwActive) getVersion() int {
	return ca.version
}

func (ca SetFwActive) setVersion(version int) {
	ca.version = version
}

type SetCreator struct {
	address common.Address
	obj     *stateObject
	creator common.Address
	version int
}

func NewSetCreator(object *stateObject, addr common.Address, creator common.Address) *SetCreator {
	return &SetCreator{
		address: addr,
		obj:     object,
		creator: creator,
	}
}

func (ca SetCreator) change(s *StateDB) {
	ca.obj.setContractCreator(ca.creator)
}

func (ca SetCreator) getAddr() common.Address {
	return ca.address
}

func (ca SetCreator) getObject() *stateObject {
	return ca.obj
}

func (ca SetCreator) getVersion() int {
	return ca.version
}

func (ca SetCreator) setVersion(version int) {
	ca.version = version
}
