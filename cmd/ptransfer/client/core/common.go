package core

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/Venachain/Venachain/log"

	"github.com/Venachain/Venachain/cmd/ptransfer/client/crypto"
	"github.com/Venachain/Venachain/cmd/ptransfer/client/utils"
	"github.com/Venachain/Venachain/cmd/ptransfer/client/venachainclient"
	utl "github.com/Venachain/Venachain/cmd/utils"
	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/common/hexutil"
	"github.com/Venachain/Venachain/crypto/bn256"
	"gopkg.in/urfave/cli.v1"
)

var pc *venachainclient.PClient
var ErrNonceUsed = errors.New("nonce is used")

//New ZSCparams: zsc,address,
func newZscInfo(c *cli.Context) (*ZCS, common.Address, *Account, string, *big.Int, *big.Int) {
	z, home := initial(c)
	account := loadAccount(c)
	value := c.String(ValueFlag.Name)
	l, r := getInterval(c)
	return z, home, account, value, l, r
}

func initial(c *cli.Context) (*ZCS, common.Address) {
	var err error

	// 1. read config file
	configInit()

	// 2. get url
	url := c.String(UrlFlag.Name)
	if url == "" {
		url = config.Url
	}

	// set url
	pc, err = venachainclient.SetupClient(url)
	if err != nil {
		utl.Fatalf(err.Error())
	}

	// 3. get contract address
	contract := c.String(ContractFlag.Name)
	if contract == "" {
		contract = config.Contract
	}

	if !utils.IsMatch(contract, "address") {
		utl.Fatalf("invalid transaction sender address")
	}

	// 4. set contract abi
	funcAbi, err := Asset("privacy_contract/PToken_sol_PToken.abi")
	if err != nil {
		utl.Fatalf("abi file not found")
	}
	z := newZCS(contract, funcAbi, "evm")

	// 5. get Tx sender
	address := c.String(TxSenderFlag.Name)
	if address == "" {
		address = config.From
	}

	if !utils.IsMatch(address, "address") {
		utl.Fatalf("invalid transaction sender address")
	}
	home := common.HexToAddress(address)

	//6. initial log handler
	verbosity := c.Int(verbosityFlag.Name)
	if verbosity == 0 {
		verbosity = config.Verbosity
	}
	log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(verbosity), log.StreamHandler(os.Stdout, log.TerminalFormat(false))))

	// 7. refresh config.json file
	if c.Bool(ConfigFlag.Name) {
		genConfig(c)
	}

	return z, home
}

func loadAccount(c *cli.Context) *Account {
	var account = NewAccount()
	accountFile := c.String(AccountFlag.Name)
	accountBytes, err := utils.ParseFileToBytes(accountFile)
	if err != nil {
		utl.Fatalf(err.Error())
	}

	err = account.Unmarshal(accountBytes)
	if err != nil {
		utl.Fatalf(err.Error())
	}
	log.Debug("load account", "balance", account.State.Balance)
	return account
}

func storeAccount(origin Account, filePath string) error {

	dataBytes, err := origin.Marshal()
	log.Debug("account data", "account", dataBytes[:])
	if err != nil {
		return fmt.Errorf("account marshal error: %v\n", err)
	}
	err = writeAccountFile(filePath, dataBytes)
	if err != nil {
		return err
	}

	return nil
}

func writeAccountFile(filename string, content []byte) error {

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(content) //写入文件(字节数组)
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}

// ============================ common utils ============================
func getInterval(c *cli.Context) (*big.Int, *big.Int) {

	l := big.NewInt(c.Int64(LeftIntervalFlag.Name))
	r := big.NewInt(c.Int64(RightIntervalFlag.Name))
	if r.Cmp(big.NewInt(0)) == 0 {
		r = new(big.Int).Sub(crypto.MAX, big.NewInt(1))

	}
	log.Debug("get interval l & r", "value", l, r)
	return l, r
}

func checkParams(acc *Account, wdValue *big.Int) error {
	//1. check value <= available balance
	if wdValue.Cmp(acc.State.Balance) == 1 {
		return errors.New("You don't have enough money")
	}

	//2.check the nonce = false
	if acc.State.NonceUsed {
		return ErrNonceUsed
	}

	return nil
}

