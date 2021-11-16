package vm

import (
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

const (
	verifyBPSuccess          CodeType = 0
	verifyBPFailed           CodeType = 1
	getProofFalse            CodeType = 2
	alreadySetPid 			 CodeType = 3

)

type SCBulletProofWrapper struct {
	base *BulletProofManager
}

func (s SCBulletProofWrapper) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.SCBulletProofGas
}

func (s SCBulletProofWrapper) Run(input []byte) ([]byte, error) {
	fnName, ret, err := execSC(input, s.AllExportFns())
	if err != nil {
		if fnName == "" {
			fnName = "Notify"
		}
		s.base.emitEvent(fnName, operateFail, err.Error())
	}

	return ret, err
}

// for access control
func (s *SCBulletProofWrapper) AllExportFns() SCExportFns {
	return SCExportFns{
		"verifyProof": s.verifyProof,
		"verifyProofByRange": s.verifyProofByRange,
		"getResult":  s.getResult,
		"getProof":  s.getProof,
	}
}

func NewBPWrapper(db StateDB) *SCBulletProofWrapper {
	return &SCBulletProofWrapper{NewBP(db)}
}

func (s *SCBulletProofWrapper) verifyProof(proof string,pid string) (int32,error) {
	var res bool
	var err error
	if res,err = s.base.verify(proof, pid); nil != err {
		switch err {
		case errBPGetFalseResult:
			return int32(verifyBPFailed), err
		default:
			return int32(verifyBPFailed), err
		}
	}
	fmt.Println("the result of the proof is:",res)
	return int32(verifyBPSuccess), nil
}
func (s *SCBulletProofWrapper) verifyProofByRange(userid, proof ,pid,scope string) (int32,error) {
	var res bool
	var err error
	if res,err = s.base.verifyByRange(userid,proof, pid,scope); nil != err {
		switch err {
		//case errEvidenceNotFound:
		//	return int32(getEvidenceNotExist), err
		default:
			return int32(verifyBPFailed), err
		}
	}
	fmt.Println("the result of the proof is:",res)
	return int32(verifyBPSuccess), nil
}

func (s *SCBulletProofWrapper) getResult(pid string) (string,error) {
	res, err := s.base.getResult(pid)
	if err != nil {
		return "", err
	}

	return newSuccessResult(res).String(),nil
}

func (s *SCBulletProofWrapper) getProof(pid string) (string,error)  {
	res, err := s.base.getProof(pid)
	if err != nil {
		return "", err
	}
	return res,nil
}
