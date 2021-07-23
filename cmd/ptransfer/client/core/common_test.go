package core

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ptransfer/client/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ptransfer/client/utils"
)

func TestStoreAccount(t *testing.T) {
	a := NewAccount()
	var err error
	a.Key, err = crypto.NewKeyPair(rand.Reader)
	if err != nil {
		fmt.Println("key generates failed")
	}
	err = storeAccount(*a, defaultFile)

	f, err := ioutil.ReadFile(defaultFile)
	if err != nil {
		fmt.Println("read fail", err)
	}
	fmt.Println(string(f))
}

func TestAccountLoad(t *testing.T) {
	var account = NewAccount()

	accountBytes, err := utils.ParseFileToBytes(defaultFile)
	if err != nil {
		fmt.Println("ParseFileToBytes failed")
	}

	err = account.Unmarshal(accountBytes)
	if err != nil {
		fmt.Println("unmarshalJson failed")
	}
	p, _ := account.Key.GetPublicKey()
	fmt.Println(account.Key.GetPrivateKey(), p)
	fmt.Println(account)
}
