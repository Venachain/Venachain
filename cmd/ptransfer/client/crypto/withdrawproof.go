package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/bn256"
	"github.com/PlatONEnetwork/PlatONE-Go/rlp"
	"math/big"
)

type WithdrawStatement struct {
	cLn    *bn256.G1
	cRn    *bn256.G1
	pub    *bn256.G1
	epoch  *big.Int
	sender []byte
	u      *bn256.G1
}

type WithdrawWitness struct {
	priv  *big.Int
	vDiff *AggBpWitness
}

type WithdrawProof struct {
	A       *bn256.G1
	S       *bn256.G1
	T1      *bn256.G1
	T2      *bn256.G1
	tHat    *big.Int
	mu      *big.Int
	c       *big.Int
	ssk     *big.Int
	sb      *big.Int
	stau    *big.Int
	ipProof *InnerProductProof
}

type WdProofMarshal struct {
	A    []byte
	S    []byte
	T1   []byte
	T2   []byte
	That *big.Int
	Mu   *big.Int
	C    *big.Int
	Ssk  *big.Int
	Sb   *big.Int
	Stau *big.Int
	Ia   *big.Int
	Ib   *big.Int
	LS   []byte
	RS   []byte
}

func NewWithdrawWit(Priv *big.Int, Vdiff *AggBpWitness) *WithdrawWitness {
	return &WithdrawWitness{priv: Priv, vDiff: Vdiff}
}

func NewWithdrawStatement(cLn, cRn, pub, u *bn256.G1, epoch *big.Int, sender []byte) *WithdrawStatement {
	return &WithdrawStatement{
		cLn:    cLn,
		cRn:    cRn,
		pub:    pub,
		epoch:  epoch,
		sender: sender,
		u:      u,
	}

}
func GenerateBurnyz(A, S, cLn, cRn, pub *bn256.G1, epoch *big.Int, sender []byte) (y, z *big.Int) {
	message := cLn.Marshal()
	message = append(message, cRn.Marshal()...)
	message = append(message, pub.Marshal()...)
	message = append(message, epoch.Bytes()...)
	message = append(message, sender...)
	digest := sha256.Sum256(message)
	digestTmp := append(digest[:], A.Marshal()...)
	yTmp := append(digestTmp, S.Marshal()...)
	y = GenerateChallenge(yTmp, ORDER)
	zTmp := append(yTmp, y.Bytes()...)
	z = GenerateChallenge(zTmp, ORDER)
	return y, z
}

//tEval = T1^x T2^{x^2}
func CalculatetEval(T1, T2 *bn256.G1, x *big.Int) *bn256.G1 {
	T1x := new(bn256.G1).ScalarMult(T1, x)
	T2x2 := new(bn256.G1).ScalarMult(T2, new(big.Int).Mod(new(big.Int).Mul(x, x), ORDER))
	return new(bn256.G1).Add(T1x, T2x2)
}

func GenerateWithdrawCh(x *big.Int, Rpub, Rb, RtHat, Ru *bn256.G1) *big.Int {
	//generate c = H(x||Rpub||Rb||RtHat||Ru)
	message := x.Bytes()
	message = append(message, Rpub.Marshal()...)
	message = append(message, Rb.Marshal()...)
	message = append(message, RtHat.Marshal()...)
	message = append(message, Ru.Marshal()...)
	dig := sha256.Sum256(message)
	c := hashToInt(dig[:], BN256())
	return c
}

