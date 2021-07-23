package crypto

import (
	"crypto/sha256"
	"errors"
	"io"
	"math/big"
	"sync"

	"github.com/PlatONEnetwork/PlatONE-Go/crypto/bn256"
)

type Signature struct {
	R *big.Int
	S *big.Int
}

var one = new(big.Int).SetInt64(1)
var zero = new(big.Int).SetInt64(0)
var MAX = new(big.Int).SetInt64(4294967296)

type param struct {
	P       *big.Int // the order of the underlying field
	N       *big.Int // the order of the base point
	B       *big.Int // the constant of the curve equation
	Gx, Gy  *big.Int // (x,y) of the base point
	BitSize int      // the size of the underlying field
	Name    string
}

func randFieldElement(p param, rand io.Reader) (k *big.Int, err error) {
	b := make([]byte, p.BitSize/8+8)
	_, err = io.ReadFull(rand, b)
	if err != nil {
		return
	}

	k = new(big.Int).SetBytes(b)
	n := new(big.Int).Sub(p.N, one)
	k.Mod(k, n)
	k.Add(k, one)
	return
}

// GenerateKey generates private key
func GenerateKey(p param, rand io.Reader) (*big.Int, error) {
	k, err := randFieldElement(p, rand)
	if err != nil {
		return nil, err
	}
	return k, nil
}

func getPublicKey(p param, sk *big.Int) (*bn256.G1, error) {
	pub := new(bn256.G1).ScalarMult(G, sk)
	return pub, nil
}

func hashToInt(hash []byte, p param) *big.Int {
	orderBits := p.N.BitLen()
	orderBytes := (orderBits + 7) / 8
	if len(hash) > orderBytes {
		hash = hash[:orderBytes]
	}
	ret := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - orderBits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}
	return ret
}

func SchnorrTest(rand io.Reader, sk *big.Int, pk *bn256.G1, msg []byte) (*Signature, error) {
	key := new(KeyPair)
	key.sk = sk
	key.pk = pk
	return SchnorrSign(rand, key, msg)
}

func SchnorrSign(rand io.Reader, key *KeyPair, msg []byte) (*Signature, error) {
	//1. generate temporary secret key and public key
	k, err := randFieldElement(BN256(), rand)
	if err != nil || k.Cmp(zero) == 0 {
		return nil, err
	}
	K, err := getPublicKey(BN256(), k)
	if err != nil {
		return nil, err
	}

	//2. encode K,pk as []byte
	KTemp := K.Marshal()
	pkTemp := (*key).pk.Marshal()

	//3. compute h = H(msg||KTemp||pkTemp)
	message := append(msg, KTemp...)
	message = append(message, pkTemp...)
	digest := sha256.Sum256(message)

	//4. encode h into bigInt r：hashToInt
	r := hashToInt(digest[:], BN256())
	r.Mod(r, Bn256.N)

	//5. compute signature s = k + sk * r
	s := new(big.Int).Add(k, new(big.Int).Mul(r, (*key).sk))
	s.Mod(s, ORDER)
	return &Signature{r, s}, nil
}

func SchnorrVerify(pub *bn256.G1, msg []byte, sigma *Signature) bool {
	//1.签名检查
	if sigma.R.Cmp(big.NewInt(0)) == 0 || sigma.S.Cmp(big.NewInt(0)) == 0 {
		return false
	}
	if sigma.R.Cmp(ORDER) == 1 || sigma.S.Cmp(ORDER) == 1 {
		return false
	}
	//2. compute K = s*G - r * pk
	rMulPk := new(bn256.G1).ScalarMult(pub, sigma.R)
	nrMulPk := rMulPk.Neg(rMulPk)
	sMulG := new(bn256.G1).ScalarMult(G, sigma.S)
	K := new(bn256.G1).Add(sMulG, nrMulPk)
	//3. encode K and pk as []byte
	KTemp := K.Marshal()
	pkTemp := pub.Marshal()

	//4. compute challenge = H(msg||K||pk)
	message := append(msg, KTemp...)
	message = append(message, pkTemp...)
	digest := sha256.Sum256(message)
	challenge := hashToInt(digest[:], BN256())
	challenge.Mod(challenge, Bn256.N)

	//5.判断challenge ==? c
	if challenge.Cmp(sigma.R) == 0 {
		return true
	}
	return false
}

// encode signature into []byte
func SignMarshal(sigma *Signature) ([]byte, error) {
	rEncode := sigma.R.Bytes()
	sEncode := sigma.S.Bytes()
	rLen := len(rEncode)
	sLen := len(sEncode)
	if rLen != 32 || sLen != 32 {
		return nil, errors.New("Invalid signature")
	}
	res := append(rEncode, sEncode...)
	return res, nil
}

func SignMarshal1(sigma *Signature) ([]byte, []byte, error) {
	rEncode := sigma.R.Bytes()
	sEncode := sigma.S.Bytes()
	return rEncode, sEncode, nil
}

//decode []byte as signature
func SignUnMarshal(res []byte) (*Signature, error) {
	if len(res) != 64 {
		return nil, errors.New("Invalid input")
	}
	r := new(big.Int).SetBytes(res[:32])
	s := new(big.Int).SetBytes(res[32:])
	return &Signature{r, s}, nil
}

var Bn256 param
var initbn sync.Once

func initbn256() {
	Bn256 = param{Name: "BN-256"}
	Bn256.P, _ = new(big.Int).SetString("21888242871839275222246405745257275088696311157297823662689037894645226208583", 10)
	Bn256.N, _ = new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)
	Bn256.B, _ = new(big.Int).SetString("3", 10)
	Bn256.Gx, _ = new(big.Int).SetString("14bcc435f49d130d189737f9762feb25c44ef5b886bef833e31a702af6be4748", 16) //待定
	Bn256.Gy, _ = new(big.Int).SetString("10cd33954522ad058f00a2553fd4e10d859fe125997e98adba777910dddc5322", 16) //待定
	Bn256.BitSize = 256
}

func BN256() param {
	initbn.Do(initbn256)
	return Bn256
}
