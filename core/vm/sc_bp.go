package vm

import (
	"bytes"
	"encoding/binary"
	"math/big"

	"github.com/Venachain/Venachain/crypto"
	"github.com/Venachain/Venachain/rlp"

	cryptobp "github.com/Venachain/Venachain/cmd/ptransfer/client/crypto"
	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/common/syscontracts"
)

var aggbpparam *cryptobp.AggBpStatement

// BulletProofManager
type BulletProofManager struct {
	stateDB      StateDB
	contractAddr common.Address
	caller       common.Address // caller = Contract.CallerAddress
	blockNumber  *big.Int       // blockNumber = evm.BlockNumber
}

type Response struct {
	Userid string `json:"userid"`
	Proof  string `json:"proof"`
	Scope  string `json:"scope"`
}

func init() {
	aggbpparam = cryptobp.GenerateAggBpStatement(2, 16)
}

func NewBP(db StateDB) *BulletProofManager {
	return &BulletProofManager{
		stateDB:      db,
		contractAddr: syscontracts.BulletProofAddress,
		blockNumber:  big.NewInt(0),
	}
}

// verify 只返回结果，具体查看verifyByRange
func (bp *BulletProofManager) verify(proof string, pid string) (bool, error) {
	//var result string
	param := aggbpparam
	res, err := cryptobp.AggBpVerify_s(proof, param)
	if err != nil {
		return false, err
	}
	if res == false {
		return res, errBPGetFalseResult
	}

	// set State
	bp.stateDB.SetState(bp.contractAddr, []byte(pid), boolToByte(res))

	bp.emitNotifyEvent(verifyBPSuccess, "the verification result is true")

	return res, nil
}

func (bp *BulletProofManager) verifyByRange(userid, proof, pid, scope string) (bool, error) {
	range_hash := crypto.Keccak256([]byte(scope))
	// 同时证明2个数在0到2^(16)-1 之内
	param := cryptobp.GenerateAggBpStatement_range(2, 16, range_hash)
	response := Response{
		Userid: userid,
		Proof:  proof,
		Scope:  scope,
	}
	if bp.getState(pid) != nil {
		bp.emitNotifyEvent(alreadySetPid, "the pid is already set")
		return false, errAlreadySetPid
	}
	res, err := cryptobp.AggBpVerify_s(proof, param)
	if err != nil {
		return false, err
	}
	if res == false {
		bp.emitNotifyEvent(getProofFalse, "the verification result is false")
		return res, errBPGetFalseResult
	}

	// set State
	valueInBytes, err := rlp.EncodeToBytes(response)
	if err != nil {
		return false, err
	}
	bp.stateDB.SetState(bp.contractAddr, []byte(pid), valueInBytes)
	bp.emitNotifyEvent(verifyBPSuccess, "the verification result is true")
	return res, nil
}

func (bp *BulletProofManager) getResult(pid string) (*Response, error) {
	var result Response
	value := bp.getState(pid)
	if len(value) == 0 {
		return nil, errEvidenceNotFound
	}
	err := rlp.DecodeBytes(value, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (bp *BulletProofManager) getProof(pid string) (string, error) {

	return "", nil
}

func (bp *BulletProofManager) emitNotifyEvent(code CodeType, msg string) {
	topic := "Notify"
	bp.emitEvent(topic, code, msg)
}

func (bp *BulletProofManager) emitEvent(topic string, code CodeType, msg string) {
	emitEvent(bp.contractAddr, bp.stateDB, bp.blockNumber.Uint64(), topic, code, msg)
}

func (bp *BulletProofManager) getState(pid string) []byte {
	return bp.stateDB.GetState(bp.contractAddr, []byte(pid))
}

func boolToByte(parm bool) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, parm)
	return buf.Bytes()
}
