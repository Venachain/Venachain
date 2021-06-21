package crypto

import (
	"crypto/rand"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	sk, err := GenerateKey(BN256(), rand.Reader)
	if err != nil {
		t.Log(err)
	}
	skTemp := hexutil.EncodeBig(sk)
	t.Log(skTemp)
}

func TestSign(t *testing.T) {
	key, _ := NewKeyPair(rand.Reader)
	msg := []byte{88, 66, 123, 12, 33, 45, 56}
	sigma, err := SchnorrSign(rand.Reader, key, msg)
	if err != nil {
		t.Log(err)
	}
	res := SchnorrVerify(key.pk, msg, sigma)
	t.Log(res)
}

func TestSignMarshal(t *testing.T) {
	key, _ := NewKeyPair(rand.Reader)
	msg := []byte{88, 66, 123}
	sigma, err := SchnorrSign(rand.Reader, key, msg)
	encode, err := SignMarshal(sigma)
	if err != nil {
		t.Log(err)
	} else {
		t.Log(len(encode))
	}
	res, err := SignUnMarshal(encode)
	if err != nil {
		t.Log(err)
	} else {
		t.Log(res.R)
		t.Log(res.S)
	}
}
