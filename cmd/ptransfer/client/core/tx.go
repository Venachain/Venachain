package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/Venachain/Venachain/log"

	"github.com/Venachain/Venachain/accounts/abi"

	"github.com/Venachain/Venachain/crypto/bn256"

	"github.com/Venachain/Venachain/cmd/ptransfer/client/crypto"
	cry "github.com/Venachain/Venachain/crypto"

	"github.com/Venachain/Venachain/cmd/ptransfer/client/packet"
	"github.com/Venachain/Venachain/cmd/ptransfer/client/utils"
	"github.com/Venachain/Venachain/common"
)

var ErrGetReceipt = errors.New("get transaction receipt error")
var errorSig = cry.Keccak256([]byte("Error(string)"))[:4]

type ZCS struct {
	contract string
	funcAbi  []byte
	vm       string
	call     *packet.ContractDataGen
}

func newZCS(contract string, funcAbi []byte, vm string) *ZCS {
	return &ZCS{
		contract: contract,
		funcAbi:  funcAbi,
		vm:       vm,
	}
}

func (z *ZCS) setMethod(funcName string, funcParams []string) {

	// new an contract call, set the interpreter(wasm or evm contract)
	data := packet.NewData(funcName, funcParams, z.funcAbi)
	call := packet.NewContractDataGen(data, "")
	call.SetInterpreter(z.vm)

	z.call = call
}

func (z *ZCS) Call() ([]interface{}, error) {
	return z.Send(common.HexToAddress(""), "", false)
}

func (z *ZCS) Send(from common.Address, gas string, isSync bool) ([]interface{}, error) {
	to := common.HexToAddress(z.contract)
	tx := packet.NewTxParams(from, &to, "", gas, "", "")

	result, isTxHash, err := pc.MessageCall(z.call, nil, tx)
	if err != nil {
		return nil, err
	}

	if isSync && isTxHash {
		res, err := pc.GetReceiptByPolling(result[0].(string))
		if err != nil {
			return result, ErrGetReceipt
		}

		receiptBytes, _ := json.Marshal(res)
		receiptStr := utils.PrintJson(receiptBytes)
		fmt.Println(receiptStr)

		recpt := z.call.ReceiptParsing(res)
		if recpt.Status != packet.TxReceiptSuccessMsg {
			revertRes, _ := pc.GetRevertMsg(tx, recpt.BlockNumber)
			if len(revertRes) >= 4 {
				recpt.Err, _ = unpackError(revertRes)
				return nil, errors.New(recpt.Err)
			}

			return nil, errors.New("failed: unknown error")
		}

		result[0] = recpt.Status
	}

	return result, nil
}

func unpackError(res []byte) (string, error) {
	var revStr string

	if !bytes.Equal(res[:4], errorSig) {
		return "<not revert string>", errors.New("not a revert string")
	}

	typ, _ := abi.NewTypeV2("string", "", nil)
	err := abi.Arguments{{Type: typ}}.UnpackV2(&revStr, res[4:])
	if err != nil {
		return "<invalid revert string>", err
	}

	return revStr, nil
}

// =============== wrapper of (z *ZCS) Send/Call ======================
// related to the methods defined in the privacy contract

func (z *ZCS) getEpochLength() (int64, error) {
	z.setMethod("epochLength", nil)
	result, err := z.Call()
	if err != nil {
		return 0, err
	}

	return result[0].(*big.Int).Int64(), nil
}

// ===================================================

type simAccReturn struct {
	X []byte
	Y []byte
}

func (z *ZCS) simulateAccount(keyPairs []*bn256.G1, epoch int64) ([]crypto.Ciphertext, error) {
	var temp = make([][2]*simAccReturn, 0)
	var res = make([]crypto.Ciphertext, 0)

	epochStr := strconv.FormatInt(epoch, 10)
	funcParams := []string{genPubKeysStr(keyPairs), epochStr}
	log.Debug("simulate account: func params", "para", funcParams)
	z.setMethod("simulateAccounts", funcParams)
	result, err := z.Call()
	if err != nil {
		res = append(res, crypto.Ciphertext{})
		return res, err
	}
	log.Debug("simulate account: call result", "result", result)
	tempBytes, _ := json.Marshal(result[0])
	_ = json.Unmarshal(tempBytes, &temp)

	for _, data := range temp {
		if isAllZeros(data[0].X) {
			res = append(res, crypto.Ciphertext{})
			return res, errors.New("get nil account ciphertext")
		}

		r := new(crypto.Ciphertext)
		r.C = PointToBn256(data[0])
		r.D = PointToBn256(data[1])

		res = append(res, *r)
	}
	log.Debug("simulate account: ", "result", res)
	return res, nil
}

