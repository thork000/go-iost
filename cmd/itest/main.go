package main

import (
	"os"

	"github.com/iost-official/go-iost/itest"
	log "github.com/sirupsen/logrus"
)

func main() {
	itest := itest.New()
	if err := itest.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
