package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testContractAddr = "0x942affd352030020d1d4e60160e99045f0c9cc21"
	testPassphrase   = "123456"
)

var (
	txSender                       string
	testCnsPostBody                string
	testCnsPostBodyLackEndPoint    string
	testCnsPostBodyLackInterpreter string
	testCnsPostBodyLackMethod      string
)

// ================== Contract Name Service =========================
func initRouterCnsTestdata() {
	testCnsPostBody = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\": {\"name\": \"tofu\", \"version\": \"0.0.0.1\", \"address\": \"" + testContractAddr + "\"},\"interpreter\": \"wasm\"}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\",\"passphrase\":\"" + testPassphrase + "\"}}"

	testCnsPostBodyLackEndPoint = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\": {\"name\": \"tofu\", \"version\": \"0.0.0.1\", \"address\": \"" + testContractAddr + "\"},\"interpreter\": \"wasm\"}," +
		"\"rpc\":{\"endPoint\": \"\"}}"

	testCnsPostBodyLackInterpreter = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"data\": {\"name\": \"tofu\", \"version\": \"0.0.0.1\", \"address\": \"" + testContractAddr + "\"},\"interpreter\": \"\"}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\",\"passphrase\":\"" + testPassphrase + "\"}}"

	testCnsPostBodyLackMethod = "{\"tx\":{\"from\": \"" + txSender + "\", \"gas\":\"0x10\"}," +
		"\"contract\":{\"method\":\"\", \"data\": {\"name\": \"tofu\", \"version\": \"0.0.0.1\", \"address\": \"" + testContractAddr + "\"},\"interpreter\": \"wasm\"}," +
		"\"rpc\":{\"endPoint\": \"http://127.0.0.1:6791\",\"passphrase\":\"" + testPassphrase + "\"}}"
}

// test-data init
func TestMain(m *testing.M) {
	txSender = getTestTxSender()
	if len(txSender) == 0 {
		println("no keyfile in ", defaultKeyfile)
		return
	}
	initRouterCnsTestdata()
	initRoutContractTest()
	initRouterFwTest()
	initRouterNodeTest()
	initRouterRoleTest()
	initRouterSysconfigTest()
	m.Run()
}

// get txSender from keyfile in dir of defaultKeyfile
func getTestTxSender() string {
	// get file name
	fileList, err := ioutil.ReadDir(defaultKeyfile)
	if err != nil {
		return ""
	}
	fileName := ""
	for _, file := range fileList {
		if file.IsDir() {
			continue
		}
		fileName = file.Name()
		break
	}
	if len(fileName) == 0 {
		return ""
	}
	path := defaultKeyfile + "/" + fileName

	// parse txSender address
	var keyfile = new(struct {
		Address string `json:"address"`
	})
	keyJson, err := ioutil.ReadFile(path)
	if err != nil {
		return ""
	}

	err = json.Unmarshal(keyJson, keyfile)
	if err != nil {
		return ""
	}

	// add 0x prefix if txSender address doesn't have
	if !strings.HasPrefix(keyfile.Address, "0x") {
		keyfile.Address = fmt.Sprintf("0x%s", keyfile.Address)
	}
	return keyfile.Address
}

func TestCnsHandlers(t *testing.T) {
	testCase := []struct {
		method       string
		path         string
		body         string
		expectedCode int
	}{
		// cns
		{"POST", "/cns/components", testCnsPostBody, 200},
		{"POST", "/cns/components", testCnsPostBodyLackEndPoint, 400},
		{"POST", "/cns/components", testCnsPostBodyLackInterpreter, 200},
		{"POST", "/cns/components", testCnsPostBodyLackMethod, 400},

		{"GET", "/cns/components?name=tofu&endPoint=http://127.0.0.1:6791", "", 200},
		{"GET", "/cns/components?page-num=1&page-size=2", "", 200},
		{"GET", "/cns/components?page-size=2", "", 200},
		{"GET", "/cns/components", "", 200},
		{"GET", "/cns/components/state?name=tofu&endPoint=http://127.0.0.1:6791", "", 200},

		{"PUT", "/cns/mappings/tofu", testCnsPostBody, 200},
		{"GET", "/cns/mappings/tofu?version=0.0.0.1&endPoint=http://127.0.0.1:6791", "", 200},

		// error cases
		{"GET", "/cns/components/state?name=tofu&address=" + testContractAddr + "&endPoint=http://127.0.0.1:6791", "", 400},
		{"GET", "/cns/components?name=@tofu", "", 400},
		{"GET", "/cns/components?origin=0x0023", "", 400},
		{"GET", "/cns/components?name=0&page-size=2", "", 400},
	}

	router := genRestRouters()

	for _, data := range testCase {
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(data.body)
		req, _ := http.NewRequest(data.method, data.path, body)
		req.Header.Set("content-type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, data.expectedCode, w.Code)
		t.Log(w.Body)
	}
}