func isAllZeros(b []byte) bool {
	for _, data := range b {
		if data != 0 {
			return false
		}
	}

	return true
}

func PointToBn256(data *simAccReturn) *bn256.G1 {
	pBytes := append(data.X, data.Y...)
	point := new(bn256.G1)
	point.Unmarshal(pBytes)
	return point
}

// ===================================================

func (z *ZCS) deposit(keyPairs *crypto.KeyPair, home common.Address, value string) ([]interface{}, error) {
	pubKey, _ := keyPairs.GetPublicKey()
	funcParams := []string{genPubKeyStr(pubKey), value}

	z.setMethod("fund", funcParams)

	result, err := z.Send(home, "", true)
	if err != nil {
		return nil, err
	}
	log.Debug("register: call result", "result", result)
	return result, nil
}

func (z *ZCS) withdraw(pubKey, NonceU *bn256.G1, wdProof *crypto.WithdrawProof, home common.Address, value string) ([]interface{}, error) {

	funcParams := []string{genPubKeyStr(pubKey), value, genPubKeyStr(NonceU), wdProof.WdProofMarshal()}

	//send the transaction
	z.setMethod("withdraw", funcParams)

	//res is the return value, which includes proof verifier
	result, err := z.Send(home, "", true)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (z *ZCS) transfer(
	cVector []*bn256.G1, dVector *bn256.G1, pubKeys []*bn256.G1, NonceU *bn256.G1, tfProof *crypto.TransferProof, home common.Address) ([]interface{}, error) {
	tfParams := []string{
		genPubKeysStr(cVector),
		genPubKeyStr(dVector),
		genPubKeysStr(pubKeys),
		genPubKeyStr(NonceU),
		crypto.TfProofMarshal(tfProof),
	}

	z.setMethod("transfer", tfParams)

	return z.Send(home, "", true)
}

func (z *ZCS) getPublicKey(index *big.Int) (*bn256.G1, error) {
	var res = new(simAccReturn)

	funcParams := []string{index.String()}
	z.setMethod("getPublicKey", funcParams)
	result, err := z.Call()
	if err != nil {
		return nil, err
	}

	tempBytes, _ := json.Marshal(result[0])
	_ = json.Unmarshal(tempBytes, res)

	return PointToBn256(res), nil
}

// ========== wrapper of z.getEpochLength =============

func (z *ZCS) currentEpoch() (int64, int64, error) {
	epochLen, err := z.getEpochLength()
	if err != nil {
		return 0, 0, err
	}

	log.Debug("get epoch length", "epoch length", epochLen)
	//compute current epoch = gepoch, and compute u = sk * gepoch
	blockHeight, err := pc.GetBlockNumber()
	if err != nil {
		return 0, 0, err
	}
	log.Debug("get block number", "height", blockHeight)
	return currentEpoch(int64(blockHeight), epochLen)
}

func currentEpoch(blockHeight, epochLen int64) (int64, int64, error) {
	if blockHeight < 0 {
		return 0, 0, errors.New("Invalid blockHeight")
	}

	return (blockHeight + 1) / epochLen, epochLen - (blockHeight+1)%epochLen, nil
}

// ========== wrapper of z.simulateAccount =============
func (z *ZCS) isAccountRegistered(pub *bn256.G1, epoch int64) bool {
	res, err := z.simulateAccount([]*bn256.G1{pub}, epoch)
	if err != nil || reflect.ValueOf(res[0]).IsZero() {
		return false
	}

	return true
}

func (z *ZCS) GetAccCipher(decoy []*bn256.G1, epoch int64) (cLn, cRn []*bn256.G1, err error) {

	res, err := z.simulateAccount(decoy, epoch)
	if err != nil {
		return cLn, cRn, err
	}

	for _, data := range res {
		cLn = append(cLn, data.C)
		cRn = append(cRn, data.D)
	}

	return cLn, cRn, nil
}