func SigmaProve(g, h, cRn *bn256.G1, epoch, x, z, taux *big.Int, wit *WithdrawWitness) (c, ssk, sb, stau *big.Int, err error) {
	//generate random numer ksk, kb, ktau
	rsk, err := rand.Int(rand.Reader, ORDER)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	rb, err := rand.Int(rand.Reader, ORDER)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	rtau, err := rand.Int(rand.Reader, ORDER)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	//generate R_pub, R_b, R_t, R_tHat, R_u
	Rpub := new(bn256.G1).ScalarMult(g, rsk)
	zzrsk := new(big.Int).Mul(rsk, new(big.Int).Mod(new(big.Int).Mul(z, z), ORDER))
	zzrsk.Mod(zzrsk, ORDER)
	Rb := new(bn256.G1).Add(new(bn256.G1).ScalarMult(g, rb), new(bn256.G1).ScalarMult(cRn, zzrsk))
	RtHat := new(bn256.G1).Add(new(bn256.G1).Neg(new(bn256.G1).ScalarMult(g, rb)), new(bn256.G1).ScalarMult(h, rtau))
	gepoch := MapIntoGroup("zether" + epoch.String())
	Ru := new(bn256.G1).ScalarMult(gepoch, rsk)
	//generate c = H(x||Rpub||Rb||RtHat||Ru)
	//calculate c
	c = GenerateWithdrawCh(x, Rpub, Rb, RtHat, Ru)
	//generate ssk=rsk+c*priv; sb=rb+c*vdiff*z^2; stau=rtau+c*taux
	ssk = new(big.Int).Add(rsk, new(big.Int).Mod(new(big.Int).Mul(c, wit.priv), ORDER))
	ssk.Mod(ssk, ORDER)
	czz := new(big.Int).Mul(c, new(big.Int).Mod(new(big.Int).Mul(z, z), ORDER))
	czz.Mod(czz, ORDER)
	sb = new(big.Int).Add(rb, new(big.Int).Mod(new(big.Int).Mul(czz, wit.vDiff.v[0]), ORDER))
	sb.Mod(sb, ORDER)
	stau = new(big.Int).Add(rtau, new(big.Int).Mod(new(big.Int).Mul(c, taux), ORDER))
	stau.Mod(stau, ORDER)
	return c, ssk, sb, stau, nil
}

func WithdrawProve(cLn, cRn, pub *bn256.G1, epoch *big.Int, sender []byte, witness *WithdrawWitness) (*WithdrawProof, error) {
	aggBP := NewAggBpStatement(int64(1), AggBp())
	wdProof := new(WithdrawProof)
	//compute A and S
	aL, aR, alpha, A, _ := GenerateAggA(aggBP, witness.vDiff)
	sL, sR, rho, S, _ := GenerateAggS(aggBP)
	wdProof.A = A
	wdProof.S = S
	//calculate challenge y,z
	y, z := GenerateBurnyz(wdProof.A, wdProof.S, cLn, cRn, pub, epoch, sender)
	//compute t1, t2 and commitment T1, T2
	T1, T2, tau1, tau2, err := GenerateAggT1T2(aggBP, aL, aR, sL, sR, y, z)
	if err != nil {
		return nil, err
	}
	wdProof.T1 = T1
	wdProof.T2 = T2
	//calculate x = H(T1||T2||z)
	x := Generatex(T1, T2, z)
	//compute l,r,tHat
	lx, rx, tHat, err := GenerateAggtHat(aL, aR, sL, sR, x, y, z, aggBP.bpParam.n, aggBP.m)
	if err != nil {
		return nil, err
	}
	wdProof.tHat = tHat
	//compute tauX = tau1*x + tau2*x^2
	tauxcoff := []*big.Int{big.NewInt(0), tau1, tau2}
	tauX := PolyEvaluate(tauxcoff, x)
	//compute mu = alpha + x*rho
	mu := new(big.Int).Mod(new(big.Int).Add(alpha, new(big.Int).Mul(x, rho)), ORDER)
	wdProof.mu = mu
	// generate sigma proof
	c, ssk, sb, stau, err := SigmaProve(aggBP.bpParam.g, aggBP.bpParam.h, cRn, epoch, x, z, tauX, witness)
	if err != nil {
		return nil, err
	}
	wdProof.c = c
	wdProof.ssk = ssk
	wdProof.sb = sb
	wdProof.stau = stau
	//compute innerproductproof
	//calculate hprime and P, len(hprime) = n*m
	n := aggBP.bpParam.n
	m := aggBP.m
	hprime := GenerateHprime(aggBP.bpParam.hVector, y)
	P, err := UpdateAggP(A, S, aggBP.bpParam.h, aggBP.bpParam.gVector, hprime, x, y, z, mu, n, m)
	if err != nil {
		return nil, err
	}
	//generate u, which can no rely on c, you can use any group point
	str := c.String()
	u := MapIntoGroup(str)
	//innerproduct proof
	ipProof, err := IpProve(n*m, aggBP.bpParam.gVector, hprime, P, u, lx, rx, tHat, c)
	if err != nil {
		return nil, err
	}
	wdProof.ipProof = ipProof

	return wdProof, nil
}

