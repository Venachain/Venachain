package vm

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/params"
)

const (
	paillierSuccess CodeType = 0
	paillierFailed  CodeType = 1

	paillierWeightAddType = "paillierWeightAdd"
	paillierAddType       = "paillierAdd"
	paillierMulType       = "paillierMul"

	paillierResult = "paillierResult:%s"
)

type SCPaillierWrapper struct {
	base *PaillierManager
}

func (s SCPaillierWrapper) RequiredGas(input []byte) uint64 {
	if common.IsBytesEmpty(input) {
		return 0
	}
	return params.SCPaillierProofGas
}

func (s SCPaillierWrapper) Run(input []byte) ([]byte, error) {
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
func (s *SCPaillierWrapper) AllExportFns() SCExportFns {
	return SCExportFns{
		"paillierWeightAdd": s.paillierWeightAdd,
		"paillierAdd":       s.paillierAdd,
		"paillierMul":       s.paillierMul,
	}
}

func NewPLWrapper(db StateDB) *SCPaillierWrapper {
	return &SCPaillierWrapper{NewPL(db)}
}

func (s *SCPaillierWrapper) paillierWeightAdd(args string, arr string, pubKey string) (string, error) {
	var intArr []uint
	split := strings.Split(arr, ",")
	for _, i := range split {
		l, err := strconv.Atoi(i)
		if err != nil {
			return "", err
		}
		intArr = append(intArr, uint(l))
	}
	res, err := s.base.paillierWeightAdd(args, intArr, pubKey)
	if err != nil {
		return "", err
	}
	data := struct {
		Args string
		Arr  string
	}{Args: args, Arr: arr}
	if err := s.base.savePaillier(paillierWeightAddType, pubKey, data, res); err != nil {
		return "", err
	}
	s.base.emitNotifyEvent(paillierSuccess, fmt.Sprintf(paillierResult, res))
	fmt.Println("the result of PaillierWeightAdd is:", res)
	return res, err
}

func (s *SCPaillierWrapper) paillierAdd(args string, pubKey string) (string, error) {
	res, err := s.base.paillierAdd(args, pubKey)
	if err != nil {
		return "", err
	}
	data := struct {
		Args string
	}{Args: args}
	if err := s.base.savePaillier(paillierAddType, pubKey, data, res); err != nil {
		return "", err
	}
	s.base.emitNotifyEvent(paillierSuccess, fmt.Sprintf(paillierResult, res))
	fmt.Println("the result of the PaillierAdd is:", res)
	return res, err
}

func (s *SCPaillierWrapper) paillierMul(arg string, scalar string, pubKey string) (string, error) {
	intScalar, err := strconv.Atoi(scalar)
	if err != nil {
		return "", err
	}
	res, err := s.base.paillierMul(arg, uint(intScalar), pubKey)
	if err != nil {
		return "", err
	}
	data := struct {
		Args   string
		Scalar string
	}{Args: arg, Scalar: scalar}
	if err := s.base.savePaillier(paillierMulType, pubKey, data, res); err != nil {
		return "", err
	}
	s.base.emitNotifyEvent(paillierSuccess, fmt.Sprintf(paillierResult, res))
	fmt.Println("the result of the PaillierMul is:", res)
	return res, err
}
