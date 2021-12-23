package state

import (
	"math/big"

	"github.com/Venachain/Venachain/common"
)

type ReadOp struct {
	ContractAddress common.Address //合约地址
	Key             []byte         //
	KeyTrie         string
	Value           []byte
	Version         int //txSimulator中用到的版本号
}
type ReadSet []*ReadOp

type WriteOp struct {
	object          *stateObject
	ContractAddress common.Address //合约地址
	Key             []byte         //
	KeyTrie         string
	KeyValue        []byte
	ValueKey        common.Hash
	Value           []byte
	Version         int //txSimulator中用到的版本号
}
type WriteSet []*WriteOp

type BalanceOp struct {
	object          *stateObject
	ContractAddress common.Address //合约地址
	Amount          *big.Int
	Version         int //txSimulator中用到的版本号
}

type BalanceSet []*BalanceOp
