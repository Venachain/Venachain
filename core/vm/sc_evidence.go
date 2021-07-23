package vm

import (
	"math/big"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/syscontracts"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

type SCEvidence struct {
	stateDB      StateDB
	contractAddr common.Address
	caller       common.Address
	blockNumber  *big.Int
}

type Evidence struct {
	Owner        common.Address
	EvidenceHash common.Hash
	EvidenceSig  string
	Timestamp    *big.Int
}

func NewSCEvidence(db StateDB) *SCEvidence {
	return &SCEvidence{
		stateDB:      db,
		contractAddr: syscontracts.NodeManagementAddress, //todo add evidence address
		blockNumber:  big.NewInt(0),
	}
}
func NewEvidence() *Evidence {
	return &Evidence{
		Owner:        ZeroAddress,
		EvidenceHash: common.Hash{},
		EvidenceSig:  "",
		Timestamp:    big.NewInt(0),
	}
}

func (e *SCEvidence) saveEvidence(id string, hash string, sig string) error {
	//if !e.checkSignature(hash, sig) {
	//	return errAuthenticationFailed
	//}
	evidence := NewEvidence()
	evidence.EvidenceHash = common.BytesToHash([]byte(hash))
	evidence.EvidenceSig = sig
	evidence.Owner = e.caller
	evidence.Timestamp = e.blockNumber
	valueInBytes, err := rlp.EncodeToBytes(evidence)
	if err != nil {
		return err
	}
	if e.getState(id) != nil {
		return errAlreadySetEvidence
	}
	e.setState(id, valueInBytes)
	return nil
}

func (e *SCEvidence) checkSignature(hash string, sig string) bool {
	pub, err := crypto.SigToPub(crypto.Keccak256([]byte(hash)), []byte(sig))
	if err != nil {
		log.Error("Ecrecover Fail", "hash =", hash, "signature =", sig)
		return false
	}
	if e.caller != crypto.PubkeyToAddress(*pub) {
		log.Error("Authentication failed！！！")
		return false
	}
	return true
}

func (e *SCEvidence) getEvidenceById(id string) (*Evidence, error) {
	value := e.getState(id)
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
