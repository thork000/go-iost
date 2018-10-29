package itest

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/iost-official/go-iost/account"
	"github.com/iost-official/go-iost/common"
	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/crypto"
	"github.com/iost-official/go-iost/rpc"

	"google.golang.org/grpc"
)

var client rpc.ApisClient

func genAccounts(num int) (accounts []*account.Account, err error) {
	for i := 0; i < num; i++ {
		acc, err := account.NewAccount(genSeckey(), crypto.Secp256k1)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func initRPC(server string) error {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	client = rpc.NewApisClient(conn)
	return nil
}

func sendTx(stx *tx.Tx) ([]byte, error) {
	resp, err := client.SendRawTx(context.Background(), &rpc.RawTxReq{Data: stx.Encode()})
	if err != nil {
		return nil, err
	}
	return []byte(resp.Hash), nil
}

func transfer(src *account.Account, tgt *account.Account, val int64) ([]byte, error) {
	action := tx.NewAction("iost.system", "Transfer", fmt.Sprintf(`["%v", "%v", %v]`, src.ID, tgt.ID, val))
	rtx := tx.NewTx([]*tx.Action{&action}, [][]byte{}, 1000, 1, time.Now().Add(time.Second*time.Duration(300)).UnixNano())
	stx, err := tx.SignTx(rtx, src)
	if err != nil {
		return nil, err
	}
	txHash, err := sendTx(stx)
	if err != nil {
		return nil, err
	}
	return txHash, nil
}

func getBalance(id string) (int64, error) {
	req := rpc.GetBalanceReq{ID: id, UseLongestChain: true}
	value, err := client.GetBalance(context.Background(), &req)
	if err != nil {
		return 0, err
	}
	return value.Balance, nil
}

func saveBytes(buf []byte) string {
	return common.Base58Encode(buf)
}

func loadBytes(s string) []byte {
	if s[len(s)-1] == 10 {
		s = s[:len(s)-1]
	}
	buf := common.Base58Decode(s)
	return buf
}

func genSeckey() []byte {
	seckey := make([]byte, 32)
	_, err := rand.Read(seckey)
	if err != nil {
		fmt.Printf("Failed to random seckey, %v\n", err)
		return nil
	}
	return seckey
}
