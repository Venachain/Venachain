package crypto

import (
	"fmt"
	bn256 "github.com/PlatONEnetwork/PlatONE-Go/crypto/bn256/cloudflare"
	"math/big"
	"testing"
)

func TestVectorCommitment_Commit(t *testing.T) {

}

func TestVectorDecompose(t *testing.T) {
	v := make([]*big.Int, 3)
	v[0] = big.NewInt(14)
	v[1] = big.NewInt(5)
	v[2] = big.NewInt(3)
	res := VectorDecompose(v, 2, 8, 3)
	t.Log(res)
}

func TestLPolyCoeff(t *testing.T) {
	var i, n, m int64
	n = 3
	m = 2
	aL := make([]*big.Int, n*m)
	for i = 1; i <= 6; i++ {
		aL[i-1] = big.NewInt(i + 1)
	}
	z := big.NewInt(3)
	l0, _ := LPolyCoeff(aL, z, n*m)
	t.Log(l0) //[21888242871839275222246405745257275088548364400416034343698204186575808495616 0 1 2 3 4]
}

func TestRpolyCoeff(t *testing.T) {
	var i, n, m int64
	n = 3
	m = 2
	aR := make([]*big.Int, n*m)
	sR := make([]*big.Int, n*m)
	for i = 1; i <= 6; i++ {
		aR[i-1] = big.NewInt(i)
		sR[i-1] = big.NewInt(i * i)
	}
	y := big.NewInt(2)
	z := big.NewInt(3)
	r0, r1, _ := RpolyCoeff(aR, sR, y, z, n, m)
	t.Log(r0) //[13 28 60 83 182 396]
	t.Log(r1) //[1 8 36 128 400 1152]
}

func TestComputet0(t *testing.T) {
	var i, n, m int64
	n = 3
	m = 2
	aL := make([]*big.Int, n*m)
	aR := make([]*big.Int, n*m)
	sR := make([]*big.Int, n*m)
	for i = 1; i <= 6; i++ {
		aL[i-1] = big.NewInt(i + 1)
		aR[i-1] = big.NewInt(i)
		sR[i-1] = big.NewInt(i * i)
	}
	y := big.NewInt(2)
	z := big.NewInt(3)
	l0, _ := LPolyCoeff(aL, z, n*m)
	r0, _, _ := RpolyCoeff(aR, sR, y, z, n, m)
	res, _ := VectorInnerProduct(l0, r0)
	t0, _ := Computet0(aL, aR, sR, y, z, n, m)
	t.Log(t0) //2343
	if t0.Cmp(res) == 0 {
		t.Log(true)
	}
}

func TestComputet1(t *testing.T) {
	var i, n, m int64
	n = 3
	m = 2
	aL := make([]*big.Int, n*m)
	aR := make([]*big.Int, n*m)
	sL := make([]*big.Int, n*m)
	sR := make([]*big.Int, n*m)
	for i = 1; i <= 6; i++ {
		aL[i-1] = big.NewInt(i + 1)
		aR[i-1] = big.NewInt(i)
		sL[i-1] = big.NewInt(i)
		sR[i-1] = big.NewInt(i * i)
	}
	y := big.NewInt(2)
	z := big.NewInt(3)
	l0, _ := LPolyCoeff(aL, z, n*m)
	r0, r1, _ := RpolyCoeff(aR, sR, y, z, n, m)
	left, _ := VectorInnerProduct(l0, r1)
	right, _ := VectorInnerProduct(sL, r0)
	res := new(big.Int).Add(left, right)
	t1, _ := Computet1(aL, aR, sL, sR, y, z, n, m)
	t.Log(t1) //9966
	if t1.Cmp(res) == 0 {
		t.Log(true)
	}
}

func TestComputet2(t *testing.T) {
	var i, n, m int64
	n = 3
	m = 2
	aR := make([]*big.Int, n*m)
	sL := make([]*big.Int, n*m)
	sR := make([]*big.Int, n*m)
	for i = 1; i <= 6; i++ {
		aR[i-1] = big.NewInt(i)
		sL[i-1] = big.NewInt(i)
		sR[i-1] = big.NewInt(i * i)
	}
	y := big.NewInt(2)
	z := big.NewInt(3)
	_, r1, _ := RpolyCoeff(aR, sR, y, z, n, m)

	res, _ := VectorInnerProduct(sL, r1)
	t2, _ := Computet2(sL, aR, sR, y, z, n, m)
	t.Log(t2) //9549
	if t2.Cmp(res) == 0 {
		t.Log(true)
	}
}

func TestComputeAggLx(t *testing.T) {
	var i, n, m int64
	n = 3
	m = 2
	aL := make([]*big.Int, n*m)
	sL := make([]*big.Int, n*m)
	for i = 1; i <= 6; i++ {
		aL[i-1] = big.NewInt(i + 1)
		sL[i-1] = big.NewInt(i)
	}
	x := big.NewInt(2)
	z := big.NewInt(3)
	lx, _ := ComputeAggLx(aL, sL, x, z, n*m)
	t.Log(lx) //[1 4 7 10 13 16]
}

func TestComputeP(t *testing.T) {
	var i, j, n, m int64
	n = 3
	m = 2
	gVector := make([]*bn256.G1, n*m)
	hPrime := make([]*bn256.G1, n*m)
	for i = 1; i <= 3; i++ {
		gVector[i-1] = new(bn256.G1).ScalarBaseMult(big.NewInt(2))
		hPrime[i-1] = new(bn256.G1).ScalarBaseMult(big.NewInt(3))
	}
	for j = 4; j <= 6; j++ {
		gVector[j-1] = new(bn256.G1).ScalarBaseMult(big.NewInt(3))
		hPrime[j-1] = new(bn256.G1).ScalarBaseMult(big.NewInt(2))
	}
	A := new(bn256.G1).ScalarBaseMult(big.NewInt(7))
	S := new(bn256.G1).ScalarBaseMult(big.NewInt(5))
	x := big.NewInt(3)
	y := big.NewInt(2)
	z := big.NewInt(2)
	h := new(bn256.G1).ScalarBaseMult(big.NewInt(2))
	mu := big.NewInt(11)
	res, _ := UpdateAggP(A, S, h, gVector, hPrime, x, y, z, mu, n, m)
	t.Log(res)
	t.Log(new(bn256.G1).ScalarBaseMult(big.NewInt(432)))
}

func TestAggBpProve(t *testing.T) {
	aggBp := new(AggBulletProof)
	t.Log(aggBp)
}

func TestAggBulletProof_AggBpVerify(t *testing.T) {
	AggBp()
	v := make([]*big.Int, 2)
	v[0] = big.NewInt(3)
	v[1] = big.NewInt(7)
	values := AggBpWitness{v}
	param := aggbpparam
	proof, _ := AggBpProve(&param, &values)
	fmt.Println("proof:", proof)
	res, _ := AggBpVerify(proof, &param)
	fmt.Println("res:", res)
}
func TestAggBulletProof_AggBpVerify_s(t *testing.T) {
	AggBp()
	v := make([]*big.Int, 2)
	v[0] = big.NewInt(3)
	v[1] = big.NewInt(16)
	//values := AggBpWitness{v}
	param := aggbpparam
	proof, _ := AggBpProve_s(&param, v)
	fmt.Println("proof:", proof)
	//p := "0x12324314"
	res, _ := AggBpVerify_s(proof, &param)

	fmt.Println("res:", res)
}