package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/PlatONEnetwork/PlatONE-Go/trie"

	"github.com/PlatONEnetwork/PlatONE-Go/log"

	"github.com/PlatONEnetwork/PlatONE-Go/rlp"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
)

type TxSimulator struct {
	stateDb    *StateDB
	tx         *types.Transaction
	hash       common.Hash
	readMap    map[string]*ReadOp
	writeMap   map[string]*WriteOp
	balanceMap map[common.Address]*BalanceOp
	logs       []*types.Log
	nonce      []common.Address
	oc         []ObjectChange //用于记录除balance变化之外的object变更
	version    int            //stateDB中已经已处理的交易数
	receipt    *types.Receipt
	err        error
	dirty      map[common.Address]*stateObject
	reTry      bool
	index      int
	txRlp      []byte
}

func NewTxSimulator(sdb *StateDB, transaction *types.Transaction) *TxSimulator {
	enc, _ := rlp.EncodeToBytes(transaction)
	return &TxSimulator{
		stateDb:    sdb,
		tx:         transaction,
		hash:       transaction.Hash(),
		version:    sdb.GetTxsLen(),
		readMap:    make(map[string]*ReadOp),
		writeMap:   make(map[string]*WriteOp),
		balanceMap: make(map[common.Address]*BalanceOp),
		dirty:      make(map[common.Address]*stateObject),
		txRlp:      enc,
	}
}

func (txSim *TxSimulator) GetHash() common.Hash {
	return txSim.hash
}

func (txSim *TxSimulator) SetReceipt(rec *types.Receipt) {
	txSim.receipt = rec
}

func (txSim *TxSimulator) GetReceipt() *types.Receipt {
	return txSim.receipt
}

func (txSim *TxSimulator) GetReadMap() map[string]*ReadOp {
	return txSim.readMap
}

func (txSim *TxSimulator) GetWriteMap() map[string]*WriteOp {
	return txSim.writeMap
}

func (txSim *TxSimulator) GetBalanceMap() map[common.Address]*BalanceOp {
	return txSim.balanceMap
}

//ReTry 用于判断是否需要重新模拟
func (txSim *TxSimulator) ReTry() bool {
	return txSim.reTry
}

//SetIndex 用于标识交易在区块内的位置
func (txSim *TxSimulator) SetIndex(index int) {
	txSim.index = index
}

//GetIndex 获取交易在区块内的位置
func (txSim *TxSimulator) GetIndex() int {
	return txSim.index
}

func (txSim *TxSimulator) getStateObject(addr common.Address) *stateObject {
	obj := txSim.stateDb.GetOrNewStateObject(addr)
	obj.CreateTrie(txSim.stateDb.db)
	txSim.dirty[addr] = obj
	return obj
}

//SetState 将设置的操作存于writeSet内
func (txSim *TxSimulator) SetState(addr common.Address, key, value []byte) {
	keyTrie, vk, _ := getKeyValue(addr, key, value)
	//log.Info("set state", "key", keyTrie, "value", string(value))
	obj, ok := txSim.dirty[addr]
	if !ok {
		obj = txSim.getStateObject(addr)
	}
	op := &WriteOp{
		ContractAddress: addr,
		object:          obj,
		Key:             key,
		KeyTrie:         keyTrie,
		ValueKey:        vk,
		Value:           value,
		Version:         txSim.version,
	}

	if len(value) != 0 {
		kv, _ := rlp.EncodeToBytes(bytes.TrimLeft(vk[:], "\x00"))
		op.KeyValue = kv
	}
	txSim.writeMap[keyTrie] = op
}

//GetState 获取数据操作
func (txSim *TxSimulator) GetState(addr common.Address, key []byte) []byte {
	keyTrie := GetKeyTrie(addr, key)
	op := &ReadOp{
		ContractAddress: addr,
		Key:             key,
		KeyTrie:         keyTrie,
		Version:         txSim.version,
	}
	//var type = 0
	if writeop, ok := txSim.writeMap[keyTrie]; ok {
		//type = 1
		op.Value = writeop.Value
	} else if readop, ok := txSim.readMap[keyTrie]; ok {
		//type = 2
		op.Value = readop.Value
	} else if state, ok := txSim.stateDb.GetStateByCache(keyTrie); ok {
		//type = 3
		op.Value = state
	} else {
		//type = 4
		op.Value = txSim.stateDb.GetStateByKeyTrie(addr, keyTrie)
	}
	//log.Info("get state", "key", keyTrie, "value", string(op.Value), "from", type)
	txSim.readMap[keyTrie] = op
	return op.Value
}

