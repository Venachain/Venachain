package crypto

import (
	"crypto/rand"
	"errors"
	"math"
	"math/big"
	"math/bits"

	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/bn256"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
)

type TransferStatement struct {
	//N is the length of Anonymous pk set and N must be 2^m
	//N int64
	AnonPk  []*bn256.G1
	CLnNew  []*bn256.G1
	CRnNew  []*bn256.G1
	CVector []*bn256.G1
	D       *bn256.G1
	NonceU  *bn256.G1
	Epoch   *big.Int
}

type TransferWitness struct {
	sk *big.Int
	//bTf is the transfer value
	bTf *big.Int
	//bDiff = b - bTf is the remain balance after the transfer tx
	bDiff *big.Int
	r     *big.Int
	l0    *big.Int
	l1    *big.Int
}

type TransferProof struct {
	BpA       *bn256.G1
	BpS       *bn256.G1
	A         *bn256.G1
	B         *bn256.G1
	CLnTilde  []*bn256.G1
	CRnTilde  []*bn256.G1
	C0Tilde   []*bn256.G1
	Dtilde    []*bn256.G1
	Tfy0Tilde []*bn256.G1
	TfgTilde  []*bn256.G1
	CXtilde   []*bn256.G1
	TfyXtilde []*bn256.G1
	Tff       []*big.Int
	TfzA      *big.Int
	BpT1      *bn256.G1
	BpT2      *bn256.G1
	BpThat    *big.Int
	BpMu      *big.Int
	Tfc       *big.Int
	TfSsk     *big.Int
	TfSr      *big.Int
	TfSb      *big.Int
	TfStau    *big.Int
	IpProof   *InnerProductProof
}

type TfSMarshal struct {
	AnonPk  []byte
	CLnNew  []byte
	CRnNew  []byte
	CVector []byte
	D       []byte
	NonceU  []byte
	Epoch   []byte
}

//Marshal TransferStatement which will be used to generate challenge
func TfStatmentMarshal(tfStatment *TransferStatement) ([]byte, error) {
	tfMarshal := new(TfSMarshal)
	tfMarshal.AnonPk = WdPointMarshal(tfStatment.AnonPk)
	tfMarshal.CLnNew = WdPointMarshal(tfStatment.CLnNew)
	tfMarshal.CRnNew = WdPointMarshal(tfStatment.CRnNew)
	tfMarshal.CVector = WdPointMarshal(tfStatment.CVector)
	tfMarshal.D = tfStatment.D.Marshal()
	tfMarshal.NonceU = tfStatment.NonceU.Marshal()
	tfMarshal.Epoch = tfStatment.Epoch.Bytes()
	res, err := rlp.EncodeToBytes(tfMarshal)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//many out of many proof: commitments to bits
func commitToBits(m int64, tfWit *TransferWitness, agBp *AggBpStatement) (*bn256.G1, *bn256.G1, *big.Int, *big.Int, []*big.Int, []*big.Int, error) {
	//compute A = Com(a||d||e, r_A), B = Com(b||c||f, r_B)
	rA, err := rand.Int(rand.Reader, ORDER)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.New("generate random number failed")
	}
	rB, err := rand.Int(rand.Reader, ORDER)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, errors.New("generate random number failed")
	}
	//randomly generate 2m value
	a := sampleRandomVector(2 * m)
	//compute b
	l0Bits := Decompose(tfWit.l0, 2, m)
	l1Bits := Decompose(tfWit.l1, 2, m)
	b := append(l0Bits, l1Bits...)
	//compute c
	negTwob := VectorScalarMul(b, big.NewInt(-2))
	var i int64
	for i = 0; i < 2*m; i++ {
		negTwob[i] = new(big.Int).Add(big.NewInt(1), negTwob[i])
	}
	c, err := VectorHadamard(a, negTwob)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	//compute d
	aa, err := VectorHadamard(a, a)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	d := VectorScalarMul(aa, big.NewInt(-1))
	//compute e
	e := make([]*big.Int, 2)
	e[0] = new(big.Int).Mul(a[0], a[m])
	e[1] = e[0]
	//compute f
	f := make([]*big.Int, 2)
	idLeft := new(big.Int).Mul(b[0], big.NewInt(m))
	f[0] = a[idLeft.Int64()]
	idRight := new(big.Int).Mul(b[m], big.NewInt(m))
	f[1] = new(big.Int).Neg(a[idRight.Int64()])
	//compute a||d||e
	ad := append(a, d...)
	ade := append(ad, e...)
	length := len(ade)
	Atmp, err := VectorScalarMulSum(ade, agBp.bpParam.gVector[:length])
	A := new(bn256.G1).Add(Atmp, new(bn256.G1).ScalarMult(agBp.bpParam.h, rA))
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	//compute b||c||f
	bc := append(b, c...)
	bcf := append(bc, f...)
	Len := len(bcf)
	if Len != length {
		return nil, nil, nil, nil, nil, nil, errVectorLength
	}
	Btmp, err := VectorScalarMulSum(bcf, agBp.bpParam.gVector[:Len])
	B := new(bn256.G1).Add(Btmp, new(bn256.G1).ScalarMult(agBp.bpParam.h, rB))
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	return A, B, rA, rB, a, b, nil
}

//generate challenge v = H(statementHash||BpA||BpS||A||B)
func GenerateTFv(tfStatement *TransferStatement, BpA, BpS, A, B *bn256.G1) (*big.Int, error) {
	tfS, err := TfStatmentMarshal(tfStatement)
	if err != nil {
		return nil, err
	}
	BpAS := append(BpA.Marshal(), BpS.Marshal()...)
	AB := append(A.Marshal(), B.Marshal()...)
	msg := append(tfS, BpAS...)
	msg = append(msg, AB...)
	v := GenerateChallenge(msg, ORDER)
	return v, nil
}

//compute F0k0 = (1-b0k) * W - a0k or F1k0 where a can be a0 or a1, lBits can be l0Bits or l1Bits
func TfPolyCoeffF0(a []*big.Int, lBits []*big.Int, m int64) ([][2]*big.Int, error) {
	if int64(len(a)) != m || int64(len(lBits)) != m {
		return nil, errVectorLength
	}
	res := make([][2]*big.Int, m)
	var i int64
	for i = 0; i < m; i++ {
		res[i][0] = new(big.Int).Neg(a[i])
		res[i][1] = new(big.Int).Add(big.NewInt(1), new(big.Int).Neg(lBits[i]))
	}
	return res, nil
}

