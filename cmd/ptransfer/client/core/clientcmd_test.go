package core

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ptransfer/client/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/bn256"
)

func TestNewAcc(t *testing.T) {
	addr := common.Hex2Bytes("72abce0e6eebff6fb02758b2181504ef490d2c77")
	fmt.Println("addr", addr)
	msg := common.LeftPadBytes(addr, 32)
	keypair, _ := crypto.NewKeyPair(rand.Reader)
	a := NewAccount()
	a.Key = keypair
	signature, _ := crypto.SchnorrSign(rand.Reader, a.Key, msg)
	sigma1, sigma2, _ := crypto.SignMarshal1(signature)
	sig1 := hexutil.Encode(sigma1)
	sig2 := hexutil.Encode(sigma2)

	pubKey, _ := a.Key.GetPublicKey()
	//pub := pubKey.String()
	pub := pubKey.Marshal()
	x := hexutil.Encode(pub[:32])
	y := hexutil.Encode(pub[32:])
	ret := "[\"" + x + "\",\"" + y + "\"]"
	fmt.Println(ret)
	fmt.Println(sig1, sig2)

	priv := hexutil.Encode(keypair.GetPrivateKey().Bytes())
	fmt.Println("private key:", priv)
}

func TestNewAccFixed(t *testing.T) {
	addr := common.Hex2Bytes("28cb4c2c5f3e18132cfea922018f2c8b6149f2c5")
	fmt.Println("addr", addr)
	msg := common.LeftPadBytes(addr, 32)

	sk, _ := new(big.Int).SetString("27d0b9fa68ca84d865f3bd77eb5929407b6788fd91881a6e0436efe6612b1e16", 16)
	pk := new(bn256.G1).ScalarMult(crypto.G, sk)

	signature, _ := crypto.SchnorrTest(rand.Reader, sk, pk, msg)
	sigma1, sigma2, _ := crypto.SignMarshal1(signature)
	sig1 := hexutil.Encode(sigma1)
	sig2 := hexutil.Encode(sigma2)

	pub := pk.Marshal()
	x := hexutil.Encode(pub[:32])
	y := hexutil.Encode(pub[32:])
	ret := "[\"" + x + "\",\"" + y + "\"]"

	fmt.Println(ret)
	fmt.Println(sig1, sig2)

	priv := hexutil.Encode(sk.Bytes())
	fmt.Println("private key:", priv)
}