func (txSim *TxSimulator) AddLog(log *types.Log) {
	log.TxHash = txSim.hash
	txSim.logs = append(txSim.logs, log)
}

func (txSim *TxSimulator) GetLogs(hash common.Hash) []*types.Log {
	return txSim.logs
}

func (txSim *TxSimulator) Logs() []*types.Log {
	return txSim.logs
}

//AddBalance 模拟交易的balance增加
func (txSim *TxSimulator) AddBalance(addr common.Address, amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	obj, ok := txSim.dirty[addr]
	if !ok {
		obj = txSim.stateDb.GetOrNewStateObjectSafe(addr)
		txSim.dirty[addr] = obj
	}
	op := &BalanceOp{
		ContractAddress: addr,
		object:          obj,
		Version:         txSim.version,
	}
	var value *big.Int

	if balanceOp, ok := txSim.balanceMap[addr]; ok {
		//优先从txSim的balanceMap中获取之前已经做过得更变
		value = balanceOp.Amount
	} else if balance, ok := txSim.stateDb.GetBalanceByCache(addr); ok {
		//从stateDB的balanceMap中获取之前已经做过得更变
		value = balance
	} else {
		//从stateDB的DB中过去balance的值
		if obj.Balance() != nil {
			value = obj.Balance()
		} else {
			value = common.Big0
		}
	}

	//优先从stateDB的balanceMap中获取之前已经做过得更变
	balance, ok := txSim.stateDb.GetBalanceByCache(addr)
	if ok {
		value = balance
	} else {
		//从stateDB的DB中过去balance的值
		if obj.Balance() != nil {
			value = obj.Balance()
		} else {
			value = common.Big0
		}
	}
	op.Amount = new(big.Int).Add(value, amount)
	txSim.balanceMap[addr] = op
}

//SubBalance 模拟交易的balance减少
func (txSim *TxSimulator) SubBalance(addr common.Address, amount *big.Int) {
	if amount.Sign() == 0 {
		return
	}
	obj, ok := txSim.dirty[addr]
	if !ok {
		obj = txSim.stateDb.GetOrNewStateObjectSafe(addr)
		txSim.dirty[addr] = obj
	}
	op := &BalanceOp{
		ContractAddress: addr,
		object:          obj,
		Version:         txSim.version,
	}
	var value *big.Int

	if balanceOp, ok := txSim.balanceMap[addr]; ok {
		//优先从txSim的balanceMap中获取之前已经做过得更变
		value = balanceOp.Amount
	} else if balance, ok := txSim.stateDb.GetBalanceByCache(addr); ok {
		//从stateDB的balanceMap中获取之前已经做过得更变
		value = balance
	} else {
		//从stateDB的DB中过去balance的值
		if obj.Balance() != nil {
			value = obj.Balance()
		} else {
			value = common.Big0
		}
	}
	op.Amount = new(big.Int).Sub(value, amount)
	txSim.balanceMap[addr] = op
}

//SetBalance 模拟交易的balance设置
func (txSim *TxSimulator) SetBalance(addr common.Address, amount *big.Int) {
	obj, ok := txSim.dirty[addr]
	if !ok {
		obj = txSim.stateDb.GetOrNewStateObjectSafe(addr)
		if obj.Balance().Cmp(amount) == 0 {
			return
		}
		txSim.dirty[addr] = obj
	} else {
		if obj.Balance().Cmp(amount) == 0 {
			return
		}
	}

	op := &BalanceOp{
		ContractAddress: addr,
		object:          obj,
		Version:         txSim.version,
		Amount:          amount,
	}
	txSim.balanceMap[addr] = op
}

//SetBalance 模拟交易的balance设置
func (txSim *TxSimulator) GetBalance(addr common.Address) *big.Int {
	//优先从stateDB的balanceMap中获取之前已经做过得更变
	balance, ok := txSim.stateDb.GetBalanceByCache(addr)
	if ok {
		return balance
	} else {
		//从stateDB的DB中过去balance的值
		return txSim.stateDb.GetBalance(addr)
	}
}