func WithdrawVerify(statement *WithdrawStatement, proof *WithdrawProof) (bool, error) {
	aggBP := NewAggBpStatement(int64(1), AggBp())
	//calculate nonce y,z
	y, z := GenerateBurnyz(proof.A, proof.S, statement.cLn, statement.cRn, statement.pub, statement.epoch, statement.sender)
	// generate delta(y,z)
	delta := ComputeAggDelta(y, z, aggBP.bpParam.n, aggBP.m)
	//calculate t = tHat - delta(y,z)
	t := new(big.Int).Sub(proof.tHat, delta)
	//calculate x
	x := Generatex(proof.T1, proof.T2, z)
	//calculate tEval = T1*x+T2*{x^2}
	tEval := CalculatetEval(proof.T1, proof.T2, x)
	//Ay = g^ssk.y^{-c}
	gsk := new(bn256.G1).ScalarMult(G, proof.ssk)
	negc := new(big.Int).Sub(ORDER, proof.c)
	yNegc := new(bn256.G1).ScalarMult(statement.pub, negc)
	Ay := new(bn256.G1).Add(gsk, yNegc)
	// Ab = g^sb.[CRn^ssk.CLn^-c]^{z^2}
	gsb := new(bn256.G1).ScalarMult(aggBP.bpParam.g, proof.sb)
	CRnssk := new(bn256.G1).ScalarMult(statement.cRn, proof.ssk)
	CLnNegc := new(bn256.G1).ScalarMult(statement.cLn, negc)
	Ab := new(bn256.G1).Add(gsb, new(bn256.G1).ScalarMult(new(bn256.G1).Add(CRnssk, CLnNegc), new(big.Int).Mod(new(big.Int).Mul(z, z), ORDER)))
	// At = {[g^t.tEval^-1]^c}.h^stau.g^{-sb}
	tEvalNegc := new(bn256.G1).ScalarMult(tEval, negc)
	gtNegc := new(bn256.G1).ScalarMult(aggBP.bpParam.g, new(big.Int).Mod(new(big.Int).Mul(t, proof.c), ORDER))
	At := new(bn256.G1).Add(tEvalNegc, gtNegc)
	At = new(bn256.G1).Add(At, new(bn256.G1).ScalarMult(aggBP.bpParam.h, proof.stau))
	At = new(bn256.G1).Add(At, new(bn256.G1).ScalarMult(G, new(big.Int).Sub(ORDER, proof.sb)))
	//calculate Au
	gepoch := MapIntoGroup("zether" + statement.epoch.String())
	Au := new(bn256.G1).Add(new(bn256.G1).ScalarMult(gepoch, proof.ssk), new(bn256.G1).ScalarMult(statement.u, negc))
	//fmt.Printf("verify Au : %v\n", Au)
	//calculate c
	c := GenerateWithdrawCh(x, Ay, Ab, At, Au)
	if c.Cmp(proof.c) != 0 {
		return false, errors.New("Sigma protocol challenge equality failure.")
	}
	//compute innerproductproof
	//calculate hprime and P, len(hprime) = n*m
	n := aggBP.bpParam.n
	m := aggBP.m
	hprime := GenerateHprime(aggBP.bpParam.hVector, y)
	P, err := UpdateAggP(proof.A, proof.S, aggBP.bpParam.h, aggBP.bpParam.gVector, hprime, x, y, z, proof.mu, n, m)
	if err != nil {
		return false, err
	}
	//generate u, which can no rely on c, you can use any group point
	str := c.String()
	u := MapIntoGroup(str)
	ipVerifier, err := IpVerify(n*m, aggBP.bpParam.gVector, hprime, P, u, proof.ipProof, proof.tHat, c)
	if err != nil {
		return false, err
	}
	return ipVerifier, nil
}

