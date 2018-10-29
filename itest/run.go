package itest

import (
	"fmt"
	"math/rand"
	"time"

	"gopkg.in/urfave/cli.v1"
)

func run(c *cli.Context) error {
	nodeID := c.GlobalInt("id")
	nodeNum := c.GlobalInt("node")
	accNum := c.GlobalInt("account")
	trxNum := c.GlobalInt("transaction")

	rand.Seed(c.GlobalInt64("seed"))

	total := nodeNum * accNum
	accounts, err := genAccounts(total)
	if err != nil {
		return err
	}

	err = initRPC(c.GlobalString("rpc"))
	if err != nil {
		return err
	}

	const maxTransfer int64 = 1e9

	for i := 0; i < nodeID*trxNum; i++ {
		rand.Intn(total)
		rand.Intn(total)
		rand.Int63n(maxTransfer)
	}

	start := time.Now()
	for i := 0; i < trxNum; i++ {
		_, err = transfer(accounts[rand.Intn(total)], accounts[rand.Intn(total)], rand.Int63n(maxTransfer))
		if err != nil {
			return err
		}
	}
	elapsed := time.Since(start)

	fmt.Printf("%v tx/s\n", float64(trxNum)/elapsed.Seconds())

	return nil
}
