package crypto

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/Venachain/Venachain/crypto/bn256"
	"github.com/Venachain/Venachain/rlp"
)

func TestTfStatmentMarshal(t *testing.T) {
	tfMarshal := new(TfSMarshal)
	var n int64 = 4
	AnonPk := make([]*bn256.G1, n)
	CLnNew := make([]*bn256.G1, n)
	CRnNew := make([]*bn256.G1, n)
	CVector := make([]*bn256.G1, n)
	D := new(bn256.G1).ScalarMult(G, big.NewInt(10))
	NonceU := new(bn256.G1).ScalarMult(G, big.NewInt(2))
	epoch := new(big.Int).Set(big.NewInt(200))
	var i int64
	for i = 0; i < n; i++ {
		key, _ := NewKeyPair(rand.Reader)
		AnonPk[i] = key.pk
		cipher, _ := Enc(rand.Reader, AnonPk[i], big.NewInt(i+1))
		CLnNew[i] = cipher.C
		CRnNew[i] = cipher.D
		CVector[i] = new(bn256.G1).ScalarMult(G, big.NewInt(i))
	}
	tfMarshal.AnonPk = WdPointMarshal(AnonPk)
	tfMarshal.CLnNew = WdPointMarshal(CLnNew)
	tfMarshal.CRnNew = WdPointMarshal(CRnNew)
	tfMarshal.CVector = WdPointMarshal(CVector)
	tfMarshal.D = D.Marshal()
	tfMarshal.NonceU = NonceU.Marshal()
	tfMarshal.Epoch = epoch.Bytes()
	t.Log(tfMarshal)
	res, _ := rlp.EncodeToBytes(tfMarshal)
	out := new(TfSMarshal)
	rlp.DecodeBytes(res, out)
	t.Log(res)
	t.Log(out)
}

func TestCommitToBits(t *testing.T) {
	//in our system m must be less than 16
	var m int64 = 2
	tfWit := new(TransferWitness)
	//l0 and l1 must be even and odd
	tfWit.l0 = big.NewInt(2)
	tfWit.l1 = big.NewInt(3)
	agBp := AggBp()
	A, B, _, _, a, b, _ := commitToBits(m, tfWit, &agBp)
	t.Log(A)
	t.Log(B)
	t.Log(len(a))
	t.Log(b)
}

func TestTfPolyCoeffF0(t *testing.T) {
	var m int64 = 8
	l0 := big.NewInt(15)
	a := sampleRandomVector(m)
	l0Bits := Decompose(l0, 2, m)
	res, _ := TfPolyCoeffF0(a, l0Bits, m)
	var i int64
	for i = 0; i < m; i++ {
		t.Log(a[i])
		t.Log(res[i])
	}
}

func TestTfPolyCoeffF1(t *testing.T) {
	var m int64 = 8
	l0 := big.NewInt(15)
	a := sampleRandomVector(m)
	l0Bits := Decompose(l0, 2, m)
	res, _ := TfPolyCoeffF1(a, l0Bits, m)
	var i int64
	for i = 0; i < m; i++ {
		t.Log(a[i])
		t.Log(res[i])
	}
}

func TestTfPolyP(t *testing.T) {
	var m int64 = 3
	var N int64 = 8
	l0 := big.NewInt(5)
	//a := sampleRandomVector(m)
	a := make([]*big.Int, m)
	a[0] = big.NewInt(5)
	a[1] = big.NewInt(3)
	a[2] = big.NewInt(2)
	l0Bits := Decompose(l0, 2, m)
	F0k0Coeff, _ := TfPolyCoeffF0(a, l0Bits, m)
	F0k1Coeff, _ := TfPolyCoeffF1(a, l0Bits, m)
	polyP0 := TfPolyP(F0k0Coeff, F0k1Coeff, N, m)
	var i int64
	for i = 0; i < N; i++ {
		t.Log(polyP0[i])
	}
	fmt.Println(ORDER)
}

func TestTfSampleRandomVector(t *testing.T) {
	var m int64 = 2
	phi, chi, psi, omega := TfSampleRandomVector(m)
	t.Log(phi)
	t.Log(chi)
	t.Log(psi)
	t.Log(omega)
}

