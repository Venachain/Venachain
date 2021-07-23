package crypto

import (
	"crypto/sha256"
	"errors"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/bn256/cloudflare"
	"math/big"
)

type InnerProductStatement struct {
	//range bit length
	n int64
	// two generators slice: g, h
	gVector []*bn256.G1
	hVector []*bn256.G1
	//a point P
	p *bn256.G1
	// a fix point
	u *bn256.G1
}

type InnerProductWitness struct {
	// as, bs st. <as,bs> = c where c is the instance
	as []*big.Int
	bs []*big.Int
}

type InnerProductProof struct {
	//n == 1, proof result: a,b
	a *big.Int
	b *big.Int
	//len(LS)=len(RS)=log_2(n)
	LS []*bn256.G1
	RS []*bn256.G1
}

//compute res[i] = {gVector[i]}^x, i in [1,n]
func VectorScalarExp(gVector []*bn256.G1, x *big.Int) []*bn256.G1 {
	length := len(gVector)
	x.Mod(x, ORDER)
	res := make([]*bn256.G1, length)
	for i := 0; i < length; i++ {
		res[i] = new(bn256.G1).ScalarMult(gVector[i], x)
	}
	return res
}

func Bij(i, j uint64) (int64, error) {
	if i < 1 || j < 1 {
		return 0, errors.New("value i is invalid!")
	}
	x := i - 1
	if 1&(x>>(j-1)) == 1 {
		return 1, nil
	}
	return -1, nil
}

/*
 instance = (gVector, hVector, P, c=<aVector, bVector>); witness = (aVector, bVector)
 after protocol1: P'= P + x * c * u, where x is challenge, c is the innerproduct
 instance' = (gVector, hVector, u*x, P'); witness' = (aVector, bVector)
*/
func Protocol1(instance *InnerProductStatement, c *big.Int, previousChallenge *big.Int) (*InnerProductStatement, *big.Int) {
	//1. generate the protocol1 challenge x = H(previousChallenge)
	h := sha256.Sum256(previousChallenge.Bytes())
	x := hashToInt(h[:], BN256())
	x.Mod(x, ORDER)
	//2. compute u * x
	ux := new(bn256.G1).ScalarMult(instance.u, x)
	//3. compute u * x * c = ux * c
	uxc := new(bn256.G1).ScalarMult(ux, c)
	//4. compute P'= P + uxMulc
	Pprime := new(bn256.G1).Add(uxc, instance.p)
	//5. 赋值(gVector, hVector, P', uMulx)
	return &InnerProductStatement{instance.n, instance.gVector, instance.hVector, Pprime, ux}, x
}

func IpProve(n int64, GVector, HVector []*bn256.G1, p, u *bn256.G1, l, r []*big.Int, tHat *big.Int, challenge *big.Int) (proof *InnerProductProof, err error) {
	instance := new(InnerProductStatement)
	instance.n = n
	instance.gVector = GVector
	instance.hVector = HVector
	instance.p = p
	instance.u = u
	witness := new(InnerProductWitness)
	witness.as = l
	witness.bs = r
	proof, err = InnerProductProver(instance, witness, tHat, challenge)
	if err != nil {
		return nil, errors.New("Generate innerProductProof failed!")
	}
	return proof, nil
}

func IpVerify(n int64, gVector, hVector []*bn256.G1, p, u *bn256.G1, ipProof *InnerProductProof, tHat, previousChallenge *big.Int) (bool, error) {
	ipStatement := new(InnerProductStatement)
	ipStatement.n = n
	ipStatement.gVector = gVector
	ipStatement.hVector = hVector
	ipStatement.p = p
	ipStatement.u = u
	res, err := InnerProductVerifier(ipStatement, ipProof, tHat, previousChallenge)
	if err != nil {
		return false, nil
	}
	return res, nil

}

func InnerProductProver(instance *InnerProductStatement, witness *InnerProductWitness, tHat *big.Int, previousChallenge *big.Int) (proof *InnerProductProof, err error) {
	//1. params check
	if instance.n <= 0 {
		return nil, errors.New("n must be greater than zero")
	}
	if int64(len(instance.gVector)) != instance.n || int64(len(instance.hVector)) != instance.n {
		return nil, errVectorLength
	}
	if len(witness.as) != len(witness.bs) {
		return nil, errVectorLength
	}
	// n must be power of 2
	if IsPowOfTwo(instance.n) == false {
		return nil, errors.New("n must be power of 2")
	}

	//2. protocol 1
	ipStatement, ch := Protocol1(instance, tHat, previousChallenge)

	//3. protocol 2
	//3.1. recursive generate proof : now instance = (n, gVector,hVector,Pprime, uMulx )=ipStatement
	Ls := make([]*bn256.G1, 0)
	Rs := make([]*bn256.G1, 0)
	proof, err = recursiveIpProof(witness, ipStatement, Ls, Rs, ch)
	if err != nil {
		return nil, errors.New("generate innerProduct proof failed!")
	}
	return proof, nil
}