func (txSim *TxSimulator) GetCode(addr common.Address) []byte {
	return txSim.stateDb.GetCode(addr)
}

//CreateAccount 记录创建账户的操作
func (txSim *TxSimulator) CreateAccount(address common.Address) {
	po := txSim.stateDb.getStateObject(address)
	no := newObject(txSim.stateDb, address, Account{})
	no.setNonce(0)
	txSim.stateDb.setStateObjectSafe(no)
	txSim.dirty[address] = no
	log.Debug("TxSimulator CreateAccount add oc ")
	txSim.oc = append(txSim.oc, NewCreateAccount(po, no))
}

func (txSim *TxSimulator) GetNonce(address common.Address) uint64 {
	return txSim.stateDb.GetNonce(address)
}

//SetNonce 记录nonce有变更的地址,鉴于nonce的变更都是对原有的nonce+1,这里只记录需要变更的地址
func (txSim *TxSimulator) SetNonce(address common.Address, u uint64) {
	txSim.getStateObject(address)
	txSim.nonce = append(txSim.nonce, address)
}

//AddNonce 记录nonce有变更的地址,鉴于nonce的变更都是对原有的nonce+1,这里只记录需要变更的地址
func (txSim *TxSimulator) AddNonce(address common.Address) {
	txSim.getStateObject(address)
	txSim.nonce = append(txSim.nonce, address)
}

func (txSim *TxSimulator) GetCodeHash(address common.Address) common.Hash {
	return txSim.stateDb.GetCodeHash(address)
}

//SetCode 记录设置代码的操作
func (txSim *TxSimulator) SetCode(address common.Address, bytes []byte) {
	log.Debug("TxSimulator set code add oc ")
	stateObject := txSim.getStateObject(address)
	txSim.oc = append(txSim.oc, NewSetCode(stateObject, address, bytes))
}

func (txSim *TxSimulator) GetCodeSize(address common.Address) int {
	return txSim.stateDb.GetCodeSize(address)
}

func (txSim *TxSimulator) GetAbiHash(address common.Address) common.Hash {
	return txSim.stateDb.GetAbiHash(address)
}

func (txSim *TxSimulator) GetAbi(address common.Address) []byte {
	return txSim.stateDb.GetAbi(address)
}

//SetAbi 记录设置abi的操作
func (txSim *TxSimulator) SetAbi(address common.Address, bytes []byte) {
	stateObject := txSim.getStateObject(address)
	txSim.oc = append(txSim.oc, NewSetAbi(stateObject, address, bytes))
}

func (txSim *TxSimulator) AddRefund(u uint64) {
}

func (txSim *TxSimulator) SubRefund(u uint64) {
}

func (txSim *TxSimulator) GetRefund() uint64 {
	return txSim.stateDb.GetRefund()
}

func (txSim *TxSimulator) GetCommittedState(address common.Address, key []byte) []byte {
	stateObject := txSim.stateDb.getStateObject(address)
	if stateObject != nil {
		var buffer bytes.Buffer
		buffer.WriteString(address.String())
		buffer.WriteString(string(key))
		key := buffer.String()
		value := stateObject.GetCommittedStateNoCache(txSim.stateDb.db, key)
		return value
	}
	return []byte{}
}

//Suicide 查看是否需要进行自杀操作，如果需要则记录自杀操作，并记录balance变更的操作
func (txSim *TxSimulator) Suicide(address common.Address) bool {
	stateObject := txSim.stateDb.getStateObject(address)
	if stateObject == nil {
		return false
	}
	txSim.dirty[address] = stateObject
	opSuicide := NewSuicide(stateObject, address)
	txSim.oc = append(txSim.oc, opSuicide)
	op := &BalanceOp{
		ContractAddress: address,
		Version:         txSim.version,
		Amount:          new(big.Int),
	}
	txSim.balanceMap[address] = op
	return true
}

func (txSim *TxSimulator) HasSuicided(address common.Address) bool {
	return txSim.stateDb.HasSuicided(address)
}

func (txSim *TxSimulator) Exist(address common.Address) bool {
	return txSim.stateDb.Exist(address)
}

func (txSim *TxSimulator) Empty(address common.Address) bool {
	return txSim.stateDb.Empty(address)
}

