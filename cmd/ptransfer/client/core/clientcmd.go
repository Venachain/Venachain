package core

import (
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ptransfer/client/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/log"
	"math/big"
	"os"
	"reflect"
	"regexp"

	"github.com/PlatONEnetwork/PlatONE-Go/cmd/ptransfer/client/crypto"
	utl "github.com/PlatONEnetwork/PlatONE-Go/cmd/utils"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common/hexutil"
	"github.com/PlatONEnetwork/PlatONE-Go/crypto/bn256"
	"gopkg.in/urfave/cli.v1"
)

var (
	RegisterCmd = cli.Command{
		Name:   "register",
		Usage:  "register a new account",
		Action: register,
		Flags:  registerCmdFlags,
		Description: `
		client register --bc-owner ... --contract ... --url ...

Register a new account in the privacy token contract`,
	}

	DepositCmd = cli.Command{
		Name:   "deposit",
		Usage:  "deposit amount of token",
		Action: deposit,
		Flags:  depositCmdFlags,
		Description: `
		client deposit --account ... --value ... --bc-owner ... --contract ... --url ...

Deposit amount of token from bc token to privacy token contract.
the operation may be failed if the amount of token is insufficient in bc token contract`,
	}

	TransferCmd = cli.Command{
		Name:   "transfer",
		Usage:  "invoke contract function",
		Action: transfer,
		Flags:  transferCmdFlags,
		Description: `
		client transfer --account ... --receiver ... --value ... --bc-owner ... --contract ... --url ... 

Transfer privacy token from an account to another account
The --receiver should also be an account registered in the privacy token
`,
	}

	WithdrawCmd = cli.Command{
		Name:   "withdraw",
		Usage:  "withdraw amount of token",
		Action: withdraw,
		Flags:  withdrawCmdFlags,
		Description: `
		client withdraw --account ... --value ... --abi ... --bc-owner ... --contract ... --url ...

Withdraw amount of token from privacy token to bc token contract.
the operation may be failed if the amount of token is insufficient in privacy token contract`,
	}

	QueryCmd = cli.Command{
		Name:   "query",
		Usage:  "query the amount of the privacy token in plaintext",
		Action: query,
		Flags:  queryCmdFlags,
		Description: `
		client query --account ... --bc-owner ... --contract ... --url ... 

query the AVAILABLE amount of privacy token at current epoch`,
	}

	ConfigCmd = cli.Command{
		Name:   "config",
		Usage:  "generate a config file to eliminate the ",
		Action: genConfig,
		Flags:  configCmdFlags,
		Description: `
		client config --bc-owner ... --contract ... --url ... 

set the value of the config file`,
	}
)

const (
	defaultFile = "./privacyAccount.json"

	threshold = 4
)

func register(c *cli.Context) {
	z, home := initial(c)

	// New account and key pair
	a := NewAccount()

	keypair, err := crypto.NewKeyPair(rand.Reader)
	if err != nil {
		utl.Fatalf(err.Error())
	}

	log.Debug("initial registration","param", c.String(OutputFlag.Name))
	a.Key = keypair
	// check file name
	filePath := c.String(OutputFlag.Name)
	_, err = os.Stat(filePath)
	if err == nil || os.IsExist(err) {
		utl.Fatalf(fmt.Sprintf("local account file is existed. \n"))
	}
	if filePath == "null" {
		utl.Fatalf(fmt.Sprintf("filename is null.\n"))
	}
	if len(filePath) > 256{
		utl.Fatalf(fmt.Sprintf("filename is too long.\n"))
	}
	reg := "^.\\/(\\w+\\/?)+.json$"
	r, _ := regexp.Match(reg,[]byte(filePath))
	log.Debug("filename check","result", r)
	if !r {
		utl.Fatalf(fmt.Sprintf("filename is illegal.\n"))
	}

	// Store account locally

	err = storeAccount(*a, filePath)
	if err != nil {
		utl.Fatalf(fmt.Sprintf("store account locally failed: %v\n", err))
	}

	// register account on ZSC
	res, err := Register(z, home, a)
	if err != nil {
		if err == ErrGetReceipt {
			utl.Fatalf("error: %s, tx hash: %v", err.Error(), res[0])
		}

		utl.Fatalf(fmt.Sprintf("register error: %s", err.Error()))
	}
	log.Debug("register: call result","result", res)
	fmt.Printf("register success\n")
}

