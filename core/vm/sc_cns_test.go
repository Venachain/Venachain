package vm

import (
	"math/big"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/core/state"
	"github.com/PlatONEnetwork/PlatONE-Go/core/types"
	"github.com/stretchr/testify/assert"
)

const (
	testAddr1 = "0x0000000000000000000000000000000000000123"
	testAddr2 = "0x0000000000000000000000000000000000000456"
	testAddr3 = "0x0000000000000000000000000000000000000789"
	testAddr4 = "0x0000000000000000000000000000000000000101"

	testOrigin = "0x0000000000000000000000000000000000000afb"
	testCaller = "0x0000000000000000000000000000000000000afc"

	testName = "tofu"
)

func TestLatestVersion(t *testing.T) {
	testCase := []struct {
		ver1 string
		ver2 string
	}{
		{"0.0.0.0", "0.0.2.0"},
		{"1.0.0.0", "0.0.0.1"},
	}

	for _, data := range testCase {
		if verCompare(data.ver1, data.ver2) == 1 {
			t.Logf("ver1 %s is larger than ver2 %s\n", data.ver1, data.ver2)
		} else {
			t.Logf("ver1 %s is smaller than ver2 %s\n", data.ver1, data.ver2)
		}
	}
}

func TestSerializeCnsInfo(t *testing.T) {
	cnsInfoArray := make([]*ContractInfo, 0)
	cnsInfo := newContractInfo("tofu", "0.0.0.1", "0x123", "0x123")
	cnsInfoArray = append(cnsInfoArray, cnsInfo, cnsInfo, cnsInfo)

	sBytes, err := serializeCnsInfo(codeOk, msgOk, cnsInfoArray)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%s\n", sBytes)
}

var (
	cns       *CnsManager
	db        *mockStateDB
	testCases []*ContractInfo
	key       = make([][]byte, 0)
)

// inital the cns with some pre prepared data
func TestMain(m *testing.M) {
	db = newMockStateDB()
	addr := common.HexToAddress("")

	cns = &CnsManager{
		cMap:       NewCnsMap(db, &addr),
		callerAddr: common.HexToAddress(testCaller),
		origin:     common.HexToAddress(testOrigin),
		isInit:     -1,
	}

	testCases = []*ContractInfo{
		{
			Name:    testName,
			Version: "0.0.0.1",
			Address: testAddr1,
			Origin:  testOrigin,
		},
		{
			Name:    testName,
			Version: "0.0.0.2",
			Address: testAddr2,
			Origin:  testOrigin,
		},
		{
			Name:    testName,
			Version: "0.0.0.3",
			Address: testAddr3,
			Origin:  testOrigin,
		},
		{
			Name:    testName,
			Version: "0.0.0.4",
			Address: testAddr4,
			Origin:  testOrigin,
		},
		{
			Name:    "bob",
			Version: "0.0.0.1",
			Address: "0x123",
			Origin:  testOrigin,
		},
	}

	for _, data := range testCases {
		value, err := data.encode()
		if err != nil {
			// m.Fatalf(err.Error())
		}

		k := getSearchKey(data.Name, data.Version)
		cns.cMap.insert(k, value)
		cns.cMap.updateLatestVer(data.Name, data.Version)

		key = append(key, k)
	}

	m.Run()
}

func TestCnsManager_cMap(t *testing.T) {

	assert.Equal(t, key[1], cns.cMap.getKey(1), "cns getKey FAILED")
	assert.Equal(t, testCases[0], cns.cMap.find(key[0]), "cns find() FAILED")
	assert.Equal(t, len(testCases), cns.cMap.total(), "cns total() FAILED")

	t.Log(db.mockDB)
}

func TestCnsManager_cnsRegister(t *testing.T) {
	result, err := cns.cnsRegister("alice", "0.0.0.1", testAddr1)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, success, result, "cnsRegister FAILED")
	t.Log(db.mockDB)
}

func TestCnsManager_getContractAddress(t *testing.T) {

	testCasesSub := []struct {
		name     string
		version  string
		expected string
	}{
		{testName, "0.0.0.2", testAddr2},
		{testName, "latest", testAddr4},
	}

	for _, data := range testCasesSub {
		result, err := cns.getContractAddress(data.name, data.version)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, data.expected, result, "getContractAddress FAILED")
		t.Log(result)
	}

}

func TestCnsManager_cnsRecall(t *testing.T) {

	curVersion := cns.cMap.getLatestVer(testName)

	result, err := cns.cnsRedirect(testName, testCases[2].Version)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, success, result, "cnsRecall FAILED")

	actVersion := cns.cMap.getLatestVer(testName)
	expVersion := testCases[2].Version
	assert.Equal(t, expVersion, actVersion, "cnsRecall FAILED")

	t.Logf("before: %s, after cnsRecall: %s\n", curVersion, actVersion)
}

