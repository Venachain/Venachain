package crypto

import (
	"crypto/rand"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/bn256"
	"math/big"
	"regexp"
	"testing"
)

func TestEnc(t *testing.T) {
	value := big.NewInt(123)
	key, _ := NewKeyPair(rand.Reader)
	sk := key.sk
	pk := key.pk
	c, _ := Enc(rand.Reader, pk, value)
	fmt.Println("ciphertest:", c.C, c.D)
	d := Dec(c, sk)
	fmt.Println("message g^b:", d)
	fmt.Println("message g^b:", new(bn256.G1).ScalarMult(G, value))
}

func TestReadBalance(t *testing.T) {
	value := big.NewInt(3000000)
	msg := new(bn256.G1).ScalarMult(G, value)
	a := big.NewInt(2000000)
	b := big.NewInt(4294967296)
	v, err := ReadBalance(msg, a, b)
	fmt.Println("value:", v)
	fmt.Println("err:", err)
}

func TestMarshal(t *testing.T) {
	c1 := new(bn256.G1).ScalarMult(G, big.NewInt(14))
	c2 := new(bn256.G1).ScalarMult(G, big.NewInt(16))
	c := Ciphertext{c1, c2}
	fmt.Println("before:", c)
	r, _ := c.Marshal()
	fmt.Println("marshal:", r)
	err := c.UnMarshal(r)
	fmt.Println("err:", err)
	fmt.Println("unmarshal:", c)
}
func TestDec(t *testing.T) {
	name := "./privacy_Account.json"
	//res,err := regexp.Match(`^[a-z0-9A-Z\p{Han}]+(_[a-z0-9A-Z\p{Han}]+)*$`,[]byte(name))
	reg := "^.\\/(\\w+\\/?)+.json$"
	res,err := regexp.Match(reg,[]byte(name))

	if err != nil{
    	fmt.Println(err)
	}
	//if name != `^[a-z0-9A-Z\p{Han}]+(_[a-z0-9A-Z\p{Han}]+)*$`{
	//	fmt.Println("error:filename is illegal")
	//}
	//if name =="null"{
	//	fmt.Println("error:filename is null")
	//}
	fmt.Println("name:",name)
	fmt.Println("res:",res)
}