func Register(z *ZCS, home common.Address, a *Account) ([]interface{}, error) {
	//compute the Schnorr signature where the message is contract address
	addrData, _ := hexutil.Decode(z.contract)
	log.Debug("register: contract address","contract", addrData)
	msg := common.LeftPadBytes(addrData, 32)
	signature, err := crypto.SchnorrSign(rand.Reader, a.Key, msg)
	if err != nil {
		return nil, errors.New("Schnorr signature failed")
	}

	sigma1, sigma2, _ := crypto.SignMarshal1(signature)
	sig1 := hexutil.Encode(sigma1)
	sig2 := hexutil.Encode(sigma2)

	pubKey, err := a.Key.GetPublicKey()
	if err != nil {
		return nil, err
	}

	funcParams := []string{genPubKeyStr(pubKey), sig1, sig2}
	log.Debug("register: func params","para", funcParams)
	z.setMethod("register", funcParams)
	return z.Send(home, "", true)
}

func deposit(c *cli.Context) {
	z, home, account, value, l, r := newZscInfo(c)
	res, err := Deposit(z, home, account, value, l, r)
	if err != nil {
		utl.Fatalf(fmt.Sprintf("deposit error: %s", err.Error()))
	}

	if err != nil {
		if err == ErrGetReceipt {
			utl.Fatalf("error: %s, tx hash: %v", err.Error(), res[0])
		}

		utl.Fatalf(fmt.Sprintf("deposit error: %s", err.Error()))
	}

	// update balance
	/// bigValue, _ := new(big.Int).SetString(value, 10)
	/// account.State.Balance.Add(account.State.Balance, bigValue)
	fmt.Println("deposit success")

	// record the result to the file
	err = storeAccount(*account, c.String(AccountFlag.Name))
	if err != nil {
		utl.Fatalf(fmt.Sprintf("store account locally failed: %v\n", err))
	}

}

func Deposit(z *ZCS, home common.Address, acc *Account, value string, l, r *big.Int) ([]interface{}, error) {

	//get the account balance, now the account state is changed
	epoch, _, _ := z.currentEpoch()
	log.Debug("deposit:","epoch", epoch)
	_, _, err := acc.recover(z, epoch, l, r)
	if err != nil {
		return nil, err
	}

	return z.deposit(acc.Key, home, value)
}

func transfer(c *cli.Context) {
 	z, home, account, value, l, r := newZscInfo(c)
	txPub := c.String(TransferPubFlag.Name)
	decoyNum := c.Int(DecoyNumFlag.Name)

	if decoyNum < 2 || decoyNum > 10 {
		utl.Fatalf("the decoy number should in the range [2,10]")
	}
	pubKeyStr, err := decodePubKey(txPub)
	if err != nil{
		utl.Fatalf(fmt.Sprintf("transfer error: %s", err.Error()))
	}

	err = Transfer(z, home, account, value, decoyNum, l, r, pubKeyStr)
	if err != nil {
		utl.Fatalf(fmt.Sprintf("transfer error: %s", err.Error()))
	}

	fmt.Println("transfer success")

	err = storeAccount(*account, c.String(AccountFlag.Name))
	if err != nil {
		utl.Fatalf(fmt.Sprintf("store account locally failed: %v\n", err))
	}
}

