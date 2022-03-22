package vm

import (
	"fmt"
	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/params"
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

func (s *SCPaillierWrapper) paillierWeightAdd(args string, arr []uint, pubKey string) (string, error) {
	res := s.base.paillierWeightAdd(args, arr, pubKey)
	fmt.Println("the result of PaillierWeightAdd is:", res)
	return res, nil
}

func (s *SCPaillierWrapper) paillierAdd(args string, pubKey string) (string, error) {
	res := s.base.paillierAdd(args, pubKey)
	fmt.Println("the result of the PaillierAdd is:", res)
	return res, nil
}

func (s *SCPaillierWrapper) paillierMul(arg string, scalar uint, pubKey string) (string, error) {
	res := s.base.paillierMul(arg, scalar, pubKey)
	fmt.Println("the result of the PaillierMul is:", res)
	return res, nil
}
