package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	bn256 "github.com/Venachain/Venachain/crypto/bn256/cloudflare"
)

func TestVectorAdd(t *testing.T) {

	a := make([]*big.Int, 3)
	b := make([]*big.Int, 3)
	a[0] = new(big.Int).SetInt64(2)
	a[1] = new(big.Int).SetInt64(1)
	a[2] = new(big.Int).SetInt64(3)
	b[0] = new(big.Int).SetInt64(1)
	b[1] = new(big.Int).SetInt64(4)
	b[2] = new(big.Int).SetInt64(2)

	c, err := VectorAdd(a, b)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(c)

}

func TestVectorSub(t *testing.T) {

	a := make([]*big.Int, 3)
	b := make([]*big.Int, 3)
	a[0] = new(big.Int).SetInt64(2)
	a[1] = new(big.Int).SetInt64(1)
	a[2] = new(big.Int).SetInt64(3)
	b[0] = new(big.Int).SetInt64(1)
	b[1] = new(big.Int).SetInt64(4)
	b[2] = new(big.Int).SetInt64(2)

	c, err := VectorSub(a, b)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(c)

}

func TestVectorNeg(t *testing.T) {

	a := make([]*big.Int, 3)
	a[0] = new(big.Int).SetInt64(2)
	a[1] = new(big.Int).SetInt64(1)
	a[2] = new(big.Int).SetInt64(3)

	c := VectorNeg(a)

	fmt.Println(c)

}

func TestVectorInnerProduct(t *testing.T) {

	a := make([]*big.Int, 4)
	b := make([]*big.Int, 4)
	a[0] = new(big.Int).SetInt64(10)
	a[1] = new(big.Int).SetInt64(22)
	a[2] = new(big.Int).SetInt64(40)
	a[3] = new(big.Int).SetInt64(96)

	b[0] = new(big.Int).SetInt64(3)
	b[1] = new(big.Int).SetInt64(2)
	b[2] = new(big.Int).SetInt64(7)
	b[3] = new(big.Int).SetInt64(9)
	//b[2] = new(big.Int).SetInt64(2)

	c, err := VectorInnerProduct(a, b)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(c)

}

func TestVectorScalarMul(t *testing.T) {

	a := make([]*big.Int, 3)
	a[0] = new(big.Int).SetInt64(2)
	a[1] = new(big.Int).SetInt64(1)
	a[2] = new(big.Int).SetInt64(3)
	b := big.NewInt(2)

	c := VectorScalarMul(a, b)

	fmt.Println(c)

}

func TestVectorHadamard(t *testing.T) {

	a := make([]*big.Int, 3)
	b := make([]*big.Int, 3)
	a[0] = new(big.Int).SetInt64(2)
	a[1] = new(big.Int).SetInt64(1)
	a[2] = new(big.Int).SetInt64(3)
	b[0] = new(big.Int).SetInt64(1)
	b[1] = new(big.Int).SetInt64(4)
	b[2] = new(big.Int).SetInt64(2)

	c, err := VectorHadamard(a, b)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(c)

}

func TestNewPedersenCommitment(t *testing.T) {

	result := new(bn256.G1)
	a := new(bn256.G1).ScalarBaseMult(big.NewInt(2))
	fmt.Println(a)
	result.Add(a, result)
	fmt.Println(result)

}

func TestDecompose(t *testing.T) {

	a := big.NewInt(5)
	result := Decompose(a, 2, 32)
	fmt.Println(result)

}

func TestPowerOf(t *testing.T) {

	a := big.NewInt(3)
	result := PowerOf(a, 4)

	fmt.Println(result)

}

func TestPolyEvaluate(t *testing.T) {

	Fx := make([]*big.Int, 3)
	//Fx[0] = new(big.Int).SetInt64(1)
	//Fx[1] = new(big.Int).SetInt64(1)
	//Fx[2] = new(big.Int).SetInt64(1)
	Fx[0] = new(big.Int).Sub(ORDER, new(big.Int).SetInt64(98))
	Fx[1] = new(big.Int).SetInt64(322)
	Fx[2] = new(big.Int).SetInt64(168)
	result := PolyEvaluate(Fx, new(big.Int).SetInt64(2))
	fmt.Println(result)
}

func TestPolyProduct(t *testing.T) {

	ax := make([]*big.Int, 3)
	bx := make([]*big.Int, 6)
	ax[0] = new(big.Int).SetInt64(13)
	ax[1] = new(big.Int).SetInt64(0)
	ax[2] = new(big.Int).SetInt64(1)
	bx[0] = new(big.Int).SetInt64(5)
	bx[1] = new(big.Int).SetInt64(0)
	bx[2] = new(big.Int).SetInt64(0)
	bx[3] = new(big.Int).SetInt64(13)
	bx[4] = new(big.Int).SetInt64(0)
	bx[5] = new(big.Int).SetInt64(1)
	result := PolyProduct(ax, bx, ORDER)
	fmt.Println(result)
}

