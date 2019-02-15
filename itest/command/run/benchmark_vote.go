package run

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/itest"
	"github.com/urfave/cli"
	"golang.org/x/sync/semaphore"
)

// Voter defines an account who votes.
type Voter struct {
	*itest.Account

	initVotes map[string]int64 // the votes he has voted when itest starts

	incVotes map[string]int64 // the votes he has voted after starting
	rw       sync.RWMutex

	initBonus float64 // the bonus he hasn't withdrawed when itest starts
}

func (v *Voter) recordVotes(candidate string, number int64) {
	v.rw.Lock()
	v.incVotes[candidate] = v.incVotes[candidate] + number
	v.rw.Unlock()
}

func (v *Voter) getIncVotes(candidate string) int64 {
	v.rw.RLock()
	defer v.rw.RUnlock()

	return v.incVotes[candidate]
}

// Candidate defines a candidate account.
type Candidate struct {
	*itest.Account

	initVotes int64 // the votes he has received when itest starts

	initBonus float64 // the bonus he hasn't withdrawed when itest starts
}

// VoteManager manages accounts' voting and withdrawal.
type VoteManager struct {
	it *itest.ITest

	voters   map[string]*Voter
	voterIDs []string
	voterRW  sync.RWMutex

	candidates   map[string]*Candidate
	candidateIDs []string
	candRW       sync.RWMutex

	hashCh chan *hashItem
	quitCh chan struct{}
}

// NewVoteManager returns a new VoteManager instance.
func NewVoteManager(it *itest.ITest, voters []*itest.Account, candidates []*itest.Account) *VoteManager {
	vm := &VoteManager{
		it:           it,
		voters:       make(map[string]*Voter),
		voterIDs:     make([]string, 0, len(voters)),
		candidates:   make(map[string]*Candidate),
		candidateIDs: make([]string, 0, len(candidates)),
		hashCh:       make(chan *hashItem, 10000),
		quitCh:       make(chan struct{}),
	}
	for _, voter := range voters {
		vm.voters[voter.ID] = &Voter{
			Account:   voter,
			initVotes: make(map[string]int64),
			incVotes:  make(map[string]int64),
		}
	}
	for _, candidate := range candidates {
		vm.candidates[candidate.ID] = &Candidate{Account: candidate}
	}
	return vm
}

func (vm *VoteManager) stop() {
	close(vm.quitCh)
}

// getVotes gets a voter's votes map.
func (vm *VoteManager) getVotes(voterID string) (map[string]int64, error) {
	ret := make(map[string]int64)
	data, _, _, err := vm.it.GetRandClient().GetContractStorage("vote.iost", "u_1", voterID, true)
	if err != nil {
		return nil, fmt.Errorf("calling GetContractStorage failed. %v", err)
	}
	if data == "null" {
		return ret, nil
	}
	votesMap := make(map[string][]interface{})
	err = json.Unmarshal([]byte(data), &votesMap)
	if err != nil {
		return nil, fmt.Errorf("json decode `%s` failed. %v", string(data), err)
	}
	for candidate, voteInfo := range votesMap {
		if len(voteInfo) != 3 {
			return nil, fmt.Errorf("vote info length not equal to 3. candID=%v, voteInfo=%v", candidate, voteInfo)
		}
		voteAmount := voteInfo[0].(string) // let it panic if it's not a string
		voteInt, err := strconv.ParseInt(voteAmount, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing %s to int64 failed. %v", voteAmount, err)
		}
		ret[candidate] = voteInt
	}

	return ret, nil
}

// getReceivedVotes gets a candidate's received votes.
func (vm *VoteManager) getReceivedVotes(accID string) (int64, error) {
	data, _, _, err := vm.it.GetRandClient().GetContractStorage("vote.iost", "v_1", accID, true)
	if err != nil {
		return 0, fmt.Errorf("calling GetContractStorage failed. %v", err)
	}
	if data == "null" {
		return 0, nil
	}
	j, err := simplejson.NewJson([]byte(data))
	if err != nil {
		return 0, fmt.Errorf("json decode `%s` failed. %v", string(data), err)
	}
	votes, err := j.Get("votes").String()
	if err != nil {
		return 0, fmt.Errorf("getting votes from json str `%s` failed. %v", string(data), err)
	}
	ret, err := strconv.ParseInt(votes, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing %s to int64 failed. %v", votes, err)
	}
	return ret, nil
}

