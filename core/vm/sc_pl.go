package vm

import (
	"math/big"
	"strings"

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

func NewPL(db StateDB) *PaillierManager {
	return &PaillierManager{
		stateDB:      db,
		contractAddr: syscontracts.PaillierAddress,
		blockNumber:  big.NewInt(0),
	}
}

// PaillierWeightAdd
func (he *PaillierManager) paillierWeightAdd(arg string, arr []uint, pubKey string) string {
	args := strings.Split(arg, ",")
	res := paillier.PaillierWeightAdd(args, arr, pubKey)
	return res
}

// PaillierAdd
func (he *PaillierManager) paillierAdd(arg string, pubKey string) string {
	args := strings.Split(arg, ",")
	res := paillier.PaillierAdd(args, pubKey)
	return res
}

// PaillierMul
func (he *PaillierManager) paillierMul(arg string, scalar uint, pubKey string) string {
	res := paillier.PaillierMul(arg, scalar, pubKey)
	return res
}

func (he *PaillierManager) emitNotifyEvent(code CodeType, msg string) {
	topic := "Notify"
	he.emitEvent(topic, code, msg)
}

func (he *PaillierManager) emitEvent(topic string, code CodeType, msg string) {
	emitEvent(he.contractAddr, he.stateDB, he.blockNumber.Uint64(), topic, code, msg)
}
