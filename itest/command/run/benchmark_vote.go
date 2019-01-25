package run

import (
	"context"
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

// Candidate defines a candidate account.
type Candidate struct {
	*itest.Account

	initVotes int64 // the votes he has received when itest starts

	availableBonus atomic.Float64 // the bonus he hasn't withdrawed
}

// Voter defines an account who votes.
type Voter struct {
	*itest.Account

	initVotes map[string]int64 // the votes he has voted when itest starts
	incVotes  map[string]int64 // the votes he has voted after starting
}

// CandidateManager manages candidate's votes and withdrawal.
type CandidateManager struct {
	candidates   map[string]*Candidate
	candidateIDs []string
	it           *itest.ITest
}

// VoteManager manages accounts' voting and withdrawal.
type VoteManager struct {
	voters   map[string]*Voter
	voterIDs []string
	rw       sync.RWMutex
}

// NewCandidateManager returns a new CandidateManager instance.
func NewCandidateManager(candidates []*itest.Account, it *itest.ITest) *CandidateManager {
	cm := &CandidateManager{
		candidates:   make(map[string]*Candidate),
		candidateIDs: make([]string, 0, len(candidates)),
		it:           it,
	}
	for _, candidate := range candidates {
		cm.candidates[candidate.ID] = &Candidate{Account: candidate}
		cm.candidateIDs = append(cm.candidateIDs, candidate.ID)
	}
	return cm
}

func (cm *CandidateManager) getVotes(accID string) int64 {
	cm.it.GetRandomClient()
}

// Recover recovers candidates' votes and bonus.
func (cm *CandidateManager) Recover() {

}

// NewVoteManager returns a new VoteManager instance.
func NewVoteManager() *VoteManager {
	return &VoteManager{
		votes: make(map[string]int),
		bonus: make(map[string]float64),
	}
}

// AddVotes add votes to an account.
func (vm *VoteManager) AddVotes(acc *itest.Account, votes int) {
	vm.rw.Lock()
	defer vm.rw.Unlock()

	vm.votes[acc.ID] = vm.votes[acc.ID] + votes
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