//RevertToSnapshot 用于回退模拟交易
func (txSim *TxSimulator) RevertToSnapshot(i int) {
	if txSim.tx.Try() {
		log.Debug("tx call vm err ,retry", "hash", txSim.tx.Hash())
		txSim.reTry = true
	} else {
		log.Debug("tx call vm err ,but still add ", "hash", txSim.tx.Hash())
		txSim.balanceMap = make(map[common.Address]*BalanceOp)
		txSim.writeMap = make(map[string]*WriteOp)
		txSim.readMap = make(map[string]*ReadOp)
		txSim.oc = make([]ObjectChange, 0)
	}

}

func (txSim *TxSimulator) Snapshot() int {
	return txSim.stateDb.Snapshot()
}

//AddPreimage debug情况下用来记录sha3操作的中间值，不做处理
func (txSim *TxSimulator) AddPreimage(hash common.Hash, bytes []byte) {
}

//ForEachStorage 暂无调用 不做处理
func (txSim *TxSimulator) ForEachStorage(address common.Address, f func(common.Hash, common.Hash) bool) {
}

//FwAdd 将防火墙的添加最终演变成防火墙的设置操作，并记录该设置操作
func (txSim *TxSimulator) FwAdd(contractAddr common.Address, action Action, list []FwElem) {
	stateObject := txSim.getStateObject(contractAddr)
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
	txSim.oc = append(txSim.oc, NewSetFwData(stateObject, contractAddr, fwData))
}

//FwClear 将防火墙的清理最终演变成防火墙的设置操作，并记录该设置操作
func (txSim *TxSimulator) FwClear(contractAddr common.Address, action Action) {
	stateObject := txSim.getStateObject(contractAddr)
	fwData := stateObject.FwData()
	switch action {
	case reject:
		fwData.DeniedList = make(map[string]bool)
	case accept:
		fwData.AcceptedList = make(map[string]bool)
	}
	txSim.oc = append(txSim.oc, NewSetFwData(stateObject, contractAddr, fwData))
}

//FwDel 将防火墙的删除最终演变成防火墙的设置操作，并记录该设置操作
func (txSim *TxSimulator) FwDel(contractAddr common.Address, action Action, list []FwElem) {
	stateObject := txSim.getStateObject(contractAddr)
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
	txSim.oc = append(txSim.oc, NewSetFwData(stateObject, contractAddr, fwData))
}

//FwSet 处理数据并记录防火墙的设置操作
func (txSim *TxSimulator) FwSet(contractAddr common.Address, action Action, list []FwElem) {
	stateObject := txSim.getStateObject(contractAddr)
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
	txSim.oc = append(txSim.oc, NewSetFwData(stateObject, contractAddr, fwData))
}

//SetFwStatus 拆分成防火墙的数据设置操作和防火墙的活跃操作，并记录
func (txSim *TxSimulator) SetFwStatus(contractAddr common.Address, status FwStatus) {
	stateObject := txSim.getStateObject(contractAddr)
	fwData := NewFwData()
	for _, addr := range status.RejectedList {
		fwData.DeniedList[addr.FuncName+":"+addr.Addr.String()] = true
	}
	for _, addr := range status.AcceptedList {
		fwData.AcceptedList[addr.FuncName+":"+addr.Addr.String()] = true
	}
	txSim.oc = append(txSim.oc, NewSetFwData(stateObject, contractAddr, fwData))
	active := uint64(0)
	if status.Active {
		active = uint64(1)
	}
	txSim.oc = append(txSim.oc, NewSetFwActive(stateObject, contractAddr, active))
}

func (txSim *TxSimulator) GetFwStatus(contractAddr common.Address) FwStatus {
	return txSim.stateDb.GetFwStatus(contractAddr)
}

//SetContractCreator 优先判断地址是否是合约，如果是则为合约设置创建人，并记录该操作
func (txSim *TxSimulator) SetContractCreator(contractAddr common.Address, creator common.Address) {
	stateObject := txSim.stateDb.GetOrNewStateObjectSafe(contractAddr)
	stateObject.CreateTrie(txSim.stateDb.db)
	txSim.dirty[contractAddr] = stateObject
	log.Debug("TxSimulator SetContractCreator add oc ")
	txSim.oc = append(txSim.oc, NewSetCreator(stateObject, contractAddr, creator))
}

