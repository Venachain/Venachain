package crypto

import (
	"errors"
	"math/big"
	"sync"

	"github.com/Venachain/Venachain/common/hexutil"
	"github.com/Venachain/Venachain/crypto/bn256"
	"github.com/Venachain/Venachain/rlp"
)

type AggBpStatement struct {
	//m is the number of values
	m       int64
	bpParam BulletProofParams
}

type AggBpWitness struct {
	v []*big.Int
}

type AggBulletProof struct {
	//a tuple of V[i]=g*v[i] + h*gamma[i]
	V                 []*bn256.G1
	A                 *bn256.G1
	S                 *bn256.G1
	T1                *bn256.G1
	T2                *bn256.G1
	taux              *big.Int
	mu                *big.Int
	tHat              *big.Int
	innerProductProof *InnerProductProof
}

type IPproofMarshal struct {
	A  *big.Int
	B  *big.Int
	LS []byte
	RS []byte
}

//innerproof marshal
func innerProductProofMarshal(proof *InnerProductProof) []byte {
	ipM := new(IPproofMarshal)
	ipM.A = new(big.Int).Set(proof.a)
	ipM.B = new(big.Int).Set(proof.b)
	ipM.LS = WdPointMarshal(proof.LS)
	ipM.RS = WdPointMarshal(proof.RS)
	res, err := rlp.EncodeToBytes(ipM)
	if err != nil {
		return nil
	}
	//str := hexutil.Encode(res)
	return res
}

//innerproof unmarshal
func innerProductProofUnMarshal(str []byte) (*InnerProductProof, error) {
	proof := new(InnerProductProof)
	ipM := new(IPproofMarshal)
	rlp.DecodeBytes(str, ipM)
	proof.a = ipM.A
	proof.b = ipM.B
	proof.LS, _ = WdPointUnMarshal(ipM.LS)
	proof.RS, _ = WdPointUnMarshal(ipM.RS)
	return proof, nil
}

//aggproof marshal and unmarshal method
func AggProofMarshal(proof *AggBulletProof) (string, error) {
	apM := new(ABProofMarshal)
	apM.V = WdPointMarshal(proof.V)
	apM.A = proof.A.Marshal()
	apM.S = proof.S.Marshal()
	apM.T1 = proof.T1.Marshal()
	apM.T2 = proof.T2.Marshal()
	apM.Taux = new(big.Int).Set(proof.taux)
	apM.Mu = new(big.Int).Set(proof.mu)
	apM.THat = new(big.Int).Set(proof.tHat)
	apM.InnerProductProof = innerProductProofMarshal(proof.innerProductProof)
	res, _ := rlp.EncodeToBytes(apM)
	str := hexutil.Encode(res)
	return str, nil

}
func AggProofUnMarshal(str string) (*AggBulletProof, error) {
	res, err := hexutil.Decode(str)
	if err != nil {
		return nil, err
	}
	apM := new(ABProofMarshal)
	proof := new(AggBulletProof)
	err = rlp.DecodeBytes(res, apM)
	if err != nil {
		return nil, err
	}
	//if apM == nil{
	//	return err, nil
	//}
	proof.V, _ = WdPointUnMarshal(apM.V)
	proof.A = new(bn256.G1)
	proof.A.Unmarshal(apM.A)
	proof.S = new(bn256.G1)
	proof.S.Unmarshal(apM.S)
	proof.T1 = new(bn256.G1)
	proof.T1.Unmarshal(apM.T1)
	proof.T2 = new(bn256.G1)
	proof.T2.Unmarshal(apM.T2)
	proof.taux = apM.Taux
	proof.mu = apM.Mu
	proof.tHat = apM.THat
	proof.innerProductProof, _ = innerProductProofUnMarshal(apM.InnerProductProof)
	return proof, nil
}

//aggbulletproof marshal struct
type ABProofMarshal struct {
	//a tuple of V[i]=g*v[i] + h*gamma[i]
	V    []byte
	A    []byte
	S    []byte
	T1   []byte
	T2   []byte
	Taux *big.Int
	//taux              []byte
	Mu                *big.Int
	THat              *big.Int
	InnerProductProof []byte
}

//generate string proof
func AggBpProve_s(instance *AggBpStatement, wit []*big.Int) (string, error) {
	witness := &AggBpWitness{wit}
	proof, err := AggBpProve(instance, witness)
	if err != nil {
		return "Generate proof error!", err
	}
	res, _ := AggProofMarshal(proof)
	return res, nil
}