func TestMultiVectorT(t *testing.T) {
	var m int64 = 2
	var N int64 = 4
	l0 := big.NewInt(2)
	//a := sampleRandomVector(m)
	a := make([]*big.Int, m)
	a[0] = big.NewInt(5)
	a[1] = big.NewInt(11)
	l0Bits := Decompose(l0, 2, m)
	F0k0Coeff, _ := TfPolyCoeffF0(a, l0Bits, m)
	F0k1Coeff, _ := TfPolyCoeffF1(a, l0Bits, m)
	polyP0 := TfPolyP(F0k0Coeff, F0k1Coeff, N, m)
	t.Log(polyP0)
	res := MultiVectorT(polyP0)
	t.Log(res)
}

func TestVectorCommitMuliExp(t *testing.T) {
	var m int64 = 2
	var N int64 = 4
	l0 := big.NewInt(2)
	//a := sampleRandomVector(m)
	a := make([]*big.Int, m)
	a[0] = big.NewInt(5)
	a[1] = big.NewInt(11)
	l0Bits := Decompose(l0, 2, m)
	F0k0Coeff, _ := TfPolyCoeffF0(a, l0Bits, m)
	F0k1Coeff, _ := TfPolyCoeffF1(a, l0Bits, m)
	polyP0 := TfPolyP(F0k0Coeff, F0k1Coeff, N, m)
	cipher := make([]*bn256.G1, N)
	for i := int64(0); i < N; i++ {
		cipher[i] = G
	}
	publ0 := new(bn256.G1).ScalarMult(G, big.NewInt(2))
	phi := make([]*big.Int, m)
	phi[0] = big.NewInt(3)
	phi[1] = big.NewInt(2)
	res, _ := VectorCommitMuliExp(polyP0, cipher, m, publ0, phi)
	t.Log(res)
	test1 := new(bn256.G1).ScalarMult(G, big.NewInt(6))
	t.Log(test1)
	test2 := new(bn256.G1).ScalarMult(G, big.NewInt(4))
	t.Log(test2)
}

func TestCXtilde(t *testing.T) {
	v := big.NewInt(3)
	btf := big.NewInt(10)
	var m int64 = 3
	var N int64 = 8
	l0 := big.NewInt(2)
	//a := sampleRandomVector(m)
	a0 := make([]*big.Int, m)
	a0[0] = big.NewInt(5)
	a0[1] = big.NewInt(11)
	a0[2] = big.NewInt(9)
	l0Bits := Decompose(l0, 2, m)
	F0k0Coeff, _ := TfPolyCoeffF0(a0, l0Bits, m)
	F0k1Coeff, _ := TfPolyCoeffF1(a0, l0Bits, m)
	polyP0 := TfPolyP(F0k0Coeff, F0k1Coeff, N, m)

	l1 := big.NewInt(3)
	a1 := make([]*big.Int, m)
	a1[0] = big.NewInt(10)
	a1[1] = big.NewInt(9)
	a1[2] = big.NewInt(9)
	l1Bits := Decompose(l1, 2, m)
	F1k0Coeff, _ := TfPolyCoeffF0(a1, l1Bits, m)
	F1k1Coeff, _ := TfPolyCoeffF1(a1, l1Bits, m)
	polyP1 := TfPolyP(F1k0Coeff, F1k1Coeff, N, m)
	omega := make([]*big.Int, m)
	omega[0] = big.NewInt(10)
	omega[1] = big.NewInt(4)
	omega[2] = big.NewInt(4)
	D := new(bn256.G1).ScalarMult(G, big.NewInt(3))
	res, _ := CXtilde(v, btf, polyP0, polyP1, N, m, omega, D, l0.Int64(), l1.Int64())
	t.Log(res)
	//test1 := new(bn256.G1).ScalarMult(G, big.NewInt(-16570))
	//t.Log(test1)
}

func TestTFf(t *testing.T) {
	var m int64 = 2
	l0 := big.NewInt(2)
	//a := sampleRandomVector(m)
	a := make([]*big.Int, m)
	a[0] = big.NewInt(5)
	a[1] = big.NewInt(11)
	l0Bits := Decompose(l0, 2, m)
	F0k1Coeff, _ := TfPolyCoeffF1(a, l0Bits, m)
	omega := big.NewInt(2)
	res := TFf(F0k1Coeff, omega)
	for i := int64(0); i < m; i++ {
		t.Log(F0k1Coeff[i])
		t.Log(res[i])
	}
}