//F0k1 or F1k1, where a can be a0 or a1, lBits can be l0Bits or l1Bits
func TfPolyCoeffF1(a []*big.Int, lBits []*big.Int, m int64) ([][2]*big.Int, error) {
	if int64(len(a)) != m || int64(len(lBits)) != m {
		return nil, errVectorLength
	}
	res := make([][2]*big.Int, m)
	var i int64
	for i = 0; i < m; i++ {
		res[i][0] = a[i]
		res[i][1] = lBits[i]
	}
	return res, nil
}

/*
 TfPolyP is to compute Polynomial P0 or P1, where the input is F0k0Coeff and F0k1Coeff
 or F1k0Coeff and F1k1Coeff, the result is polynomials of length N, and every polynomial
 coefficient length is m, namely res =[N][m]*big.Int
*/
func TfPolyP(Flk0Coeff, Flk1Coeff [][2]*big.Int, N, m int64) [][]*big.Int {
	//construct the result which is N*m array, res[i] is a Polynomial of degree m
	res := make([][]*big.Int, N)
	var i, j int64
	for j = 0; j < N; j++ {
		res[j] = make([]*big.Int, 2)
		for i := int64(0); i < 2; i++ {
			res[j][i] = big.NewInt(0)
		}
	}
	for i = 0; i < N; i++ {
		bini := Decompose(big.NewInt(i), 2, m)
		if bini[0].Cmp(big.NewInt(1)) == 0 {
			res[i][0] = res[i][0].Add(res[i][0], Flk1Coeff[0][0])
			res[i][1] = res[i][1].Add(res[i][1], Flk1Coeff[0][1])
		} else { //bini[0] = 0
			res[i][0] = res[i][0].Add(res[i][0], Flk0Coeff[0][0])
			res[i][1] = res[i][1].Add(res[i][1], Flk0Coeff[0][1])
		}
		for k := int64(1); k < m; k++ {
			if bini[k].Cmp(big.NewInt(1)) == 0 {
				res[i] = PolyProduct(Flk1Coeff[k][:], res[i], ORDER)
			} else {
				res[i] = PolyProduct(Flk0Coeff[k][:], res[i], ORDER)
			}
		}
	}
	return res
}

func TfSampleRandomVector(m int64) (phi, chi, psi, omega []*big.Int) {
	phi = sampleRandomVector(m)
	chi = sampleRandomVector(m)
	psi = sampleRandomVector(m)
	omega = sampleRandomVector(m)
	return
}

//multidimensional vector transpose
func MultiVectorT(vector [][]*big.Int) [][]*big.Int {
	row := int64(len(vector))
	column := int64(len(vector[0]))
	res := make([][]*big.Int, column)
	for i := int64(0); i < column; i++ {
		res[i] = make([]*big.Int, row)
		for j := int64(0); j < row; j++ {
			res[i][j] = new(big.Int).SetBytes(vector[j][i].Bytes())
		}
	}
	return res
}

//P0 is N*m, len(C) = N, MultiVectorT(P0) is m*N
// this function can compute proof CLnk, CRnk, C0k, Tfy0k
func VectorCommitMuliExp(P0 [][]*big.Int, Cipher []*bn256.G1, m int64, Publ0 *bn256.G1, phi []*big.Int) ([]*bn256.G1, error) {
	if int64(len(phi)) != m {
		return nil, errVectorLength
	}
	res := make([]*bn256.G1, m)
	P0T := MultiVectorT(P0)
	for k := int64(0); k < m; k++ {
		if len(P0T[0]) != len(Cipher) {
			return nil, errVectorLength
		}
		tmp, err := VectorScalarMulSum(P0T[k], Cipher)
		if err != nil {
			return nil, err
		}
		res[k] = new(bn256.G1).Add(tmp, new(bn256.G1).ScalarMult(Publ0, phi[k]))
	}
	return res, nil
}

//compute Dtilde or tfgTilde
func VectorScalarMulG(G *bn256.G1, randNum []*big.Int) []*bn256.G1 {
	m := len(randNum)
	res := make([]*bn256.G1, m)
	for i := 0; i < m; i++ {
		res[i] = new(bn256.G1).ScalarMult(G, randNum[i])
	}
	return res
}

//compute proof CXk
func CXtilde(v, bTf *big.Int, P0 [][]*big.Int, P1 [][]*big.Int, N, m int64, omega []*big.Int, D *bn256.G1, l0, l1 int64) ([]*bn256.G1, error) {
	if int64(len(omega)) != m {
		return nil, errVectorLength
	}
	var i, k, j int64
	kexi := make([]*big.Int, N)
	kexi[0] = big.NewInt(1)
	for i = 1; i < N; i++ {
		kexi[i] = new(big.Int).Exp(v, big.NewInt(i-1), ORDER)
	}
	CXk := make([]*bn256.G1, m)
	for k = 0; k < m; k++ {
		if int64(len(P0)) != N || int64(len(P1)) != N || int64(len(P0[0])) != m+1 || int64(len(P1[0])) != m+1 {
			return nil, errVectorLength
		}
		sum0 := big.NewInt(0)
		sum1 := big.NewInt(0)
		for j = 0; j <= N/2-1; j++ {
			l02j := new(big.Int).Mod(big.NewInt(l0-2*j), big.NewInt(N)).Int64()
			l12j := new(big.Int).Mod(big.NewInt(l1-2*j), big.NewInt(N)).Int64()
			tmpSub := new(big.Int).Mod(new(big.Int).Sub(P0[l12j][k], P0[l02j][k]), ORDER)
			tmpMul := new(big.Int).Mod(new(big.Int).Mul(kexi[2*j], bTf), ORDER)
			sum0 = new(big.Int).Add(new(big.Int).Mod(new(big.Int).Mul(tmpMul, tmpSub), ORDER), sum0)
			P1Sub := new(big.Int).Mod(new(big.Int).Sub(P1[l12j][k], P1[l02j][k]), ORDER)
			P1Mul := new(big.Int).Mod(new(big.Int).Mul(kexi[2*j+1], bTf), ORDER)
			sum1 = new(big.Int).Add(new(big.Int).Mod(new(big.Int).Mul(P1Sub, P1Mul), ORDER), sum1)
		}
		gbp := new(bn256.G1).ScalarMult(G, new(big.Int).Mod(new(big.Int).Add(sum0, sum1), ORDER))
		CXk[k] = new(bn256.G1).Add(gbp, new(bn256.G1).ScalarMult(D, omega[k]))

	}
	return CXk, nil
}