func recursiveIpProof(witness *InnerProductWitness, instance *InnerProductStatement, Ls, Rs []*bn256.G1, ch *big.Int) (*InnerProductProof, error) {
	ipProof := new(InnerProductProof)
	//1. compute n == 1, return a, b; verifier can check P = g^ah^b
	n := int64(len(witness.as))
	if n == 1 {
		ipProof.a = witness.as[0]
		ipProof.b = witness.bs[0]
		ipProof.LS = Ls
		ipProof.RS = Rs
		return ipProof, nil
	}
	//2. n != 1, compute a[:n'], a[n':], b[:n'], b[n':], g[:n'],g[n':],h[:n'],h[n':]
	nPrime := n / 2
	asLeft := witness.as[:nPrime]
	asRight := witness.as[nPrime:]
	bsLeft := witness.bs[:nPrime]
	bsRight := witness.bs[nPrime:]
	gLeft := instance.gVector[:nPrime]
	gRight := instance.gVector[nPrime:]
	hLeft := instance.hVector[:nPrime]
	hRight := instance.hVector[nPrime:]
	//3. compute cL=<asLeft, bsRight>, cR=<asRight, bsLeft>
	cL, err := VectorInnerProduct(asLeft, bsRight)
	if err != nil {
		return nil, err
	}
	cR, err := VectorInnerProduct(asRight, bsLeft)
	if err != nil {
		return nil, err
	}
	//4. compute L
	L1, err := VectorScalarMulSum(asLeft, gRight)
	if err != nil {
		return nil, errVectorLength
	}
	L2, err := VectorScalarMulSum(bsRight, hLeft)
	if err != nil {
		return nil, errVectorLength
	}
	L3 := new(bn256.G1).ScalarMult(instance.u, cL)
	L := new(bn256.G1).Add(L1, L2)
	L = new(bn256.G1).Add(L3, L)
	//5. compute R
	R1, err := VectorScalarMulSum(asRight, gLeft)
	if err != nil {
		return nil, errVectorLength
	}
	R2, err := VectorScalarMulSum(bsLeft, hRight)
	if err != nil {
		return nil, errVectorLength
	}
	R3 := new(bn256.G1).ScalarMult(instance.u, cR)
	R := new(bn256.G1).Add(R1, R2)
	R = new(bn256.G1).Add(R3, R)
	//6. compute challenge x
	x := Generatex(L, R, ch)
	//7. compute gPrime, hPrime, PPrime, aPrime, bPrime
	//7.1. compute x^(-1)
	xInv := new(big.Int).ModInverse(x, ORDER)
	//7.2. compute gPrime and hPrime
	gPrime, err := VectorEcAdd(VectorScalarExp(gLeft, xInv), VectorScalarExp(gRight, x))
	if err != nil {
		return nil, err
	}
	hPrime, err := VectorEcAdd(VectorScalarExp(hLeft, x), VectorScalarExp(hRight, xInv))
	if err != nil {
		return nil, err
	}
	//7.3. firstly compute asPrime, bsPrime, so that we do not need to compute x^2, which may change x
	aPrimeLeft := VectorScalarMul(asLeft, x)
	aPrimeRight := VectorScalarMul(asRight, xInv)
	aPrime, err := VectorAdd(aPrimeLeft, aPrimeRight)
	if err != nil {
		return nil, err
	}
	bPrimeLeft := VectorScalarMul(bsLeft, xInv)
	bPrimeRight := VectorScalarMul(bsRight, x)
	bPrime, err := VectorAdd(bPrimeLeft, bPrimeRight)
	if err != nil {
		return nil, err
	}
	//7.4. compute PPrime
	//compute x^2 and x^2 * L
	xx := new(big.Int).Mul(x, x)
	xx.Mod(xx, ORDER)
	PPrime := new(bn256.G1).ScalarMult(L, xx)
	PPrime = new(bn256.G1).Add(PPrime, instance.p)
	//compute x^(-2) = x^(-1) * x^(-1)
	xxInv := new(big.Int).Mul(xInv, xInv)
	xxInv = new(big.Int).Mod(xxInv, ORDER)
	//compute PPrime = x^2*L + P + x^(-2)*R
	PPrime = new(bn256.G1).Add(new(bn256.G1).ScalarMult(R, xxInv), PPrime)

	//8. generate recursive parameters
	witnessPrime := &InnerProductWitness{aPrime, bPrime}
	instancePrime := &InnerProductStatement{int64(nPrime), gPrime, hPrime, PPrime, instance.u}
	Ls = append(Ls, L)
	Rs = append(Rs, R)
	//9. recursive generate proofs
	ipProof, err = recursiveIpProof(witnessPrime, instancePrime, Ls, Rs, x)
	if err != nil {
		return nil, err
	}
	return ipProof, nil
}

