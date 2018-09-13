package main

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var command = `./iwallet -s "54.95.136.154:30002" call Contract5ikJitzkyfxTwG8k2W87zrMGv21cGHYYH3ySoBMcpRBz transfer '["IOSTjBxx7sUJvmxrMiyjEQnz9h5bfNrXwLinkoL9YvWjnrGdbKnBP","IOST4TCYbe4mjfmKKtF5J3QQ1mxA74UafgmCCzDftk4svVQK7aTbv",100]' --expiration 100000 -l 10000 -p 1 -k ./pri`

func shellRun(s string) (string, error) {
	cmd := exec.Command("/bin/bash", "-c", s)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	return out.String(), err
}

func main() {
	args := os.Args

	tps, _ := strconv.Atoi(args[1])

	c := time.Tick(time.Second)

	for _ = range c {
		for i := 0; i < tps; i++ {
			go func() {
				rs, err := shellRun(command)
				if err != nil {
					panic(err)
				}
				println("txHash: ", rs)
			}()
		}
	}
}
