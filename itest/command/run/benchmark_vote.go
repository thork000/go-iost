package run

import (
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/iost-official/go-iost/itest"
	"github.com/urfave/cli"
)

type CandidateManager struct {
	candidates   []*itest.Account
	candidateMap map[string]int
	rw           sync.RWMutex
}

func (cm *CandidateManager) RandomCandidate() *itest.Account {
	cm.rw.RLock()
	defer cm.rw.RUnlock()

	if len(cm.candidates) == 0 {
		return nil
	}
	return cm.candidates[rand.Intn(len(cm.candidates))]
}

func (cm *CandidateManager) IsCandidate(acc *itest.Account) bool {
	cm.rw.RLock()
	defer cm.rw.RUnlock()

	_, exist := cm.candidateMap[acc.ID]
	return exist
}

func (cm *CandidateManager) AddCandidate(acc *itest.Account) {
	cm.rw.Lock()
	defer cm.rw.Unlock()

	if _, exist := cm.candidateMap[acc.ID]; !exist {
		cm.candidates = append(cm.candidates, acc)
		cm.candidateMap[acc.ID] = len(cm.candidates) - 1
	}
}

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
	accountMap := make(map[string]*itest.Account)
	for _, acc := range accounts {
		accountMap[acc.ID] = acc
	}
	tps := c.Int("tps")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	return nil
}