func TestMapIntoGroup(t *testing.T) {

	result := MapIntoGroup("venachain" + "g" + string(rune(23)))
	fmt.Println(result)
	Bp()
	x := new(bn256.G1).ScalarMult(bpparam.h, big.NewInt(2))
	y := new(bn256.G1).ScalarMult(bpparam.h, big.NewInt(3))
	z := new(bn256.G1).ScalarMult(bpparam.h, big.NewInt(5))
	r := new(bn256.G1).Add(x, y)
	fmt.Println(r)
	fmt.Println(z)
}

func TestNonceSample(t *testing.T) {
	msg := []byte{'h', 'e'}
	n := big.NewInt(5)
	res := GenerateChallenge(msg, n)
	fmt.Println(res)
}
func TestComputeAR(t *testing.T) {
	aL := make([]*big.Int, 4)
	for i := 0; i < 4; i++ {
		aL[i] = new(big.Int).Add(big.NewInt(int64(i-1)), one)
	}
	aR := ComputeAR(aL)
	fmt.Println(aL)
	fmt.Println(aR)
}
func TestNewVectorCommitment(t *testing.T) {
	Bp()
	a := make([]*big.Int, int(bpparam.n))
	b := make([]*big.Int, int(bpparam.n))
	for i := 0; i < int(bpparam.n); i++ {
		a[i] = one
		b[i] = one
	}
	c := NewVectorCommitment(bpparam, a, b)
	result, _ := c.Commit()
	for i := 0; i < 32; i++ {
		fmt.Println(c.p.gVector[i])
		fmt.Println(c.p.hVector[i])
	}
	fmt.Println(c)
	fmt.Println(result)
}
func TestUpdateP(t *testing.T) {
	Bp()
	x := new(big.Int).SetInt64(1)
	y := new(big.Int).SetInt64(1)
	z := new(big.Int).SetInt64(1)
	//n := new(big.Int).SetInt64(1)
	mu := new(big.Int).SetInt64(1)
	A := new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	S := new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	hprime := make([]*bn256.G1, 2)
	hprime[0] = new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	hprime[1] = new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	g := make([]*bn256.G1, 2)
	g[0] = new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	g[1] = new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	h := new(bn256.G1).ScalarBaseMult(big.NewInt(1))
	res := UpdateP(A, S, g, hprime, h, x, z, y, mu)
	fmt.Println("res:", res)
	g7 := new(bn256.G1).ScalarBaseMult(new(big.Int).SetInt64(7))
	neg1 := new(big.Int).Sub(ORDER, big.NewInt(1))
	neg := new(bn256.G1).ScalarBaseMult(neg1)
	neg3 := new(bn256.G1).ScalarMult(neg, new(big.Int).SetInt64(3))
	g7.Add(g7, neg3)
	fmt.Println("g7+3gneg:", g7)
}
func TestPolyCoefficientsT1(t *testing.T) {
	al := make([]*big.Int, 4)
	al[0] = new(big.Int).SetInt64(1)
	al[1] = new(big.Int).SetInt64(0)
	al[2] = new(big.Int).SetInt64(1)
	al[3] = new(big.Int).SetInt64(1)
	ar := make([]*big.Int, 4)
	ar[0] = new(big.Int).SetInt64(0)
	ar[1] = new(big.Int).SetInt64(-1)
	ar[2] = new(big.Int).SetInt64(0)
	ar[3] = new(big.Int).SetInt64(0)

	sl := make([]*big.Int, 4)
	sl[0] = new(big.Int).SetInt64(2)
	sl[1] = new(big.Int).SetInt64(2)
	sl[2] = new(big.Int).SetInt64(4)
	sl[3] = new(big.Int).SetInt64(5)
	sr := make([]*big.Int, 4)
	sr[0] = new(big.Int).SetInt64(2)
	sr[1] = new(big.Int).SetInt64(3)
	sr[2] = new(big.Int).SetInt64(2)
	sr[3] = new(big.Int).SetInt64(3)
	y := big.NewInt(2)
	z := big.NewInt(2)
	res, _ := PolyCoefficientsT1(sl, sr, al, ar, y, z)
	fmt.Println(res)

}
func TestPolyCoefficientsT2(t *testing.T) {
	sl := make([]*big.Int, 4)
	sl[0] = new(big.Int).SetInt64(2)
	sl[1] = new(big.Int).SetInt64(2)
	sl[2] = new(big.Int).SetInt64(4)
	sl[3] = new(big.Int).SetInt64(5)
	sr := make([]*big.Int, 4)
	sr[0] = new(big.Int).SetInt64(2)
	sr[1] = new(big.Int).SetInt64(3)
	sr[2] = new(big.Int).SetInt64(2)
	sr[3] = new(big.Int).SetInt64(3)
	y := big.NewInt(2)
	//yn := PowerOf(y,2)
	t2 := PolyCoefficientsT2(sl, sr, y)
	fmt.Println(t2)
}
func TestPolyXl(t *testing.T) {
	sl := make([]*big.Int, 4)
	sl[0] = new(big.Int).SetInt64(2)
	sl[1] = new(big.Int).SetInt64(2)
	sl[2] = new(big.Int).SetInt64(4)
	sl[3] = new(big.Int).SetInt64(5)
	al := make([]*big.Int, 4)
	al[0] = new(big.Int).SetInt64(1)
	al[1] = new(big.Int).SetInt64(0)
	al[2] = new(big.Int).SetInt64(1)
	al[3] = new(big.Int).SetInt64(1)
	z := big.NewInt(2)
	x := big.NewInt(2)
	fmt.Println(ComputeLx(al, sl, z, x))
	n2 := PowerOf(big.NewInt(2), 4)
	v, _ := VectorInnerProduct(al, n2)
	fmt.Println("v:", v)
}
func TestPolyXr(t *testing.T) {
	sr := make([]*big.Int, 4)
	sr[0] = new(big.Int).SetInt64(2)
	sr[1] = new(big.Int).SetInt64(3)
	sr[2] = new(big.Int).SetInt64(2)
	sr[3] = new(big.Int).SetInt64(3)
	ar := make([]*big.Int, 4)
	ar[0] = new(big.Int).SetInt64(0)
	ar[1] = new(big.Int).SetInt64(-1)
	ar[2] = new(big.Int).SetInt64(0)
	ar[3] = new(big.Int).SetInt64(0)
	z := big.NewInt(2)
	x := big.NewInt(2)
	y := big.NewInt(2)
	fmt.Println(ComputeRx(ar, sr, z, y, x))
}
func TestDelta(t *testing.T) {
	z := big.NewInt(2)
	y := big.NewInt(2)
	d := ComputeDelta(y, z, 4)
	fmt.Println(d)
	fmt.Println(d.Sub(d, ORDER))
}

