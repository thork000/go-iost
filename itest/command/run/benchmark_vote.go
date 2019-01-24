package run

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/itest"
	"github.com/uber-go/atomic"
	"github.com/urfave/cli"
)

// VoteManager manages accounts' voting and withdrawal.
type VoteManager struct {
	voteTimes map[string]int
	rw        sync.RWMutex
}

// RandomCandidate returns a random candidate.
func (vm *VoteManager) RandomCandidate() *itest.Account {
	cm.rw.RLock()
	defer cm.rw.RUnlock()

	if len(cm.candidates) == 0 {
		return nil
	}
	return cm.candidates[rand.Intn(len(cm.candidates))]
}

// IsCandidate returns whether account is a candidate.
func (cm *CandidateManager) IsCandidate(acc *itest.Account) bool {
	cm.rw.RLock()
	defer cm.rw.RUnlock()

	_, exist := cm.candidateMap[acc.ID]
	return exist
}

// AddCandidate adds a candidate.
func (cm *CandidateManager) AddCandidate(acc *itest.Account) {
	cm.rw.Lock()
	defer cm.rw.Unlock()

	if _, exist := cm.candidateMap[acc.ID]; !exist {
		cm.candidates = append(cm.candidates, acc)
		cm.candidateMap[acc.ID] = len(cm.candidates) - 1
	}
}

// RemoveCandidate removes a candidate.
func (cm *CandidateManager) RemoveCandidate(acc *itest.Account) {
	cm.rw.Lock()
	defer cm.rw.Unlock()

	if i, exist := cm.candidateMap[acc.ID]; exist {
		delete(cm.candidateMap, acc.ID)
		cm.candidates = append(cm.candidates[0:i], cm.candidates[i+1:]...)
	}
}

// BenchmarkVoteCommand is the subcommand for benchmark of vote_producer.iost.
var BenchmarkVoteCommand = cli.Command{
	Name:      "benchmarkVote",
	ShortName: "benchVt",
	Usage:     "Run vote benchmark by given tps",
	Flags:     BenchmarkVoteFlags,
	Action:    BenchmarkVoteAction,
}

// BenchmarkVoteFlags is the list of flags for benchmark.
var BenchmarkVoteFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "tps",
		Value: 20,
		Usage: "The expected ratio of transactions per second",
	},
}

func checkVoteReceipt(ctx context.Context, it *itest.ITest, items <-chan *hashItem) {
	var counter atomic.Int64
	var failed atomic.Int64
	for i := 0; i < 64; i++ {
		go func() {
			select {
			case <-ctx.Done():
				println("checkreceipt done")
				return
			case item := <-items:
				r, err := it.GetRandomClient().CheckTransactionWithTimeout(item.hash, item.expire)
				counter.Inc()
				if err != nil {
					ilog.Errorf("check transaction failed, txHash=%v, err=%v", item.hash, err)
					failed.Inc()
				}
				c, f := counter.Load(), failed.Load()
				if c%1000 == 0 {
					ilog.Infof("check %d receipts, %d success, %d failed", c, c-f, f)
				}
			}
		}()
	}
}

func sendVoteTx(ctx context.Context, it *itest.ITest, accounts []itest.Account, candidates []string, tps int) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	select {
	case <-ctx.Done():
		return
	case <-ticker.C:
		trxs := make([]*itest.Transaction, 0, tps)
		for i := 0; i < tps; i++ {

		}
	}
}

// BenchmarkVoteAction is the action of benchmark.
var BenchmarkVoteAction = func(c *cli.Context) error {
	it, err := itest.Load(c.GlobalString("keys"), c.GlobalString("config"))
	if err != nil {
		return err
	}
	accountFile := c.GlobalString("account")
	accounts, err := itest.LoadAccounts(accountFile)
	if err != nil {
		if err := AccountCaseAction(c); err != nil {
			return err
		}
		if accounts, err = itest.LoadAccounts(accountFile); err != nil {
			return err
		}
	}
	tps := c.Int("tps")
	hashCh := make(chan *hashItem, 4*tps*int(itest.Timeout.Seconds()))
	ctx, cancel := context.WithCancel(context.Background())

	checkVoteReceipt(ctx, it, hashCh)
	sendVoteTx(ctx, it, accounts, tps)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sig
	cancel()
	ilog.Info("quit vote benckmark, wait all goroutines to be closed for one second")
	time.Sleep(time.Second)
	return nil
}