func TestVectorLine(t *testing.T) {
	m := int64(2)
	point := new(bn256.G1).ScalarMult(G, big.NewInt(10))
	pub := new(bn256.G1).Set(G)
	phi := make([]*big.Int, m)
	phi[0] = big.NewInt(5)
	phi[1] = big.NewInt(11)
	w := big.NewInt(3)
	res, _ := VectorLine(point, pub, phi, w, m)
	t.Log(res)
	testRes := new(bn256.G1).ScalarMult(G, big.NewInt(52))
	t.Log(testRes)
}

func TestVectorShift(t *testing.T) {
	m := int64(7)
	shift := int64(6)
	vector := make([]*big.Int, m)
	for i := int64(0); i < m; i++ {
		vector[i] = big.NewInt(2*i + 1)
	}
	t.Log(vector)
	res := VectorShift(vector, shift)
	t.Log(res)
}

func TestVectorPolyEvaluate(t *testing.T) {
	var m int64 = 2
	var N int64 = 4
	l0 := big.NewInt(2)
	//a := sampleRandomVector(m)
	a := make([]*big.Int, m)
	a[0] = big.NewInt(5)
	a[1] = big.NewInt(11)
	l0Bits := Decompose(l0, 2, m)
	F0k0Coeff, _ := TfPolyCoeffF0(a, l0Bits, m)
	F0k1Coeff, _ := TfPolyCoeffF1(a, l0Bits, m)
	PolyCoeff := TfPolyP(F0k0Coeff, F0k1Coeff, N, m)
	t.Log(PolyCoeff)
	x := big.NewInt(2)
	res := VectorPolyEvaluate(PolyCoeff, x)
	t.Log(res)
}

func TestTfyXline(t *testing.T) {
	var m int64 = 2
	var N int64 = 4
	l0 := big.NewInt(2)
	//a := sampleRandomVector(m)
	a0 := make([]*big.Int, m)
	a0[0] = big.NewInt(5)
	a0[1] = big.NewInt(11)
	l0Bits := Decompose(l0, 2, m)
	F0k0Coeff, _ := TfPolyCoeffF0(a0, l0Bits, m)
	F0k1Coeff, _ := TfPolyCoeffF1(a0, l0Bits, m)
	polyP0Coeff := TfPolyP(F0k0Coeff, F0k1Coeff, N, m)

	l1 := big.NewInt(3)
	a1 := make([]*big.Int, m)
	a1[0] = big.NewInt(10)
	a1[1] = big.NewInt(9)
	l1Bits := Decompose(l1, 2, m)
	F1k0Coeff, _ := TfPolyCoeffF0(a1, l1Bits, m)
	F1k1Coeff, _ := TfPolyCoeffF1(a1, l1Bits, m)
	polyP1Coeff := TfPolyP(F1k0Coeff, F1k1Coeff, N, m)

	omega := make([]*big.Int, m)
	omega[0] = big.NewInt(10)
	omega[1] = big.NewInt(4)

	AnonPk := make([]*bn256.G1, N)
	AnonPk[0] = new(bn256.G1).ScalarMult(G, big.NewInt(11))
	AnonPk[1] = new(bn256.G1).ScalarMult(G, big.NewInt(2))
	AnonPk[2] = new(bn256.G1).ScalarMult(G, big.NewInt(3))
	AnonPk[3] = new(bn256.G1).ScalarMult(G, big.NewInt(5))
	v := big.NewInt(2)
	challOmega := big.NewInt(2)
	res, _ := TfyXline(polyP0Coeff, polyP1Coeff, N, m, AnonPk, v, challOmega, omega)
	t.Log(res)
	testRes := new(bn256.G1).ScalarMult(G, big.NewInt(-4267))
	t.Log(testRes)
}

func TestVerifierF0Coeff(t *testing.T) {
	m := 10
	f := make([]*big.Int, m)
	for i := 0; i < m; i++ {
		//f[i], _ = rand.Int(rand.Reader, ORDER)
		f[i] = big.NewInt(int64(i + 20))
	}
	//t.Log(f)
	omega := big.NewInt(10)
	res := VerifierF0Coeff(f, omega)
	t.Log(res)
}

