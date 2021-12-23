package crypto

import (
	"fmt"
	"io"
	"math/big"

	"github.com/Venachain/Venachain/common/hexutil"
	"github.com/Venachain/Venachain/crypto/bn256"
)

type KeyPair struct {
	sk *big.Int
	pk *bn256.G1
}

func NewKeyPair(rand io.Reader) (*KeyPair, error) {
	privateKey, err := GenerateKey(BN256(), rand)
	if err != nil {
		return nil, err
	}
	publicKey, err := getPublicKey(BN256(), privateKey)
	key := &KeyPair{
		sk: privateKey,
		pk: publicKey,
	}
	return key, nil
}

func (key *KeyPair) GetPublicKey() (*bn256.G1, error) {
	if key.pk == nil {
		key.pk, _ = getPublicKey(BN256(), key.sk)
	}
	return key.pk, nil
}

func (key *KeyPair) GetPrivateKey() *big.Int {
	return key.sk
}

func (key *KeyPair) NewPublicKey(s string) error {
	t, err := hexutil.Decode(s)
	if err != nil {
		return fmt.Errorf("Decode failed err=%v\n", err)
	}

	_, err = key.pk.Unmarshal(t)
	if err != nil {
		return fmt.Errorf("unmarshal pk failed err=%v\n", err)
	}

	return nil
}

func (key *KeyPair) NewPrivateKey(a *big.Int) {
	key.sk = a
}
