package itest

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
)

func validate(c *cli.Context) error {
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

	const initBalance int64 = 1e14
	balances := make([]int64, total)
	for i := 0; i < total; i++ {
		balances[i] = initBalance
	}

	err = initRPC(c.GlobalString("rpc"))
	if err != nil {
		return err
	}

	f, err := os.Open(c.GlobalString("dump"))
	if err != nil {
		return err
	}
	defer f.Close()

	var src []int
	var tgt []int
	var val []int

	const gas int64 = 303

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		params := strings.Split(scanner.Text(), ",")
		s, err := strconv.Atoi(params[0])
		if err != nil {
			return err
		}
		t, err := strconv.Atoi(params[1])
		if err != nil {
			return err
		}
		v, err := strconv.Atoi(params[2])
		if err != nil {
			return err
		}
		balances[s] -= int64(v) + gas
		balances[t] += int64(v)

		src = append(src, s)
		tgt = append(tgt, t)
		val = append(val, v)
	}

	for i := nodeID * trxNum; i < (nodeID+1)*trxNum; i++ {
		_, err := transfer(accounts[src[i]], accounts[tgt[i]], int64(val[i]))
		if err != nil {
			return err
		}
	}

	time.Sleep(45 * time.Second)

	for i := 0; i < total; i++ {
		balance, err := getBalance(accounts[i].ID)
		if err != nil {
			return err
		}
		if balance != balances[i] {
			log.Warn(fmt.Sprintf("Account: %v Expected: %v Actual: %v", accounts[i].ID, balances[i], balance))
		}
	}

	return nil
}