func TestCnsManager_ifRegisteredByName(t *testing.T) {
	testCasesSub := []struct {
		name     string
		expected int
	}{
		{testName, registered},
		{"tom", unregistered},
	}

	for _, data := range testCasesSub {
		result, err := cns.ifRegisteredByName(data.name)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, data.expected, result, "ifRegisteredByName FAILED")
	}
}

func TestCnsManager_getRegisteredContractsByRange(t *testing.T) {
	result, err := cns.getRegisteredContractsByRange(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("the result is %s\n", result)
}

//==================================mock======================================

func newMockStateDB() *mockStateDB {
	return &mockStateDB{
		mockDB: make(map[common.Address]map[string][]byte),
	}
}

type mockStateDB struct {
	mockDB map[common.Address]map[string][]byte
}

func (m *mockStateDB) GetState(addr common.Address, key []byte) []byte {

	return m.mockDB[addr][string(key)]
}

func (m *mockStateDB) SetState(addr common.Address, key []byte, value []byte) {

	if m.mockDB[addr] == nil {
		m.mockDB[addr] = make(map[string][]byte)
	}

	m.mockDB[addr][string(key)] = value
}

func (m *mockStateDB) GetContractCreator(contractAddr common.Address) common.Address {
	return common.HexToAddress(testOrigin)
}

func (m *mockStateDB) CreateAccount(common.Address) {
	panic("implement me")
}

func (m *mockStateDB) SubBalance(common.Address, *big.Int) {
	panic("implement me")
}

func (m *mockStateDB) AddBalance(common.Address, *big.Int) {
	panic("implement me")
}

func (m *mockStateDB) GetBalance(common.Address) *big.Int {
	panic("implement me")
}

func (m *mockStateDB) GetNonce(common.Address) uint64 {
	panic("implement me")
}

func (m *mockStateDB) SetNonce(common.Address, uint64) {
	panic("implement me")
}

func (m *mockStateDB) GetCodeHash(common.Address) common.Hash {
	panic("implement me")
}

func (m *mockStateDB) GetCode(common.Address) []byte {
	panic("implement me")
}

func (m *mockStateDB) SetCode(common.Address, []byte) {
	panic("implement me")
}

func (m *mockStateDB) GetCodeSize(common.Address) int {
	panic("implement me")
}

func (m *mockStateDB) GetAbiHash(common.Address) common.Hash {
	panic("implement me")
}

func (m *mockStateDB) GetAbi(common.Address) []byte {
	panic("implement me")
}

func (m *mockStateDB) SetAbi(common.Address, []byte) {
	panic("implement me")
}

func (m *mockStateDB) AddRefund(uint64) {
	panic("implement me")
}

func (m *mockStateDB) SubRefund(uint64) {
	panic("implement me")
}

func (m *mockStateDB) GetRefund() uint64 {
	panic("implement me")
}

func (m *mockStateDB) GetCommittedState(common.Address, []byte) []byte {
	panic("implement me")
}

func (m *mockStateDB) Suicide(common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) HasSuicided(common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) Exist(common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) Empty(common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) RevertToSnapshot(int) {
	panic("implement me")
}

func (m *mockStateDB) Snapshot() int {
	panic("implement me")
}

func (m *mockStateDB) AddLog(*types.Log) {
	panic("implement me")
}

func (m *mockStateDB) AddPreimage(common.Hash, []byte) {
	panic("implement me")
}

func (m *mockStateDB) ForEachStorage(common.Address, func(common.Hash, common.Hash) bool) {
	panic("implement me")
}

func (m *mockStateDB) FwAdd(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m *mockStateDB) FwClear(contractAddr common.Address, action state.Action) {
	panic("implement me")
}

func (m *mockStateDB) FwDel(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m *mockStateDB) FwSet(contractAddr common.Address, action state.Action, list []state.FwElem) {
	panic("implement me")
}

func (m *mockStateDB) SetFwStatus(contractAddr common.Address, status state.FwStatus) {
	panic("implement me")
}

func (m *mockStateDB) GetFwStatus(contractAddr common.Address) state.FwStatus {
	panic("implement me")
}

func (m *mockStateDB) SetContractCreator(contractAddr common.Address, creator common.Address) {
	panic("implement me")
}

func (m *mockStateDB) OpenFirewall(contractAddr common.Address) {
	panic("implement me")
}

func (m *mockStateDB) CloseFirewall(contractAddr common.Address) {
	panic("implement me")
}

func (m *mockStateDB) IsFwOpened(contractAddr common.Address) bool {
	panic("implement me")
}

func (m *mockStateDB) FwImport(contractAddr common.Address, data []byte) error {
	panic("implement me")
}
