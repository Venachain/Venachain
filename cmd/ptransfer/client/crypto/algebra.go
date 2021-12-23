package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/crypto"
	"github.com/Venachain/Venachain/crypto/bn256"
	"github.com/pkg/errors"
)

var errVectorLength = errors.New("Invalid vector length")
var errBnUnmarshal = errors.New("bn256: not enough data")
var ORDER = BN256().N

//compute a global generator G instead of BN128 generator g
var G = MapIntoGroup("g")

/*
PedersenCommitment structure
*/
type PedersenCommitment struct {
	p      BulletProofParams
	value  *big.Int
	random *big.Int
}

/*
VectorCommitment structure
*/
type VectorCommitment struct {
	p      BulletProofParams
	aL     []*big.Int
	aR     []*big.Int
	random *big.Int
}

/*
VectorAdd computes vector addition: a[]+b[]
*/
func VectorAdd(a, b []*big.Int) ([]*big.Int, error) {
	if len(a) != len(b) {
		return nil, errVectorLength
	}
	result := make([]*big.Int, len(a))
	for i := 0; i < len(a); i++ {
		result[i] = new(big.Int).Add(a[i], b[i])
		result[i] = new(big.Int).Mod(result[i], ORDER)
	}

	return result, nil
}

/*
VectorScalarMul computes vector scalar multiplication: b*a[]
*/
func VectorScalarMul(a []*big.Int, b *big.Int) []*big.Int {
	result := make([]*big.Int, len(a))
	for i := 0; i < len(a); i++ {
		result[i] = new(big.Int).Mul(a[i], b)
		result[i] = new(big.Int).Mod(result[i], ORDER)
	}
	return result
}

/*
VectorNeg computes vector neg: -a[]
*/
func VectorNeg(a []*big.Int) []*big.Int {
	result := VectorScalarMul(a, big.NewInt(-1))
	return result
}

/*
VectorSub computes vector sub: a[] - b[]
*/
func VectorSub(a, b []*big.Int) ([]*big.Int, error) {
	if len(a) != len(b) {
		return nil, errVectorLength
	}
	result := make([]*big.Int, len(a))
	for i := 0; i < len(a); i++ {
		result[i] = new(big.Int).Sub(a[i], b[i])
		result[i] = new(big.Int).Mod(result[i], ORDER)
	}
	return result, nil
}

/*
VectorEcAdd computes vector add: a[] + b[],where a[i] is *bn256.G1
*/
func VectorEcAdd(a, b []*bn256.G1) ([]*bn256.G1, error) {
	if len(a) != len(b) {
		return nil, errVectorLength
	}
	result := make([]*bn256.G1, len(a))
	for i := 0; i < len(a); i++ {
		result[i] = new(bn256.G1).Add(a[i], b[i])
	}

	return result, nil
}

/*
VectorInnerProduct computes vector innerproduct: <a[],b[]>
*/
func VectorInnerProduct(a, b []*big.Int) (*big.Int, error) {
	if len(a) != len(b) {
		return nil, errVectorLength
	}
	result := big.NewInt(0)
	for i := 0; i < len(a); i++ {
		result = new(big.Int).Add(result, new(big.Int).Mul(a[i], b[i]))
		result = new(big.Int).Mod(result, ORDER)
	}
	return result, nil
}

/*
VectorHadamard computes vector Hadamard: a[]。b[]
*/
func VectorHadamard(a, b []*big.Int) ([]*big.Int, error) {
	if len(a) != len(b) {
		return nil, errVectorLength
	}
	result := make([]*big.Int, len(a))
	for i := 0; i < len(a); i++ {
		result[i] = new(big.Int).Mul(a[i], b[i])
		result[i] = new(big.Int).Mod(result[i], ORDER)
	}
	return result, nil
}

/*
NewPedersenCommitment returns a PedersenCommitment
*/
func NewPedersenCommitment(p BulletProofParams, x *big.Int) *PedersenCommitment {
	r, _ := randFieldElement(BN256(), rand.Reader)
	c := &PedersenCommitment{
		p:      p,
		value:  x,
		random: r,
	}
	return c
}

