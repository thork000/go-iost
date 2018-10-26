package iserver

import (
	"os"
	"fmt"
	"os/exec"
	"strconv"
	"math/rand"
	"time"
	"context"
	"testing"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/vm"
	"github.com/iost-official/go-iost/core/global"
	"github.com/iost-official/go-iost/ilog"
	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/metrics"
	"github.com/iost-official/go-iost/rpc"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/common"

	"github.com/stretchr/testify/suite"
)

const (
	DBPATH = "storage"
)

type IServerTestSuite struct {
	suite.Suite
	conf *common.Config
	iserver *IServer
}

func initMetrics(metricsConfig *common.MetricsConfig) error {
	if metricsConfig == nil || !metricsConfig.Enable {
		return nil
	}
	err := metrics.SetPusher(metricsConfig.PushAddr, metricsConfig.Username, metricsConfig.Password)
	if err != nil {
		return err
	}
	metrics.SetID(metricsConfig.ID)
	return metrics.Start()
}

func getLogLevel(l string) ilog.Level {
	switch l {
	case "debug":
		return ilog.LevelDebug
	case "info":
		return ilog.LevelInfo
	case "warn":
		return ilog.LevelWarn
	case "error":
		return ilog.LevelError
	case "fatal":
		return ilog.LevelFatal
	default:
		return ilog.LevelDebug
	}
}

func initLogger(logConfig *common.LogConfig) {
	if logConfig == nil {
		return
	}
	logger := ilog.New()
	if logConfig.AsyncWrite {
		logger.AsyncWrite()
	}
	if logConfig.ConsoleLog != nil && logConfig.ConsoleLog.Enable {
		consoleWriter := ilog.NewConsoleWriter()
		consoleWriter.SetLevel(getLogLevel(logConfig.ConsoleLog.Level))
		logger.AddWriter(consoleWriter)
	}
	if logConfig.FileLog != nil && logConfig.FileLog.Enable {
		fileWriter := ilog.NewFileWriter(logConfig.FileLog.Path)
		fileWriter.SetLevel(getLogLevel(logConfig.FileLog.Level))
		logger.AddWriter(fileWriter)
	}
	ilog.InitLogger(logger)
}

func sendTx(server *rpc.GRPCServer, src *account.Account, tgt *account.Account, val int64) ([]byte, error) {
	action := tx.NewAction("iost.system", "Transfer", fmt.Sprintf(`["%v", "%v", %v]`, src.ID, tgt.ID, val))

	const gasLimit = 1000
	const gasPrice = 1
	deadline := time.Now().Add(time.Duration(300) * time.Second)

	rtx := tx.NewTx([]*tx.Action{&action}, [][]byte{}, gasLimit, gasPrice, deadline.UnixNano())
	stx, err := tx.SignTx(rtx, src)
	if err != nil {
		return nil, err
	}

	resp, err := server.SendRawTx(context.Background(), &rpc.RawTxReq{Data: stx.Encode()})
	if err != nil {
		return nil, err
	}
	return []byte(resp.Hash), nil
}

func getBalance(server *rpc.GRPCServer, id string) (int64, error) {
	req := rpc.GetBalanceReq{ID: id, UseLongestChain: true}
	value, err := server.GetBalance(context.Background(), &req)
	if err != nil {
		return 0, err
	}
	return value.Balance, nil
}

func (suite *IServerTestSuite) SetupTest() {
	configfile := os.Getenv("GOPATH") + "/src/github.com/iost-official/go-iost/config/iserver.yml"

	conf := common.NewConfig(configfile)

	initLogger(conf.Log)
	ilog.Infof("Config Information:\n%v", conf.YamlString())
	ilog.Infof("build time:%v", global.BuildTime)
	ilog.Infof("git hash:%v", global.GitHash)

	vm.SetUp(conf.VM)

	err := initMetrics(conf.Metrics)
	suite.Nil(err, "init metrics failed.")

	iserver := New(conf)

	err = iserver.Start()
	suite.Nil(err)

	suite.conf = conf
	suite.iserver = iserver
}

func (suite *IServerTestSuite) TestGetNodeInfo() {
	server := suite.iserver.grpc
	nodeInfoRes, err := server.GetNodeInfo(context.Background(), nil)
	suite.Nil(err, "Failed to get NodeInfo from gRPC server.")

	suite.Equal(global.BuildTime, nodeInfoRes.BuildTime)
	suite.Equal(global.GitHash, global.GitHash)
	suite.Equal(0, nodeInfoRes.Network.PeerCount)
}

func (suite *IServerTestSuite) TestGetBalance() {
	server := suite.iserver.grpc
	conf := suite.conf

	acc, err := account.NewAccount(common.Base58Decode(conf.ACC.SecKey), crypto.NewAlgorithm(conf.ACC.Algorithm))
	suite.Nil(err, "Failed to generate the genesis account.")

	balance, err := getBalance(server, acc.ID)
	suite.Nil(err, "Failed tp get balance of the genesis account.")

	v := common.LoadYamlAsViper(conf.Genesis)
	genesisConfig := &common.GenesisConfig{}
	err = v.Unmarshal(genesisConfig)
	suite.Nil(err, "Unable to decode into GenesisConfig.")

	total, err := strconv.ParseInt(genesisConfig.WitnessInfo[1], 10, 64)
	suite.Nil(err, "Wrong WitnessInfo")

	suite.Equal(total, balance)
}

func (suite *IServerTestSuite) TestSendTx() {
	server := suite.iserver.grpc
	conf := suite.conf

	genesisAcc, err := account.NewAccount(common.Base58Decode(conf.ACC.SecKey), crypto.NewAlgorithm(conf.ACC.Algorithm))
	suite.Nil(err, "Failed tp get balance of the genesis account.")

	var accounts []*account.Account
	for i := 0; i < 100; i++ {
		acc, err := account.NewAccount(nil, crypto.Ed25519)
		suite.Nil(err, "Failed to create new account.")
		accounts = append(accounts, acc)
	}

	for _, acc := range accounts {
		_, err = sendTx(server, genesisAcc, acc, rand.Int63n(1e10))
		suite.Nil(err, "Unable to send transaction.")
	}

	v := common.LoadYamlAsViper(conf.Genesis)
	genesisConfig := &common.GenesisConfig{}
	err = v.Unmarshal(genesisConfig)
	suite.Nil(err, "Unable to decode into GenesisConfig.")

	total, err := strconv.ParseInt(genesisConfig.WitnessInfo[1], 10, 64)
	suite.Nil(err, "Wrong WitnessInfo")

	var sum int64 = 0
	for _, acc := range accounts {
		balance, err := getBalance(server, acc.ID)
		suite.Nil(err, "Failed to get balance.")
		sum += balance
	}

	suite.Equal(total, sum, "Some IOST has vanished!")
}

func (suite *IServerTestSuite) TearDownTest() {
	suite.iserver.Stop()
	ilog.Stop()

	cmd := exec.Command("rm", "-r", DBPATH)
	err := cmd.Run()
	suite.Nil(err, "Failed to delete storage folder.")
}

func TestIServerTestSuite(t *testing.T) {
	suite.Run(t, new(IServerTestSuite))
}
