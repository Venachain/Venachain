package crypto

import (
	"errors"
	//"errors"
	"math/big"
	"sync"

	"github.com/Venachain/Venachain/crypto/bn256"
	//"test"
)

/*
 * Copyright (C) 2020 WXBlockchain PlatONE
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

/*
This file contains the implementation of the Bulletproofs scheme proposed in the paper:
Bulletproofs: Short Proofs for Confidential Transactions and More
Benedikt Bunz, Jonathan Bootle, Dan Boneh, Andrew Poelstra, Pieter Wuille and Greg Maxwell
Asiacrypt 2008
*/

/*
BulletProofParams is the structure that stores the parameters for
the Zero Knowledge Proof system.
*/
type BulletProofParams struct {
	// n is the bit-length of the range
	n int64
	//G is the elliptic curve generator
	g *bn256.G1
	//H is a new generator
	h       *bn256.G1
	gVector []*bn256.G1
	hVector []*bn256.G1
}

/*
BulletProof structure contains the elements that are necessary for the verification
of the Zero Knowledge Proof.
*/
type BulletProof struct {
	V                 *bn256.G1
	A                 *bn256.G1
	S                 *bn256.G1
	T1                *bn256.G1
	T2                *bn256.G1
	taux              *big.Int
	mu                *big.Int
	tHat              *big.Int
	innerProductProof *InnerProductProof
}

//compute bulletproof challenge y and z
func Generateyz(A, S *bn256.G1) (y, z *big.Int) {
	yTmp := append(A.Marshal(), S.Marshal()...)
	y = GenerateChallenge(yTmp, ORDER)
	zTmp := append(yTmp, y.Bytes()...)
	z = GenerateChallenge(zTmp, ORDER)
	return y, z
}

/*
l(X) = (aL - z *1^n) +sL *X
*/
func ComputeLx(aL, sL []*big.Int, z, x *big.Int) ([]*big.Int, error) {
	n := len(aL)
	zn := VectorScalarMul(PowerOf(one, int64(n)), z)
	result, err := VectorSub(aL, zn)
	if err != nil {
		return nil, errors.New("calculate Vectorsub failed!")
	}
	result, err = VectorAdd(result, VectorScalarMul(sL, x))
	if err != nil {
		return nil, errors.New("calculate Vectoraddx failed!")
	}
	return result, nil
}

/*
rPloyX = y^n 。 (aR + z . 1^n + x . sR ) + z^2 * 2^n
*/
func ComputeRx(aR, sR []*big.Int, z, y, x *big.Int) ([]*big.Int, error) {
	n := len(aR)
	zn := VectorScalarMul(PowerOf(one, int64(n)), z)
	//aR + z . 1^n
	aRz, err := VectorAdd(aR, zn)
	if err != nil {
		return nil, errors.New("calculate Vectoradd failed!")
	}
	//x . sR
	xsR := VectorScalarMul(sR, x)
	yn := PowerOf(y, int64(n))
	// (aR + z . 1^n + x . sR )
	tmp, err := VectorAdd(aRz, xsR)
	if err != nil {
		return nil, errors.New("calculate Vectoradd failed!")
	}
	result, err := VectorHadamard(yn, tmp)
	if err != nil {
		return nil, errors.New("calculate Vectorhadamard failed!")
	}
	//z^2 * 2^n
	z2 := new(big.Int).Mul(z, z)
	z2n := VectorScalarMul(PowerOf(big.NewInt(2), int64(n)), z2.Mod(z2, ORDER))
	result, err = VectorAdd(result, z2n)
	if err != nil {
		return nil, errors.New("calculate Vectoradd failed!")
	}
	return result, nil
}

/*
delta(y,z) = (z-z^2) . < 1^n, y^n > - z^3 . < 1^n, 2^n >
*/
func ComputeDelta(y, z *big.Int, n int64) *big.Int {
	//z^2
	zz := new(big.Int).Mul(z, z)
	zz.Mod(zz, ORDER)
	//z^3
	z3 := new(big.Int).Mul(zz, z)
	z3.Mod(z3, ORDER)
	//1^n
	onen := PowerOf(one, int64(n))
	//2^n
	twon := PowerOf(big.NewInt(2), int64(n))
	//y^n
	yn := PowerOf(y, int64(n))
	//< 1^n, y^n >
	OneYn, _ := VectorInnerProduct(onen, yn)
	//< 1^n, 2^n>
	OneTwon, _ := VectorInnerProduct(onen, twon)
	// z^3 . < 1^n, 2^n >
	z3OneTwon := new(big.Int).Mul(z3, OneTwon)
	z3OneTwon.Mod(z3OneTwon, ORDER)
	//(z-z^2)
	result := new(big.Int).Sub(z, zz)
	result.Mod(result, ORDER)
	//(z-z^2) . < 1^n, y^n>
	result.Mul(result, OneYn)
	result.Mod(result, ORDER)
	result.Sub(result, z3OneTwon)
	result.Mod(result, ORDER)
	return result
}