// getAvailableBonus gets an account's bonus that has not been withdrawed.
func (vm *VoteManager) getAvailableBonus(acc *itest.Account, isCandidate bool) (float64, error) {
	var abi = "getVoterBonus"
	if isCandidate {
		abi = "getCandidateBonus"
	}
	act := tx.NewAction("vote_producer.iost", abi, fmt.Sprintf(`["%v"]`, acc.ID))
	t := itest.NewTransaction([]*tx.Action{act})
	trx, err := acc.Sign(t)
	if err != nil {
		return 0, fmt.Errorf("sign tx failed. %v", err)
	}
	hash, err := vm.it.GetRandClient().SendTransaction(trx, false)
	if err != nil {
		return 0, fmt.Errorf("send tx failed. %v", err)
	}
	_, receipt, err := vm.it.GetRandClient().CheckTransactionWithTimeout(hash, time.Now().Add(time.Second*80))
	if len(receipt.Returns) == 0 {
		return 0, fmt.Errorf("no return from %s ABI. receipt=%+v", abi, *receipt)
	}
	j, err := simplejson.NewJson([]byte(receipt.Returns[0]))
	if err != nil {
		return 0, fmt.Errorf("json decode `%s` failed. %v", string(receipt.Returns[0]), err)
	}
	str, err := j.GetIndex(0).String()
	if err != nil {
		return 0, fmt.Errorf("get returns failed. %v", err)
	}
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing %s to float64 failed. %v", str, err)
	}
	return f, nil
}

// initVoter recovers voters' votes and bonus.
func (vm *VoteManager) initVoter() {
	var wg sync.WaitGroup
	wg.Add(len(vm.voters))

	var brokenVoters sync.Map
	sem := semaphore.NewWeighted(500)
	for _, voter := range vm.voters {
		sem.Acquire(context.Background(), 1)

		voter := voter // create new variable to bind current variable to closure, odder but easier than passing parameters
		go func() {
			defer sem.Release(1)
			defer wg.Done()

			voteMap, err := vm.getVotes(voter.ID)
			if err != nil {
				ilog.Errorf("getting voter %s's votes map failed. err=%v", voter.ID, err)
				brokenVoters.Store(voter.ID, struct{}{})
			} else {
				voter.initVotes = voteMap
			}

			initBonus, err := vm.getAvailableBonus(voter.Account, false)
			if err != nil {
				ilog.Errorf("getting voter %s's init bonus failed. err=%v", voter.ID, err)
				brokenVoters.Store(voter.ID, struct{}{})
			} else {
				voter.initBonus = initBonus
			}
		}()
	}
	wg.Wait()

	brokenVoters.Range(func(k, _ interface{}) bool {
		delete(vm.voters, k.(string))
		return true
	})
	for id := range vm.voters {
		vm.voterIDs = append(vm.voterIDs, id)
	}
}

// initCandidate recovers candidates' votes and bonus.
func (vm *VoteManager) initCandidate() {
	var wg sync.WaitGroup
	wg.Add(len(vm.candidates))

	var brokenCandidates sync.Map
	for _, candidate := range vm.candidates {
		candidate := candidate // create new variable to bind current variable to closure, odder but easier than passing parameters

		// given that candidates won't be too many, otherwise we should use semaphore to limit the goroutines.
		go func() {
			defer wg.Done()

			initVotes, err := vm.getReceivedVotes(candidate.ID)
			if err != nil {
				ilog.Errorf("getting candidate %s's initial received votes failed. err=%v", candidate.ID, err)
				brokenCandidates.Store(candidate.ID, struct{}{})
			} else {
				candidate.initVotes = initVotes
			}

			initBonus, err := vm.getAvailableBonus(candidate.Account, true)
			if err != nil {
				ilog.Errorf("getting candidate %s's initial bonus failed. err=%v", candidate.ID, err)
				brokenCandidates.Store(candidate.ID, struct{}{})
			} else {
				candidate.initBonus = initBonus
			}
		}()
	}
	wg.Wait()

	brokenCandidates.Range(func(k, _ interface{}) bool {
		delete(vm.candidates, k.(string))
		return true
	})
	for id := range vm.candidates {
		vm.candidateIDs = append(vm.candidateIDs, id)
	}
}

