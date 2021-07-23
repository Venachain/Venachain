package core

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"math/big"
	"reflect"

	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ptransfer/client/crypto"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/bn256"
)

type UserState struct {
	Balance *big.Int
	/// Pending      *big.Int
	NonceUsed    bool
	LastRollOver int64
}

type Account struct {
	State *UserState
	Key   *crypto.KeyPair
}

type AccountCopy struct {
	State   *UserState
	KeyPair *KeyPair
}

type KeyPair struct {
	Sk string
	Pk string
}

func NewAccount() *Account {
	a := &Account{
		State: NewUserState(),
		Key:   &crypto.KeyPair{},
	}

	keypair, _ := crypto.NewKeyPair(rand.Reader)
	a.Key = keypair

	return a
}

func NewAccountCopy() *AccountCopy {
	a := &AccountCopy{
		State:   NewUserState(),
		KeyPair: new(KeyPair),
	}
	return a
}

func NewUserState() *UserState {
	u := &UserState{
		Balance: big.NewInt(0),
		/// Pending:      big.NewInt(0),
		NonceUsed:    false,
		LastRollOver: 0,
	}
	return u
}

func (a *Account) Marshal() ([]byte, error) {
	copyAcc := copyAccount(a)
	return json.Marshal(copyAcc)
}

//The functions with "copy" mean that copy the lower case "crypt.Keypair" to  the upper case "Keypair" for json.Marshal
// only accepts the upper case value.
func copyAccount(origin *Account) *AccountCopy {
	accCopy := NewAccountCopy()
	accCopy.State = origin.State
	accCopy.KeyPair.Sk = hexutil.Encode(origin.Key.GetPrivateKey().Bytes())
	pubKey, _ := origin.Key.GetPublicKey()
	accCopy.KeyPair.Pk = genPubKeyStr(pubKey)

	return accCopy
}

func (a *Account) Unmarshal(b []byte) error {
	accCopy := NewAccountCopy()
	err := json.Unmarshal(b, accCopy)
	if err != nil {
		return err
	}

	a.State = accCopy.State

	// give accountcopy.Keypair to account.Key
	skBytes, err := hexutil.Decode(accCopy.KeyPair.Sk)
	if err != nil{
		return err
	}
	sk := new(big.Int).SetBytes(skBytes)
	a.Key.NewPrivateKey(sk)

	pubKeyStr, err := decodePubKey(accCopy.KeyPair.Pk)
	if err != nil{
		return err
	}
	err = a.Key.NewPublicKey(pubKeyStr)
	if err != nil {
		return err
	}

	return nil
}

func (a *Account) recover(z *ZCS, epoch int64, l, r *big.Int) (*bn256.G1, *bn256.G1, error) {
	var err error
	sk := a.Key.GetPrivateKey()
	log.Debug("recover private key","sk", sk.String())
	pub, err := a.Key.GetPublicKey()
	if err != nil{
		return nil, nil, err
	}
	log.Debug("recover public key","pk", pub.String())
	result, err := z.simulateAccount([]*bn256.G1{pub}, epoch)
	if err != nil {
		 return nil, nil, err
	}
	if reflect.ValueOf(result[0]).IsZero() {
		return nil, nil, errors.New("unknown account")
	}

	gb := crypto.Dec(&result[0], sk)
	log.Debug("get gb","value", gb.String())
	a.State.Balance, err = crypto.ReadBalance(gb, l, r)
	if err != nil {
		return nil, nil, err
	}
	log.Debug("get balance","value", a.State.Balance)
	return result[0].C, result[0].D, nil
}

func (a *Account) update(currentEpoch int64) {
	log.Debug("update account","value", currentEpoch)
	log.Debug("update account","state", a.State.LastRollOver)
	if a.State.LastRollOver < currentEpoch {
		/// a.State.Balance.Add(a.State.Balance, a.State.Pending)
		/// a.State.Pending.SetInt64(0)
		a.State.NonceUsed = false
		a.State.LastRollOver = currentEpoch
	}
}