func TestVerifierPolyCoeff(t *testing.T) {
	m := 3
	N := 8
	flk1 := make([]*big.Int, m)
	for i := 0; i < m; i++ {
		//f[i], _ = rand.Int(rand.Reader, ORDER)
		flk1[i] = big.NewInt(int64(i + 10))
	}
	omega := big.NewInt(20)
	flk0 := VerifierF0Coeff(flk1, omega)
	poly := VerifierPolyCoeff(flk0, flk1, int64(N), int64(m))
	t.Log(poly) //[720 720 880 880 1080 1080 1320 1320]
}

/*
func TestVerifierCommit(t *testing.T) {
	//var m int64 = 2
	//var N int64 = 4
	//tfWit := new(TransferWitness)
	////l0 and l1 must be even and odd
	//tfWit.l0 = big.NewInt(2)
	//tfWit.l1 = big.NewInt(3)
	//agBp := AggBp()
	//A, B, rA, rB, a, b, _ := commitToBits(m, tfWit, &agBp)
	//F0k1, _ := TfPolyCoeffF1(a[:m], b[:m], m)
	//F1k1, _ := TfPolyCoeffF1(a[m:], b[m:], m)
	//
	//omega := big.NewInt(20)
	//f0k := VectorPolyEvaluate(F0k1[:][:], omega)
	//f1k := VectorPolyEvaluate(F1k1, omega)
	//f := append(f0k, f1k...)
	//zA := new(big.Int).Add(new(big.Int).Mul(rB, omega), rA)
	//zA.Mod(zA, ORDER)
	//instance := AggBp()
	//VerifierCommit(f, A, B, m, zA, omega, &instance)
}
*/

func TestVerifierVectorTilde(t *testing.T) {
	m := int64(4)
	vectorTilde := make([]*bn256.G1, m)
	for i := int64(0); i < m; i++ {
		vectorTilde[i] = new(bn256.G1).ScalarMult(G, big.NewInt(i+1))
	}
	omega := big.NewInt(3)
	res, _ := VerifierVectorTilde(vectorTilde, omega)
	t.Log(res)
	testRes := new(bn256.G1).ScalarMult(G, big.NewInt(-142))
	t.Log(testRes)
}

func TestVerifierVectorLine(t *testing.T) {
	m := int64(3)
	vectorTilde := make([]*bn256.G1, m)
	for i := int64(0); i < m; i++ {
		vectorTilde[i] = new(bn256.G1).ScalarMult(G, big.NewInt(i+1))
	}
	omega := big.NewInt(3)

	N := int64(8)
	vectorPoint := make([]*bn256.G1, N)
	coeff := make([]*big.Int, N)
	for j := int64(0); j < N; j++ {
		vectorPoint[j] = new(bn256.G1).ScalarMult(G, big.NewInt(j+1))
		coeff[j] = new(big.Int).Mul(big.NewInt(2), big.NewInt(j))
	}
	res, _ := VerifierVectorLine(vectorPoint, vectorTilde, coeff, omega)
	t.Log(res)
	testRes := new(bn256.G1).ScalarMult(G, big.NewInt(302))
	t.Log(testRes)
}

func TestGenerateKexi(t *testing.T) {
	v := big.NewInt(3)
	N := int64(8)
	res := GenerateKexi(v, N)
	t.Log(res)
}

func TestVectorMuliExpShift(t *testing.T) {
	N := int64(4)
	vectorPoint := make([]*bn256.G1, N)
	polyP0 := make([]*big.Int, N)
	polyP1 := make([]*big.Int, N)
	for j := int64(0); j < N; j++ {
		vectorPoint[j] = new(bn256.G1).ScalarMult(G, big.NewInt(j+1))
		polyP0[j] = new(big.Int).Mul(big.NewInt(2), big.NewInt(j))
		polyP1[j] = new(big.Int).Mul(big.NewInt(1), big.NewInt(j+1))
	}
	v := big.NewInt(3)
	kexi := GenerateKexi(v, N)
	res, _ := VectorMuliExpShift(vectorPoint, kexi, polyP0, polyP1, N)
	t.Log(res)
	testRes := new(bn256.G1).ScalarMult(G, big.NewInt(340))
	t.Log(testRes)
}