//generate omega
func GenerateTfOmega(v *big.Int, CLnTilde, CRnTilde, C0Tilde, Dtilde, Tfy0Tilde, TfgTilde, CXtilde, TfyXtilde []*bn256.G1) *big.Int {
	msg := append(v.Bytes(), WdPointMarshal(CLnTilde)...)
	msg = append(msg, WdPointMarshal(CRnTilde)...)
	msg = append(msg, WdPointMarshal(C0Tilde)...)
	msg = append(msg, WdPointMarshal(Dtilde)...)
	msg = append(msg, WdPointMarshal(Tfy0Tilde)...)
	msg = append(msg, WdPointMarshal(TfgTilde)...)
	msg = append(msg, WdPointMarshal(CXtilde)...)
	msg = append(msg, WdPointMarshal(TfyXtilde)...)
	omega := GenerateChallenge(msg, ORDER)
	return omega
}

//compute proof TFf []*big.Int where flk = Flk1(omega)
func TFf(F1Coeff [][2]*big.Int, omega *big.Int) (f []*big.Int) {
	m := len(F1Coeff[:])
	f = make([]*big.Int, m)
	for i := 0; i < m; i++ {
		f[i] = new(big.Int).Add(F1Coeff[i][0], new(big.Int).Mul(omega, F1Coeff[i][1]))
		f[i].Mod(f[i], ORDER)
	}
	return f
}

//compute CRnLine, DLine, y0Line, gLine
func VectorLine(point *bn256.G1, pub *bn256.G1, r []*big.Int, w *big.Int, m int64) (*bn256.G1, error) {
	if int64(len(r)) != m {
		return nil, errVectorLength
	}
	wm := new(big.Int).Exp(w, big.NewInt(m), ORDER)
	tmp1 := new(bn256.G1).ScalarMult(point, wm)
	var k int64
	sum := big.NewInt(0)
	for k = 0; k < m; k++ {
		wk := new(big.Int).Exp(w, big.NewInt(k), ORDER)
		negRwk := new(big.Int).Mod(new(big.Int).Neg(new(big.Int).Mul(r[k], wk)), ORDER)
		sum = new(big.Int).Add(negRwk, sum)
		sum.Mod(sum, ORDER)
	}
	tmp2 := new(bn256.G1).ScalarMult(G, sum)
	res := new(bn256.G1).Add(tmp1, tmp2)
	return res, nil
}

//compute circularly Shift vector
func VectorShift(vector []*big.Int, shift int64) []*big.Int {
	length := int64(len(vector))
	if shift > length {
		return nil
	}
	res := make([]*big.Int, length)
	var i int64
	for i = 0; i < length; i++ {
		if i < shift {
			res[i] = vector[i-shift+length]
		} else {
			res[i] = vector[i-shift]
		}
	}
	return res
}

//Given polynomials coefficients, compute polynomial value
func VectorPolyEvaluate(PolyCoeff [][]*big.Int, x *big.Int) []*big.Int {
	resLen := len(PolyCoeff)
	res := make([]*big.Int, resLen)
	for i := 0; i < resLen; i++ {
		res[i] = PolyEvaluate(PolyCoeff[i], x)
	}
	return res
}

// compute TfyXline
func TfyXline(PolyP0Coeff [][]*big.Int, PolyP1Coeff [][]*big.Int, N, m int64, AnonPk []*bn256.G1, v, challOmega *big.Int, omega []*big.Int) (*bn256.G1, error) {
	if int64(len(AnonPk)) != N || int64(len(omega)) != m {
		return nil, errVectorLength
	}
	var i, j, k int64
	kexi := make([]*big.Int, N)
	kexi[0] = big.NewInt(1)
	for i = 1; i < N; i++ {
		kexi[i] = new(big.Int).Exp(v, big.NewInt(i-1), ORDER)
	}
	//compute P_l,i(omega)
	PolyP0 := VectorPolyEvaluate(PolyP0Coeff, challOmega)
	PolyP1 := VectorPolyEvaluate(PolyP1Coeff, challOmega)

	sum0 := new(bn256.G1).ScalarBaseMult(big.NewInt(0))
	sum1 := new(bn256.G1).ScalarBaseMult(big.NewInt(0))
	for j = 0; j <= N/2-1; j++ {
		res0, err := VectorScalarMulSum(VectorShift(PolyP0, 2*j), AnonPk)
		if err != nil {
			return nil, err
		}
		res0 = new(bn256.G1).ScalarMult(res0, kexi[2*j])
		sum0 = new(bn256.G1).Add(res0, sum0)

		res1, err := VectorScalarMulSum(VectorShift(PolyP1, 2*j), AnonPk)
		res1 = new(bn256.G1).ScalarMult(res1, kexi[2*j+1])
		sum1 = new(bn256.G1).Add(res1, sum1)
	}
	res := new(bn256.G1).Add(sum0, sum1)
	exp := big.NewInt(0)
	for k = 0; k < m; k++ {
		wk := new(big.Int).Exp(challOmega, big.NewInt(k), ORDER)
		negRwk := new(big.Int).Mod(new(big.Int).Neg(new(big.Int).Mul(omega[k], wk)), ORDER)
		exp = new(big.Int).Mod(new(big.Int).Add(negRwk, exp), ORDER)
	}
	tmp := new(bn256.G1).ScalarMult(G, exp)
	res = new(bn256.G1).Add(res, tmp)
	return res, nil
}

//generate transfer proof c
func GenerateTfC(x *big.Int, Ay, AD, Ab, AX, At, Au *bn256.G1) *big.Int {
	//generate c = H(x||Ay||AD||Ab||AX||At||Au)
	message := x.Bytes()
	message = append(message, Ay.Marshal()...)
	message = append(message, AD.Marshal()...)
	message = append(message, Ab.Marshal()...)
	message = append(message, AX.Marshal()...)
	message = append(message, At.Marshal()...)
	message = append(message, Au.Marshal()...)
	c := GenerateChallenge(message, ORDER)
	return c
}

