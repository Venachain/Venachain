package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/bn256/cloudflare"
	"math/big"
	"testing"
)

func TestVectorScalarExp(t *testing.T) {
	testVector := make([]*bn256.G1, 4)
	for i := 0; i < 4; i++ {
		key, _ := NewKeyPair(rand.Reader)
		testVector[i] = key.pk
		t.Log(new(bn256.G1).ScalarMult(testVector[i], big.NewInt(2)))
	}
	res := VectorScalarExp(testVector, big.NewInt(2))
	t.Log(res)
}

func TestBij(t *testing.T) {
	var i, j uint64
	for i = 1; i <= 4; i++ {
		for j = 1; j <= 4; j++ {
			res, err := Bij(i, j)
			if err != nil {
				t.Log(err)
			} else {
				t.Logf("i = %v, j = %v, res = %v\n", i, j, res)
			}
		}
	}
}

func TestComputeX(t *testing.T) {
	key, _ := NewKeyPair(rand.Reader)
	res := Generatex(key.pk, key.pk, big.NewInt(3333))
	t.Log(res)
}

func TestProtocol1(t *testing.T) {
	gVector := make([]*bn256.G1, 4)
	hVector := make([]*bn256.G1, 4)
	key, _ := NewKeyPair(rand.Reader)
	instance := new(InnerProductStatement)
	instance.n = 4
	instance.gVector = gVector
	instance.hVector = hVector
	instance.p = key.pk
	instance.u = key.pk
	c := big.NewInt(4)
	previousCh := big.NewInt(2)

	h := sha256.Sum256(previousCh.Bytes())
	ch := hashToInt(h[:], BN256()) //JP o=H(previousChallenge)
	ch.Mod(ch, ORDER)
	t.Logf("the challenge is %v\n", ch)

	statement, _ := Protocol1(instance, c, previousCh)
	t.Log(statement.p)

	result := new(bn256.G1).ScalarMult(key.pk, new(big.Int).Add(one, new(big.Int).Mul(ch, c)))
	t.Log(result)
}

func TestInnerProductProver(t *testing.T) {
	//test n == 1 proof and Verify
	instance := new(InnerProductStatement)
	witness := new(InnerProductWitness)
	l := sampleRandomVector(1)
	r := sampleRandomVector(1)
	witness.as = l
	witness.bs = r
	tHat, err := VectorInnerProduct(l, r)
	if err != nil {
		t.Log(err)
	}

	g := make([]*bn256.G1, 1)
	g[0] = new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	h := make([]*bn256.G1, 1)
	h[0] = new(bn256.G1).ScalarBaseMult(big.NewInt(3))
	key, _ := NewKeyPair(rand.Reader)
	u := key.pk
	ga := new(bn256.G1).ScalarMult(g[0], l[0])
	hb := new(bn256.G1).ScalarMult(h[0], r[0])
	P := new(bn256.G1).Add(ga, hb)
	instance.n = 1
	instance.gVector = g
	instance.hVector = h
	instance.p = P
	instance.u = u

	ch := sha256.Sum256([]byte{123})
	challenge := hashToInt(ch[:], BN256())
	proof, err := InnerProductProver(instance, witness, tHat, challenge)
	if err != nil {
		t.Log(err)
	}

	res, _ := InnerProductVerifier(instance, proof, tHat, challenge)
	t.Log(res)

}

func TestInnerProductProver2(t *testing.T) {
	//test n = 2, innerproductproof and verifier
	//1. setup witness
	var n int64 = 32
	witness := new(InnerProductWitness)
	l := make([]*big.Int, n)
	r := make([]*big.Int, n)
	for i := 0; int64(i) < n; i++ {
		l[i] = new(big.Int).Set(big.NewInt(2))
		r[i] = new(big.Int).Set(big.NewInt(3))
	}
	witness.as = l
	witness.bs = r
	tHat, err := VectorInnerProduct(l, r)
	if err != nil {
		t.Log(err)
	}

	//2. setup instance
	instance := new(InnerProductStatement)
	g := make([]*bn256.G1, n)
	h := make([]*bn256.G1, n)
	u := new(bn256.G1).ScalarBaseMult(big.NewInt(2))
	var i int64
	for i = 0; i < n; i++ {
		g[i] = new(bn256.G1).ScalarBaseMult(big.NewInt(i + 1))
		h[i] = new(bn256.G1).ScalarBaseMult(big.NewInt(i + 1))
	}
	//compute ga = g*a; hb = h * b
	ga, err := VectorScalarMulSum(l, g)
	if err != nil {
		t.Logf("generate ga failed %v\n", err)
	}
	hb, err := VectorScalarMulSum(r, h)
	if err != nil {
		t.Logf("generate hb failed %v\n", err)
	}
	P := new(bn256.G1).Add(ga, hb)

	instance.n = n
	instance.gVector = g
	instance.hVector = h
	instance.p = P
	instance.u = u

	//3. setup challenge
	//ch := sha256.Sum256([]byte{123})
	//challenge := hashToInt(ch[:],BN256())
	challenge := big.NewInt(3)

	//4. generate inner product proof
	proof, err := InnerProductProver(instance, witness, tHat, challenge)
	if err != nil {
		t.Log(err)
	}
	t.Log(proof)

	//5. verify inner product proof
	res, _ := InnerProductVerifier(instance, proof, tHat, challenge)
	t.Log(res)
}

func TestInnerProductProver3(t *testing.T) {
	//test n=32; using bpparam
	//1. setup witness
	param := Bp()
	n := param.n
	witness := new(InnerProductWitness)
	l := sampleRandomVector(n)
	r := sampleRandomVector(n)
	witness.as = l
	witness.bs = r

	//2. setup instance: g, h, u, tHat, P
	instance := new(InnerProductStatement)
	tHat, _ := VectorInnerProduct(l, r)
	g := param.gVector
	h := param.gVector
	k, _ := randFieldElement(BN256(), rand.Reader)
	u := new(bn256.G1).ScalarBaseMult(k)
	//2.1. compute ga = g*a; hb = h * b; and P = a*g+b*h
	ga, _ := VectorScalarMulSum(l, g)
	hb, _ := VectorScalarMulSum(r, h)
	P := new(bn256.G1).Add(ga, hb)

	instance.n = n
	instance.gVector = g
	instance.hVector = h
	instance.p = P
	instance.u = u

	//3. setup challenge
	challenge, _ := randFieldElement(BN256(), rand.Reader)

	//4. generate innerproductproof
	proof, err := InnerProductProver(instance, witness, tHat, challenge)
	if err != nil {
		t.Log(err)
	}
	t.Logf("proof.a = %v\n", proof.a)
	t.Logf("proof.b = %v\n", proof.b)
	t.Logf("proof.LS = %v\n", proof.LS)
	t.Logf("proof.RS = %v\n", proof.RS)

	//5. verify inner product proof
	res, _ := InnerProductVerifier(instance, proof, tHat, challenge)
	t.Log(res)
}