/*
Commit returns a PedersenCommitment: V = r*h + v*g
*/
func (commit *PedersenCommitment) Commit() *bn256.G1 {
	Hr := new(bn256.G1).ScalarMult(commit.p.h, commit.random)
	Gv := new(bn256.G1).ScalarMult(commit.p.g, commit.value)
	result := new(bn256.G1).Add(Hr, Gv)
	return result
}

/*
NewVectorCommitment returns a VectorCommitment
*/
func NewVectorCommitment(p BulletProofParams, a, b []*big.Int) *VectorCommitment {
	r, _ := randFieldElement(BN256(), rand.Reader)
	c := &VectorCommitment{
		p:      p,
		aL:     a,
		aR:     b,
		random: r,
	}
	return c
}

/*
Commit returns a VectorCommitment: V = r*h + aL*G + aR*H
*/
func (commit *VectorCommitment) Commit() (*bn256.G1, error) {
	Hr := new(bn256.G1).ScalarMult(commit.p.h, commit.random)
	if len(commit.aL) != len(commit.p.gVector) {
		return nil, errVectorLength
	}
	result := Hr
	for i := 0; i < len(commit.aL); i++ {
		result = new(bn256.G1).Add(result, new(bn256.G1).ScalarMult(commit.p.gVector[i], commit.aL[i]))
	}
	for i := 0; i < len(commit.aR); i++ {
		result = new(bn256.G1).Add(result, new(bn256.G1).ScalarMult(commit.p.hVector[i], commit.aR[i]))
	}
	return result, nil
}

/*
PolyEvaluate returns a polynomial: c[0]+c[1]*x + c[2]*x^2 +...
*/
func PolyEvaluate(coefficients []*big.Int, x *big.Int) *big.Int {
	result := new(big.Int).Set(coefficients[0])
	current := big.NewInt(1)
	for i := 1; i < len(coefficients); i++ {
		current.Mul(current, x)
		current.Mod(current, ORDER)
		result.Add(result, new(big.Int).Mul(current, coefficients[i]))
		result.Mod(result, ORDER)

	}
	return result
}

/*
PolyEvaluateField returns a polynomial: c[0]+c[1]*x + c[2]*x^2 +...
*/
func PolyEvaluateField(coefficients []*big.Int, x *big.Int) *big.Int {
	result := new(big.Int).Set(coefficients[0])
	current := big.NewInt(1)
	for i := 1; i < len(coefficients); i++ {
		current.Mul(current, x)
		current.Mod(current, BN256().P)
		result.Add(result, new(big.Int).Mul(current, coefficients[i]))
		result.Mod(result, BN256().P)

	}
	return result
}

func PolyProduct(a, b []*big.Int, field *big.Int) []*big.Int {
	alen := len(a)
	blen := len(b)
	result := PowerOf(big.NewInt(0), int64(alen+blen-1))
	for i := 0; i < alen; i++ {
		for j := 0; j < blen; j++ {
			tmp := new(big.Int).Mul(a[i], b[j])
			tmp.Mod(tmp, field)
			result[i+j] = new(big.Int).Mod(new(big.Int).Add(result[i+j], tmp), field)
		}
	}

	return result
}

/*
a^n =(a^0,a^1...,a^{n-1})
*/
func PowerOf(a *big.Int, n int64) []*big.Int {
	result := make([]*big.Int, n)
	current := big.NewInt(1)
	if n == 0 {
		return result
	} else if n < 0 {
		return nil
	}
	if a.Cmp(zero) == 0 {
		result[0] = big.NewInt(0)
	} else {
		result[0] = current
	}
	var i int64
	for i = 1; i < n; i++ {
		current = new(big.Int).Mul(current, a)
		current = new(big.Int).Mod(current, ORDER)
		result[i] = current
	}
	return result
}

