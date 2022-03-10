package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/rlp"
)

func TestParseConfig(t *testing.T) {
	//param := TxParams{}
	param := DeployParams{}
	parseConfigJson(configPath)

	fmt.Println(param.Gas)
	fmt.Println(param.GasPrice)
	fmt.Println(param.From)
}

func TestParseFuncFromAbi(t *testing.T) {

	dir, _ := os.Getwd()
	filePath := dir + "/demo01.cpp.abi.json"
	funcDesc, _ := parseFuncFromAbi(filePath, "transfer")

	fmt.Println(funcDesc.Name)
	fmt.Println(funcDesc.Inputs)
	fmt.Println(funcDesc.Outputs)
	fmt.Println(len(funcDesc.Constant))
}

func TestParseAbiFromJson(t *testing.T) {

	dir, _ := os.Getwd()
	filePath := dir + "/demo01.cpp.abi.json"
	a, e := parseAbiFromJson(filePath)
	if e != nil {
		t.Fatalf("parse abi json error! \n， %s", e.Error())
	}
	fmt.Println(a)
	marshal, _ := json.Marshal(a)
	fmt.Println(string(marshal))
}

func TestHttpPostTransfer(t *testing.T) {
	//params := TxParams{
	//	From:"0xfb8c2fa47e84fbde43c97a0859557a36a5fb285b",
	//	To:"0x9f75627b1436b506eafc96bf70bcd2ff88f715e2",
	//	Value:"0x2710",
	//	Gas: "0x76c0",
	//	GasPrice:"0x9184e72a000",
	//}
	//url := "http://localhost:8545"
	param := JsonParam{
		Jsonrpc: "2.0",
		Method:  "eth_sendTransaction",
		//Params:[]TxParams{},
		Id: 1,
	}
	s, e := HttpPost(param)
	if e != nil {
		t.Fatal("test http post error .\n" + e.Error())
	}
	fmt.Println(s)

}

func TestHttpPostDeploy(t *testing.T) {
	deployParams := DeployParams{
		From:     "0xfb8c2fa47e84fbde43c97a0859557a36a5fb285b",
		Gas:      "0x400000",
		GasPrice: "0x9184e72a000",
	}

	params := make([]interface{}, 1)
	params[0] = deployParams
	param := JsonParam{
		Jsonrpc: "2.0",
		Method:  "eth_sendTransaction",
		Params:  params,
		Id:      1,
	}

	r, e := HttpPost(param)
	if e != nil {
		t.Fatal("test http post error .\n" + e.Error())
	}

	var resp = Response{}
	err := json.Unmarshal([]byte(r), &resp)
	if err != nil {
		t.Fatalf("parse result error ! \n %s", err.Error())
	}

	if resp.Error.Code != 0 {
		t.Fatalf("send transaction error ,error:%v", resp.Error.Message)
	}
	fmt.Printf("trasaction hash: %s\n", resp.Result)

	ch := make(chan string, 1)
	go GetTransactionReceipt(resp.Result, ch)

	select {
	case address := <-ch:
		fmt.Printf("contract address:%s\n", address)
	case <-time.After(time.Second * 100):
		fmt.Printf("get contract receipt timeout...more than 100 second.\n")
	}

}

func TestHttpCallContact(t *testing.T) {

	//url := "http://localhost:8545"
	param1 := uint(33)
	b := new(bytes.Buffer)
	rlp.Encode(b, param1)

	params := TxParams{
		From:     "0xfb8c2fa47e84fbde43c97a0859557a36a5fb285b",
		To:       "0xace6bdba54c8c359e70f541bfc1cabaf0244b916",
		Value:    "0x2710",
		Gas:      "0x76c00",
		GasPrice: "0x9184e72a000",
		//Data:"0x60fe47b10000000000000000000000000000000000000000000000000000000000000011",
	}

	param := JsonParam{
		Jsonrpc: "2.0",
		Method:  "eth_sendTransaction",
		Params:  []TxParams{params},
		Id:      1,
	}
	paramJson, _ := json.Marshal(param)
	fmt.Println(string(paramJson))
	s, e := HttpPost(param)
	if e != nil {
		t.Fatal("test http post error .\n" + e.Error())
	}
	fmt.Println(s)

}

func TestGetFuncParam(t *testing.T) {
	//f := "set(\"1\",\"b\",1.2)"
	f := "set()"
	s, strings := GetFuncNameAndParams(f)
	fmt.Println(s)
	fmt.Println(len(strings))

	//funcName := string(f[0:strings.Index(f, "(")])
	//fmt.Println(funcName)
	//
	//paramString := string(f[strings.Index(f, "(")+1 : strings.LastIndex(f, ")")])
	//fmt.Println(paramString)
	//
	//params := strings.Split(paramString, ",")
	//for _, param := range params {
	//	if strings.HasPrefix(param, "\"") {
	//		i, err := strconv.Atoi(param[strings.Index(param, "\"")+1 : strings.LastIndex(param, "\"")])
	//		fmt.Println(err)
	//		fmt.Println(i)
	//	}
	//}
	//fmt.Println(params)
}

func TestAAA(t *testing.T) {
	//funcName:= "transfer"
	//funcParams := "transfer(\"0x60ceca9c\",\"0x60ceca\",100)"
	//encodeParam(dir,funcName,funcParams)

	byts := []byte("0x00000000000000000000000000000000000000c5")
	fmt.Print(byts)
	//fmt.Printf(string(byts))

	//toAddr :=common.Address{}
	//toAddr.SetBytes([]byte("0x43355c787c50b647c425f594b441d4bd751951c1"))
	//fmt.Printf(toAddr.Hex())
	//
	//
	//toAddr2 :=common.Address{}
	//decode, _ := hexutil.Decode("0x43355c787c50b647c425f594b441d4bd751951c1")
	//toAddr2.SetBytes(decode)
	//fmt.Printf(toAddr2.Hex())

}
func TestBBB(t *testing.T) {
	dir := "D:\\resource\\venachain\\contract\\venachain-contract\\temp\\contracta.cpp.abi.json"
	funcName := "atransfer2"
	funcParams := "atransfer2(\"eeeeeee\",\"ffffff\",3333)"
	//funcParams := "transfer(\"0x43355c787c50b647c425f594b441d4bd751951c1\")"

	encodeParam(dir, funcName, funcParams)

}

func TestCCC(t *testing.T) {
	b := common.BytesToHash(common.Int64ToBytes(int64(1231)))
	fmt.Println(bytes.Equal(b[:24], make([]byte, 24)))
	fmt.Println(b[24:])
	fmt.Print(common.Int64ToBytes(int64(1231)))

}
