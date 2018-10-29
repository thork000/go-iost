package main

import (
	"log"
	"os"

	"github.com/iost-official/go-iost/itest"
)

func main() {
	itest := itest.New()
	if err := itest.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