//given flk1, compute the flk0 = w - flk1, where f is flk1, challOmega is w
func VerifierF0Coeff(f []*big.Int, challOmega *big.Int) []*big.Int {
	length := len(f)
	flk0 := make([]*big.Int, length)
	for i := 0; i < length; i++ {
		flk0[i] = new(big.Int).Mod(new(big.Int).Sub(challOmega, f[i]), ORDER)
	}
	return flk0
}

//Given flk0 and flk1 where their length is m instead of 2m, the output is polyli = \PI(flkk_i)
func VerifierPolyCoeff(flk0, flk1 []*big.Int, N, m int64) []*big.Int {
	poly := make([]*big.Int, N)
	for i := int64(0); i < N; i++ {
		poly[i] = big.NewInt(1)
		bini := Decompose(big.NewInt(i), 2, m)
		for k := int64(0); k < m; k++ {
			if bini[k].Cmp(big.NewInt(0)) == 0 {
				poly[i] = new(big.Int).Mod(new(big.Int).Mul(flk0[k], poly[i]), ORDER)
			} else {
				poly[i] = new(big.Int).Mod(new(big.Int).Mul(flk1[k], poly[i]), ORDER)
			}
		}
	}
	return poly
}

func VerifierCommit(f []*big.Int, A, B *bn256.G1, m int64, zA, challOmega *big.Int, aggBp *AggBpStatement) (bool, error) {
	if int64(len(f)) != 2*m {
		return false, errVectorLength
	}
	//	compute B^wA
	BwA := new(bn256.G1).Add(A, new(bn256.G1).ScalarMult(B, challOmega))
	resLeft := new(big.Int).SetBytes(BwA.Marshal())
	wnegf := VerifierF0Coeff(f, challOmega)
	fwf, err := VectorHadamard(f, wnegf)
	if err != nil {
		return false, err
	}
	tmp1 := new(big.Int).Mod(new(big.Int).Mul(f[0], f[m]), ORDER)
	tmp2 := new(big.Int).Mod(new(big.Int).Mul(wnegf[0], wnegf[m]), ORDER)
	input := append(f, fwf...)
	input = append(input, tmp1)
	input = append(input, tmp2)
	length := len(input)
	com, err := VectorScalarMulSum(input, aggBp.bpParam.gVector[:length])
	hzA := new(bn256.G1).ScalarMult(aggBp.bpParam.h, zA)
	res := new(bn256.G1).Add(com, hzA)
	resRight := new(big.Int).SetBytes(res.Marshal())
	if resLeft.Cmp(resRight) == 0 {
		return true, nil
	}
	return false, errors.New("commit verify failed")
}

//this function is to compute \Pi(vectorTilde[k] *(-omega^k))
func VerifierVectorTilde(vectorTilde []*bn256.G1, omega *big.Int) (*bn256.G1, error) {
	m := len(vectorTilde)
	negwk := make([]*big.Int, m)
	for k := 0; k < m; k++ {
		negwk[k] = new(big.Int).Sub(ORDER, new(big.Int).Exp(omega, big.NewInt(int64(k)), ORDER))
	}
	res, err := VectorScalarMulSum(negwk, vectorTilde)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//compute CLnLine which includes MultiExp and VectorTilde algorithms, line 104-109
func VerifierVectorLine(vectorPoint, vectorTilde []*bn256.G1, coeff []*big.Int, omega *big.Int) (*bn256.G1, error) {
	if len(vectorPoint) != len(coeff) {
		return nil, errVectorLength
	}
	multiExp, err := VectorScalarMulSum(coeff, vectorPoint)
	if err != nil {
		return nil, err
	}
	tilde, err := VerifierVectorTilde(vectorTilde, omega)
	if err != nil {
		return nil, err
	}
	res := new(bn256.G1).Add(multiExp, tilde)
	return res, nil
}

//generate kexi = [1,1,v,v^2,..,v^(N-2)]
func GenerateKexi(v *big.Int, N int64) []*big.Int {
	kexi := make([]*big.Int, N)
	kexi[0] = big.NewInt(1)
	for i := int64(1); i < N; i++ {
		kexi[i] = new(big.Int).Exp(v, big.NewInt(i-1), ORDER)
	}
	return kexi
}

//compute CXline and yXline left part which includes MultiExp and shift algorithm, where polyP0= P0i, polyP1=P1i
func VectorMuliExpShift(vectorPoint []*bn256.G1, kexi, polyP0, polyP1 []*big.Int, N int64) (*bn256.G1, error) {
	if int64(len(vectorPoint)) != N || int64(len(polyP0)) != N || int64(len(polyP1)) != N || int64(len(kexi)) != N {
		return nil, errVectorLength
	}
	var j int64
	sum0 := new(bn256.G1).ScalarBaseMult(big.NewInt(0))
	sum1 := new(bn256.G1).ScalarBaseMult(big.NewInt(0))
	for j = 0; j <= N/2-1; j++ {
		res0, err := VectorScalarMulSum(VectorShift(polyP0, 2*j), vectorPoint)
		if err != nil {
			return nil, err
		}
		res0 = new(bn256.G1).ScalarMult(res0, kexi[2*j])
		sum0 = new(bn256.G1).Add(res0, sum0)

		res1, err := VectorScalarMulSum(VectorShift(polyP1, 2*j), vectorPoint)
		res1 = new(bn256.G1).ScalarMult(res1, kexi[2*j+1])
		sum1 = new(bn256.G1).Add(res1, sum1)
	}
	res := new(bn256.G1).Add(sum0, sum1)
	return res, nil
}

func ComputeAt(aggBp *AggBpStatement, c, omega, tHat, stau, delta, sb, x *big.Int, m int64, T1, T2 *bn256.G1) *bn256.G1 {
	wm := new(big.Int).Exp(omega, big.NewInt(m), ORDER)
	wmc := new(big.Int).Mod(new(big.Int).Mul(wm, c), ORDER)
	tmp1 := new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, new(big.Int).Mod(new(big.Int).Mul(wmc, tHat), ORDER)), new(bn256.G1).ScalarMult(aggBp.bpParam.h, stau))
	a := new(bn256.G1).ScalarMult(G, new(big.Int).Mod(new(big.Int).Mul(wmc, delta), ORDER))
	b := new(bn256.G1).ScalarMult(G, sb)
	tmp2 := new(bn256.G1).Add(a, b)
	T1x := new(bn256.G1).ScalarMult(T1, x)
	T2xx := new(bn256.G1).ScalarMult(T2, new(big.Int).Mod(new(big.Int).Mul(x, x), ORDER))
	tmp3 := new(bn256.G1).ScalarMult(new(bn256.G1).Add(T1x, T2xx), wmc)
	At := new(bn256.G1).Add(tmp1, new(bn256.G1).Neg(new(bn256.G1).Add(tmp2, tmp3)))
	return At
}