func (vm *VoteManager) checkTxLoop() {
	for i := 0; i < 64; i++ {
		go func() {
			for {
				select {
				case <-vm.quitCh:
					return
				case item := <-vm.hashCh:
					t, _, err := vm.it.GetRandClient().CheckTransactionWithTimeout(item.hash, item.expire)
					if err != nil {
						ilog.Errorf("check transaction failed, txHash=%v, err=%v", item.hash, err)
						continue
					}
					for _, action := range t.Actions {
						if action.Contract == "vote_producer.iost" &&
							(action.ActionName == "vote" || action.ActionName == "unvote") {

							var params []string // [from, to, amount]
							err := json.Unmarshal([]byte(action.Data), &params)
							if err != nil {
								ilog.Errorf("json decode `%s` failed. err=%v, txHash=%v", action.Data, err, item.hash)
								continue
							}
							if len(params) != 3 {
								ilog.Errorf("%s abi's parameter length is %d, not equal to 3. txHash=%v", action.ActionName, len(params), item.hash)
								continue
							}
							amount, err := strconv.ParseInt(params[2], 10, 64)
							if err != nil {
								ilog.Errorf("parsing %s to int64 failed. err=%v, txHash=%v", params[2], err, item.hash)
								continue
							}
							if action.ActionName == "unvote" {
								amount = -amount
							}
							voter := vm.getVoter(params[0])
							if voter != nil {
								voter.recordVotes(params[1], amount)
							}
						}
					}
				}
			}
		}()
	}
}

func (vm *VoteManager) getVoter(id string) *Voter {
	vm.voterRW.RLock()
	defer vm.voterRW.RUnlock()

	return vm.voters[id]
}

func (vm *VoteManager) getRandomNVoters(n int) []*Voter {
	voters := make([]*Voter, 0, n)

	vm.voterRW.RLock()
	defer vm.voterRW.RUnlock()

	for i := 0; i < n; i++ {
		voterID := vm.voterIDs[rand.Intn(len(vm.voterIDs))]
		voters = append(voters, vm.voters[voterID])
	}
	return voters
}

func (vm *VoteManager) getCandidate(id string) *Candidate {
	vm.voterRW.RLock()
	defer vm.voterRW.RUnlock()

	return vm.candidates[id]
}

func (vm *VoteManager) getRandomNCandidates(n int) []*Candidate {
	candidates := make([]*Candidate, 0, n)

	vm.candRW.RLock()
	defer vm.candRW.RUnlock()

	for i := 0; i < n; i++ {
		candidateID := vm.candidateIDs[rand.Intn(len(vm.candidateIDs))]
		candidates = append(candidates, vm.candidates[candidateID])
	}
	return candidates
}

// RandomVoteN randomly picks n voters and votes for n random candidates.
func (vm *VoteManager) RandomVoteN(n int) {
	voters := vm.getRandomNVoters(n)
	candidates := vm.getRandomNCandidates(n)
	txs := make([]*itest.Transaction, 0, n)

	for i := 0; i < n; i++ {
		act := tx.NewAction("vote_producer.iost", "vote", fmt.Sprintf(`["%v", "%v", "%v"]`, voters[i].ID, candidates[i].ID, strconv.Itoa(1+rand.Intn(10))))

		// vote : unvote = 4 : 1
		if rand.Intn(5) == 0 {
			votes := voters[i].getIncVotes(candidates[i].ID) + voters[i].initVotes[candidates[i].ID]
			if votes > 0 {
				act = tx.NewAction("vote_producer.iost", "unvote", fmt.Sprintf(`["%v", "%v", "%v"]`, voters[i].ID, candidates[i].ID, strconv.Itoa(1+rand.Intn(int(votes)))))
			}
		}
		t := itest.NewTransaction([]*tx.Action{act})
		signedTx, err := voters[i].Sign(t)
		if err != nil {
			ilog.Errorf("signing tx failed. err=%v, accID=%v, tx=%+v", err, voters[i].ID, t)
			continue
		}
		txs = append(txs, signedTx)
	}

	hashList, errList := vm.it.SendTransactionN(txs, false)
	if len(errList) > 0 {
		ilog.Errorf("send transactionN error list: %v", errList)
	}
	expire := time.Now().Add(itest.Timeout)
	for _, hash := range hashList {
		select {
		case vm.hashCh <- &hashItem{hash: hash, expire: expire}:
		default:
			ilog.Warnf("hash channel full")
		}
	}
}

