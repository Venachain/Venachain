package vm

import (
	"encoding/json"
	"math/big"
	"strings"

	"github.com/Venachain/Venachain/rlp"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/common/syscontracts"
	"github.com/Venachain/Venachain/crypto/paillier"
)

// PaillierManager
type PaillierManager struct {
	stateDB      StateDB
	contractAddr common.Address
	caller       common.Address // caller = Contract.CallerAddress
	blockNumber  *big.Int       // blockNumber = evm.BlockNumber
}

type Paillier struct {
	PaillierType string      `json:"paillierType"`
	PubKey       string      `json:"pubKey"`
	Data         interface{} `json:"data"`
	Timestamp    *big.Int    `json:"timestamp"`
}

func NewPL(db StateDB) *PaillierManager {
	return &PaillierManager{
		stateDB:      db,
		contractAddr: syscontracts.PaillierAddress,
		blockNumber:  big.NewInt(0),
	}
}

func NewPaillier(paillierType string) *Paillier {
	return &Paillier{
		PaillierType: paillierType,
	}
}

// PaillierWeightAdd
func (he *PaillierManager) paillierWeightAdd(arg string, arr []uint, pubKey string) (string, error) {
	args := strings.Split(arg, ",")
	return paillier.PaillierWeightAdd(args, arr, pubKey)
}

// PaillierAdd
func (he *PaillierManager) paillierAdd(arg string, pubKey string) (string, error) {
	args := strings.Split(arg, ",")
	return paillier.PaillierAdd(args, pubKey)
}

// PaillierMul
func (he *PaillierManager) paillierMul(arg string, scalar uint, pubKey string) (string, error) {
	return paillier.PaillierMul(arg, scalar, pubKey)
}

func (he *PaillierManager) savePaillier(paillierType, pubKey string, data interface{}, result string) error {
	pai := NewPaillier(paillierType)
	pai.PubKey = pubKey
	pai.Data = data
	pai.Timestamp = he.blockNumber
	key, _ := json.Marshal(pai)
	valueInBytes, err := rlp.EncodeToBytes(result)
	if err != nil {
		return err
	}
	he.stateDB.SetState(he.contractAddr, key, valueInBytes)
	return nil
}

func (he *PaillierManager) emitNotifyEvent(code CodeType, msg string) {
	topic := "Notify"
	he.emitEvent(topic, code, msg)
}

func (he *PaillierManager) emitEvent(topic string, code CodeType, msg string) {
	emitEvent(he.contractAddr, he.stateDB, he.blockNumber.Uint64(), topic, code, msg)
}