//generate transfer proof
func TfProver(AnonPk, CLnNew, CRnNew, CVector []*bn256.G1, D, NonceU *bn256.G1, epoch *big.Int, sk, bTf, bDiff, r, l0, l1 *big.Int) (*TransferProof, error) {
	tfStatement := new(TransferStatement)
	tfStatement.AnonPk = AnonPk
	tfStatement.CLnNew = CLnNew
	tfStatement.CRnNew = CRnNew
	tfStatement.CVector = CVector
	tfStatement.D = D
	tfStatement.NonceU = NonceU
	tfStatement.Epoch = epoch

	tfWit := new(TransferWitness)
	tfWit.sk = sk
	tfWit.bTf = bTf
	tfWit.bDiff = bDiff
	tfWit.r = r
	tfWit.l0 = l0
	tfWit.l1 = l1

	tfProof, err := TransferProve(tfStatement, tfWit)
	if err != nil {
		return nil, err
	}
	return tfProof, nil
}

func TransferProve(tfStatement *TransferStatement, tfWit *TransferWitness) (*TransferProof, error) {
	proof := new(TransferProof)
	//1. use witness bTf and bDiff to and aggbulletproof to generate A and S
	bpWit := new(AggBpWitness)
	bpV := make([]*big.Int, 2)
	bpV[0] = tfWit.bTf
	bpV[1] = tfWit.bDiff
	bpWit.v = bpV
	instance := AggBp()
	aL, aR, alpha, BpA, err := GenerateAggA(&instance, bpWit)
	if err != nil {
		return nil, err
	}
	sL, sR, rho, BpS, err := GenerateAggS(&instance)
	if err != nil {
		return nil, err
	}
	proof.BpA = BpA
	proof.BpS = BpS
	//2. compute Anonymous pk set length which must be power of 2, m is the N bit length
	N := int64(len(tfStatement.AnonPk))
	Ncheck := IsPowOfTwo(N)
	if Ncheck == false {
		return nil, errors.New("Anonymous pk set length is invalid")
	}
	m := int64(bits.Len(uint(N))) - 1
	//3. many out of many proof: commitments to bits, generate proof A and B
	A, B, rA, rB, Randa, Witb, err := commitToBits(m, tfWit, &instance)
	if err != nil {
		return nil, err
	}
	proof.A = A
	proof.B = B
	//4. generate challenge v
	v, err := GenerateTFv(tfStatement, BpA, BpS, A, B)
	if err != nil {
		return nil, err
	}
	//5.construct polynomial for witness l0
	//5.1. F0k1 = a0k+b0k * W
	F0k1Coeff, _ := TfPolyCoeffF1(Randa[:m], Witb[:m], m)
	//5.2. F0k0 = -a0k+(1-b0k) * W
	F0k0Coeff, _ := TfPolyCoeffF0(Randa[:m], Witb[:m], m)
	//6.construct polynomial for witness l1
	//6.1. F1k1 = a1k+b1k * W
	F1k1Coeff, _ := TfPolyCoeffF1(Randa[m:], Witb[m:], m)
	//6.2. F1k0 = -a1k+(1-b1k) * W
	F1k0Coeff, _ := TfPolyCoeffF0(Randa[m:], Witb[m:], m)
	//7.compute polynomial P0 and P1
	PolyP0 := TfPolyP(F0k0Coeff, F0k1Coeff, N, m)
	PolyP1 := TfPolyP(F1k0Coeff, F1k1Coeff, N, m)

	//8. generate randomly vector
	phi, chi, psi, omega := TfSampleRandomVector(m)
	//9. compute CLnTilde, CRnTilde, C0Tilde, Dtilde, Tfy0Tilde,TfgTilde, CXTilde, TfyXTilde
	CLnTilde, err := VectorCommitMuliExp(PolyP0, tfStatement.CLnNew, m, tfStatement.AnonPk[tfWit.l0.Int64()], phi)
	CRnTilde, err := VectorCommitMuliExp(PolyP0, tfStatement.CRnNew, m, G, phi)
	C0Tilde, err := VectorCommitMuliExp(PolyP0, tfStatement.CVector, m, tfStatement.AnonPk[tfWit.l0.Int64()], chi)
	Dtilde := VectorScalarMulG(G, chi)
	Tfy0Tilde, err := VectorCommitMuliExp(PolyP0, tfStatement.AnonPk, m, new(bn256.G1).ScalarMult(G, tfWit.sk), psi)
	TfgTilde := VectorScalarMulG(G, psi)
	CXtilde, err := CXtilde(v, tfWit.bTf, PolyP0, PolyP1, N, m, omega, tfStatement.D, tfWit.l0.Int64(), tfWit.l1.Int64())
	TfyXtilde := VectorScalarMulG(G, omega)
	proof.CLnTilde = CLnTilde
	proof.CRnTilde = CRnTilde
	proof.C0Tilde = C0Tilde
	proof.Dtilde = Dtilde
	proof.Tfy0Tilde = Tfy0Tilde
	proof.TfgTilde = TfgTilde
	proof.CXtilde = CXtilde
	proof.TfyXtilde = TfyXtilde
	//10. generate challenge omega = H(v||CLnTilde||CRnTilde||C0Tilde||Dtilde||Tfy0Tilde||TfgTilde||CXtilde||TfyXtilde)
	challOmega := GenerateTfOmega(v, CLnTilde, CRnTilde, C0Tilde, Dtilde, Tfy0Tilde, TfgTilde, CXtilde, TfyXtilde)
	//11. compute f0 and f1 and zA
	f0 := TFf(F0k1Coeff, challOmega)
	f1 := TFf(F1k1Coeff, challOmega)
	proof.Tff = append(f0, f1...)
	zA := new(big.Int).Mod(new(big.Int).Add(rA, new(big.Int).Mul(rB, challOmega)), ORDER)
	proof.TfzA = zA
	//12. compute challenge y and z
	chally := GenerateChallenge(challOmega.Bytes(), ORDER)
	challz := GenerateChallenge(chally.Bytes(), ORDER)
	//13. generate T1 and T2
	T1, T2, tau1, tau2, err := GenerateAggT1T2(&instance, aL, aR, sL, sR, chally, challz)
	proof.BpT1 = T1
	proof.BpT2 = T2
	//14. generate challenge x
	challx := Generatex(T1, T2, challz)
	//15. compute tHat, tauX, and mu
	lx, rx, tHat, err := GenerateAggtHat(aL, aR, sL, sR, challx, chally, challz, instance.bpParam.n, instance.m)
	tauXcoeff := []*big.Int{big.NewInt(0), tau1, tau2}
	tauX := PolyEvaluate(tauXcoeff, challx)
	mu := new(big.Int).Mod(new(big.Int).Add(alpha, new(big.Int).Mul(challx, rho)), ORDER)
	proof.BpThat = tHat
	proof.BpMu = mu
	//16. compute CRnLine, Dline, y0Line, gLine
	CRnLine, err := VectorLine(tfStatement.CRnNew[tfWit.l0.Int64()], G, phi, challOmega, m)
	DLine, err := VectorLine(tfStatement.D, G, chi, challOmega, m)
	//y0Line, err :=  VectorLine(tfStatement.AnonPk[tfWit.l0.Int64()],tfStatement.AnonPk[tfWit.l0.Int64()], psi, challOmega, m)
	//fmt.Printf("transferproof - y0Line is : %v\n", y0Line)
	gLine, err := VectorLine(G, G, psi, challOmega, m)
	yXline, err := TfyXline(PolyP0, PolyP1, N, m, tfStatement.AnonPk, v, challOmega, omega)
	//17. compute Ay, AD, Ab, AX, At, Au
	ksk, _ := rand.Int(rand.Reader, ORDER)
	kr, _ := rand.Int(rand.Reader, ORDER)
	kb, _ := rand.Int(rand.Reader, ORDER)
	ktau, _ := rand.Int(rand.Reader, ORDER)
	Ay := new(bn256.G1).ScalarMult(gLine, ksk)
	AD := new(bn256.G1).ScalarMult(G, kr)
	zz := new(big.Int).Mod(new(big.Int).Mul(challz, challz), ORDER)
	zzz := new(big.Int).Mod(new(big.Int).Mul(zz, challz), ORDER)
	gkb := new(bn256.G1).ScalarMult(G, kb)
	negDzz := new(bn256.G1).ScalarMult(DLine, new(big.Int).Neg(zz))
	Ab := new(bn256.G1).Add(gkb, new(bn256.G1).ScalarMult(new(bn256.G1).Add(negDzz, new(bn256.G1).ScalarMult(CRnLine, zzz)), ksk))
	AX := new(bn256.G1).ScalarMult(yXline, kr)
	At := new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, new(big.Int).Neg(kb)), new(bn256.G1).ScalarMult(instance.bpParam.h, ktau))
	Gepoch := MapIntoGroup("zether" + tfStatement.Epoch.String())
	Au := new(bn256.G1).ScalarMult(Gepoch, ksk)
	c := GenerateTfC(challx, Ay, AD, Ab, AX, At, Au)
	proof.Tfc = c
	//18. compute ssk, sr, sb, stau
	ssk := new(big.Int).Mod(new(big.Int).Add(ksk, new(big.Int).Mul(c, tfWit.sk)), ORDER)
	sr := new(big.Int).Mod(new(big.Int).Add(kr, new(big.Int).Mul(c, tfWit.r)), ORDER)
	cwm := new(big.Int).Mod(new(big.Int).Mul(c, new(big.Int).Exp(challOmega, big.NewInt(m), ORDER)), ORDER)
	tfbzz := new(big.Int).Mod(new(big.Int).Mul(tfWit.bTf, zz), ORDER)
	bzzz := new(big.Int).Mod(new(big.Int).Mul(tfWit.bDiff, zzz), ORDER)
	sb := new(big.Int).Mod(new(big.Int).Add(kb, new(big.Int).Mod(new(big.Int).Mul(cwm, new(big.Int).Mod(new(big.Int).Add(tfbzz, bzzz), ORDER)), ORDER)), ORDER)
	stau := new(big.Int).Mod(new(big.Int).Add(ktau, new(big.Int).Mod(new(big.Int).Mul(cwm, tauX), ORDER)), ORDER)
	proof.TfSsk = ssk
	proof.TfSr = sr
	proof.TfSb = sb
	proof.TfStau = stau
	//19. generate ipproof
	ipn := instance.bpParam.n
	ipm := instance.m
	hprime := GenerateHprime(instance.bpParam.hVector, chally)
	P, err := UpdateAggP(BpA, BpS, instance.bpParam.h, instance.bpParam.gVector, hprime, challx, chally, challz, mu, ipn, ipm)
	if err != nil {
		return nil, err
	}
	//generate u, which can no rely on c, you can use any group point
	str := c.String()
	u := MapIntoGroup(str)
	//innerproduct proof
	ipProof, err := IpProve(ipn*ipm, instance.bpParam.gVector, hprime, P, u, lx, rx, tHat, c)
	if err != nil {
		return nil, err
	}
	proof.IpProof = ipProof

	return proof, nil
}

