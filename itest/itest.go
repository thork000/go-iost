package itest

import (
	"os"

	"gopkg.in/urfave/cli.v1"
	"gopkg.in/urfave/cli.v1/altsrc"
)

type ITest struct {
	*cli.App
}

func flags() []cli.Flag {
	return []cli.Flag{
		altsrc.NewInt64Flag(cli.Int64Flag{
			Name:  "seed, s",
			Value: 1,
			Usage: "Initialize the random number generator with `SEED`",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "id, i",
			Value: 0,
			Usage: "Specify the `ID` of this node",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "node, n",
			Value: 1,
			Usage: "`NUM` of test nodes",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "account, a",
			Value: 100,
			Usage: "`NUM` of accounts per node",
		}),
		altsrc.NewIntFlag(cli.IntFlag{
			Name:  "transaction, t",
			Value: 1000,
			Usage: "`NUM` of transactions per node",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "key, k",
			Value: "1rANSfcRzr4HkhbUFZ7L1Zp69JZZHiDDq5v7dNSbbEqeU4jxy3fszV4HGiaLQEyqVpS1dKT9g7zCVRxBVzuiUzB",
			Usage: "Initialize accounts using an account whose private key is `KEY`",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "rpc, r",
			Value: "localhost:30002",
			Usage: "Set rpc address to `ADDR:PORT`",
		}),
		altsrc.NewStringFlag(cli.StringFlag{
			Name:  "dump, d",
			Value: "transactions.txt",
			Usage: "`FILE` contains generated transactions",
		}),
		cli.StringFlag{
			Name:  "config, c",
			Value: "config.yml",
			Usage: "Load configuration from this `FILE`",
		},
	}
}

func New() *ITest {
	app := cli.NewApp()
	app.Name = "itest"
	app.Usage = "tool for testing the IOST test net"
	app.Authors = []cli.Author{
		{
			Name:  "Wei Zhang",
			Email: "wei@iost.io",
		},
	}
	app.Flags = flags()
	app.Before = func(c *cli.Context) error {
		if _, err := os.Stat(c.String("config")); os.IsNotExist(err) {
			return nil
		}
		return altsrc.InitInputSourceWithContext(app.Flags, altsrc.NewYamlSourceFromFlagFunc("config"))(c)
	}

	app.Commands = []cli.Command{
		{
			Name:    "prepare",
			Aliases: []string{"p"},
			Usage:   "Prepare test data",
			Action:  prepare,
		},
		{
			Name:    "validate",
			Aliases: []string{"v"},
			Usage:   "Run validation test",
			Action:  validate,
		},
		{
			Name:    "run",
			Aliases: []string{"r"},
			Usage:   "Run performance test",
			Action:  run,
		},
	}

	return &ITest{App: app}
}