func TestComputeAt(t *testing.T) {
	aggBp := AggBp()
	c := big.NewInt(10)
	omega := big.NewInt(2)
	tHat := big.NewInt(8)
	stau := big.NewInt(4)
	delta := big.NewInt(20)
	sb := big.NewInt(11)
	x := big.NewInt(5)
	m := int64(4)
	T1 := new(bn256.G1).ScalarMult(G, big.NewInt(10))
	T2 := new(bn256.G1).ScalarMult(G, big.NewInt(2))
	res := ComputeAt(&aggBp, c, omega, tHat, stau, delta, sb, x, m, T1, T2)
	t.Log(res)
	testRes := new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, big.NewInt(-17931)), new(bn256.G1).ScalarMult(aggBp.bpParam.h, stau))
	t.Log(testRes)
}

func TestTransferProve(t *testing.T) {
	N := int64(512)
	tfWit := new(TransferWitness)
	tfWit.sk, _ = rand.Int(rand.Reader, ORDER)
	tfWit.r, _ = rand.Int(rand.Reader, ORDER)
	tfWit.l0 = big.NewInt(3)
	tfWit.l1 = big.NewInt(2)
	tfWit.bTf = big.NewInt(10)
	tfWit.bDiff = big.NewInt(10)

	tfStatement := new(TransferStatement)
	AnonPk := make([]*bn256.G1, N)
	CLnNew := make([]*bn256.G1, N)
	CRnNew := make([]*bn256.G1, N)
	CVector := make([]*bn256.G1, N)
	D := new(bn256.G1).ScalarMult(G, tfWit.r)
	epoch := big.NewInt(200)
	Gepoch := MapIntoGroup("zether" + epoch.String())
	NonceU := new(bn256.G1).ScalarMult(Gepoch, tfWit.sk)

	for i := int64(0); i < N; i++ {
		if tfWit.l0.Cmp(big.NewInt(i)) == 0 {
			AnonPk[i] = new(bn256.G1).ScalarMult(G, tfWit.sk)
		} else {
			AnonPk[i] = new(bn256.G1).ScalarMult(G, big.NewInt(i*i+2))
		}
	}

	for i := int64(0); i < N; i++ {
		if tfWit.l0.Cmp(big.NewInt(i)) == 0 {
			CVector[i] = new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, new(big.Int).Neg(tfWit.bTf)), new(bn256.G1).ScalarMult(AnonPk[i], tfWit.r))
		}
		if tfWit.l1.Cmp(big.NewInt(i)) == 0 {
			CVector[i] = new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, tfWit.bTf), new(bn256.G1).ScalarMult(AnonPk[i], tfWit.r))
		}
		if tfWit.l0.Cmp(big.NewInt(i)) != 0 && tfWit.l1.Cmp(big.NewInt(i)) != 0 {
			CVector[i] = new(bn256.G1).ScalarMult(AnonPk[i], tfWit.r)
		}
	}

	for i := int64(0); i < N; i++ {
		if tfWit.l0.Cmp(big.NewInt(i)) == 0 {
			cipher, _ := Enc(rand.Reader, AnonPk[i], tfWit.bDiff)
			CLnNew[i] = cipher.C
			CRnNew[i] = cipher.D
		} else {
			ciphertext, _ := Enc(rand.Reader, AnonPk[i], big.NewInt(i+10))
			CLnNew[i] = ciphertext.C
			CRnNew[i] = ciphertext.D
		}
	}

	tfStatement = &TransferStatement{AnonPk, CLnNew, CRnNew, CVector, D, NonceU, epoch}
	tfProof, _ := TransferProve(tfStatement, tfWit)
	res, _ := TransferVerify(tfStatement, tfProof)
	t.Log(res)
}