//verify transfer proof
func TfVerifier(AnonPk, CLnNew, CRnNew, CVector []*bn256.G1, D, NonceU *bn256.G1, epoch *big.Int, proof *TransferProof) (bool, error) {
	tfStatement := new(TransferStatement)
	tfStatement.AnonPk = AnonPk
	tfStatement.CLnNew = CLnNew
	tfStatement.CRnNew = CRnNew
	tfStatement.CVector = CVector
	tfStatement.D = D
	tfStatement.NonceU = NonceU
	tfStatement.Epoch = epoch

	return TransferVerify(tfStatement, proof)
}

func TransferVerify(tfStatement *TransferStatement, tfProof *TransferProof) (bool, error) {
	v, err := GenerateTFv(tfStatement, tfProof.BpA, tfProof.BpS, tfProof.A, tfProof.B)
	if err != nil {
		return false, err
	}
	challOmega := GenerateTfOmega(v, tfProof.CLnTilde, tfProof.CRnTilde, tfProof.C0Tilde, tfProof.Dtilde, tfProof.Tfy0Tilde, tfProof.TfgTilde, tfProof.CXtilde, tfProof.TfyXtilde)
	m := int64(len(tfProof.Tff)) / 2
	N := int64(math.Pow(float64(2), float64(m)))
	instance := AggBp()
	comVerifier, _ := VerifierCommit(tfProof.Tff, tfProof.A, tfProof.B, m, tfProof.TfzA, challOmega, &instance)
	if comVerifier == false {
		return false, errors.New("commitment verify failed")
	}
	//compute CLnLine, CRnLine, C0Line, DLine, Tfy0Line, TfgLine
	f0k0 := VerifierF0Coeff(tfProof.Tff[:m], challOmega)
	polyP0 := VerifierPolyCoeff(f0k0, tfProof.Tff[:m], N, m)
	f1k0 := VerifierF0Coeff(tfProof.Tff[m:], challOmega)
	polyP1 := VerifierPolyCoeff(f1k0, tfProof.Tff[m:], N, m)
	CLnLine, _ := VerifierVectorLine(tfStatement.CLnNew, tfProof.CLnTilde, polyP0, challOmega)
	CRnLine, _ := VerifierVectorLine(tfStatement.CRnNew, tfProof.CRnTilde, polyP0, challOmega)
	C0Line, _ := VerifierVectorLine(tfStatement.CVector, tfProof.C0Tilde, polyP0, challOmega)
	Dtmp, _ := VerifierVectorTilde(tfProof.Dtilde, challOmega)
	Dwm := new(bn256.G1).ScalarMult(tfStatement.D, new(big.Int).Exp(challOmega, big.NewInt(m), ORDER))
	Dline := new(bn256.G1).Add(Dtmp, Dwm)
	y0Line, _ := VerifierVectorLine(tfStatement.AnonPk, tfProof.Tfy0Tilde, polyP0, challOmega)
	gTmp, _ := VerifierVectorTilde(tfProof.TfgTilde, challOmega)
	gwm := new(bn256.G1).ScalarMult(G, new(big.Int).Exp(challOmega, big.NewInt(m), ORDER))
	gLine := new(bn256.G1).Add(gTmp, gwm)
	kexi := GenerateKexi(v, N)
	CXtmp1, _ := VectorMuliExpShift(tfStatement.CVector, kexi, polyP0, polyP1, N)
	CXtmp2, _ := VerifierVectorTilde(tfProof.CXtilde, challOmega)
	CXline := new(bn256.G1).Add(CXtmp1, CXtmp2)
	yXtmp1, _ := VectorMuliExpShift(tfStatement.AnonPk, kexi, polyP0, polyP1, N)
	yXtmp2, _ := VerifierVectorTilde(tfProof.TfyXtilde, challOmega)
	yXline := new(bn256.G1).Add(yXtmp1, yXtmp2)
	//compute Ay, AD, Ab, AX, At, Au
	Ay := new(bn256.G1).Add(new(bn256.G1).ScalarMult(gLine, tfProof.TfSsk), new(bn256.G1).ScalarMult(y0Line, new(big.Int).Neg(tfProof.Tfc)))
	AD := new(bn256.G1).Add(new(bn256.G1).ScalarMult(G, tfProof.TfSr), new(bn256.G1).ScalarMult(tfStatement.D, new(big.Int).Neg(tfProof.Tfc)))
	chally := GenerateChallenge(challOmega.Bytes(), ORDER)
	challz := GenerateChallenge(chally.Bytes(), ORDER)
	zz := new(big.Int).Mod(new(big.Int).Mul(challz, challz), ORDER)
	negzz := new(big.Int).Neg(zz)
	zzz := new(big.Int).Mod(new(big.Int).Mul(zz, challz), ORDER)
	AbTmp1 := new(bn256.G1).ScalarMult(new(bn256.G1).Add(new(bn256.G1).ScalarMult(Dline, negzz), new(bn256.G1).ScalarMult(CRnLine, zzz)), tfProof.TfSsk)
	AbTmp2 := new(bn256.G1).ScalarMult(new(bn256.G1).Add(new(bn256.G1).ScalarMult(C0Line, negzz), new(bn256.G1).ScalarMult(CLnLine, zzz)), new(big.Int).Neg(tfProof.Tfc))
	gsb := new(bn256.G1).ScalarMult(G, tfProof.TfSb)
	Ab := new(bn256.G1).Add(gsb, new(bn256.G1).Add(AbTmp1, AbTmp2))
	AX := new(bn256.G1).Add(new(bn256.G1).ScalarMult(yXline, tfProof.TfSr), new(bn256.G1).ScalarMult(CXline, new(big.Int).Neg(tfProof.Tfc)))
	delta := ComputeAggDelta(chally, challz, instance.bpParam.n, instance.m)
	challx := Generatex(tfProof.BpT1, tfProof.BpT2, challz)
	At := ComputeAt(&instance, tfProof.Tfc, challOmega, tfProof.BpThat, tfProof.TfStau, delta, tfProof.TfSb, challx, m, tfProof.BpT1, tfProof.BpT2)
	Gepoch := MapIntoGroup("zether" + tfStatement.Epoch.String())
	Au := new(bn256.G1).Add(new(bn256.G1).ScalarMult(Gepoch, tfProof.TfSsk), new(bn256.G1).ScalarMult(tfStatement.NonceU, new(big.Int).Neg(tfProof.Tfc)))
	c := GenerateTfC(challx, Ay, AD, Ab, AX, At, Au)
	if c.Cmp(tfProof.Tfc) != 0 {
		return false, errors.New("verify proof.tfc failed")
	}
	//compute innerproductproof
	//calculate hprime and P, len(hprime) = n*m
	ipn := instance.bpParam.n
	ipm := instance.m
	hprime := GenerateHprime(instance.bpParam.hVector, chally)
	P, err := UpdateAggP(tfProof.BpA, tfProof.BpS, instance.bpParam.h, instance.bpParam.gVector, hprime, challx, chally, challz, tfProof.BpMu, ipn, ipm)
	if err != nil {
		return false, err
	}
	//generate u, which can no rely on c, you can use any group point
	str := c.String()
	ipu := MapIntoGroup(str)
	ipVerifier, err := IpVerify(ipn*ipm, instance.bpParam.gVector, hprime, P, ipu, tfProof.IpProof, tfProof.BpThat, c)
	if err != nil {
		return false, err
	}
	if ipVerifier != true {
		return false, errors.New("verify iproof failed")
	}
	return true, nil
}

