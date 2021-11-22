package vm

import (
	"math/big"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

type SCEvidence struct {
	stateDB      StateDB
	contractAddr common.Address
	caller       common.Address
	blockNumber  *big.Int
}

//type Evidence struct {
//	Owner        common.Address
//	EvidenceHash common.Hash
//	EvidenceSig  string
//	Timestamp    *big.Int
//}

type Evidence struct {
	Owner        common.Address
	EvidenceValue  string
	Timestamp    *big.Int
}

func NewSCEvidence(db StateDB) *SCEvidence {
	return &SCEvidence{
		stateDB:      db,
		contractAddr: syscontracts.EvidenceManagementAddress,
		blockNumber:  big.NewInt(0),
	}
}
func NewEvidence() *Evidence {
	return &Evidence{
		Owner:        ZeroAddress,
		//EvidenceHash: common.Hash{},
		EvidenceValue:  "",
		Timestamp:    big.NewInt(0),
	}
}

// isUpdateEvidence 判断是否需要对Evidence 进行更新。
func (e *SCEvidence) setEvidence(key string, value string, isUpdateEvidence bool) error {
	evidence := NewEvidence()
	evidence.EvidenceValue = value
	evidence.Owner = e.caller
	evidence.Timestamp = e.blockNumber
	valueInBytes, err := rlp.EncodeToBytes(evidence)
	if err != nil {
		//e.emitNotifyEvent(setEvidenceFailed,err.Error())
		return err
	}
	// 如果不需要对Evidence 进行更新
	if !isUpdateEvidence && e.getState(key) != nil {
		//e.emitNotifyEvent(setEvidenceFailed,"evidence already exists")
		return errAlreadySetEvidence
	}
	e.setState(key, valueInBytes)
	return nil
}

//func (e *SCEvidence) checkSignature(hash string, sig string) bool {
//	pub, err := crypto.SigToPub(crypto.Keccak256([]byte(hash)), []byte(sig))
//	if err != nil {
//		log.Error("Ecrecover Fail", "hash =", hash, "signature =", sig)
//		return false
//	}
//	if e.caller != crypto.PubkeyToAddress(*pub) {
//		log.Error("Authentication failed！！！")
//		return false
//	}
//	return true
//}

func (e *SCEvidence) getEvidenceById(key string) (*Evidence, error) {
	value := e.getState(key)
	if len(value) == 0 {
		return nil, errEvidenceNotFound
	}
	var evidence Evidence
	err := rlp.DecodeBytes(value, &evidence)
	if err != nil {
		return nil, err
	}
	return &evidence, nil
}

func (e *SCEvidence) setState(key string, value []byte) {
	e.stateDB.SetState(e.contractAddr, []byte(key), value)
}

func (e *SCEvidence) getState(key string) []byte {
	return e.stateDB.GetState(e.contractAddr, []byte(key))
}

func (e *SCEvidence) emitNotifyEvent(code CodeType, msg string) {
	topic := "Notify"
	e.emitEvent(topic, code, msg)
}

func (e *SCEvidence) emitEvent(topic string, code CodeType, msg string) {
	emitEvent(e.contractAddr, e.stateDB, e.blockNumber.Uint64(), topic, code, msg)
}