func TestTfProver(t *testing.T) {
	N := int64(512)
	key, _ := NewKeyPair(rand.Reader)
	sk := key.sk
	r, _ := rand.Int(rand.Reader, ORDER)
	l0 := big.NewInt(100)
	l1 := big.NewInt(77)
	bTf := big.NewInt(10)
	bDiff := big.NewInt(10)

	AnonPk := make([]*bn256.G1, N)
	CLnNew := make([]*bn256.G1, N)
	CRnNew := make([]*bn256.G1, N)
	CVector := make([]*bn256.G1, N)
	D := new(bn256.G1).ScalarMult(G, r)
	epoch := big.NewInt(200)
	Gepoch := MapIntoGroup("zether" + epoch.String())
	NonceU := new(bn256.G1).ScalarMult(Gepoch, sk)

	for i := int64(0); i < N; i++ {
		if l0.Cmp(big.NewInt(i)) == 0 {
			AnonPk[i] = new(bn256.G1).ScalarMult(G, sk)
		} else {
			AnonPk[i] = new(bn256.G1).ScalarMult(G, big.NewInt(i*i+2))
		}
	}

	for i := int64(0); i < N; i++ {
		if l0.Cmp(big.NewInt(i)) == 0 {
			CVector[i] = new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, new(big.Int).Neg(bTf)), new(bn256.G1).ScalarMult(AnonPk[i], r))
		}
		if l1.Cmp(big.NewInt(i)) == 0 {
			CVector[i] = new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, bTf), new(bn256.G1).ScalarMult(AnonPk[i], r))
		}
		if l0.Cmp(big.NewInt(i)) != 0 && l1.Cmp(big.NewInt(i)) != 0 {
			CVector[i] = new(bn256.G1).ScalarMult(AnonPk[i], r)
		}
	}

	for i := int64(0); i < N; i++ {
		if l0.Cmp(big.NewInt(i)) == 0 {
			cipher, _ := Enc(rand.Reader, AnonPk[i], bDiff)
			CLnNew[i] = cipher.C
			CRnNew[i] = cipher.D
		} else {
			ciphertext, _ := Enc(rand.Reader, AnonPk[i], big.NewInt(i+10))
			CLnNew[i] = ciphertext.C
			CRnNew[i] = ciphertext.D
		}
	}

	proof, err := TfProver(AnonPk, CLnNew, CRnNew, CVector, D, NonceU, epoch, sk, bTf, bDiff, r, l0, l1)
	if err != nil {
		t.Log(err)
	}
	res, err := TfVerifier(AnonPk, CLnNew, CRnNew, CVector, D, NonceU, epoch, proof)
	t.Log(res)
}

func TestTfProofMarshal(t *testing.T) {
	N := int64(16)
	key, _ := NewKeyPair(rand.Reader)
	sk := key.sk
	r, _ := rand.Int(rand.Reader, ORDER)
	l0 := big.NewInt(8)
	l1 := big.NewInt(11)
	bTf := big.NewInt(10)
	bDiff := big.NewInt(10)

	AnonPk := make([]*bn256.G1, N)
	CLnNew := make([]*bn256.G1, N)
	CRnNew := make([]*bn256.G1, N)
	CVector := make([]*bn256.G1, N)
	D := new(bn256.G1).ScalarMult(G, r)
	epoch := big.NewInt(200)
	Gepoch := MapIntoGroup("zether" + epoch.String())
	NonceU := new(bn256.G1).ScalarMult(Gepoch, sk)

	for i := int64(0); i < N; i++ {
		if l0.Cmp(big.NewInt(i)) == 0 {
			AnonPk[i] = new(bn256.G1).ScalarMult(G, sk)
		} else {
			AnonPk[i] = new(bn256.G1).ScalarMult(G, big.NewInt(i*i+2))
		}
	}

	for i := int64(0); i < N; i++ {
		if l0.Cmp(big.NewInt(i)) == 0 {
			CVector[i] = new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, new(big.Int).Neg(bTf)), new(bn256.G1).ScalarMult(AnonPk[i], r))
		}
		if l1.Cmp(big.NewInt(i)) == 0 {
			CVector[i] = new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, bTf), new(bn256.G1).ScalarMult(AnonPk[i], r))
		}
		if l0.Cmp(big.NewInt(i)) != 0 && l1.Cmp(big.NewInt(i)) != 0 {
			CVector[i] = new(bn256.G1).ScalarMult(AnonPk[i], r)
		}
	}

	for i := int64(0); i < N; i++ {
		if l0.Cmp(big.NewInt(i)) == 0 {
			cipher, _ := Enc(rand.Reader, AnonPk[i], bDiff)
			CLnNew[i] = cipher.C
			CRnNew[i] = cipher.D
		} else {
			ciphertext, _ := Enc(rand.Reader, AnonPk[i], big.NewInt(i+10))
			CLnNew[i] = ciphertext.C
			CRnNew[i] = ciphertext.D
		}
	}

	proof, err := TfProver(AnonPk, CLnNew, CRnNew, CVector, D, NonceU, epoch, sk, bTf, bDiff, r, l0, l1)
	if err != nil {
		t.Log(err)
	}
	marshal := TfProofMarshal(proof)
	t.Logf("proof encode result is %v\n", marshal)

	unmarshal, err := TfProofUnMarshal(marshal)
	if err != nil {
		t.Logf("the unmarshal error is %v\n", err)
	}

	res, err := TfVerifier(AnonPk, CLnNew, CRnNew, CVector, D, NonceU, epoch, unmarshal)
	t.Log(res)
}