//the following struct is only used for encode transferproof into a string, since the transferproof struct has many *bn256.G1 and *big.Int

//tfProofMarshal struct
type TfPMarshal struct {
	BpA       []byte
	BpS       []byte
	A         []byte
	B         []byte
	CLnTilde  []byte
	CRnTilde  []byte
	C0Tilde   []byte
	Dtilde    []byte
	Tfy0Tilde []byte
	TfgTilde  []byte
	CXtilde   []byte
	TfyXtilde []byte
	Tff       []*big.Int
	TfzA      *big.Int
	BpT1      []byte
	BpT2      []byte
	BpThat    *big.Int
	BpMu      *big.Int
	Tfc       *big.Int
	TfSsk     *big.Int
	TfSr      *big.Int
	TfSb      *big.Int
	TfStau    *big.Int
	Ia        *big.Int
	Ib        *big.Int
	LS        []byte
	RS        []byte
}

//transferproof marshal and unmarshal method
func TfProofMarshal(TfProof *TransferProof) string {
	tfM := new(TfPMarshal)
	tfM.BpA = TfProof.BpA.Marshal()
	tfM.BpS = TfProof.BpS.Marshal()
	tfM.A = TfProof.A.Marshal()
	tfM.B = TfProof.B.Marshal()
	tfM.CLnTilde = WdPointMarshal(TfProof.CLnTilde)
	tfM.CRnTilde = WdPointMarshal(TfProof.CRnTilde)
	tfM.C0Tilde = WdPointMarshal(TfProof.C0Tilde)
	tfM.Dtilde = WdPointMarshal(TfProof.Dtilde)
	tfM.Tfy0Tilde = WdPointMarshal(TfProof.Tfy0Tilde)
	tfM.TfgTilde = WdPointMarshal(TfProof.TfgTilde)
	tfM.CXtilde = WdPointMarshal(TfProof.CXtilde)
	tfM.TfyXtilde = WdPointMarshal(TfProof.TfyXtilde)
	tfM.Tff = TfProof.Tff
	tfM.TfzA = new(big.Int).Set(TfProof.TfzA)
	tfM.BpT1 = TfProof.BpT1.Marshal()
	tfM.BpT2 = TfProof.BpT2.Marshal()
	tfM.BpThat = new(big.Int).Set(TfProof.BpThat)
	tfM.BpMu = new(big.Int).Set(TfProof.BpMu)
	tfM.Tfc = new(big.Int).Set(TfProof.Tfc)
	tfM.TfSsk = new(big.Int).Set(TfProof.TfSsk)
	tfM.TfSr = new(big.Int).Set(TfProof.TfSr)
	tfM.TfSb = new(big.Int).Set(TfProof.TfSb)
	tfM.TfStau = new(big.Int).Set(TfProof.TfStau)
	tfM.Ia = new(big.Int).Set(TfProof.IpProof.a)
	tfM.Ib = new(big.Int).Set(TfProof.IpProof.b)
	tfM.LS = WdPointMarshal(TfProof.IpProof.LS)
	tfM.RS = WdPointMarshal(TfProof.IpProof.RS)
	res, err := rlp.EncodeToBytes(tfM)
	/// fmt.Printf("the marshal proof is :%v\n", TfProof)
	if err != nil {
		return "Marshal transfer proof failed"
	}
	str := hexutil.Encode(res)
	return str
}

