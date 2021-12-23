package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/Venachain/Venachain/cmd/ptransfer/client/core"

	"github.com/Venachain/Venachain/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	app = utils.NewApp("", "the privacy token command line interface")
)

func init() {

	// Initialize the CLI app
	app.Commands = []cli.Command{
		core.RegisterCmd,
		core.DepositCmd,
		core.TransferCmd,
		core.WithdrawCmd,
		core.QueryCmd,
	}
	sort.Sort(cli.CommandsByName(app.Commands))
	app.Commands = append(app.Commands, core.ConfigCmd)
	app.After = func(ctx *cli.Context) error {
		return nil
	}
}

//go:generate go-bindata -pkg core -o ./core/bindata.go ./privacy_contract/...
func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