func Transfer(z *ZCS, home common.Address, acc *Account, value string, decoyNum int, l, r *big.Int, txPub string) error {

	epoch, remainBlocks, err := z.currentEpoch()

	if err != nil {
		return err
	}
	log.Debug("transfer:","epoch", epoch)
	log.Debug("transfer:","remain block", remainBlocks)
	_, _, err = acc.recover(z, epoch, l, r)
	if err != nil {
		return err
	}

	acc.update(epoch)

	// 1. Param check
	// 1.1 convert and check value param
	tfValue, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return errors.New("SetString: error")
	}

	err = checkParams(acc, tfValue)
	if err != nil {
		if err == ErrNonceUsed {
			return fmt.Errorf("%v: remain %d blocks switching to last epoch", err, remainBlocks)
		}

		return err
	}

	// 1.2 check: you can not transfer to yourself (currently unsupported)
	pub, err := acc.Key.GetPublicKey()
	if err != nil{
		return err
	}
	if hexutil.Encode(pub.Marshal()) == txPub {
		return errors.New("you can not transfer to yourself")
	}

	//1.3. decode txPub as receiver public key which is a point
	receiverPub := new(bn256.G1)
	rpub, err := hexutil.Decode(txPub)
	if err != nil {
		return errors.New("decode receiver pk failed")
	}
	receiverPub.Unmarshal(rpub)

	// 2. decoy
	var decoy Decoy
	var check = make(map[string]int, 0)
	//var check = make(map[*bn256.G1]int, 0)

	if !z.isAccountRegistered(receiverPub, epoch) {
		return errors.New("the receiver's public key is not valid")
	}

	decoy = append(decoy, pub)
	decoy = append(decoy, receiverPub)
	check[hexutil.Encode(pub.Marshal())], check[txPub] = 1, 1

	for i := 0; i < 2<<(decoyNum-1)-2; {
		randNum, _ := rand.Int(rand.Reader, crypto.MAX)
		pubKey, _ := z.getPublicKey(randNum)
		dpub := hexutil.Encode(pubKey.Marshal())
		check[dpub] += 1

		if check[dpub] > 1 {
			if check[dpub] > threshold {
				return errors.New("fetch decoy members failed")
			}

			continue
		}

		i++
		decoy = append(decoy, pubKey)
	}

	// 2.3. check:the decoy size must be a power of two
	AnonLen := len(decoy)
	if !isPowerOfTwo(AnonLen) {
		return errors.New("the decoy size must be power of two")
	}

	// 2.4. shuffle the decoy
	// the sender and receiver index must be in the opposite parity
	decoy.Shuffle(pub, receiverPub)

	// 3. generate transferProof
	// 3.1 get the witness: sk, bTf, bDiff, r, l0, l1
	sk := acc.Key.GetPrivateKey()
	bTf := new(big.Int).Set(tfValue)
	bDiff := new(big.Int).Sub(acc.State.Balance, bTf)

	indexSender, indexReceiver := checkIndex(decoy, pub, txPub)
	l0 := big.NewInt(int64(indexSender))
	l1 := big.NewInt(int64(indexReceiver))
	log.Debug("transfer:","l0 index", l0)
	log.Debug("transfer:","l1 index", l1)
	// 3.2 get the instance: AnonPk, CLnNew, CRnNew, CVector, D, NonceU, epoch
	// AnonPk = decoy,
	// D = G*r, CVector[l0]=-b*G+r*pk[l0], CVector[l1]=b*G+r*pk[l1], CVector[i]= r*pk[i](i != l0 or l1)
	randNum, err := rand.Int(rand.Reader, crypto.ORDER)
	if err != nil {
		return fmt.Errorf("generate random number failed: %v", err)
	}

	D := new(bn256.G1).ScalarMult(crypto.G, randNum)
	CVector := make([]*bn256.G1, AnonLen)
	for i := 0; i < AnonLen; i++ {
		if i == indexSender {
			CVector[i] = new(bn256.G1).Add(new(bn256.G1).ScalarMult(crypto.G, new(big.Int).Neg(bTf)), new(bn256.G1).ScalarMult(decoy[indexSender], randNum))
		} else if i == indexReceiver {
			CVector[i] = new(bn256.G1).Add(new(bn256.G1).ScalarMult(crypto.G, bTf), new(bn256.G1).ScalarMult(decoy[indexReceiver], randNum))
		} else {
			CVector[i] = new(bn256.G1).ScalarMult(decoy[i], randNum)
		}
	}

	//compute epoch and NonceU
	currentepoch := big.NewInt(epoch)
	Gepoch := crypto.MapIntoGroup("zether" + currentepoch.String())
	NonceU := new(bn256.G1).ScalarMult(Gepoch, sk)

	//compute CLnNew CRnNew
	CLnOld, CRnOld, err := z.GetAccCipher(decoy, epoch)
	if err != nil{
		return err
	}

	CLnNew := make([]*bn256.G1, AnonLen)
	CRnNew := make([]*bn256.G1, AnonLen)
	for j := 0; j < AnonLen; j++ {
		CLnNew[j] = new(bn256.G1).Add(CLnOld[j], CVector[j])
		CRnNew[j] = new(bn256.G1).Add(CRnOld[j], D)
	}

	tfProof, err := crypto.TfProver(decoy, CLnNew, CRnNew, CVector, D, NonceU, currentepoch, sk, bTf, bDiff, randNum, l0, l1)
	if err != nil {
		return err
	}
	//4. transfer params and send transaction
	res, err := z.transfer(CVector, D, decoy, NonceU, tfProof, home)
	if err != nil {
		if err == ErrGetReceipt {
			return fmt.Errorf("error: %s, tx hash: %v", err.Error(), res[0])
		}

		return err
	}

	// 4.2 else the proof is valid and then update the state: balance, nonce, pending
	acc.State.Balance = new(big.Int).Sub(acc.State.Balance, tfValue)
	acc.State.NonceUsed = true
	/// acc.State.Pending = big.NewInt(0)
	acc.State.LastRollOver = epoch
	return nil
}