func topupCandidateBonus(it *itest.ITest, amount string) (string, error) {
	bank := it.GetDefaultAccount()
	act := tx.NewAction("vote_producer.iost", "topupCandidateBonus", fmt.Sprintf(`["%v", "%v"]`, amount, bank.ID))
	t := itest.NewTransaction([]*tx.Action{act})
	trx, err := bank.Sign(t)
	if err != nil {
		return "", err
	}
	return it.SendTransaction(trx, true)
}

func topupVoterBonus(it *itest.ITest, amount, accID string) (string, error) {
	bank := it.GetDefaultAccount()
	act := tx.NewAction("vote_producer.iost", "topupVoterBonus", fmt.Sprintf(`["%v", "%v", "%v"]`, accID, amount, bank.ID))
	t := itest.NewTransaction([]*tx.Action{act})
	trx, err := bank.Sign(t)
	if err != nil {
		return "", err
	}
	return it.SendTransaction(trx, true)
}

func getVoteManager(c *cli.Context, it *itest.ITest) (*VoteManager, error) {
	accountConfig := c.GlobalString("account")
	accounts, err := itest.LoadAccounts(accountConfig)
	if err != nil {
		ilog.Info("found no account config, create new account")
		if err := AccountCaseAction(c); err != nil {
			return nil, err
		}
		if accounts, err = itest.LoadAccounts(accountConfig); err != nil {
			return nil, err
		}
	}
	if len(accounts) > 10000 {
		accounts = accounts[:10000]
	}

	candConfig := c.String("candidate")
	candidates, err := itest.LoadAccounts(candConfig)
	if err != nil {
		return nil, err
	}
	vm := NewVoteManager(it, accounts, candidates)
	vm.initVoter()
	vm.initCandidate()
	vm.checkTxLoop()
	return vm, nil
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
	cli.StringFlag{
		Name:  "candidate",
		Value: "candidates.json",
		Usage: "the candidate account config file",
	},
}

// BenchmarkVoteAction is the action of vote benchmark.
var BenchmarkVoteAction = func(c *cli.Context) error {
	it, err := itest.Load(c.GlobalString("keys"), c.GlobalString("config"))
	if err != nil {
		return err
	}

	vm, err := getVoteManager(c, it)
	if err != nil {
		ilog.Errorf("getVoteManager failed. err=%v", err)
		return err
	}
	defer vm.stop()
	ilog.Infof("candidate number: %d, voter number:%d", len(vm.candidates), len(vm.voters))

	for _, c := range vm.candidates {
		ilog.Warnf("%+v, %+v", c.ID, c)
	}
	for _, v := range vm.voters {
		ilog.Warnf("%+v, %+v", v.ID, v)
	}
	return nil

	tps := c.Int("tps")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	ticker := time.NewTicker(time.Second)
L:
	for {
		select {
		case s := <-sig:
			ilog.Infof("receive quit signal: %v", s)
			break L
		case <-ticker.C:
			if rand.Intn(5) == 1 {

			} else {
				vm.RandomVoteN(tps)
			}
		}
	}

	time.Sleep(time.Second * 5)
	for _, c := range vm.candidates {
		ilog.Warnf("id:%v, %+v", c.ID, c)
	}
	ilog.Info("\n\n\n")
	for _, v := range vm.voters {
		ilog.Warnf("id:%v, %+v", v.ID, v)
	}
	// topupCandidateBonus(it, "10000")
	// hash, err := topupVoterBonus(it, "22000", "producer001")
	return nil
}