func isPowerOfTwo(n int) bool {
	count := 0
	for n != 0 {
		count += 1
		n &= n - 1
	}
	return count == 1
}

func checkPubkeys(decoy []*bn256.G1, pub *bn256.G1) bool {
	for _, pubs := range decoy {
		if hexutil.Encode(pubs.Marshal()) == hexutil.Encode(pub.Marshal()) {
			return true
		}
	}
	return false
}

// deprecated
func checkPubStr(decoy []*bn256.G1, pub string) bool {
	for _, pubs := range decoy {
		if hexutil.Encode(pubs.Marshal()) == pub {
			return true
		}
	}
	return false
}

// deprecated
func checkChainPub(chainDecoy []*bn256.G1, decoy []*bn256.G1) (bool, []*bn256.G1) {
	var unusablePubList []*bn256.G1
	i := 0
	for _, unusablePubs := range decoy {
		if !checkPubkeys(chainDecoy, unusablePubs) {
			unusablePubList = append(unusablePubList, unusablePubs)
			i++
		}
	}
	if i != 0 {
		return false, unusablePubList
	} else {
		return true, nil
	}
}

// ====================== Decoy Shuffle =========================

type Decoy []*bn256.G1

//this function should be tested later--------zbx
//the sender and receiver must be have opposite parity
func (items Decoy) Shuffle(sender, receiver *bn256.G1) {
	i := len(items) - 1
	var index0, index1 int
	for i > 0 {
		r, _ := rand.Int(rand.Reader, new(big.Int).SetInt64(int64(i+1)))
		j := r.Int64()
		//fmt.Println("rand", j)
		tmp := items[j]
		items[j] = items[i]
		items[i] = tmp
		if PointCmp(tmp, sender) == 0 {
			index0 = i
		} else if PointCmp(tmp, receiver) == 0 {
			index1 = i
		}
		i = i - 1
	}
	if index0%2 == index1%2 {
		tmp := items[index1]
		if index1%2 == 0 {
			items[index1] = items[index1+1]
			items[index1+1] = tmp
			index1 = index1 + 1
		} else {
			items[index1] = items[index1-1]
			items[index1-1] = tmp
			index1 = index1 - 1
		}
	}
}

func PointCmp(point1, point2 *bn256.G1) int {
	value1 := new(big.Int).SetBytes(point1.Marshal())
	value2 := new(big.Int).SetBytes(point2.Marshal())
	return value1.Cmp(value2)
}

func checkIndex(decoy []*bn256.G1, pub *bn256.G1, txPub string) (indexPub, indexTxPub int) {
	for i := range decoy {
		if hexutil.Encode(decoy[i].Marshal()) == hexutil.Encode(pub.Marshal()) {
			indexPub = i
		}
		if hexutil.Encode(decoy[i].Marshal()) == txPub {
			indexTxPub = i
		}
	}
	if indexPub&indexTxPub == 0 {
		return
	}
	return indexPub, indexTxPub
}

// ====================== string conversion =========================

func genPubKeysStr(pubKeys []*bn256.G1) string {
	var str = make([]string, 0)

	for _, data := range pubKeys {
		str = append(str, genPubKeyStr(data))
	}

	return "[" + strings.Join(str, ",") + "]"
}

func genPubKeyStr(pubKey *bn256.G1) string {
	pub := pubKey.Marshal()
	x := hexutil.Encode(pub[:32])
	y := hexutil.Encode(pub[32:])
	return "[" + x + "," + y + "]"
}

func decodePubKey(pubKeyStr string) (string, error) {
	index1 := strings.Index(pubKeyStr, "[")
	index2 := strings.LastIndex(pubKeyStr, "]")
	pubKeyStr = pubKeyStr[index1+1 : index2]
	arr := strings.Split(pubKeyStr, ",")
	if index2 != 134 {
		return "", errors.New("The length of receiver pk is illegal")
	}
	if len(arr) != 2 {
		return "", errors.New("Decode receiver pk failed !!!")

	}

	return arr[0] + arr[1][2:], nil
}
