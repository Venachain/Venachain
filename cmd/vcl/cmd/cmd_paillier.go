package cmd

import (
	"fmt"

	precompile "github.com/Venachain/Venachain/cmd/vcl/client/precompiled"
	"gopkg.in/urfave/cli.v1"
)

var (
	// fire wall
	PlCmd = cli.Command{
		Name:     "pl",
		Usage:    "contract paillier",
		Category: "pl",
		Subcommands: []cli.Command{
			PaillierWeightAddCmd,
			PaillierAddCmd,
			PaillierMulCmd,
		},
	}

	PaillierWeightAddCmd = cli.Command{
		Name:   "paillierWeightAdd",
		Action: paillierWeightAdd,
		Flags:  globalCmdFlags,
	}

	PaillierAddCmd = cli.Command{
		Name:   "paillierAdd",
		Action: paillierAdd,
		Flags:  globalCmdFlags,
	}

	PaillierMulCmd = cli.Command{
		Name:   "paillierMul",
		Action: paillierMul,
		Flags:  globalCmdFlags,
	}
)

func paillierWeightAdd(c *cli.Context) {
	funcName := "paillierWeightAdd"
	arg := c.Args().First()
	arr := c.Args().Get(1)
	pubKey := c.Args().Get(2)
	fmt.Println("arg", arg)
	fmt.Println("arr", arr)
	fmt.Println("pubKey", pubKey)
	result := contractCall(c, []string{arg, arr, pubKey}, funcName, precompile.PaillierAddress)
	fmt.Printf("result: %s\n", result)
}

func paillierAdd(c *cli.Context) {
	funcName := "paillierAdd"
	arg := c.Args().First()
	pubKey := c.Args().Get(1)

	result := contractCall(c, []string{arg, pubKey}, funcName, precompile.PaillierAddress)
	fmt.Printf("result: %s\n", result)
}

func paillierMul(c *cli.Context) {
	funcName := "paillierMul"
	arg := c.Args().First()
	scalar := c.Args().Get(1)
	pubKey := c.Args().Get(2)
	result := contractCall(c, []string{arg, scalar, pubKey}, funcName, precompile.PaillierAddress)
	fmt.Printf("result: %s\n", result)
}