func WdPointMarshal(point []*bn256.G1) []byte {
	pLen := len(point)
	res := make([]byte, 0)
	for i := 0; i < pLen; i++ {
		res = append(res, point[i].Marshal()...)
	}
	return res
}

func WdPointUnMarshal(res []byte) ([]*bn256.G1, error) {
	resLen := len(res)
	if resLen%64 != 0 {
		return nil, errVectorLength
	}
	n := resLen / 64
	point := make([]*bn256.G1, n)
	for i := 0; i < n; i++ {
		point[i] = new(bn256.G1)
		_, err := point[i].Unmarshal(res[i*(resLen/n) : (i+1)*(resLen/n)])
		if err != nil {
			return nil, err
		}
	}
	return point, nil
}

func (wdProof *WithdrawProof) WdProofMarshal() string {
	wdM := new(WdProofMarshal)
	wdM.A = wdProof.A.Marshal()
	wdM.S = wdProof.S.Marshal()
	wdM.T1 = wdProof.T1.Marshal()
	wdM.T2 = wdProof.T2.Marshal()
	wdM.That = new(big.Int).Set(wdProof.tHat)
	wdM.Mu = new(big.Int).Set(wdProof.mu)
	wdM.C = new(big.Int).Set(wdProof.c)
	wdM.Ssk = new(big.Int).Set(wdProof.ssk)
	wdM.Sb = new(big.Int).Set(wdProof.sb)
	wdM.Stau = new(big.Int).Set(wdProof.stau)
	wdM.Ia = new(big.Int).Set(wdProof.ipProof.a)
	wdM.Ib = new(big.Int).Set(wdProof.ipProof.b)
	wdM.LS = WdPointMarshal(wdProof.ipProof.LS)
	wdM.RS = WdPointMarshal(wdProof.ipProof.RS)
	res, err := rlp.EncodeToBytes(wdM)
	if err != nil {
		return "Marshal proof failed"
	}
	str := hexutil.Encode(res)
	return str
}

func WdProofUnMarshal(str string) (*WithdrawProof, error) {
	res, err := hexutil.Decode(str)
	if err != nil {
		return nil, err
	}
	wdM := new(WdProofMarshal)
	rlp.DecodeBytes(res, wdM)
	wdProof := new(WithdrawProof)
	ipProof := new(InnerProductProof)
	wdProof.A = new(bn256.G1)
	wdProof.A.Unmarshal(wdM.A)
	wdProof.S = new(bn256.G1)
	wdProof.S.Unmarshal(wdM.S)
	wdProof.T1 = new(bn256.G1)
	wdProof.T1.Unmarshal(wdM.T1)
	wdProof.T2 = new(bn256.G1)
	wdProof.T2.Unmarshal(wdM.T2)
	wdProof.tHat = wdM.That
	wdProof.mu = wdM.Mu
	wdProof.c = wdM.C
	wdProof.ssk = wdM.Ssk
	wdProof.sb = wdM.Sb
	wdProof.stau = wdM.Stau
	wdProof.ipProof = ipProof
	ipProof.a = wdM.Ia
	ipProof.b = wdM.Ib
	ipProof.LS, _ = WdPointUnMarshal(wdM.LS)
	ipProof.RS, _ = WdPointUnMarshal(wdM.RS)
	wdProof.ipProof = ipProof
	return wdProof, nil
}