//the proof input is string
func AggBpVerify_s(proof string, instance *AggBpStatement) (bool, error) {
	aggproof, err := AggProofUnMarshal(proof)
	if err != nil {
		return false, err
	}
	res, err := AggBpVerify(aggproof, instance)
	if err != nil {
		return false, err
	}
	return res, nil
}

//only for withdrawproof of clientcmd
func NewAggBpWitness(Wit []*big.Int) *AggBpWitness {
	return &AggBpWitness{v: Wit}
}

//compute commitment for a tuple of v[i], i in [1,m]
func (aggreBp *AggBulletProof) VectorvCommit(param BulletProofParams, v []*big.Int, m int64) (vectorV []*bn256.G1, vectorGamma []*big.Int, err error) {
	//1. params check
	if int64(len(v)) != m {
		return nil, nil, errors.New("invalid input params!")
	}
	//2. generate PedersenCommitment for all v and generate commit for all v
	vVectorCommit := make([]*PedersenCommitment, m)
	vectorV = make([]*bn256.G1, m)
	vectorGamma = make([]*big.Int, m)
	var i int64
	for i = 0; i < m; i++ {
		vVectorCommit[i] = NewPedersenCommitment(param, v[i])
		vectorV[i] = vVectorCommit[i].Commit()
		vectorGamma[i] = vVectorCommit[i].random
	}
	return vectorV, vectorGamma, nil
}

//compute aL, where <2^n, aL[(j-1)n:jn-1]> = v[j]
func VectorDecompose(v []*big.Int, u, n, m int64) []*big.Int {
	aL := make([]*big.Int, 0)
	var j int64
	for j = 1; j <= m; j++ {
		tmp := Decompose(v[j-1], u, n)
		aL = append(aL, tmp...)
	}
	return aL
}

//compute polynomial l(X) coefficients l0, since l1 = sL, we do not add into this function
func LPolyCoeff(aL []*big.Int, z *big.Int, nm int64) (lZero []*big.Int, err error) {
	if int64(len(aL)) != nm {
		return nil, errVectorLength
	}
	lZero = make([]*big.Int, nm)
	znm := VectorScalarMul(PowerOf(one, nm), z)
	lZero, err = VectorSub(aL, znm)
	if err != nil {
		return nil, err
	}
	return lZero, nil
}

// compute polynomial r(x) coefficients
func RpolyCoeff(aR, sR []*big.Int, y, z *big.Int, n, m int64) (rZero []*big.Int, rOne []*big.Int, err error) {
	//0. param check
	nm := n * m
	if int64(len(aR)) != nm || int64(len(sR)) != nm {
		return nil, nil, errVectorLength
	}
	//1. compute aR + z * 1^nm
	znm := VectorScalarMul(PowerOf(one, nm), z)
	aRznm, err := VectorAdd(aR, znm)
	if err != nil {
		return nil, nil, err
	}
	//2. compute y^nm and y^nm O (aR + z*1^nm)
	ynm := PowerOf(y, nm)
	rZeroTmp, err := VectorHadamard(ynm, aRznm)
	if err != nil {
		return nil, nil, err
	}
	// 3. compute \sum[z^(j+1) * (0^((j-1)*n)||2^n||0^((m-j)*n))]
	var j int64
	rZeroPrime := PowerOf(big.NewInt(0), nm)
	for j = 1; j <= m; j++ {
		zVector := make([]*big.Int, 0)
		zExp := new(big.Int).Exp(z, big.NewInt(1+j), ORDER)
		index0 := (j - 1) * n
		index1 := (m - j) * n
		left := PowerOf(big.NewInt(0), index0)
		mid := PowerOf(big.NewInt(2), n)
		right := PowerOf(big.NewInt(0), index1)
		zVector = append(zVector, left...)
		zVector = append(zVector, mid...)
		zVector = append(zVector, right...)
		tmp := VectorScalarMul(zVector, zExp)
		rZeroPrime, _ = VectorAdd(tmp, rZeroPrime)
	}
	rZero, err = VectorAdd(rZeroTmp, rZeroPrime)
	if err != nil {
		return nil, nil, err
	}
	rOne, err = VectorHadamard(ynm, sR)
	if err != nil {
		return nil, nil, err
	}
	return rZero, rOne, nil
}