/*
t1: < aL - z.1^n, y^n 。 sR > + < sL, y^n 。 (aR + z . 1^n) > +  <sL, z^2 * 2^n >
*/
func PolyCoefficientsT1(sL, sR, aL, aR []*big.Int, y, z *big.Int) (*big.Int, error) {
	n := int64(len(sL))
	// z. 1^n = {z,z,z....z}
	zn := VectorScalarMul(PowerOf(one, int64(n)), z)
	yn := PowerOf(y, int64(n))
	//aL - z.1^n
	aLz, err := VectorSub(aL, zn)
	if err != nil {
		return nil, errors.New("calculate vectorsub failed!")
	}
	// y^n 。 sR
	ynsR, err := VectorHadamard(yn, sR)
	if err != nil {
		return nil, errors.New("calculate vectorHadamard failed!")
	}
	// < aL - z.1^n, y^n . sR >
	sp1, err := VectorInnerProduct(aLz, ynsR)
	if err != nil {
		return nil, errors.New("calculate VectorInnerProduct failed!")
	}
	// aR + z . 1^n
	aRz, err := VectorAdd(aR, zn)
	if err != nil {
		return nil, errors.New("calculate Vectoradd failed!")
	}
	// y^n 。 (aR + z . 1^n)
	ynaRz, err := VectorHadamard(yn, aRz)
	if err != nil {
		return nil, errors.New("calculate Vectorhadmard failed!")
	}
	// < sL, y^n . (aR + z . 1^n) >
	sp2, err := VectorInnerProduct(sL, ynaRz)
	if err != nil {
		return nil, errors.New("calculate VectorInnerProduct failed!")
	}
	// 2^n
	n2 := PowerOf(big.NewInt(2), int64(n))
	// z^2 * 2^n
	z2 := new(big.Int).Mul(z, z)
	z2.Mod(z2, ORDER)
	zn2 := VectorScalarMul(n2, z2)
	//<sL, z^2 * 2^n >
	sp3, err := VectorInnerProduct(sL, zn2)
	if err != nil {
		return nil, errors.New("calculate VectorInnerProduct failed!")
	}
	result := new(big.Int).Add(sp1, sp2)
	result.Mod(result, ORDER)
	result.Add(result, sp3)
	result.Mod(result, ORDER)
	return result, nil
}

/*
compute t2: < sL, y^n 。 sR >
*/
func PolyCoefficientsT2(sL, sR []*big.Int, y *big.Int) *big.Int {
	n := int64(len(sL))
	yn := PowerOf(y, int64(n))
	ymulsR, _ := VectorHadamard(yn, sR)
	result, _ := VectorInnerProduct(sL, ymulsR)
	result.Mod(result, ORDER)
	return result
}

/*
GenerateHprime is responsible for computing generators in the following format:
[h_1, h_2^(y^-1), ..., h_n^(y^(-n+1))], where [h_1, h_2, ..., h_n] is the original
vector of generators. This method is used both by prover and verifier.
*/
func GenerateHprime(h []*bn256.G1, y *big.Int) []*bn256.G1 {
	n := len(h)
	hprime := make([]*bn256.G1, n)
	negy := new(big.Int).ModInverse(y, ORDER) //y^{-1}
	yn := PowerOf(negy, int64(n))
	for i := 0; i < n; i++ {
		hprime[i] = new(bn256.G1).ScalarMult(h[i], yn[i])
	}
	return hprime
}

