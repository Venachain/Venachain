package core

import (
	"gopkg.in/urfave/cli.v1"
)

var (
	/*
		AbiFlag = cli.StringFlag{
			Name:  "abi",
			Usage: "abi file of the privacy token contract.",
		}*/
	UrlFlag = cli.StringFlag{
		Name:  "url",
		Usage: "url of remote nodes, format: http://<ip>:<port>",
	}
	ContractFlag = cli.StringFlag{
		Name:  "contract",
		Usage: "contract address of privacy token contract",
	}
	TxSenderFlag = cli.StringFlag{
		Name:  "bc-owner",
		Usage: "account registered in the bc token contract",
	}
	AccountFlag = cli.StringFlag{
		Name:  "account",
		Usage: "account registered in the privacy token contract",
	}
	TransferPubFlag = cli.StringFlag{
		Name:  "receiver",
		Usage: "public key of the receiver for receiving privacy token",
	}
	ValueFlag = cli.StringFlag{
		Name:  "value",
		Usage: "amount of token",
	}
	DecoyNumFlag = cli.IntFlag{
		Name:  "decoy-num",
		Usage: "2^n of members in the decoy, where n is a number from 2 to 6. e.g. --decoy-num 2 means 4 decoy members",
		Value: 3,
	}
	LeftIntervalFlag = cli.Int64Flag{
		Name:  "l-invl",
		Usage: "left interval of the balance. it is used for revealing the balance from ciphertext",
		/// Value: 0,
	}
	RightIntervalFlag = cli.Int64Flag{
		Name:  "r-invl",
		Usage: "right interval of the balance. it is used for revealing the balance from ciphertext",
		/// Value: 4294967295,
	}
	OutputFlag = cli.StringFlag{
		Name:  "o",
		Usage: "specify the account file name",
		Value: defaultFile,
	}
	ConfigFlag = cli.BoolFlag{
		Name:  "config",
		Usage: "set the value of the config flags provided to the config.json file",
	}
	verbosityFlag = cli.IntFlag{
		Name:  "v",
		Usage: "Logging verbosity: 0=crit, 1=error, 2=warn, 3=info, 4=debug, 5=trace",
		Value: 0,
	}

	registerCmdFlags = []cli.Flag{
		/// AbiFlag,
		ConfigFlag,
		ContractFlag,
		UrlFlag,
		TxSenderFlag,
		verbosityFlag,

		OutputFlag,
	}

	depositCmdFlags = []cli.Flag{
		/// AbiFlag,
		ConfigFlag,
		ContractFlag,
		UrlFlag,
		TxSenderFlag,
		verbosityFlag,

		AccountFlag,
		ValueFlag,
		LeftIntervalFlag,
		RightIntervalFlag,
	}

	transferCmdFlags = []cli.Flag{
		/// AbiFlag,
		ConfigFlag,
		ContractFlag,
		UrlFlag,
		TxSenderFlag,
		verbosityFlag,

		AccountFlag,
		ValueFlag,
		LeftIntervalFlag,
		RightIntervalFlag,
		TransferPubFlag,

		DecoyNumFlag,
	}

	withdrawCmdFlags = []cli.Flag{
		/// AbiFlag,
		ConfigFlag,
		ContractFlag,
		UrlFlag,
		TxSenderFlag,
		verbosityFlag,

		AccountFlag,
		ValueFlag,
		LeftIntervalFlag,
		RightIntervalFlag,
	}

	queryCmdFlags = []cli.Flag{
		/// AbiFlag,
		ConfigFlag,
		ContractFlag,
		UrlFlag,
		TxSenderFlag,
		verbosityFlag,

		AccountFlag,
		LeftIntervalFlag,
		RightIntervalFlag,
	}

	configCmdFlags = []cli.Flag{
		UrlFlag,
		ContractFlag,
		TxSenderFlag,
		verbosityFlag,
	}
)