/*
Decompose receives as input a bigint x and outputs an array of integers such that
x = sum(xi.u^i), i.e. it returns the decomposition of x into base u.
*/
func Decompose(v *big.Int, u int64, l int64) []*big.Int {
	result := make([]*big.Int, l)
	tmp := new(big.Int).Set(v)
	for i := int64(0); i < l; i++ {
		result[i] = new(big.Int).Mod(tmp, big.NewInt(u))
		tmp.Div(tmp, big.NewInt(u))
		if tmp.Cmp(zero) != 1 {
			i++
			for i < l {
				result[i] = big.NewInt(0)
				i++
			}
			break
		}
	}
	return result
}

/*
MapIntoGroup returns a valid elliptic curve point given as input a string.
*/
//https://datatracker.ietf.org/doc/draft-irtf-cfrg-hash-to-curve/?include_text=1
func MapIntoGroup(s string) *bn256.G1 {
	md := crypto.Keccak256([]byte(s))
	d := new(big.Int).SetBytes(md)
	x := d.Mod(d, BN256().P)
	coff := make([]*big.Int, 4)
	coff[0] = big.NewInt(3)
	coff[1] = big.NewInt(0)
	coff[2] = big.NewInt(0)
	coff[3] = big.NewInt(1)
	result := new(bn256.G1)
	for {
		ySquare := PolyEvaluateField(coff, x)
		y := new(big.Int).ModSqrt(ySquare, BN256().P)
		if y != nil {
			tmp := make([]byte, 64)
			x32 := common.LeftPadBytes(x.Bytes(), 32)
			y32 := common.LeftPadBytes(y.Bytes(), 32)
			tmp = append(x32, y32...)
			_, err := result.Unmarshal(tmp)
			if err != nil {
				fmt.Println(err)
				x.Add(x, big.NewInt(1))
				continue
			}
			return result
		}
		x.Add(x, big.NewInt(1))
		//fmt.Println(x)
	}
	return nil
}

/*
aR = aL - 1^n
*/
func ComputeAR(x []*big.Int) []*big.Int {
	result := make([]*big.Int, len(x))
	for i := 0; i < len(x); i++ {
		result[i] = new(big.Int).Sub(x[i], one)
		result[i].Mod(result[i], ORDER)
	}
	return result
}

/*
randomly choose vector sL,sR in (Zp)^n, generate num random field element,field = BN256.order
*/
func sampleRandomVector(n int64) []*big.Int {
	s := make([]*big.Int, n)
	for i := int64(0); i < n; i++ {
		s[i], _ = rand.Int(rand.Reader, ORDER)
	}
	return s
}

//compute challenge
func GenerateChallenge(msg []byte, order *big.Int) *big.Int {
	digest := sha256.Sum256(msg)
	c := hashToInt(digest[:], BN256())
	return c
}

//n ?= 2^x
func IsPowOfTwo(n int64) bool {
	if (n & (n - 1)) == 0 {
		return true
	}
	return false
}

//sVector[i]^-1=sInv[i]
func VectorInv(sVector []*big.Int) (sInv []*big.Int) {
	length := len(sVector)
	sInv = make([]*big.Int, length)
	for i := 0; i < length; i++ {
		sInv[i] = new(big.Int).ModInverse(sVector[i], ORDER)
		if sInv[i] == nil {
			return nil
		}
	}
	return sInv
}

//compute sum(aVector[i] * gVector[i])
func VectorScalarMulSum(aVector []*big.Int, gVector []*bn256.G1) (*bn256.G1, error) {
	if len(aVector) != len(gVector) {
		return nil, errVectorLength
	}
	n := len(aVector)
	res := new(bn256.G1).ScalarBaseMult(big.NewInt(0))
	for i := 0; i < n; i++ {
		aVector[i].Mod(aVector[i], ORDER)
		tmp := new(bn256.G1).ScalarMult(gVector[i], aVector[i])
		res = new(bn256.G1).Add(tmp, res)
	}
	return res, nil
}

//compute challenge x = H(T1||T2||z)
func Generatex(T1, T2 *bn256.G1, z *big.Int) *big.Int {
	msg := append(T1.Marshal(), T2.Marshal()...)
	msg = append(msg, z.Bytes()...)
	dig := sha256.Sum256(msg)
	x := hashToInt(dig[:], BN256())
	x.Mod(x, ORDER)
	return x
}