/*
UpdateP compute P ,where P = A .S^x. g^(-z).hprime ^{ z.y^n + z^2.2^n} * h{-mu}
*/
func UpdateP(A, S *bn256.G1, g, hprime []*bn256.G1, h *bn256.G1, x, z, y, mu *big.Int) *bn256.G1 {
	result := new(bn256.G1).ScalarBaseMult(new(big.Int).SetInt64(0))
	sx := new(bn256.G1).ScalarMult(S, x)
	result = new(bn256.G1).Add(A, sx)
	n := len(g)
	gnegz := make([]*bn256.G1, n)
	//negz := new(big.Int).ModInverse(z,ORDER)
	negz := new(big.Int).Sub(ORDER, z)
	pgnegz := new(bn256.G1).ScalarBaseMult(new(big.Int).SetInt64(0))
	//pgnegz = new(bn256.G1).ScalarMult(g[0],negz)
	for i := 0; i < n; i++ {
		gnegz[i] = new(bn256.G1).ScalarMult(g[i], negz)
		pgnegz = new(bn256.G1).Add(pgnegz, gnegz[i])
	}
	result = new(bn256.G1).Add(result, pgnegz)
	// z.y^n
	yn := PowerOf(y, int64(n))
	zyn := VectorScalarMul(yn, z)
	//  z^2 * 2^n
	z2 := new(big.Int).Mul(z, z)
	z2.Mod(z2, ORDER)
	n2 := PowerOf(big.NewInt(2), int64(n))
	zn2 := VectorScalarMul(n2, z2)
	//z.y^n + z^2.2^n
	zy2, _ := VectorAdd(zyn, zn2)
	ph := new(bn256.G1).ScalarBaseMult(big.NewInt(0))
	hvector := make([]*bn256.G1, n)
	for i := 0; i < n; i++ {
		hvector[i] = new(bn256.G1).ScalarMult(hprime[i], zy2[i])
		//ph = new(bn256.G1).Add(ph,hvector[i])
		//ph = hvector[i]
		ph = new(bn256.G1).Add(ph, hvector[i])
	}
	result = new(bn256.G1).Add(result, ph)
	//-mu
	//negmu := new(big.Int).ModInverse(mu,ORDER)
	negmu := new(big.Int).Sub(ORDER, mu)
	hnegmu := new(bn256.G1).ScalarMult(h, negmu)
	result = new(bn256.G1).Add(result, hnegmu)
	return result
}

// Prove computes the ZK rangeproof.
func Prove(witness *big.Int, params BulletProofParams) (*BulletProof, error) {
	var proof = BulletProof{}
	//prove-0 : commitment to v -> V
	vcommit := NewPedersenCommitment(params, witness)
	V := vcommit.Commit()
	gamma := vcommit.random
	proof.V = V
	// prove-1 : compute A and S
	//1.1 A = h^{alpha}g^{aL}h^{aR}
	aL := Decompose(witness, 2, params.n)
	aR := ComputeAR(aL)
	aLRcommit := NewVectorCommitment(params, aL, aR)
	alpha := aLRcommit.random
	A, err := aLRcommit.Commit()
	if err != nil {
		return nil, errors.New("generate commitment failed!")
	}
	proof.A = A
	// 1.2 S = h^{rho}g^{sL}h^{sR}
	sL := sampleRandomVector(params.n)
	sR := sampleRandomVector(params.n)
	sLRcommit := NewVectorCommitment(params, sL, sR)
	rho := sLRcommit.random
	S, err := sLRcommit.Commit()
	if err != nil {
		return nil, errors.New("generate commitment failed!")
	}
	proof.S = S
	//challenge-1 : compute challenge y and z
	y, z := Generateyz(A, S)
	//prove-2 : compute t1, t2 and commitment T1, T2
	//2.1 compute t1, t2
	t1, err := PolyCoefficientsT1(sL, sR, aL, aR, y, z)
	if err != nil {
		return nil, errors.New("calculate PolyCoefficientsT1 failed!")
	}
	t2 := PolyCoefficientsT2(sL, sR, y)
	//2.2 compute T1, T2
	t1commit := NewPedersenCommitment(params, t1)
	tau1 := t1commit.random
	T1 := t1commit.Commit()
	proof.T1 = T1
	t2commit := NewPedersenCommitment(params, t2)
	tau2 := t2commit.random
	T2 := t2commit.Commit()
	proof.T2 = T2
	//challenge-2 : compute challenge x = H(T1||T2||z)
	x := Generatex(T1, T2, z)
	//prove-3 : compute tauX, mu, tHat
	//prove-3.1  : compute tau_x = tau2*x^2 + tau1*x + z^2*gamma
	zz := new(big.Int).Mul(z, z)
	zz.Mod(zz, ORDER)
	zzgamma := new(big.Int).Mul(zz, gamma)
	zzgamma.Mod(zzgamma, ORDER)
	var taucoff = []*big.Int{zzgamma, tau1, tau2}
	taux := PolyEvaluate(taucoff, x)
	proof.taux = taux
	//prove-3.2 : compute mu
	var mucoff = []*big.Int{alpha, rho}
	mu := PolyEvaluate(mucoff, x)
	proof.mu = mu
	//prove-3.3 : compute tHat
	delta := ComputeDelta(y, z, params.n)
	// t0 = v.z^2 + delta(y,z)
	t0 := new(big.Int).Add(delta, new(big.Int).Mul(witness, zz))
	t0.Mod(t0, ORDER)
	var tcoff = []*big.Int{t0, t1, t2}
	that := PolyEvaluate(tcoff, x)
	proof.tHat = that
	lpoly, err := ComputeLx(aL, sL, z, x)
	if err != nil {
		return nil, errors.New("calculate Polyxl failed!")
	}
	rpoly, err := ComputeRx(aR, sR, z, y, x)
	if err != nil {
		return nil, errors.New("calculate Polyxr failed!")
	}
	//prove-4 : compute innerproductproof
	hprime := GenerateHprime(params.hVector, y)
	P := UpdateP(A, S, params.gVector, hprime, params.h, x, z, y, mu)
	ipproof, err := IpProve(params.n, params.gVector, hprime, P, params.h, lpoly, rpoly, that, x)
	if err != nil {
		return nil, errors.New("generate ipproof failed!")
	}
	proof.innerProductProof = ipproof
	return &proof, nil
}

