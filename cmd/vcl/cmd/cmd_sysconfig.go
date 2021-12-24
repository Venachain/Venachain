package cmd

import (
	"fmt"
	"log"
	"reflect"
	"strconv"

	"github.com/Venachain/Venachain/cmd/utils"
	precompile "github.com/Venachain/Venachain/cmd/vcl/client/precompiled"
	cmd_common "github.com/Venachain/Venachain/cmd/vcl/common"
	"github.com/Venachain/Venachain/core/vm"
	"gopkg.in/urfave/cli.v1"
)

const (
	txUseGas    = "use-gas" // IsTxUseGas
	txNotUseGas = "not-use"

	conAudit    = "audit"
	conNotAudit = "not-audit"

	checkPerm    = "with-perm"
	notCheckPerm = "without-perm"

	prodEmp    = "allow-empty"
	notProdEmp = "notallow-empty"
)

var (
	SysConfigCmd = cli.Command{
		Name:  "sysconfig",
		Usage: "Manage the system configurations",

		Subcommands: []cli.Command{
			setCfg,
			getCfg,
		},
	}

	setCfg = cli.Command{
		Name:   "set",
		Usage:  "set the system configurations",
		Action: setSysConfig,
		Flags:  sysConfigCmdFlags,
	}

	getCfg = cli.Command{
		Name:   "get",
		Usage:  "get the system configurations",
		Action: getSysConfig,
		Flags:  getSysConfigCmdFlags,
	}
)

func setSysConfig(c *cli.Context) {

	/*
		if c.NumFlags() > 1 {
			utils.Fatalf("please set one system configuration at a time")
		}*/

	txGasLimit := c.String(TxGasLimitFlags.Name)
	blockGasLimit := c.String(BlockGasLimitFlags.Name)
	isTxUseGas := c.String(IsTxUseGasFlags.Name)
	isApproveDeployedContract := c.String(IsApproveDeployedContractFlags.Name)
	isCheckContractDeployPermission := c.String(IsCheckContractDeployPermissionFlags.Name)
	isProduceEmptyBlock := c.String(IsProduceEmptyBlockFlags.Name)
	gasContractName := c.String(GasContractNameFlags.Name)
	vrfParams := c.String(VrfParamsFlags.Name)

	// temp solution
	if len(txGasLimit)+len(blockGasLimit)+len(isTxUseGas)+len(isApproveDeployedContract)+
		len(isCheckContractDeployPermission)+len(isProduceEmptyBlock)+len(gasContractName) > 15 {
		utils.Fatalf("please set one system configuration at a time")
	}
	if txGasLimit != "" {
		setConfig(c, txGasLimit, vm.TxGasLimitKey)
	}
	if blockGasLimit != "" {
		setConfig(c, blockGasLimit, vm.BlockGasLimitKey)
	}
	if isTxUseGas != "" {
		setConfig(c, isTxUseGas, vm.IsTxUseGasKey)
	}
	if isApproveDeployedContract != "" {
		setConfig(c, isApproveDeployedContract, vm.IsApproveDeployedContractKey)
	}
	if isCheckContractDeployPermission != "" {
		setConfig(c, isCheckContractDeployPermission, vm.IsCheckContractDeployPermissionKey)
	}
	if isProduceEmptyBlock != "" {
		setConfig(c, isProduceEmptyBlock, vm.IsProduceEmptyBlockKey)
	}
	if gasContractName != "" {
		setConfig(c, gasContractName, vm.GasContractNameKey)
	}
	if vrfParams != "" {
		setConfig(c, vrfParams, vm.VrfParamsKey)
	}
}

func setConfig(c *cli.Context, param string, name string) {
	// todo: optimize the code, param check, param convert
	var funcName string
	if !checkConfigParam(param, name) {
		return
	}

	newParam, err := sysConfigConvert(param, name)
	if err != nil {
		utils.Fatalf(err.Error())
	}
	if name == "IsCheckContractDeployPermission" {
		funcName = "setCheckContractDeployPermission"
	} else {
		funcName = "set" + name
	}
	//funcName = "set" + name
	funcParams := cmd_common.CombineFuncParams(newParam)

	result := contractCall(c, funcParams, funcName, precompile.ParameterManagementAddress)
	fmt.Printf("%s\n", result)
}