func TfProofUnMarshal(str string) (*TransferProof, error) {
	res, err := hexutil.Decode(str)
	if err != nil {
		return nil, errors.New("transfer proof unmarshal failed")
	}
	tfM := new(TfPMarshal)
	rlp.DecodeBytes(res, tfM)
	tfProof := new(TransferProof)
	tfProof.BpA = new(bn256.G1)
	tfProof.BpA.Unmarshal(tfM.BpA)
	tfProof.BpS = new(bn256.G1)
	tfProof.BpS.Unmarshal(tfM.BpS)
	tfProof.A = new(bn256.G1)
	tfProof.A.Unmarshal(tfM.A)
	tfProof.B = new(bn256.G1)
	tfProof.B.Unmarshal(tfM.B)
	tfProof.CLnTilde, _ = WdPointUnMarshal(tfM.CLnTilde)
	tfProof.CRnTilde, _ = WdPointUnMarshal(tfM.CRnTilde)
	tfProof.C0Tilde, _ = WdPointUnMarshal(tfM.C0Tilde)
	tfProof.Dtilde, _ = WdPointUnMarshal(tfM.Dtilde)
	tfProof.Tfy0Tilde, _ = WdPointUnMarshal(tfM.Tfy0Tilde)
	tfProof.TfgTilde, _ = WdPointUnMarshal(tfM.TfgTilde)
	tfProof.CXtilde, _ = WdPointUnMarshal(tfM.CXtilde)
	tfProof.TfyXtilde, _ = WdPointUnMarshal(tfM.TfyXtilde)
	tfProof.Tff = tfM.Tff
	tfProof.TfzA = tfM.TfzA
	tfProof.BpT1 = new(bn256.G1)
	tfProof.BpT1.Unmarshal(tfM.BpT1)
	tfProof.BpT2 = new(bn256.G1)
	tfProof.BpT2.Unmarshal(tfM.BpT2)
	tfProof.BpThat = tfM.BpThat
	tfProof.BpMu = tfM.BpMu
	tfProof.Tfc = tfM.Tfc
	tfProof.TfSsk = tfM.TfSsk
	tfProof.TfSr = tfM.TfSr
	tfProof.TfSb = tfM.TfSb
	tfProof.TfStau = tfM.TfStau

	ipProof := new(InnerProductProof)
	ipProof.a = tfM.Ia
	ipProof.b = tfM.Ib
	ipProof.LS, _ = WdPointUnMarshal(tfM.LS)
	ipProof.RS, _ = WdPointUnMarshal(tfM.RS)
	tfProof.IpProof = ipProof
	/// fmt.Printf("the unmarshal proof is :%v\n", tfProof)
	return tfProof, nil

}
