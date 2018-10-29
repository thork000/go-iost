package itest

import (
	"fmt"
	"math/rand"
	"os"

	"gopkg.in/urfave/cli.v1"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/crypto"
)

func prepare(c *cli.Context) error {
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

	initAccount, err := account.NewAccount(loadBytes(c.GlobalString("key")), crypto.Ed25519)
	if err != nil {
		return err
	}

	err = initRPC(c.GlobalString("rpc"))
	if err != nil {
		return err
	}

	const initBalance int64 = 1e14
	for i := nodeID * accNum; i < (nodeID+1)*accNum; i++ {
		txHash, err := transfer(initAccount, accounts[i], initBalance)
		if err != nil {
			return err
		}
		fmt.Printf("[\"%v\", \"%v\", %v] txHash: %v\n", initAccount.ID, accounts[i].ID, initBalance, saveBytes(txHash))
	}

	f, err := os.Create(c.GlobalString("dump"))
	if err != nil {
		return err
	}
	defer f.Close()

	const maxTransfer int = 1e9
	for i := 0; i < total; i += accNum {
		for j := 0; j < trxNum; j++ {
			f.WriteString(fmt.Sprintf("%v,%v,%v\n", i+rand.Intn(accNum), rand.Intn(total), rand.Intn(maxTransfer)))
		}
	}

	return nil
}
