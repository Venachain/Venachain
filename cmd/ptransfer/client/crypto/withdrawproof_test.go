package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/Venachain/Venachain/common/hexutil"
	"github.com/Venachain/Venachain/crypto/bn256"
)

func TestWithdrawProve(t *testing.T) {
	key, _ := NewKeyPair(rand.Reader)
	sk := key.sk
	pk := key.pk
	//r, _ := rand.Int(rand.Reader, ORDER)
	c, _ := Enc(rand.Reader, pk, big.NewInt(10))
	cLn := c.C
	cRn := c.D
	//cLn := new(bn256.G1).Add(new(bn256.G1).ScalarMult(AggBp().bpParam.g, big.NewInt(10)), new(bn256.G1).ScalarMult(pk, r))
	//cRn := new(bn256.G1).ScalarMult(AggBp().bpParam.g, r)
	epoch := big.NewInt(4)
	sender := []byte{1, 2, 3, 4, 5}
	witness := &WithdrawWitness{
		priv:  sk,
		vDiff: &AggBpWitness{},
	}
	v := make([]*big.Int, 1)
	v[0] = big.NewInt(10) //the cLn = g*vDiff + r * pk
	//v[1] = big.NewInt(4)
	witness.vDiff.v = v
	t.Logf("witness%v:\n", v)
	//witness.vDiff.v[1] = big.NewInt(4)
	proof, _ := WithdrawProve(cLn, cRn, pk, epoch, sender, witness)
	pstr := proof.WdProofMarshal()
	newproof, _ := WdProofUnMarshal(pstr)

	gepoch := MapIntoGroup("zether" + epoch.String())
	u := new(bn256.G1).ScalarMult(gepoch, sk)
	statement := &WithdrawStatement{cLn, cRn, pk, epoch, sender, u}
	res, _ := WithdrawVerify(statement, newproof)
	t.Log(res)
}

func TestWithdrawProve2(t *testing.T) {
	priv := "1ee4808d414900fde6e8c1713098f10b42d922fa1e27ec41a911f840a046aebf"
	//priv := "2a75f45840f076b927965ad3860729bbd4753f626df484c472ca7d2831341e3f"
	sk, _ := new(big.Int).SetString(priv, 16)
	pk := new(bn256.G1).ScalarMult(G, sk)
	cLn := new(bn256.G1).Add(pk, new(bn256.G1).ScalarMult(G, big.NewInt(100)))
	cRn := G
	bTransfer := big.NewInt(20)
	cLnNew := new(bn256.G1).Add(cLn, new(bn256.G1).Neg(new(bn256.G1).ScalarMult(G, bTransfer)))
	fmt.Println(cLnNew)
	//fmt.Println(sk)
	//fmt.Println(pk)
	//fmt.Println(cLn)
	epoch := big.NewInt(200)
	sender, _ := hexutil.Decode("0x7760FaFcD09CF06b627673A9b0eA77867b6e0427")
	witness := &WithdrawWitness{
		priv:  sk,
		vDiff: &AggBpWitness{},
	}
	v := make([]*big.Int, 1)
	v[0] = big.NewInt(80) //the cLn = g*vDiff + r * pk
	witness.vDiff.v = v
	wdProof, _ := WithdrawProve(cLnNew, cRn, pk, epoch, sender, witness)
	resStr := wdProof.WdProofMarshal()
	fmt.Println(resStr)

	proof, _ := WdProofUnMarshal(resStr)
	gepoch := MapIntoGroup("zether" + epoch.String())
	u := new(bn256.G1).ScalarMult(gepoch, sk)
	fmt.Println(u)
	statement := &WithdrawStatement{cLnNew, cRn, pk, epoch, sender, u}
	res, err := WithdrawVerify(statement, proof)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)

}