func (txSim *TxSimulator) GetContractCreator(contractAddr common.Address) common.Address {
	log.Debug("GetContractCreator", "addr", contractAddr.Hex())
	return txSim.stateDb.GetContractCreator(contractAddr)
}

//OpenFirewall 记录开启防火墙的操作
func (txSim *TxSimulator) OpenFirewall(contractAddr common.Address) {
	stateObject := txSim.getStateObject(contractAddr)
	txSim.oc = append(txSim.oc, NewSetFwActive(stateObject, contractAddr, uint64(1)))
}

//CloseFirewall 记录关闭防火墙的操作
func (txSim *TxSimulator) CloseFirewall(contractAddr common.Address) {
	obj := txSim.getStateObject(contractAddr)
	txSim.oc = append(txSim.oc, NewSetFwActive(obj, contractAddr, uint64(0)))
}

func (txSim *TxSimulator) IsFwOpened(contractAddr common.Address) bool {
	stateObject := txSim.stateDb.GetOrNewStateObjectSafe(contractAddr)
	return stateObject.FwActive()
}

//FwImport 从json数据反序列化信息后，并记录防火墙的设置操作
func (txSim *TxSimulator) FwImport(contractAddr common.Address, data []byte) error {
	stateObject := txSim.getStateObject(contractAddr)
	status := FwStatus{}
	err := json.Unmarshal(data, &status)
	if err != nil {
		return errors.New("Firewall import failed")
	}
	fwData := NewFwData()
	for _, addr := range status.RejectedList {
		fwData.DeniedList[addr.FuncName+":"+addr.Addr.String()] = true
	}
	for _, addr := range status.AcceptedList {
		fwData.AcceptedList[addr.FuncName+":"+addr.Addr.String()] = true
	}
	txSim.oc = append(txSim.oc, NewSetFwData(stateObject, contractAddr, fwData))
	return nil
}

//CloneAccount 用于合约的迁移，暂时不做处理
func (txSim *TxSimulator) CloneAccount(src common.Address, dest common.Address) error {
	return txSim.stateDb.CloneAccount(src, dest)
}

//StartProcess 开始并行计算
func (self *StateDB) StartProcess() {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()
	self.process = true
	self.SetHashGenerator()
}

//IsProcess 判断是否正进行并行计算
func (self *StateDB) IsProcess() bool {
	self.rwLock.RLock()
	defer self.rwLock.RUnlock()
	return self.process
}

//StopProcess 停止并行计算
func (self *StateDB) StopProcess() {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()
	self.process = false
}

//GetStateByCache 从模拟交易的缓存中获取状态变动
func (self *StateDB) GetStateByCache(keyTrie string) ([]byte, bool) {
	//检查cache是否启用
	if self.writeMap == nil {
		return nil, false
	}
	self.rwLock.RLock()
	defer self.rwLock.RUnlock()
	// 检查cache里是否存在此key的写操作
	if op, ok := self.writeMap[keyTrie]; ok {
		return op.Value, ok
	}
	// 检查cache里是否存在此key的读操作
	if op, ok := self.readMap[keyTrie]; ok {
		return op.Value, ok
	}

	return nil, false
}

//GetBalanceByCache 从模拟交易的缓存中获取余额变动
func (self *StateDB) GetBalanceByCache(addr common.Address) (*big.Int, bool) {
	if self.balanceMap == nil {
		return nil, false
	}
	self.rwLock.RLock()
	defer self.rwLock.RUnlock()
	// 检查cache里是否存在此key
	if value, ok := self.balanceMap[addr]; ok {
		return value.Amount, ok
	}
	return nil, false
}

//GetTxsLen 获取db加入的模拟交易数
func (self *StateDB) GetTxsLen() int {
	self.rwLock.RLock()
	defer self.rwLock.RUnlock()
	return len(self.txs)
}

//AddTxSim 将模拟交易加入到db中
func (self *StateDB) AddTxSim(txSim *TxSimulator, applyCh chan *TxSimulator, withDag bool) (bool, int) { //
	//优先判断并行过程是否已经结束
	if !self.IsProcess() {
		return false, self.GetTxsLen()
	}
	txVersion := txSim.version

	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	txCount := len(self.txs)
	//如果模拟交易的交易版本和state中的交易数一致，表示没有交易在执行过程中加入，不可能存在冲突
	if txVersion == txCount {
		if withDag {
			self.addTxSimWithoutDependency(txSim)
		} else {
			self.addTxSim(txSim)
		}
		applyCh <- txSim
		return true, txCount + 1
	}
	//判断模拟交易是否存在冲突
	if self.checkConflict(txSim) {
		return false, txCount
	}
	if withDag {
		self.addTxSimWithoutDependency(txSim)
	} else {
		self.addTxSim(txSim)
	}
	applyCh <- txSim
	return true, txCount + 1
}