//compute t0 = l0 * r0
func Computet0(aL, aR, sR []*big.Int, y, z *big.Int, n, m int64) (*big.Int, error) {
	l0, err := LPolyCoeff(aL, z, n*m)
	if err != nil {
		return nil, err
	}
	r0, _, err := RpolyCoeff(aR, sR, y, z, n, m)
	if err != nil {
		return nil, err
	}
	t0, err := VectorInnerProduct(l0, r0)
	if err != nil {
		return nil, err
	}
	return t0, nil
}

//compute t1 = l0*r1+l1*r0
func Computet1(aL, aR, sL, sR []*big.Int, y, z *big.Int, n, m int64) (*big.Int, error) {
	l0, err := LPolyCoeff(aL, z, n*m)
	if err != nil {
		return nil, err
	}
	r0, r1, err := RpolyCoeff(aR, sR, y, z, n, m)
	if err != nil {
		return nil, err
	}
	Tmp1, err := VectorInnerProduct(l0, r1)
	if err != nil {
		return nil, err
	}
	Tmp2, err := VectorInnerProduct(sL, r0)
	if err != nil {
		return nil, err
	}
	t1 := new(big.Int).Add(Tmp1, Tmp2)
	t1.Mod(t1, ORDER)
	if err != nil {
		return nil, err
	}
	return t1, nil
}

//compute t2 = l1*r1
func Computet2(sL, aR, sR []*big.Int, y, z *big.Int, n, m int64) (*big.Int, error) {
	_, r1, err := RpolyCoeff(aR, sR, y, z, n, m)
	if err != nil {
		return nil, err
	}
	t2, err := VectorInnerProduct(sL, r1)
	if err != nil {
		return nil, err
	}
	return t2, nil
}

//compute l(x)
func ComputeAggLx(aL, sL []*big.Int, x, z *big.Int, nMulm int64) (lx []*big.Int, err error) {
	l0, err := LPolyCoeff(aL, z, nMulm)
	if err != nil {
		return nil, err
	}
	sLMulx := VectorScalarMul(sL, x)
	lx, err = VectorAdd(l0, sLMulx)
	if err != nil {
		return nil, err
	}
	return lx, nil
}

//compute r(x)
func ComputeAggRx(aR, sR []*big.Int, x, y, z *big.Int, n, m int64) (rx []*big.Int, err error) {
	r0, r1, err := RpolyCoeff(aR, sR, y, z, n, m)
	if err != nil {
		return nil, err
	}
	r1Multx := VectorScalarMul(r1, x)
	rx, err = VectorAdd(r0, r1Multx)
	if err != nil {
		return nil, err
	}
	return rx, nil
}

//compute P, where P = A .S^x. g^(-z).hprime ^{ z.y^nm + sum(hprime[(j-1)*n:jn-1]^(z^(j+1)*2^n))} * h{-mu}
func UpdateAggP(A, S, h *bn256.G1, gVector, hprime []*bn256.G1, x, y, z, mu *big.Int, n, m int64) (*bn256.G1, error) {
	//1. A + x*S
	ASx := new(bn256.G1).Add(A, new(bn256.G1).ScalarMult(S, x))
	//2. g^(-z) -> len = n*m
	tmp := VectorScalarExp(gVector, new(big.Int).Neg(z))
	gNegz := new(bn256.G1).ScalarBaseMult(big.NewInt(0))
	var i int64
	for i = 0; i < n*m; i++ {
		//gNegz = new(bn256.G1).Add(tmp[i], gNegz)
		gNegz.Add(tmp[i], gNegz)
	}
	//3. h' * (z*y^nm)
	ynm := PowerOf(y, n*m)
	zynm := VectorScalarMul(ynm, z)
	hPrimeZynm, err := VectorScalarMulSum(zynm, hprime)
	if err != nil {
		return nil, err
	}
	// 4. sum(hprime[(j-1)*n:jn-1]^(z^(j+1)*2^n))
	var j int64
	hPrimeSum := new(bn256.G1).ScalarBaseMult(big.NewInt(0))
	for j = 1; j <= m; j++ {
		zExp := new(big.Int).Exp(z, big.NewInt(j+1), ORDER)
		twoExpn := PowerOf(big.NewInt(2), n)
		zTwo := VectorScalarMul(twoExpn, zExp)
		indexL := (j - 1) * n
		indexR := j*n - 1
		hPrimeTmp, err := VectorScalarMulSum(zTwo, hprime[indexL:indexR+1])
		if err != nil {
			return nil, err
		}
		//hPrimeSum = new(bn256.G1).Add(hPrimeTmp, hPrimeSum)
		hPrimeSum.Add(hPrimeTmp, hPrimeSum)
	}
	//5. hPrimeRes = h'*(z*y^nm) + hPrimeSum
	hPrimeRes := new(bn256.G1).Add(hPrimeZynm, hPrimeSum)
	//6. compute h*(-mu)
	hMu := new(bn256.G1).ScalarMult(h, mu)
	hNegMu := new(bn256.G1).Neg(hMu)
	//7. compute P
	P := new(bn256.G1).Add(new(bn256.G1).Add(ASx, gNegz), new(bn256.G1).Add(hPrimeRes, hNegMu))
	return P, nil
}

