package vm

import (
	"strings"

	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
)

const (
	setEvidenceSuccess      CodeType = 0
	setEvidenceFailed       CodeType = 1
	setEvidenceAlreadyExist CodeType = 2
)

type SCEvidenceWrapper struct {
	base *SCEvidence
}

func (e *SCEvidenceWrapper) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.SCEvidenceGas
}

func (e *SCEvidenceWrapper) Run(input []byte) ([]byte, error) {
	fnName, ret, err := execSC(input, e.allExportFns())
	if err != nil {
		if fnName == "" {
			fnName = "Notify"
		}
		e.base.emitEvent(fnName, operateFail, err.Error())

		if strings.Contains(fnName, "get") {
			return MakeReturnBytes([]byte(newInternalErrorResult(err).String())), err
		}
	}
	return ret, err
}
func NewEvidenceWrapper(db StateDB) *SCEvidenceWrapper {
	return &SCEvidenceWrapper{NewSCEvidence(db)}
}

func (e *SCEvidenceWrapper) saveEvidence(key string, value string) (int, error) {
	if err := e.base.setEvidence(key, value, false); nil != err {
		switch err {
		//case errAuthenticationFailed:
		//	return int(setEvidenceFailed), err
		case errAlreadySetEvidence:
			return int(setEvidenceAlreadyExist), err
		default:
			return int(setEvidenceFailed), err
		}
	}
	e.base.emitNotifyEvent(setEvidenceSuccess,"save evidence success")
	return int(setEvidenceSuccess), nil
}

func (e *SCEvidenceWrapper) getEvidence(key string) (string, error) {
	evidence, err := e.base.getEvidenceById(key)
	if err != nil {
		return "", err
	}
	return newSuccessResult(evidence).String(), nil
}

func (e *SCEvidenceWrapper) setJsonData(key string, hash string) (int, error) {
	if err := e.base.setEvidence(key, hash, true); nil != err {
		switch err {
		//case errAuthenticationFailed:
		//	return int(setEvidenceFailed), err
		case errAlreadySetEvidence:
			return int(setEvidenceAlreadyExist), err
		default:
			return int(setEvidenceFailed), err
		}
	}
	e.base.emitNotifyEvent(setEvidenceSuccess,"set jsondata success.")
	return int(setEvidenceSuccess), nil
}

func (e *SCEvidenceWrapper) getJsonData(key string) (string, error) {
	evidence, err := e.base.getEvidenceById(key)
	if err != nil {
		return "", err
	}
	return evidence.EvidenceValue, nil
}

func (e *SCEvidenceWrapper) allExportFns() SCExportFns {
	return SCExportFns{
		"saveEvidence": e.saveEvidence,
		"getEvidence":  e.getEvidence,
		"setJsonData":  e.setJsonData,
		"getJsonData":  e.getJsonData,
	}
}
