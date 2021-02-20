package cmd

import (
	"reflect"
	"testing"

	cmd_common "github.com/PlatONEnetwork/PlatONE-Go/cmd/platonecli/common"
	"github.com/PlatONEnetwork/PlatONE-Go/common"
	"github.com/stretchr/testify/assert"
)

const (
	testConfigFilePath = "./test/test_case/config.json"
	testAccount        = "0x60ceca9c1290ee56b98d4e160ef0453f7c40d219"
)

//TODO ???
func TestWriteConfigFile(t *testing.T) {
	WriteConfigFile(testConfigFilePath, "account", testAccount)
	//writeConfigFile(TEST_CONFIG_FILE_PATH, "wrong_key", "0x000..00")
	config := ParseConfigJson(testConfigFilePath)

	t.Logf("the config values are %+v", *config)
}

func TestParamParse(t *testing.T) {
	var r string

	testCase := []struct {
		param     string
		paramName string
		result    interface{}
	}{
		{testAccount, "contract", true},
		{"Alice_02", "contract", false},
		//{"Alice.bob", "contract"},
		//{"na*&2", "contract"},
		//{"-1", "p2pPort"},
		{"123", "p2pPort", int64(123)},
		//{"123456", "p2pPort"},
		{"123456", "", "123456"},
		{"invalid", "status", 2},
		{"approve", "operation", "2"},
		{"observer", "type", 0},
	}

	for i, data := range testCase {
		result := cmd_common.ParamParse(data.param, data.paramName)
		assert.Equal(t, data.result, result, "FAILED")

		t.Logf("%s: case %d: Before: (%v) %s, After convert: (%v) %v\n", r, i, reflect.TypeOf(data.param), data.param, reflect.TypeOf(result), result)
	}
}

func TestChainParamConvert(t *testing.T) {
	testCase := []struct {
		param     string
		paramName string
		expResult interface{}
	}{
		{"0x002", "value", "0x2"},
		{"0020", "value", "0x14"},
		{"-20", "value", "-0x14"},
		{"0xFD", "value", "0xfd"}, //TODO 负数?
		{"0x020", "gas", "0x20"},
		{"002302", "gas", "0x8fe"},
		{testAccount, "to", common.HexToAddress(testAccount)},
	}

	for i, data := range testCase {
		result := cmd_common.ChainParamConvert(data.param, data.paramName)
		assert.Equal(t, data.expResult, result, "FAILED")

		t.Logf("case %d: Before: (%v) %s, After convert: (%v) %v\n", i, reflect.TypeOf(data.param), data.param, reflect.TypeOf(result), result)
	}
}

func TestParamValid(t *testing.T) {
	testCase := []struct {
		param     string
		paramName string
	}{
		{"*", "fw"},
		{testAccount, "fw"},
		{testAccount, "contract"},
		{"Alice_02", "contract"},
		//{"Alice.bob", "contract"},
		{"accept", "action"},
		//{"xxx", "action"},
		{"127.0.0.1:6791", "url"},
		{"127.0.0.1", "externalIP"},
		{"[\"nodeAdmin \"]", "roles"},
		{"fd.deng@wxblockchain.com", "email"},
		{"13240283946", "mobile"},
		{"0.0.0.1", "version"},
		{"-123", "num"},
		{"+13", "num"},
		{"12459234", "num"},
		// {"+-123", "num"},
	}

	for i, data := range testCase {
		paramValid(data.param, data.paramName)
		t.Logf("case %d: the %s \"%s\" is valid\n", i, data.paramName, data.param)
	}

}

//TODO 能否进行合并 funcparse 与 get function params
func TestFuncParse(t *testing.T) {
	funcArray := []struct {
		funcName   string
		funcParams []string
		expParams  []string
	}{
		//{"", nil},
		{"set", []string{"123", "true"}, []string{"123", "true"}},
		{"set", []string{""}, []string{""}},
		{"set", nil, nil},
		{"set()", []string{"123", "true"}, []string{"123", "true"}},
	}

	for i, data := range funcArray {
		t.Logf("case %d: \n", i)
		name, params := FuncParse(data.funcName, data.funcParams)
		assert.Equal(t, "set", name, "name parse FAILED")
		assert.Equal(t, data.expParams, params, "params parse FAILED")
		t.Logf("Before: function name: %s, function params: %s\n", data.funcName, data.funcParams)
		t.Logf("After:  function name: %s, function params: %s\n", name, params)
	}
}

func TestGetFuncParam(t *testing.T) {
	testCases := []struct {
		function  string
		expName   string
		expParams []string
	}{
		{"set()", "set", nil},
		{"set(\"1\",'b' , 1.2, true)", "set", []string{"1", "b", "1.2", "true"}},
		{"set('[\"chainAdmin\",\"nodeAdmin\"]', [\"chainAdmin\",\"nodeAdmin\"])", "set", []string{"[\"chainAdmin\",\"nodeAdmin\"]", "[\"chainAdmin\",\"nodeAdmin\"]"}},
		{"set({\"key\":\"value\"})", "set", []string{"{\"key\":\"value\"}"}},
		{"set(\"{\"key\":\"{\"name\": \"alice\", \"score\": \"[12, 25.0, 35]\"}\"}\")", "set", []string{"{\"key\":\"{\"name\":\"alice\",\"score\":\"[12,25.0,35]\"}\"}"}},
		{"set(show(), 1000 ) ", "set", []string{"show()", "1000"}},
	}

	for i, data := range testCases {
		t.Logf("case %d: %s", i, data)
		name, params := GetFuncNameAndParams(data.function)
		assert.Equal(t, data.expName, name, "name parse FAILED")
		assert.Equal(t, data.expParams, params, "params parse FAILED")

		t.Logf("result: function name: %s, function params: %s\n", name, params)
	}
}

func TestPrintJson(t *testing.T) {
	str := "{\"account\":\"\",\"url\":\"http://127.0.0.1:6794\",\"keystore\":" +
		"\"../../release/linux/data/node-0/keystore/UTC--2020-07-27T03-08-50.310696196Z--8bc9cbeac3b9e89c47b3d0f21ba93b8a6e0aa818\"}"
	result := PrintJson([]byte(str))
	t.Logf("\n%s", result)
}