//compute delta = (z-z^2)*<1^nm,y^nm>- sum[z^(j+2)*<1^n,2^n>]
func ComputeAggDelta(y, z *big.Int, n, m int64) *big.Int {
	nm := n * m
	//z^2
	zz := new(big.Int).Mul(z, z)
	zz.Mod(zz, ORDER)
	//result = (z-z^2)
	result := new(big.Int).Sub(z, zz)
	result.Mod(result, ORDER)
	//1^nm
	onenm := PowerOf(one, nm)
	//y^nm
	ynm := PowerOf(y, nm)
	//< 1^nm, y^nm >
	OneYnm, _ := VectorInnerProduct(onenm, ynm)
	//(z-z^2) * < 1^nm, y^nm >
	result.Mul(result, OneYnm)
	result.Mod(result, ORDER)
	//sum_1^m z^{j+2}<1^n,2^n>
	//< 1^n, 2^n>
	onen := PowerOf(big.NewInt(1), n)
	twon := PowerOf(big.NewInt(2), n)
	OneTwon, _ := VectorInnerProduct(onen, twon)
	sum := big.NewInt(0)
	var j int64
	for j = 1; j <= m; j++ {
		zExp := new(big.Int).Exp(z, big.NewInt(2+j), ORDER)
		tmp := new(big.Int).Mul(zExp, OneTwon)
		sum.Add(sum, tmp)
		sum.Mod(sum, ORDER)
	}
	result.Sub(result, sum)
	result.Mod(result, ORDER)
	return result
}

//compute aggregate bulletproof A
func GenerateAggA(instance *AggBpStatement, witness *AggBpWitness) (aL, aR []*big.Int, alpha *big.Int, A *bn256.G1, err error) {
	//compute aL st. <2^n, aL[(j-1)n:jn-1]> = v[j] and aR
	aL = VectorDecompose(witness.v, 2, instance.bpParam.n, instance.m)
	aR = ComputeAR(aL)
	//compute A, S
	aLRcommit := NewVectorCommitment(instance.bpParam, aL, aR)
	alpha = aLRcommit.random
	A, err = aLRcommit.Commit()
	if err != nil {
		return nil, nil, nil, nil, errVectorLength
	}
	return aL, aR, alpha, A, nil
}

//compute aggregate bulletproof S
func GenerateAggS(instance *AggBpStatement) (sL, sR []*big.Int, rho *big.Int, S *bn256.G1, err error) {
	sL = sampleRandomVector(instance.bpParam.n * instance.m)
	sR = sampleRandomVector(instance.bpParam.n * instance.m)
	sLRcommit := NewVectorCommitment(instance.bpParam, sL, sR)
	rho = sLRcommit.random
	S, err = sLRcommit.Commit()
	if err != nil {
		return nil, nil, nil, nil, errVectorLength
	}
	return sL, sR, rho, S, nil
}

//compute aggregate bulletproof challenge y and z
func GenerateAggyz(A, S *bn256.G1) (y, z *big.Int) {
	yTmp := append(A.Marshal(), S.Marshal()...)
	y = GenerateChallenge(yTmp, ORDER)
	zTmp := append(yTmp, y.Bytes()...)
	z = GenerateChallenge(zTmp, ORDER)
	return y, z
}