func InnerProductVerifier(instance *InnerProductStatement, proof *InnerProductProof, c *big.Int, previousChallenge *big.Int) (bool, error) {
	//1. params check
	if int64(len(instance.hVector)) != instance.n || int64(len(instance.gVector)) != instance.n {
		return false, errors.New("invalid instance")
	}
	if proof.a == nil || proof.b == nil {
		return false, errors.New("invalid proof")
	}
	if len(proof.LS) != len(proof.RS) {
		return false, errors.New("invalid proof length")
	}
	//2. execute protocol1: ipStatement = (n, gVector, hVector, Pprime, uMulx), ch = H(previousChallenge)
	ipStatement, ch := Protocol1(instance, c, previousChallenge)
	//3. if n==1, check P = proof.a * g + proof.b * h + c * u
	if len(ipStatement.gVector) == 1 {
		ga := new(bn256.G1).ScalarMult(ipStatement.gVector[0], proof.a)
		hb := new(bn256.G1).ScalarMult(ipStatement.hVector[0], proof.b)
		P := new(bn256.G1).Add(new(bn256.G1).Add(ga, hb), new(bn256.G1).ScalarMult(ipStatement.u, c))
		PLeft := hashToInt(ipStatement.p.Marshal(), BN256())
		PRight := hashToInt(P.Marshal(), BN256())
		if PLeft.Cmp(PRight) == 0 {
			return true, nil
		}
		return false, nil
	}
	/*
		4. compute all challenges in recursiveProof
		previousChallenge is the challenge x of protocol, while H(previousChallenge) is the recursiveIpProof challenge source,
		using H(previousChallenge) and proof.LS as well as proof.RS, we can compute all challenges in recursiveIpProof
	*/
	//4.1. compute x_0 = H(L0||R0||ch)
	nLog := len(proof.LS)
	chVector := make([]*big.Int, nLog)
	chVector[0] = Generatex(proof.LS[0], proof.RS[0], ch)
	//4.2. compute x[j] = H(L[j]||R[j]||x[j-1])
	for j := 1; j < nLog; j++ {
		chVector[j] = Generatex(proof.LS[j], proof.RS[j], chVector[j-1])
	}
	//5. compute right verify equation = Pprime + sum{LS[j]*x[j]^2 + RS[j] * x[j]^-2}
	resRight := ipStatement.p
	for j := 0; j < nLog; j++ {
		//compute x^2 and x^-2
		xx := new(big.Int).Mul(chVector[j], chVector[j])
		xx.Mod(xx, ORDER)
		xxInv := new(big.Int).ModInverse(xx, ORDER)
		L := new(bn256.G1).ScalarMult(proof.LS[j], xx)
		R := new(bn256.G1).ScalarMult(proof.RS[j], xxInv)
		res := new(bn256.G1).Add(L, R)
		resRight = new(bn256.G1).Add(res, resRight)
	}

	//6. compute left verify equation a*s*g + b*sInv*h + a*b*u
	//6.1. compute sVector: sVector[i] = \Pi(x[j]^b(i,j))
	sVector := make([]*big.Int, ipStatement.n)
	var i, j uint64
	for i = 1; i <= uint64(ipStatement.n); i++ {
		stemp := big.NewInt(1)
		for j = 1; j <= uint64(nLog); j++ {
			bij, err := Bij(i, uint64(nLog)+1-j)
			if err != nil {
				return false, err
			}
			if bij == 1 {
				stemp.Mul(chVector[j-1], stemp)
				stemp.Mod(stemp, ORDER)
			} else {
				v := new(big.Int).ModInverse(chVector[j-1], ORDER)
				stemp.Mul(v, stemp)
				stemp.Mod(stemp, ORDER)
			}
		}
		sVector[i-1] = stemp
	}
	//6.2. compute a*s*gVector
	asVector := VectorScalarMul(sVector, proof.a)
	gas, err := VectorScalarMulSum(asVector, ipStatement.gVector)
	if err != nil {
		return false, errVectorLength
	}
	//6.3. compute b*s^-1*hVector
	sInv := VectorInv(sVector)
	bsInv := VectorScalarMul(sInv, proof.b)
	hbsInv, err := VectorScalarMulSum(bsInv, ipStatement.hVector)
	if err != nil {
		return false, errVectorLength
	}
	//6.3. compute a*s*g + b*s^-1*h
	resLeft := new(bn256.G1).Add(gas, hbsInv)
	//6.4. compute u*a*b
	ab := new(big.Int).Mul(proof.a, proof.b)
	ab.Mod(ab, ORDER)
	uMulab := new(bn256.G1).ScalarMult(ipStatement.u, ab)
	//6.5. compute left verify equation a*s*g + b*sInv*h + a*b*u
	resLeft = new(bn256.G1).Add(uMulab, resLeft)

	//7. compare resLeft == resRight
	left := hashToInt(resLeft.Marshal(), BN256())
	right := hashToInt(resRight.Marshal(), BN256())
	if left.Cmp(right) == 0 {
		return true, nil
	}
	return false, errors.New("InnerProductProof verify failed!")
}