//this function is to generate test case for smart contract
func TestTfProver2(t *testing.T) {
	N := int64(4)
	sk := make([]*big.Int, N)
	sk[0], _ = new(big.Int).SetString("03ebaedfaad6ae54b1dc21432ee41dbef0f13ccad914771165e30037b1fd564c", 16)
	sk[1], _ = new(big.Int).SetString("1679aa304d337896d85a487b93dd624070e7cf6473eb5b00abe8cb54e9d3dcd8", 16)
	sk[2], _ = new(big.Int).SetString("0268df8b8490d0d46de3a8b738a7967bdad5c4b892ec615d90be43f57eadc908", 16)
	sk[3], _ = new(big.Int).SetString("188ac41ed6280882a92f63844384c0cbaa937319ffc4371c1d0708cb1d829dd0", 16)

	r, _ := rand.Int(rand.Reader, ORDER)
	fmt.Printf("the random number is %v\n", r)
	l0 := big.NewInt(1)
	l1 := big.NewInt(2)
	bTf := big.NewInt(100)
	bDiff := big.NewInt(900)

	AnonPk := make([]*bn256.G1, N)
	CLnOld := make([]*bn256.G1, N)
	CRnOld := make([]*bn256.G1, N)
	CLnNew := make([]*bn256.G1, N)
	CRnNew := make([]*bn256.G1, N)
	CVector := make([]*bn256.G1, N)
	D := new(bn256.G1).ScalarMult(G, r)
	epoch := big.NewInt(200)
	Gepoch := MapIntoGroup("zether" + epoch.String())
	NonceU := new(bn256.G1).ScalarMult(Gepoch, sk[l0.Int64()])

	balance := []int64{0, 1000, 500, 100}
	for i := int64(0); i < N; i++ {
		AnonPk[i] = new(bn256.G1).ScalarMult(G, sk[i])
		fmt.Printf("the public key is %v\n", AnonPk[i])
		CRnOld[i] = G // where the r = 1, CLnOld = b * G + y
		CLnOld[i] = new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, big.NewInt(balance[i])), AnonPk[i])
		fmt.Printf("the initial balance is %v\n", CLnOld[i])
	}

	for i := int64(0); i < N; i++ {
		if l0.Cmp(big.NewInt(i)) == 0 {
			CVector[i] = new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, new(big.Int).Neg(bTf)), new(bn256.G1).ScalarMult(AnonPk[i], r))
		}
		if l1.Cmp(big.NewInt(i)) == 0 {
			CVector[i] = new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, bTf), new(bn256.G1).ScalarMult(AnonPk[i], r))
		}
		if l0.Cmp(big.NewInt(i)) != 0 && l1.Cmp(big.NewInt(i)) != 0 {
			CVector[i] = new(bn256.G1).ScalarMult(AnonPk[i], r)
		}
		fmt.Printf("CVector is %v\n", CVector[i])
	}

	for i := int64(0); i < N; i++ {
		CLnNew[i] = new(bn256.G1).Add(CLnOld[i], CVector[i])
		CRnNew[i] = new(bn256.G1).Add(CRnOld[i], D)
		fmt.Printf("CLnNew is %v\n", CLnNew[i])
		fmt.Printf("CRnNew is %v\n", CRnNew[i])
	}

	proof, err := TfProver(AnonPk, CLnNew, CRnNew, CVector, D, NonceU, epoch, sk[l0.Int64()], bTf, bDiff, r, l0, l1)
	if err != nil {
		t.Log(err)
	}
	pRes := TfProofMarshal(proof)
	t.Logf("the proof is %v\n", pRes)
	proofNew, _ := TfProofUnMarshal(pRes)
	res, err := TfVerifier(AnonPk, CLnNew, CRnNew, CVector, D, NonceU, epoch, proofNew)
	t.Log(res)
}