func TestIsPowOfTwo(t *testing.T) {
	res := make([]bool, 34)
	var i int64
	for i = -1; i <= 32; i++ {
		res[i+1] = IsPowOfTwo(i)
	}
	std := make([]bool, 34)
	std = []bool{false, true, true, true, false, true, false, false, false, true, false, false, false, false, false, false, false,
		true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true}
	count := 0
	for j := 0; j <= 33; j++ {
		if res[j] == std[j] {
			count++
		}
	}
	if count == 34 {
		t.Log(true)
	} else {
		t.Log(false)
	}
}

func TestVectorInv(t *testing.T) {
	n := 3
	testVector := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		k, _ := randFieldElement(BN256(), rand.Reader)
		testVector[i] = k
		t.Logf("i = %v, k = %v\n", i, k)
	}
	sinv := VectorInv(testVector)
	t.Log(sinv)
	//	i = 0, k = 15400480304429980986620560008943572773487538281699763783454097972744981209303
	//  i = 1, k = 17323668381969791101142459589647960471501840186948185317897842650251264596348
	//  i = 2, k = 5204704643388599313817773844139142503305702121833958376724887578900381127003
	//sinv = [1201138316196803564936085783378793463506992750854498443146472018241387137298 20381054051253203269417959772463687208761418590181151748395233111167477454369 20180336691182748325169995926449263047952702627466549819994374829595534263665]
	value, _ := new(big.Int).SetString("15400480304429980986620560008943572773487538281699763783454097972744981209303", 10)
	value.ModInverse(value, ORDER)
	t.Log(value)
}

//test bigInt = zero, return nil
func TestVectorInv2(t *testing.T) {
	testVector := make([]*big.Int, 2)
	testVector[0] = big.NewInt(0)
	testVector[1] = big.NewInt(1)
	res := VectorInv(testVector)
	fmt.Println(ORDER)
	t.Log(res)
}

func TestVectorScalarMulSum(t *testing.T) {
	aVector := make([]*big.Int, 2)
	aVector[0] = one
	aVector[1] = big.NewInt(10)
	gVector := make([]*bn256.G1, 2)
	key, _ := NewKeyPair(rand.Reader)
	gVector[0] = key.pk
	gVector[1] = key.pk
	res1, _ := VectorScalarMulSum(aVector, gVector)
	res2 := new(bn256.G1).ScalarMult(key.pk, big.NewInt(11))
	t.Log(res1)
	t.Log(res2)
}