func withdraw(c *cli.Context) {
	z, home, account, value, l, r := newZscInfo(c)

	err := Withdraw(z, home, account, value, l, r)
	if err != nil {
		utl.Fatalf(fmt.Sprintf("withdraw error: %s", err.Error()))
	}

	fmt.Printf("withdraw success\n")

	err = storeAccount(*account, c.String(AccountFlag.Name))
	if err != nil {
		utl.Fatalf(fmt.Sprintf("store account locally failed: %v\n", err))
	}
}

func Withdraw(z *ZCS, home common.Address, acc *Account, value string, l, r *big.Int) error {

	//1. get acc state from the blockchain
	epoch, remainBlocks, err := z.currentEpoch()
	if err != nil {
		return fmt.Errorf("get current epoch failed: %v", err)
	}
	currentEpoch := big.NewInt(epoch)
	log.Debug("transfer:","epoch", epoch)
	log.Debug("transfer:","remain block", remainBlocks)
	cL, cR, err := acc.recover(z, epoch, l, r)
	if err != nil {
		return err
	}
	acc.update(epoch)

	wdValue, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return errors.New("SetString: error")
	}

	//check params
	err = checkParams(acc, wdValue)
	if err != nil {
		if err == ErrNonceUsed {
			return fmt.Errorf("%v: remain %d blocks switching to last epoch", err, remainBlocks)
		}

		return err
	}

	//generate withdrawproof
	pub, err := acc.Key.GetPublicKey()
	if err != nil{
		return err
	}
	sk := acc.Key.GetPrivateKey()

	cL = new(bn256.G1).Add(cL, new(bn256.G1).Neg(new(bn256.G1).ScalarMult(crypto.G, wdValue)))

	v := make([]*big.Int, 1)
	bDiff := new(big.Int).Sub(acc.State.Balance, wdValue)
	v[0] = bDiff

	aggWit := crypto.NewAggBpWitness(v)
	wdWit := crypto.NewWithdrawWit(sk, aggWit)

	wdProof, err := crypto.WithdrawProve(cL, cR, pub, currentEpoch, home.Bytes(), wdWit)
	if err != nil {
		return err
	}

	gepoch := crypto.MapIntoGroup("zether" + currentEpoch.String())
	NonceU := new(bn256.G1).ScalarMult(gepoch, sk)

	//burnParams=(pk, value, u=sk*gepoch, proof)
	res, err := z.withdraw(pub, NonceU, wdProof, home, value)
	if err != nil {
		if err == ErrGetReceipt {
			return fmt.Errorf("error: %s, tx hash: %v", err.Error(), res[0])
		}

		return err
	}

	//else the proof is valid and then update the state: balance, nonce, pending
	acc.State.Balance = new(big.Int).Sub(acc.State.Balance, wdValue)
	acc.State.NonceUsed = true
	acc.State.LastRollOver = epoch
	return nil
}

func query(c *cli.Context) {
	z, _, account, _, l, r := newZscInfo(c)

	result, err := Query(z, account, l, r)
	if err != nil {
		utl.Fatalf("query error, %s", err.Error())
	}

	pub, _ := account.Key.GetPublicKey()
	fmt.Printf("the balance of accout %v: %v\n", genPubKeyStr(pub), result)
}

func Query(z *ZCS, acc *Account, l, r *big.Int) (*big.Int, error) {
	var err error

	sk := acc.Key.GetPrivateKey()
	pub, _ := acc.Key.GetPublicKey()

	epoch, _, err := z.currentEpoch()
	if err != nil {
		return nil, fmt.Errorf("get current epoch failed: %v", err)
	}

	res, err := z.simulateAccount([]*bn256.G1{pub}, epoch)
	if err != nil || reflect.ValueOf(res[0]).IsZero() {
		return nil, errors.New("the account is not registered")
	}

	gb := crypto.Dec(&res[0], sk)
	balance, err := crypto.ReadBalance(gb, l, r)
	if err != nil {
		return nil, err
	}

	return balance, nil

}

func genConfig(c *cli.Context) {
	if c.String(UrlFlag.Name) != ""{
		config.Url = c.String(UrlFlag.Name)
	}
	if c.String(ContractFlag.Name) != ""{
		config.Contract = c.String(ContractFlag.Name)
	}
	if c.String(TxSenderFlag.Name) != ""{
		config.From = c.String(TxSenderFlag.Name)
	}
	if c.Int(verbosityFlag.Name) != 0{
		config.Verbosity = c.Int(verbosityFlag.Name)
	}

	//config.Contract = c.String(ContractFlag.Name)
	//config.From = c.String(TxSenderFlag.Name)

	runPath := utils.GetRunningTimePath()
	configFile := runPath + defaultConfigFilePath
	WriteConfig(configFile, config)
}
