package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/params"
	"strings"
)

const (
	setEvidenceSuccess      CodeType = 0
	setEvidenceAuthFailed   CodeType = 1
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
	return ret, nil
}
func NewEvidenceWrapper(db StateDB) *SCEvidenceWrapper {
	return &SCEvidenceWrapper{NewSCEvidence(db)}
}

func (e *SCEvidenceWrapper) saveEvidence(id string, hash string, sig string) (int, error) {
	if err := e.base.saveEvidence(id, hash, sig); nil != err {
		switch err {
		case errAuthenticationFailed:
			return int(setEvidenceAuthFailed), err
		case errAlreadySetEvidence:
			return int(setEvidenceAlreadyExist), err
		default:
			return int(setEvidenceAuthFailed), err
		}
	}

	return int(setEvidenceSuccess), nil
}

func (e *SCEvidenceWrapper) getEvidence(id string) (string, error) {
	evidence, err := e.base.getEvidenceById(id)
	if err != nil {
		return "", err
	}
	return newSuccessResult(evidence).String(), nil
}

func (e *SCEvidenceWrapper) allExportFns() SCExportFns {
	return SCExportFns{
		"saveEvidence": e.saveEvidence,
		"getEvidence":  e.getEvidence,
	}
}