//checkConflict 检查是否存在严重冲突
func (self *StateDB) checkConflict(txSim *TxSimulator) bool {
	//判断模拟交易的读集在state的写集合中是否存在相同的key，且写操作的version大于等于模拟交易的version
	for keyTrie, _ := range txSim.GetReadMap() {
		if wp, ok := self.writeMap[keyTrie]; ok {
			if wp.Version >= txSim.version {
				return true
			}
		}
	}
	//判断模拟交易的余额变更在state的余额变更集合中是否存在相同的key，且变更操作的version大于等于模拟交易的version
	for addr, _ := range txSim.GetBalanceMap() {
		if bp, ok := self.balanceMap[addr]; ok {
			if bp.Version >= txSim.version {
				return true
			}
		}
	}

	//判断模拟模拟交易的object变更在state的余额变更集合中是否存在相同的key，且变更操作的version大于等于模拟交易的version
	for _, v := range txSim.oc {
		if c, ok := self.oc[v.getAddr()]; ok {
			if c.getVersion() >= txSim.version {
				return true
			}
		}
	}

	return false
}

//addTxSim 将模拟交易加入db缓存，并判断交易依赖
func (self *StateDB) addTxSim(txSim *TxSimulator) {
	version := len(self.txs)
	txSim.version = version
	var depend types.Dependency

	if len(txSim.readMap) != 0 {
		for _, op := range txSim.readMap {
			//模拟交易中的读集中的值，在stateDB中读写集cache中的写集内是否已经存在相同的key，获取该笔写操作的version，将该version的值加入该笔交易的Dependency中。
			if wp, ok := self.writeMap[op.KeyTrie]; ok {
				depend = depend.Add(wp.Version)
			}
			op.Version = version
			self.readMap[op.KeyTrie] = op
		}
	}
	if len(txSim.writeMap) != 0 {
		for _, op := range txSim.writeMap {
			//模拟交易中的写集中的值，在stateDB中读写集cache中的写集内是否已经存在相同的key，获取该笔写操作的version，将该version的值加入该笔交易的Dependency中
			if wp, ok := self.writeMap[op.KeyTrie]; ok {
				if wp.Version != version {
					depend = depend.Add(wp.Version)
				}
			}
			//模拟交易中的写集中的值，在stateDB中读写集cache中的读集内是否已经存在相同的key，获取该笔写操作的version，将该version的值加入该笔交易的Dependency中。
			if rp, ok := self.readMap[op.KeyTrie]; ok {
				if rp.Version != version {
					depend = depend.Add(rp.Version)
				}
			}
			op.Version = version
			self.writeMap[op.KeyTrie] = op
		}
	}

	if len(txSim.balanceMap) != 0 {
		for _, op := range txSim.balanceMap {
			//模拟交易中的写集中的值，在stateDB中读写集cache中的写集内是否已经存在相同的key，获取该笔写操作的version，将该version的值加入该笔交易的Dependency中
			if wp, ok := self.balanceMap[op.ContractAddress]; ok {
				if wp.Version != version {
					depend = depend.Add(wp.Version)
				}
			}
			op.Version = version
			self.balanceMap[op.ContractAddress] = op
		}
	}

	if len(txSim.oc) != 0 {
		for _, op := range txSim.oc {
			//模拟交易中的写集中的值，在stateDB中读写集cache中的写集内是否已经存在相同的key，获取该笔写操作的version，将该version的值加入该笔交易的Dependency中
			if c, ok := self.oc[op.getAddr()]; ok {
				if c.getVersion() != version {
					depend = depend.Add(c.getVersion())
				}
			}
			op.setVersion(version)
			self.oc[op.getAddr()] = op
		}
	}

	self.dag = append(self.dag, depend)

	self.txs = append(self.txs, txSim.tx)

	self.receipts = append(self.receipts, txSim.receipt)

}

