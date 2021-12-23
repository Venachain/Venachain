package crypto

import (
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/Venachain/Venachain/crypto/bn256"
	"github.com/pkg/errors"
)

type Ciphertext struct {
	C *bn256.G1
	D *bn256.G1
}

func Newciphertext(C, D *bn256.G1) *Ciphertext {
	c := &Ciphertext{C, D}
	return c
}

func Enc(rand io.Reader, pk *bn256.G1, b *big.Int) (*Ciphertext, error) {
	if b.Cmp(Bn256.N) != -1 {
		err := errors.New("value is out of range")
		return nil, err
	}
	// 1. compute msg = g^b
	m := new(bn256.G1).ScalarMult(G, b)
	//2. generate temporary secret key
	r, err := randFieldElement(BN256(), rand)
	if err != nil {
		return nil, err
	}
	// 3. ciphertest(C,D)  C = g^b*pk^r , D = g^r
	C := new(bn256.G1).ScalarMult(pk, r)
	C = new(bn256.G1).Add(C, m)
	D := new(bn256.G1).ScalarMult(G, r)
	return &Ciphertext{C, D}, nil
}
func Dec(c *Ciphertext, sk *big.Int) *bn256.G1 {
	// 1. s = {g^r}^sk  m = s^{-1}*C
	// 1. compute msg = g^b
	s := new(bn256.G1).ScalarMult(c.D, sk)
	m := new(bn256.G1).Add(c.C, s.Neg(s))
	return m
}

func PointToInt(p *bn256.G1) *big.Int {
	num := p.Marshal()
	return new(big.Int).SetBytes(num)
}

func (c *Ciphertext) Marshal() ([]byte, error) {
	cEncode := c.C.Marshal()
	dEncode := c.D.Marshal()
	if len(cEncode) != 64 || len(dEncode) != 64 {
		return nil, errors.New("Invalid ciphertext")
	}
	res := append(cEncode, dEncode...)
	return res, nil
}

func (c *Ciphertext) UnMarshal(res []byte) error {
	if len(res) != 128 {
		return errors.New("Invalid ciphertext")
	}
	e := new(bn256.G1)
	d := new(bn256.G1)
	_, err := e.Unmarshal(res[:64])
	if err != nil {
		return errors.New("wrong Unmarshal")
	}
	_, err = d.Unmarshal(res[64:])
	if err != nil {
		return errors.New("wrong Unmarshal")
	}
	c.C, c.D = e, d
	return nil
}

func ReadBalance(m *bn256.G1, a, b *big.Int) (*big.Int, error) {
	// a>=0 && b>=0
	if a.Cmp(zero) == -1 || b.Cmp(zero) == -1 {
		err := errors.New("value is out of range")
		return nil, err
	}
	// a < b
	if a.Cmp(b) == 1 {
		a, b = b, a
	}
	// b < 2^{32}
	if b.Cmp(MAX) != -1 {
		err := errors.New("value is out of range")
		return nil, err
	}

	d := time.Duration(time.Minute * 2)
	t := time.NewTicker(d)

	go func() {
		<-t.C
		fmt.Fprintf(os.Stdout, "query time is more than 2 min.\n")
		os.Exit(1)
	}()

	defer t.Stop()
	pa := new(bn256.G1).ScalarMult(G, a)
	bigm := PointToInt(m)
	//i := new(big.Int).Set(a)
	for i := a; i.Cmp(b) != 1; {
		//pi := new(bn256.G1).ScalarMult(G, i)
		bigi := PointToInt(pa)
		if bigi.Cmp(bigm) == 0 {
			return i, nil
		} else {
			i = new(big.Int).Add(i, big.NewInt(1))
			pa = new(bn256.G1).Add(pa, G)
		}
	}

	return nil, errors.New("Read balance failed")
}
