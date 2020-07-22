package source_test

import (
	"encoding/hex"
	"fmt"
	"github.com/zenghq3/substrate-go/config"
	"github.com/zenghq3/substrate-go/rpc"
	"github.com/zenghq3/substrate-go/tx"
	"io/ioutil"
	"testing"
)

func TestLoadTypeRegistry(t *testing.T) {
	cc, err := ioutil.ReadFile(fmt.Sprintf("%s.json", config.Edgeware))
	if err != nil {
		panic(err)
	}
	fmt.Println(hex.EncodeToString(cc))
}

func TestSendTx(t *testing.T) {
	var client, err = rpc.New("http://192.168.1.14:9933", "", "")
	if err != nil {
		return
	}
	originTx := tx.CreateTransaction("5D58Xe8fMeCndF6zN4osDkHMwybv4t995rVrYnHWvr4Qwbz4", "5D58Xe8fMeCndF6zN4osDkHMwybv4t995rVrYnHWvr4Qwbz4", uint64(120), uint64(1), uint64(10))
	originTx.SetGenesisHashAndBlockHash("0x44ef51c86927a1e2da55754dba9684dd6ff9bac8c61624ffe958be656c42e036", "0x76ddef34654c3b2d9a2e355c98b99fd43bbad817595a43a9090c0d23db0a4da4", 1452065)
	originTx.SetSpecVersionAndCallId(uint32(client.SpecVersion), uint32(client.TransactionVersion), config.CallIdKusama)
	_, message, err := originTx.CreateEmptyTransactionAndMessage()
	if err != nil {
		return
	}
	sig, err := originTx.SignTransaction("0xeb0b090dcbfd2a3de74b9444138759aa3b5618edd25735c49620be039adcc6c8", message)
	if err != nil {
		return
	}
	txHex, err := originTx.GetSignTransaction(sig)
	if err != nil {
		return
	}
	fmt.Println(txHex)

	txidBytes, err := client.Rpc.SendRequest("author_submitExtrinsic", []interface{}{txHex})
	if err != nil {
		panic(err)
	}
	txid := string(txidBytes)
	fmt.Println(txid)
}
