package util

import (
	"data-manager/config"
	"math/big"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/core/types"
)

func Sender(tx *types.Transaction) (common.Address, error) {
	//first try Frontier
	signer := types.FrontierSigner{}
	addr, err := signer.Sender(tx)
	if nil == err {
		return addr, nil
	}

	addr, err = types.NewEIP155Signer(big.NewInt(0).SetUint64(config.Config.ChainConf.ID)).Sender(tx)
	if nil == err {
		return addr, nil
	}

	return types.HomesteadSigner{}.Sender(tx)
}
