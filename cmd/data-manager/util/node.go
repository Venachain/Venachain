package util

import (
	"context"
	"data-manager/config"
	"math/big"
	"time"

	"github.com/Venachain/Venachain/common"
	"github.com/Venachain/Venachain/core/types"
	"github.com/Venachain/Venachain/venaclient"
	"github.com/sirupsen/logrus"
)

type node struct {
}

var (
	DefaultNode = newNode()
)

func newNode() *node {
	return new(node)
}

func (this *node) LatestBlock() (*types.Block, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

	return newNode().Client().BlockByNumber(ctx, nil)
}

func (this *node) BlockByHeight(h uint64) (*types.Block, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

	return newNode().Client().BlockByNumber(ctx, big.NewInt(int64(h)))
}

func (this *node) TransactionReceipt(hash common.Hash) (*types.Receipt, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

	return newNode().Client().TransactionReceipt(ctx, hash)
}

func (this *node) CodeAt(address common.Address) ([]byte, error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

	return newNode().Client().CodeAt(ctx, address, nil)
}

func (this *node) Client() *venaclient.Client {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

	rawurl := config.Config.SyncConf.RandomURL()
	cli, err := venaclient.DialContext(ctx, rawurl)
	if nil != err {
		logrus.Panicln("dial eth failed")
	}

	return cli
}