//compute aggregate bulletproof T1 and T2
func GenerateAggT1T2(instance *AggBpStatement, aL, aR, sL, sR []*big.Int, y, z *big.Int) (T1, T2 *bn256.G1, tau1, tau2 *big.Int, err error) {
	t1, err := Computet1(aL, aR, sL, sR, y, z, instance.bpParam.n, instance.m)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	t2, err := Computet2(sL, aR, sR, y, z, instance.bpParam.n, instance.m)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	t1Commit := NewPedersenCommitment(instance.bpParam, t1)
	tau1 = t1Commit.random
	T1 = t1Commit.Commit()

	t2Commit := NewPedersenCommitment(instance.bpParam, t2)
	tau2 = t2Commit.random
	T2 = t2Commit.Commit()
	return T1, T2, tau1, tau2, nil
}

//compute aggregate bulletproof taux
func GenerateAggTaux(tau1, tau2, x, z *big.Int, m int64, vectorGamma []*big.Int) (tauX *big.Int) {
	tau1x := new(big.Int).Mul(tau1, x)
	tau2x := new(big.Int).Mul(tau2, new(big.Int).Mul(x, x))
	var j int64
	sum := big.NewInt(0)
	for j = 1; j <= m; j++ {
		zExp := new(big.Int).Exp(z, big.NewInt(1+j), ORDER)
		zExpGamma := new(big.Int).Mul(zExp, vectorGamma[j-1])
		sum = new(big.Int).Add(sum, zExpGamma)
	}
	tauX = new(big.Int).Add(new(big.Int).Add(tau1x, tau2x), sum)
	tauX.Mod(tauX, ORDER)
	return tauX
}

//compute aggregate bulletproof tHat
func GenerateAggtHat(aL, aR, sL, sR []*big.Int, x, y, z *big.Int, n, m int64) (lx, rx []*big.Int, tHat *big.Int, err error) {
	lx, err = ComputeAggLx(aL, sL, x, z, n*m)
	if err != nil {
		return nil, nil, nil, err
	}
	rx, err = ComputeAggRx(aR, sR, x, y, z, n, m)
	if err != nil {
		return nil, nil, nil, err
	}
	tHat, err = VectorInnerProduct(lx, rx)
	if err != nil {
		return nil, nil, nil, err
	}
	return lx, rx, tHat, nil
}

