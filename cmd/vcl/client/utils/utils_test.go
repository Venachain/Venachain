package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/Venachain/Venachain/accounts/abi"
)

const (
	testParseFile = "../test/test_case/wasm/contracta.wasm"
)

func TestParseFileToBytes(t *testing.T) {
	var testErr = errors.New("error is not nil")
	var nullFile = make([]byte, 0)

	testCase := []struct {
		path      string
		fileBytes []byte
		err       error
	}{
		{testParseFile, []byte("test"), nil},               // case 0: correct
		{"", nil, testErr},                                 // case 1: path is null
		{"../test/test_case", nil, testErr},                // case 2: file directory
		{".//", nil, testErr},                              // case 3: ?
		{"../test/test_case/config.txt", nil, testErr},     // case 4: file not exist
		{"../test/test_case/test_null.txt", nullFile, nil}, // case 5: content of file is null
		{".@\"", nil, testErr},                             // case 6: invalid input
	}

	for i, data := range testCase {
		fileBytes, err := ParseFileToBytes(data.path)

		errCorrect := (err != nil) == (data.err != nil)
		fileBytesCorrect := bytes.EqualFold(fileBytes, data.fileBytes) || len(data.fileBytes) > 0

		switch {
		case err != nil && errCorrect:
			t.Logf("case %d: test file path %s: fileBytes: %v, the error is %v\n", i, data.path, fileBytes, err.Error())
		case fileBytesCorrect:
			t.Logf("case %d: test file path %s: no error\n", i, data.path)
		default:
			t.Fail()
		}
	}
}

func TestGetFuncParams(t *testing.T) {
	testCase := "\"1\",'b' , 1.2, true"
	result := abi.GetFuncParams(testCase)

	t.Log(result)
}

func TestStructType(t *testing.T) {
	var testCase = make(map[string]interface{}, 0)
	var S struct{}
	var str = "{\"components\": [{\"internalType\": \"int32\",\"name\": \"x\",\"type\": \"int32\"},{\"internalType\": \"int32\",\"name\": \"y\",\"type\": \"int32\"}],\"internalType\": \"struct TupleTest.Point\",\"name\": \"num\",\"type\": \"tuple\"}"

	testCase["test"] = 2
	_ = json.Unmarshal([]byte(str), &S)

	t.Log(reflect.ValueOf(testCase).Kind())
	t.Log(S)
}

//TODO 重新设计测试
/*
func TestAbiParse(t *testing.T){
	testCase := []struct{
		abiPath string
		contract string
	}{
		//{"", ""},
		//{"", "__sys_UserManager"},
		//{"", CNS_PROXY_ADDRESS},
		{TEST_ABI_FILE_PATH, ""},
		{TEST_ABI_FILE_PATH, CNS_PROXY_ADDRESS},
		//{"", contract_name}, //TODO get abi on chain
		//{"", contract_address},
	}

	for i, data := range testCase{
		t.Logf("case %d: \n",i)
		_ = abiParse(data.abiPath, data.contract)

	}
}*/