//addTxSim 将模拟交易加入db缓存
func (self *StateDB) addTxSimWithoutDependency(txSim *TxSimulator) {
	version := len(self.txs)
	txSim.version = txSim.index
	if len(txSim.readMap) != 0 {
		for _, op := range txSim.readMap {
			op.Version = version
			self.readMap[op.KeyTrie] = op
		}
	}
	if len(txSim.writeMap) != 0 {
		for _, op := range txSim.writeMap {
			op.Version = version
			self.writeMap[op.KeyTrie] = op
		}
	}

	if len(txSim.balanceMap) != 0 {
		for _, op := range txSim.balanceMap {
			op.Version = version
			self.balanceMap[op.ContractAddress] = op
		}
	}
	self.txs = append(self.txs, txSim.tx)
}

//ApplyTxSim 将模拟交易的变更应用到stateObject的MPT树中，可异步进行
func (self *StateDB) ApplyTxSim(txSim *TxSimulator, isProposer bool) {
	if len(txSim.writeMap) != 0 {
		//将写集中的变更应用到stateObject的MPT树中
		for _, op := range txSim.writeMap {
			if op.ValueKey == emptyStorage {
				op.object.setError(op.object.trie.TryDelete([]byte(op.KeyTrie)))
				continue
			}
			op.object.setError(op.object.trie.TryUpdate([]byte(op.KeyTrie), op.KeyValue))
			op.object.setError(op.object.trie.TryUpdateValue(op.ValueKey.Bytes(), op.Value))
		}
	}

	if len(txSim.balanceMap) != 0 {
		//余额变更
		for _, op := range txSim.balanceMap {
			op.object.setBalance(op.Amount)
		}
	}

	if len(txSim.oc) != 0 {
		for _, op := range txSim.oc {
			op.change(self)
		}
	}

	// 每笔交易都更新nonce
	for _, addr := range txSim.nonce {
		// nonce 变更
		if obj, ok := txSim.dirty[addr]; ok {
			obj.setNonce(obj.data.Nonce + 1)
		}
	}

	// 每笔交易都更新stateObject的rlp
	for _, obj := range txSim.dirty {
		//更新stateObject
		obj.data.Root = obj.trie.Hash()
		//self.updateStateObject(obj)
		self.SetDirty(obj.address)
	}
	self.gasUsed += txSim.receipt.GasUsed
	//log.Info("add gasUsed", "gas", txSim.receipt.GasUsed, "all", self.gasUsed)
	txSim.receipt.CumulativeGasUsed = self.gasUsed

	self.txTrie.AddItem(txSim.version, txSim.txRlp)
	if isProposer {
		receiptRlp, _ := rlp.EncodeToBytes(txSim.receipt)
		self.receiptTrie.AddItem(txSim.version, receiptRlp)
	}
}

func (self *StateDB) GetDag() types.DAG {
	return self.dag
}

func (self *StateDB) GetTxs() types.Transactions {
	return self.txs
}

func (self *StateDB) GetReceipts() types.Receipts {
	return self.receipts
}

func (self *StateDB) GetGasUsed() uint64 {
	return self.gasUsed
}

func (self *StateDB) SetDirty(address common.Address) {
	self.stateObjectsDirty[address] = struct{}{}
}

func (self *StateDB) UpdateDirtyObject() {
	self.objLock.RLock()
	defer self.objLock.RUnlock()
	for addr := range self.stateObjectsDirty {
		if object, exist := self.stateObjects[addr]; exist {
			self.updateStateObject(object)
		}
	}
}

func (self *StateDB) UpdateReceiptTrie(receipts []*types.Receipt) {
	for index, v := range receipts {
		receiptRlp, _ := rlp.EncodeToBytes(v)
		self.receiptTrie.AddItem(index, receiptRlp)
	}
}

func (self *StateDB) GetTxHash() common.Hash {
	return self.txTrie.Hash()
}

func (self *StateDB) GetReceiptHash() common.Hash {
	return self.receiptTrie.Hash()
}

func (self *StateDB) SetHashGenerator() {
	if common.SysCfg.IsBlockUseTrieHash() {
		self.txTrie = trie.NewHashTrie()
		self.receiptTrie = trie.NewHashTrie()
	} else {
		self.txTrie = trie.NewHashValue()
		self.receiptTrie = trie.NewHashValue()
	}
}