//compute aggregate bulletproof InnerProductProof
func GenerateAggIpp(instance *AggBpStatement, A, S *bn256.G1, x, y, z, mu, tHat *big.Int, lx, rx []*big.Int) (*InnerProductProof, error) {
	//calculate hprime and P, len(hprime) = n*m
	n := instance.bpParam.n
	m := instance.m
	hprime := GenerateHprime(instance.bpParam.hVector, y)
	P, err := UpdateAggP(A, S, instance.bpParam.h, instance.bpParam.gVector, hprime, x, y, z, mu, n, m)
	if err != nil {
		return nil, err
	}
	//generate u
	str := x.String()
	u := MapIntoGroup(str)
	//innerproduct proof
	proof, err := IpProve(n*m, instance.bpParam.gVector, hprime, P, u, lx, rx, tHat, x)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

//generate aggregate bulletproof which can be considered as 4 rounds prove phases and 2 rounds of challenge
func AggBpProve(instance *AggBpStatement, witness *AggBpWitness) (*AggBulletProof, error) {
	aggBp := new(AggBulletProof)
	//params check n,m > 0;
	n := instance.bpParam.n
	m := instance.m
	nm := n * m
	if int64(len(instance.bpParam.gVector)) != nm || int64(len(instance.bpParam.hVector)) != nm {
		return nil, errVectorLength
	}
	if int64(len(witness.v)) != instance.m {
		return nil, errVectorLength
	}

	//prove-0 : generate AggBulletProof.V
	vectorV, vectorGamma, err := aggBp.VectorvCommit(instance.bpParam, witness.v, m)
	if err != nil {
		return nil, err
	}
	//prove-1 : compute A and S
	aL, aR, alpha, A, err := GenerateAggA(instance, witness)
	if err != nil {
		return nil, err
	}
	sL, sR, rho, S, err := GenerateAggS(instance)
	if err != nil {
		return nil, err
	}
	//challenge-1 : compute challenge y and z
	y, z := GenerateAggyz(A, S)
	//prove-2 : compute t1, t2 and commitment T1, T2
	T1, T2, tau1, tau2, err := GenerateAggT1T2(instance, aL, aR, sL, sR, y, z)
	if err != nil {
		return nil, err
	}
	//challenge-2 : compute challenge x = H(T1||T2||z), where ComputeX are define in innerproductproof file
	x := Generatex(T1, T2, z)
	//prove-3 : compute tauX, mu, tHat
	//prove-3.1  : compute tauX = tau1*x + tau2*x^2 + sum(z^(1+j)*gamma_j)
	tauX := GenerateAggTaux(tau1, tau2, x, z, m, vectorGamma)
	//prove-3.2 : compute mu
	mu := new(big.Int).Mod(new(big.Int).Add(alpha, new(big.Int).Mul(x, rho)), ORDER)
	//prove-3.3 : compute tHat
	lx, rx, tHat, err := GenerateAggtHat(aL, aR, sL, sR, x, y, z, n, m)
	if err != nil {
		return nil, err
	}
	//prove-4 : compute innerproductproof
	ipproof, err := GenerateAggIpp(instance, A, S, x, y, z, mu, tHat, lx, rx)
	if err != nil {
		return nil, err
	}
	aggBp.V = vectorV
	aggBp.A = A
	aggBp.S = S
	aggBp.T1 = T1
	aggBp.T2 = T2
	aggBp.taux = tauX
	aggBp.mu = mu
	aggBp.tHat = tHat
	aggBp.innerProductProof = ipproof
	return aggBp, nil
}

//verify proof.tHat is valid
func AggBpVerify(proof *AggBulletProof, instance *AggBpStatement) (bool, error) {
	tHatVerifier, err := proof.VerifyAggtHat(instance)
	if err != nil {
		return false, err
	}
	//fmt.Println("tHatVerifier:", tHatVerifier)
	ipVerifier, err := VerifyAggIpp(proof, instance)
	if err != nil {
		return false, err
	}
	//fmt.Println("ipVerifier:", ipVerifier)

	if tHatVerifier && ipVerifier {
		return true, nil
	}
	return false, nil
}

//verify proof.InnerProductProof is valid
func (proof *AggBulletProof) VerifyAggtHat(instance *AggBpStatement) (bool, error) {
	tHat := proof.tHat
	taux := proof.taux
	g := instance.bpParam.g
	h := instance.bpParam.h
	V := proof.V
	T1 := proof.T1
	T2 := proof.T2
	//compute challenge y and z
	yTmp := append(proof.A.Marshal(), proof.S.Marshal()...)
	y := GenerateChallenge(yTmp, ORDER)
	zTmp := append(yTmp, y.Bytes()...)
	z := GenerateChallenge(zTmp, ORDER)
	//calculate x
	x := Generatex(T1, T2, z)
	//delta
	delta := ComputeAggDelta(y, z, instance.bpParam.n, instance.m)
	//g^that * h^taux
	gtHat := new(bn256.G1).ScalarMult(g, tHat)
	htaux := new(bn256.G1).ScalarMult(h, taux)
	left := new(bn256.G1).Add(gtHat, htaux)
	//g^delta(y,z)
	gdelta := new(bn256.G1).ScalarMult(g, delta)
	//V^{z^2 * z^m}
	z2 := new(big.Int).Mul(z, z)
	z2.Mod(z2, ORDER)
	zm := PowerOf(z, instance.m)
	zmz2 := VectorScalarMul(zm, z2)
	//this function is define in InnerProductProof file
	vectorVsum, err := VectorScalarMulSum(zmz2, V)
	if err != nil {
		return false, err
	}
	right := new(bn256.G1).Add(gdelta, vectorVsum)
	// T1^x, T2^{x^2}
	T1Mulx := new(bn256.G1).ScalarMult(T1, x)
	xx := new(big.Int).Mul(x, x)
	xx.Mod(xx, ORDER)
	T2Mulxx := new(bn256.G1).ScalarMult(T2, xx)
	right = new(bn256.G1).Add(right, T1Mulx)
	right = new(bn256.G1).Add(right, T2Mulxx)
	// point to int
	l := new(big.Int).SetBytes(left.Marshal())
	r := new(big.Int).SetBytes(right.Marshal())
	if l.Cmp(r) == 0 {
		return true, nil
	}
	return false, nil
}

//verify the aggregate bulletproof
func VerifyAggIpp(proof *AggBulletProof, instance *AggBpStatement) (bool, error) {
	n := instance.bpParam.n
	m := instance.m
	nm := n * m
	//compute challenge y and z and x
	y, z := GenerateAggyz(proof.A, proof.S)
	x := Generatex(proof.T1, proof.T2, z)
	//compute h'
	hprime := GenerateHprime(instance.bpParam.hVector, y)
	//compute P
	P, err := UpdateAggP(proof.A, proof.S, instance.bpParam.h, instance.bpParam.gVector, hprime, x, y, z, proof.mu, n, m)
	if err != nil {
		return false, err
	}
	// generate u
	str := x.String()
	u := MapIntoGroup(str)
	ipVerifier, err := IpVerify(nm, instance.bpParam.gVector, hprime, P, u, proof.innerProductProof, proof.tHat, x)
	if err != nil {
		return false, nil
	}
	return ipVerifier, nil
}

//only for withdrawproof
func NewAggBpStatement(m int64, aggbppram AggBpStatement) *AggBpStatement {
	result := &AggBpStatement{
		m:       m,
		bpParam: BulletProofParams{},
	}
	result.bpParam.n = aggbpparam.bpParam.n
	nm := m * aggbpparam.bpParam.n
	result.bpParam.gVector = aggbpparam.bpParam.gVector[:nm]
	result.bpParam.hVector = aggbpparam.bpParam.hVector[:nm]
	result.bpParam.g = aggbpparam.bpParam.g
	result.bpParam.h = aggbpparam.bpParam.h

	return result
}

func GetGenerator(param AggBpStatement) (*bn256.G1, *bn256.G1, []*bn256.G1, []*bn256.G1) {
	return param.bpParam.g, param.bpParam.h, param.bpParam.gVector, param.bpParam.hVector
}

var aggbpparam AggBpStatement
var agginitbp sync.Once

func GenerateAggBpStatement(m, n int64) *AggBpStatement {
	aggbpparam = AggBpStatement{}
	aggbpparam.m = m
	aggbpparam.bpParam.n = n
	aggbpparam.bpParam.g = MapIntoGroup("g")
	aggbpparam.bpParam.h = MapIntoGroup("h")
	nm := aggbpparam.bpParam.n * aggbpparam.m
	aggbpparam.bpParam.gVector = make([]*bn256.G1, nm)
	aggbpparam.bpParam.hVector = make([]*bn256.G1, nm)
	for i := 0; i < int(nm); i++ {
		aggbpparam.bpParam.gVector[i] = MapIntoGroup("venachain" + "g" + string(rune(i)))
		aggbpparam.bpParam.hVector[i] = MapIntoGroup("venachain" + "h" + string(rune(i)))
	}
	return &aggbpparam
}

func GenerateAggBpStatement_range(m, n int64, range_hash []byte) *AggBpStatement {
	aggbpparam = AggBpStatement{}
	aggbpparam.m = m
	aggbpparam.bpParam.n = n

	aggbpparam.bpParam.g = MapIntoGroup("g" + string(range_hash))
	aggbpparam.bpParam.h = MapIntoGroup("h" + string(range_hash))
	nm := aggbpparam.bpParam.n * aggbpparam.m
	aggbpparam.bpParam.gVector = make([]*bn256.G1, nm)
	aggbpparam.bpParam.hVector = make([]*bn256.G1, nm)
	for i := 0; i < int(nm); i++ {
		aggbpparam.bpParam.gVector[i] = MapIntoGroup("venachain" + "g" + string(rune(i)))
		aggbpparam.bpParam.hVector[i] = MapIntoGroup("venachain" + "h" + string(rune(i)))
	}
	return &aggbpparam
}

func agginitbpparam() {
	aggbpparam = AggBpStatement{}
	aggbpparam.m = 2
	aggbpparam.bpParam.n = 16
	aggbpparam.bpParam.g = MapIntoGroup("g")
	aggbpparam.bpParam.h = MapIntoGroup("h")
	nm := aggbpparam.bpParam.n * aggbpparam.m
	aggbpparam.bpParam.gVector = make([]*bn256.G1, nm)
	aggbpparam.bpParam.hVector = make([]*bn256.G1, nm)
	for i := 0; i < int(nm); i++ {
		aggbpparam.bpParam.gVector[i] = MapIntoGroup("venachain" + "g" + string(rune(i)))
		aggbpparam.bpParam.hVector[i] = MapIntoGroup("venachain" + "h" + string(rune(i)))
	}
}

func AggBp() AggBpStatement {
	agginitbp.Do(agginitbpparam)
	return aggbpparam
}
