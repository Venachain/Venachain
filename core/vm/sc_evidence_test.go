package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"testing"
)

func TestSetEvidence(t *testing.T) {
	db := newMockStateDB()
	e := NewSCEvidence(db)
	address := []byte("0x6e094334323672a7f1254d7b8833649747992430")
	e.caller = common.BytesToAddress(address)
	id := "ljj"
	hash := "lijingjingwxblockchain"
	sig := "7e073d06767e9926a6aed030225bd2e691e50aa97809e7b5c52d7b9ee92a763d322e148b636cdf5670c921ed0d316d8fcb931f4750ec15c087f5438c7b2c252d00"
	err := e.saveEvidence(id, hash, sig)
	if err != nil {
		panic(err)
	}

	res, err := e.getEvidenceById("ljj")
	if err != nil {
		panic(err)
	}
	println(res)
}
