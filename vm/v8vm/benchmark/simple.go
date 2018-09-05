package main

import (
	"fmt"
	"log"
	"time"

	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/ilog"
	"github.com/iost-official/Go-IOS-Protocol/vm/database"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
	"github.com/iost-official/Go-IOS-Protocol/vm/v8vm"
	"os"
	"runtime/pprof"
)

var vmPool *v8.VMPool
var currPath = "/Users/lihaifeng/GoLang/src/github.com/iost-official/Go-IOS-Protocol/vm/v8vm/"

func init() {
	vmPool = v8.NewVMPool(3, 1020)
	vmPool.SetJSPath(currPath + "/v8/libjs/")
	vmPool.Init()
}

func MyInit(conName string) (*host.Host, *contract.Contract) {
	db := database.NewDatabaseFromPath(currPath + "simple.json")
	vi := database.NewVisitor(100, db)

	ctx := host.NewContext(nil)
	ctx.Set("gas_price", int64(1))
	var gasLimit = int64(1000000000000000)
	ctx.GSet("gas_limit", gasLimit)
	ctx.Set("contract_name", conName)
	h := host.NewHost(ctx, vi, nil, ilog.DefaultLogger())

	rawCode := `
class Contract {
	constructor() {

	}

	show() {
		//if (a == 0) return 0;
		//if (a == 1) return 1;
		//return this.show(a-1) + this.show(a-2);
		return "hello world";
	}
}
module.exports = Contract;
`

	code := &contract.Contract{
		ID:   conName,
		Code: rawCode,
	}

	//code.Code, _ = vmPool.Compile(code)

	return h, code
}

func main() {
	f, _ := os.Create(currPath + "/benchmark/cpu.prof")

	var times float64 = 100
	h, code := MyInit("simple")

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	a := time.Now()

	var i float64 = 0
	for ; i < times; i++ {
		_, _, err := vmPool.LoadAndCall(h, code, "show", 10)
		if err != nil {
			log.Fatal(err)
		}
		//println(rs[0].(string))
	}

	timeUsed := time.Since(a).Nanoseconds()
	tps := int(1000 / (float64(timeUsed) / 1000000 / times))
	fmt.Println("time used: ", time.Since(a))
	fmt.Println("each: ", tps)

	v8.Printaa()
	v8.RealRun()
}
