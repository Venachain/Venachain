package vm

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	crypto1 "github.com/Venachain/Venachain/cmd/ptransfer/client/crypto"
	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/crypto"
)

func TestBpVerify(t *testing.T) {
	db := newMockStateDB()
	address := []byte("0x3fcaa0a86dfbbe105c7ed73ca505c7a59c579667")
	e := NewBPWrapper(db)
	e.base.caller = common.BytesToAddress(address)
	pid := "1"
	proof := "0xf904bbb8802f18e25d035cd59839629f2c6c29ecbd6a48ad634e4f6f4c543b3a7ae5e480f32a19bbb5ffa51948ffd9b4b515da4881f1544190b66a862bcc342e7db15759f20cfe9cc3b3f64516cefd70e20c97a24dfd5a445c2f50b3142e8fcf911351a985048e0f71e5c8ee5932166c0dbd3a7b4853965e517687bcbc301684dd7b46a94bb840077738e7254756f8656c3a9ebf053bd99abb8e1b89f4ef179f5db10d8560a3aa0b9ef61cab42c4454c1e460f05aa7544161d20d97020112b7b3b0227e8d2d168b840205ca04366c1b48f820c85886a3063e1aeda69340d190ba639efd77362e0fee903fa7da96c71830fc2bba2cc3ad449526e8a3719d7387d40d3e2a4485ee832bdb84021a4fe181b1459ea31fa0d16c10f5e26e57f3e468c54ef22569fbb36888286e4255587c3042e1a20eb253bcb3aa651e65490d54bb0110f49cbb8327eea66c699b8401f16aa3eb0a4ea95937b70490770759a21d4a59c635e18411d2480a2dcec3b3b2ade34ce8fdfa2364235215682c9b40f658be262721cf3aba0f2d4a2ca62522da01db1304a7c85877080558291bd9209db8b61dd4efdc314f3693a0c39dda39257a0151fcfd94a604cbf7d1e7f59bc50ba6e85c622f0571206b4ef600ab6d2b31af9a01cb19eaf12e7e916f7879eef34087283cfbdda973cf0d5367414064ed3b1151bb902cbf902c8a01e84baf00e545a29ac494b153012028d07704a3aeff8e31d86d58a9ddd71097aa01e2d99d3ffea7c741c7cfb871a9223bb22a344c4c122c2bce090fe9f9323e87eb901402a16f6c6c38ece0cabda9c56524464d563a44b845442f71d7a5fcd6122ccbf502a01630a17b7337ac11dd7a668939b5541065873f1b24dfda2f515ec675c42d00e86fb02cd4f70b915d6b3d3f27601ed545829230a1bdf82ca11417a68bc009b02f36fd81d4c7cd9300903fcebc14da43acfcbbb26905b3b87bde3d6ad03d731005ffae0cb5d313e6219eeac4dfbf2ef532123370354328fbfcd87dcf784817227fb80b4e1db72a8794d190efd9f772eb6f9fc2176f6b02d97df6e161dac7ac215c8781191210da3194f88dc13c2bf38cd66dbf1f9ae305287907341c249d977045d8745677efcceff8904d203377989b5698a30151b33daf7cf657fa4cdac030a990b63000217d985ef092f85fa8f85c2c1967ff9880e22c72d3ca5be4107932eda2118db09c3f9f71eed941aca4ce74dd571714cc6437b63e3378a32561c25b9014022125cb553d49be00b14cc6c18ba7939ffdbe9ef3e1e804f0007c9243b331f86196b2fe4febd4d0014e7f4a0890c8b7cad67c88d19c55f4b763d0d52e952e106260e2c03d1e8e1490ed70977607ffb0db9a2e0c940ec9c68984d8d7541ef76e60c711212b037583c2d0b7f096ecfa327cb672d021a6a37916a887ef2231c1f2302a200f13a5ce99971eed8e1868da1b35b9d7bfe5e94848c543ca7cfbbe7b2f917a53bdcbfb14215149770e0258ad1c0d97641c77c4030cb9891ed208c70172e0c22753bac12fd497d3d6cf58e03b7baa9a32d61a586e490bda9d702c1136b450e622f5cb5582f20d0065a88b46611225741a3a8a1dd0a00b503f3e254ca16121b3f99337f7048d4119dc42590fa3c38fcbadc10547aded0caa1d1a8c4db73581cc0ddb9abe424cb7933a60a9a16c3a2b73665ebd0b63234abdb1f4408ef1eb1"

	res, err := e.verifyProof(proof, pid)
	if err != nil {
		panic(err)
	}
	assert.True(t, res == 0)
}

func TestBpVerifyByRange(t *testing.T) {
	db := newMockStateDB()
	address := []byte("0x3fcaa0a86dfbbe105c7ed73ca505c7a59c579667")
	sc := "test"
	pid := "1"
	userid := "cxh"
	e := NewBPWrapper(db)
	e.base.caller = common.BytesToAddress(address)

	range_hash := crypto.Keccak256([]byte(sc))
	param := crypto1.GenerateAggBpStatement_range(2, 16, range_hash)
	v := make([]*big.Int, 2)
	v[0] = big.NewInt(3)
	v[1] = big.NewInt(16)
	proof, _ := crypto1.AggBpProve_s(param, v)
	fmt.Println("proof", proof)

	res, _ := e.verifyProofByRange(userid, proof, pid, sc)
	assert.True(t, res == 0)
}

func TestBpGetResult(t *testing.T) {
	db := newMockStateDB()
	address := []byte("0x80989fb9a8eb623dad541c1525828484e5fab75a")
	sc := "test"
	pid := "1"
	userid := "cxh"
	e := NewBPWrapper(db)
	e.base.caller = common.BytesToAddress(address)

	range_hash := crypto.Keccak256([]byte(sc))
	param := crypto1.GenerateAggBpStatement_range(2, 16, range_hash)
	v := make([]*big.Int, 2)
	v[0] = big.NewInt(3)
	v[1] = big.NewInt(16)
	proof, _ := crypto1.AggBpProve_s(param, v)
	fmt.Println(proof)
	res, _ := e.verifyProofByRange(userid, proof, pid, sc)
	fmt.Println(res)
	rr, _ := e.getResult(pid)
	assert.True(t, rr == "test")
}