func checkConfigParam(param string, key string) bool {

	switch key {
	case "TxGasLimit":
		// number check
		num, err := strconv.ParseUint(param, 10, 0)
		if err != nil {
			log.Println("param invalid")

			return false
		}

		// param check
		isInRange := vm.TxGasLimitMinValue <= num && vm.TxGasLimitMaxValue >= num
		if !isInRange {
			fmt.Printf("the transaction gas limit should be within (%d, %d)\n",
				vm.TxGasLimitMinValue, vm.TxGasLimitMaxValue)
			return false
		}
	case "BlockGasLimit":
		num, err := strconv.ParseUint(param, 10, 0)
		if err != nil {
			log.Println("param invalid")

			return false
		}

		isInRange := vm.BlockGasLimitMinValue <= num && vm.BlockGasLimitMaxValue >= num
		if !isInRange {
			fmt.Printf("the block gas limit should be within (%d, %d)\n",
				vm.BlockGasLimitMinValue, vm.BlockGasLimitMaxValue)
			return false
		}
	default:
		if param == "" {
			return false
		}
	}

	return true
}

func getSysConfig(c *cli.Context) {

	txGasLimit := c.Bool(TxGasLimitFlags.Name)
	blockGasLimit := c.Bool(BlockGasLimitFlags.Name)
	isTxUseGas := c.Bool(IsTxUseGasFlags.Name)
	isApproveDeployedContract := c.Bool(IsApproveDeployedContractFlags.Name)
	isCheckContractDeployPermission := c.Bool(IsCheckContractDeployPermissionFlags.Name)
	isProduceEmptyBlock := c.Bool(IsProduceEmptyBlockFlags.Name)
	gasContractName := c.Bool(GasContractNameFlags.Name)
	vrfParams := c.Bool(VrfParamsFlags.Name)

	getConfig(c, txGasLimit, vm.TxGasLimitKey)
	getConfig(c, blockGasLimit, vm.BlockGasLimitKey)
	getConfig(c, isTxUseGas, vm.IsTxUseGasKey)
	getConfig(c, isApproveDeployedContract, vm.IsApproveDeployedContractKey)
	getConfig(c, isCheckContractDeployPermission, vm.IsCheckContractDeployPermissionKey)
	getConfig(c, isProduceEmptyBlock, vm.IsProduceEmptyBlockKey)
	getConfig(c, gasContractName, vm.GasContractNameKey)
	getConfig(c, vrfParams, vm.VrfParamsKey)
}

func getConfig(c *cli.Context, isGet bool, name string) {
	var funcName string
	if isGet {
		if name == "IsCheckContractDeployPermission" {
			funcName = "getCheckContractDeployPermission"
		} else {
			funcName = "get" + name
		}

		result := contractCall(c, nil, funcName, precompile.ParameterManagementAddress)

		result = sysconfigToString(result)
		str := sysConfigParsing(result, name)

		fmt.Printf("%s: %v\n", name, str)
	}
}

func sysconfigToString(param interface{}) interface{} {
	value := reflect.TypeOf(param)

	switch value.Kind() {
	case reflect.Uint64:
		return strconv.FormatUint(param.(uint64), 10)

	case reflect.Uint32:
		return strconv.FormatUint(uint64(param.(uint32)), 10)

	case reflect.String:
		return param

	default:
		panic("not support, please add the corresponding type")
	}
}

func sysConfigParsing(param interface{}, paramName string) string {
	if paramName == vm.TxGasLimitKey || paramName == vm.BlockGasLimitKey ||
		paramName == vm.GasContractNameKey || paramName == vm.VrfParamsKey {
		return param.(string)
	}

	conv := genConfigConverter(paramName)
	return conv.Parse(param)
}

func sysConfigConvert(param, paramName string) (string, error) {

	if paramName == vm.TxGasLimitKey || paramName == vm.BlockGasLimitKey || paramName == vm.VrfParamsKey || paramName == vm.GasContractNameKey {
		return param, nil
	}

	conv := genConfigConverter(paramName)
	result, err := conv.Convert(param)
	if err != nil {
		return "", err
	}

	return result.(string), nil
}

func genConfigConverter(paramName string) *cmd_common.Convert {

	switch paramName {
	case vm.IsTxUseGasKey:
		return cmd_common.NewConvert(txUseGas, txNotUseGas, "1", "0", paramName)
	case vm.IsApproveDeployedContractKey:
		return cmd_common.NewConvert(conAudit, conNotAudit, "1", "0", paramName)
	case vm.IsCheckContractDeployPermissionKey:
		return cmd_common.NewConvert(checkPerm, notCheckPerm, "1", "0", paramName)
	case vm.IsProduceEmptyBlockKey:
		return cmd_common.NewConvert(prodEmp, notProdEmp, "1", "0", paramName)
	default:
		utils.Fatalf("invalid system configuration %v", paramName)
	}

	return nil
}