/*
 Verify g^that h^taux = V^{z^2}g^delta(y,z) T1^x T2^{x^2}
*/
func VerifyThat(proof *BulletProof, params BulletProofParams) bool {
	that := proof.tHat
	taux := proof.taux
	g := params.g
	h := params.h
	V := proof.V
	T1 := proof.T1
	T2 := proof.T2
	//calculate nonce y,z,x
	y, z := Generateyz(proof.A, proof.S)
	//calculate x
	x := Generatex(T1, T2, z)
	//delta
	delta := ComputeDelta(y, z, params.n)
	//g^that .h^taux
	that.Mod(that, ORDER)
	taux.Mod(taux, ORDER)
	gthat := new(bn256.G1).ScalarMult(g, that)
	htaux := new(bn256.G1).ScalarMult(h, taux)
	left := new(bn256.G1).Add(gthat, htaux)
	//V^{z^2}
	z2 := new(big.Int).Mul(z, z)
	z2.Mod(z2, ORDER)
	vz2 := new(bn256.G1).ScalarMult(V, z2)
	//g^delta(y,z)
	delta.Mod(delta, ORDER)
	gdelta := new(bn256.G1).ScalarMult(g, delta)
	// T1^x, T2^{x^2}
	t1x := new(bn256.G1).ScalarMult(T1, x)
	x2 := new(big.Int).Mul(x, x)
	x2.Mod(x2, ORDER)
	t2x2 := new(bn256.G1).ScalarMult(T2, x2)
	right := new(bn256.G1).Add(vz2, gdelta)
	right = new(bn256.G1).Add(right, t1x)
	right = new(bn256.G1).Add(right, t2x2)
	// point to int
	l := left.Marshal()
	r := right.Marshal()
	if new(big.Int).SetBytes(l).Cmp(new(big.Int).SetBytes(r)) == 0 {
		return true
	}
	return false
}

/*
Verify returns true if and only if the proof is valid.
*/
func (proof *BulletProof) Verify(param BulletProofParams) (bool, error) {
	tHatVerifier := VerifyThat(proof, param)
	T1 := proof.T1
	T2 := proof.T2
	A := proof.A
	S := proof.S

	//calculate nonce y,z,x
	y, z := Generateyz(proof.A, proof.S)
	//calculate x
	x := Generatex(T1, T2, z)
	ipp := proof.innerProductProof
	hprime := GenerateHprime(param.hVector, y)
	p := UpdateP(A, S, param.gVector, hprime, param.h, x, z, y, proof.mu)
	ipstatement := &InnerProductStatement{param.n, param.gVector, hprime, p, param.h}
	ipVerifier, err := InnerProductVerifier(ipstatement, ipp, proof.tHat, x)
	if err != nil {
		return false, errors.New("verify inner product failed!")
	}
	if tHatVerifier && ipVerifier {
		return true, nil
	}
	return false, nil
}

var bpparam BulletProofParams
var initbp sync.Once

/*
init params
*/
func GenerateBpParam(n int64) *BulletProofParams {
	bpparam = BulletProofParams{}
	bpparam.n = n
	bpparam.g = MapIntoGroup("g")
	bpparam.h = MapIntoGroup("h")
	bpparam.gVector = make([]*bn256.G1, bpparam.n)
	bpparam.hVector = make([]*bn256.G1, bpparam.n)
	for i := 0; i < int(bpparam.n); i++ {
		bpparam.gVector[i] = MapIntoGroup("venachain" + "g" + string(rune(i)))
		bpparam.hVector[i] = MapIntoGroup("venachain" + "h" + string(rune(i)))
	}
	return &bpparam
}

func initbpparam() {
	bpparam = BulletProofParams{}
	bpparam.n = 32
	bpparam.g = MapIntoGroup("g")
	bpparam.h = MapIntoGroup("h")
	bpparam.gVector = make([]*bn256.G1, bpparam.n)
	bpparam.hVector = make([]*bn256.G1, bpparam.n)
	for i := 0; i < int(bpparam.n); i++ {
		bpparam.gVector[i] = MapIntoGroup("venachain" + "g" + string(rune(i)))
		bpparam.hVector[i] = MapIntoGroup("venachain" + "h" + string(rune(i)))
	}
}

func Bp() BulletProofParams {
	initbp.Do(initbpparam)
	return bpparam
}
