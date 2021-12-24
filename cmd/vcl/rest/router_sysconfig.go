package rest

import (
	"strings"

	precompile "github.com/Venachain/Venachain/cmd/vcl/client/precompiled"
	"github.com/gin-gonic/gin"
)

func registerSysConfigRouters(r *gin.Engine) {
	sysConf := r.Group("/sysConfig")
	{
		sysConf.PUT("/block-gas-limit", blockGasLimitHandler)
		sysConf.PUT("/tx-gas-limit", txGasLimitHandler)
		sysConf.PUT("/is-tx-use-gas", isTxUseGasHandler)
		sysConf.PUT("/is-approve-deployed-contract", isApproveDeployedContractHandler)
		sysConf.PUT("/check-contract-deploy-permission", checkContractDeployPermissionHandler)
		sysConf.PUT("/is-produce-empty-block", isProduceEmptyBlockHandler)
		sysConf.PUT("/gas-contract-name", gasContractNameHandler)
		sysConf.PUT("/vrf-params", vrfParamsHandler)

		sysConf.GET("/block-gas-limit", sysConfigGetHandler)
		sysConf.GET("/tx-gas-limit", sysConfigGetHandler)
		sysConf.GET("/is-tx-use-gas", sysConfigGetHandler)
		sysConf.GET("/is-approve-deployed-contract", sysConfigGetHandler)
		sysConf.GET("/check-contract-deploy-permission", sysConfigGetHandler)
		sysConf.GET("/is-produce-empty-block", sysConfigGetHandler)
		sysConf.GET("/gas-contract-name", sysConfigGetHandler)
		sysConf.GET("/vrf-params", sysConfigGetHandler)
	}
}

// ===================== sys config ====================
func blockGasLimitHandler(ctx *gin.Context) {
	funcParams := &struct {
		BlockGasLimit string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func txGasLimitHandler(ctx *gin.Context) {
	funcParams := &struct {
		TxGasLimit string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func isTxUseGasHandler(ctx *gin.Context) {
	funcParams := &struct {
		IsTxUseGas string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func isApproveDeployedContractHandler(ctx *gin.Context) {
	funcParams := &struct {
		IsApproveDeployedContract string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func checkContractDeployPermissionHandler(ctx *gin.Context) {
	funcParams := &struct {
		CheckContractDeployPermission string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func isProduceEmptyBlockHandler(ctx *gin.Context) {
	funcParams := &struct {
		IsProduceEmptyBlock string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func gasContractNameHandler(ctx *gin.Context) {
	funcParams := &struct {
		GasContractName string
	}{}

	sysConfigHandler(ctx, funcParams)
}

func vrfParamsHandler(ctx *gin.Context) {
	funcParams := &struct {
		VRFParams string
	}{}
	sysConfigHandler(ctx, funcParams)
}

func sysConfigHandler(ctx *gin.Context, funcParams interface{}) {
	var contractAddr = precompile.ParameterManagementAddress

	index := strings.LastIndex(ctx.FullPath(), "/")
	str := ctx.FullPath()[index+1:]
	funcName := "set" + strings.Title(UrlParamConvert(str))

	data := newContractParams(contractAddr, funcName, "wasm", nil, funcParams)

	posthandlerCommon(ctx, data)
}

func sysConfigGetHandler(ctx *gin.Context) {
	var contractAddr = precompile.ParameterManagementAddress
	endPoint := ctx.Query("endPoint")

	index := strings.LastIndex(ctx.FullPath(), "/")
	str := ctx.FullPath()[index+1:]
	funcName := "get" + strings.Title(UrlParamConvert(str))

	data := newContractParams(contractAddr, funcName, "wasm", nil, nil)
	queryHandlerCommon(ctx, endPoint, data)
}
