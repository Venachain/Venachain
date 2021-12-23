package core

import (
	"testing"
)

var (
	configPath = "D:\\resource\\venachain\\Venachain\\src\\github.com\\Venachain\\Venachain\\cmd\\ctool\\config.json"
	pkFilePath = "../test/privateKeys.txt"
)

func TestPrepareAccount(t *testing.T) {
	parseConfigJson(configPath)
	err := PrepareAccount(10, pkFilePath, "0xDE0B6B3A7640000")
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestStressTest(t *testing.T) {
	parseConfigJson(configPath)
	err := StabilityTest(pkFilePath, 1, 10)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
