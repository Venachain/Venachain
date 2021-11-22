package vm

import (
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetEvidence(t *testing.T) {
	db := newMockStateDB()
	e := NewSCEvidence(db)
	address := []byte("0x80989fb9a8eb623dad541c1525828484e5fab75a")
	e.caller = common.BytesToAddress(address)
	id := "cxh"
	value := "cxhwxblockchain"
	err := e.setEvidence(id, value, false)
	if err != nil {
		panic(err)
	}
	res, err := e.getEvidenceById("cxh")
	if err != nil {
		panic(err)
	}
	rr := res.EvidenceValue
	assert.True(t, rr == "cxhwxblockchain")
}
