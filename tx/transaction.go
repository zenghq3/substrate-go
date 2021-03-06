package tx

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	codec "github.com/zenghq3/substrate-go/codes"
	"github.com/zenghq3/substrate-go/sr25519"
	"github.com/zenghq3/substrate-go/ss58"
	"strings"
)

type Transaction struct {
	SenderPubkey       string `json:"sender_pubkey"`    // from address public key ,0x开头
	RecipientPubkey    string `json:"recipient_pubkey"` // to address public key ,0x开头
	Amount             uint64 `json:"amount"`           // 转账金额
	Nonce              uint64 `json:"nonce"`            //nonce值
	Fee                uint64 `json:"fee"`              //手续费
	BlockHeight        uint64 `json:"block_height"`     //最新区块高度
	BlockHash          string `json:"block_hash"`       //最新区块hash
	GenesisHash        string `json:"genesis_hash"`     //
	SpecVersion        uint32 `json:"spec_version"`
	TransactionVersion uint32 `json:"transaction_version"`
	CallId             string `json:"call_id"` //
}

/*
	GenesisHash string
	SpecVersion uint32
*/
func CreateTransaction(from, to string, amount, nonce, fee uint64) *Transaction {
	tx := Transaction{
		SenderPubkey:    AddressToPublicKey(from),
		RecipientPubkey: AddressToPublicKey(to),
		Amount:          amount,
		Nonce:           nonce,
		Fee:             fee,
	}
	return &tx
}

func (tx *Transaction) SetGenesisHashAndBlockHash(genesisHash, blockHash string, blockNumber uint64) {
	tx.GenesisHash = Remove0X(genesisHash)
	tx.BlockHash = Remove0X(blockHash)
	tx.BlockHeight = blockNumber
}

func (tx *Transaction) SetSpecVersionAndCallId(specVersion, transactionVersion uint32, callId string) {
	tx.SpecVersion = specVersion
	tx.TransactionVersion = transactionVersion
	tx.CallId = callId

}
func (tx *Transaction) CreateEmptyTransactionAndMessage() (string, string, error) {
	tp, err := tx.NewTxPayload()
	if err != nil {
		return "", "", err
	}

	return tx.ToJSONString(), tp.ToBytesString(), nil
}

func (*Transaction) SignTransaction(private, message string) (string, error) {
	message = Remove0X(message)
	messageBytes, err := hex.DecodeString(message)
	if err != nil {
		return "", err
	}
	private = Remove0X(private)
	priv, err1 := hex.DecodeString(private)
	if err1 != nil {
		return "", err1
	}
	sig, err2 := sr25519.Sign(priv, messageBytes)
	if err2 != nil {
		return "", err2
	}
	if len(sig) != 64 {
		return "", errors.New("sign fail,sig length is not equal 64")
	}
	return hex.EncodeToString(sig), nil
}

func (tx *Transaction) NewTxPayload() (*TxPayLoad, error) {
	var tp TxPayLoad
	method, err := NewMethodTransfer(tx.RecipientPubkey, tx.Amount)

	if err != nil {
		return nil, err
	}

	tp.Method, err = method.ToBytes(tx.CallId)

	if err != nil {
		return nil, err
	}

	if tx.BlockHeight == 0 {
		return nil, errors.New("invalid block height")
	}

	tp.Era = GetEra(tx.BlockHeight)

	if tx.Nonce == 0 {
		tp.Nonce = []byte{0}
	} else {
		nonce, err := codec.Encode(Compact_U32, uint64(tx.Nonce))
		if err != nil {
			return nil, err
		}
		tp.Nonce, _ = hex.DecodeString(nonce)
	}

	if tx.Fee == 0 {
		//return nil, errors.New("a none zero fee must be payed")
		tp.Fee = []byte{0}
	} else {
		fee, err := codec.Encode(Compact_U32, uint64(tx.Fee))
		if err != nil {
			return nil, err
		}
		tp.Fee, _ = hex.DecodeString(fee)
	}

	specv := make([]byte, 4)
	binary.LittleEndian.PutUint32(specv, tx.SpecVersion)
	tp.SpecVersion = specv
	// 2020/6/15 add transaction version
	transV := make([]byte, 4)
	binary.LittleEndian.PutUint32(transV, tx.TransactionVersion)
	tp.TransactionVersion = transV

	genesis, err := hex.DecodeString(tx.GenesisHash)
	if err != nil || len(genesis) != 32 {
		return nil, errors.New("invalid genesis hash")
	}

	tp.GenesisHash = genesis

	block, err := hex.DecodeString(tx.BlockHash)
	if err != nil || len(block) != 32 {
		return nil, errors.New("invalid block hash")
	}

	tp.BlockHash = block

	return &tp, nil
}

const calPeriod = 64

func GetEra(height uint64) []byte {
	return []byte{0x00}
	phase := height % calPeriod
	index := uint64(6)
	trailingZero := index - 1

	var encoded uint64
	if trailingZero > 1 {
		encoded = trailingZero
	} else {
		encoded = 1
	}

	if trailingZero < 15 {
		encoded = trailingZero
	} else {
		encoded = 15
	}

	encoded += phase / 1 << 4

	first := byte(encoded >> 8)
	second := byte(encoded & 0xff)

	return []byte{second, first}
}

func (tx *Transaction) ToJSONString() string {
	j, _ := json.Marshal(tx)

	return string(j)
}
func AddressToPublicKey(address string) string {
	if address == "" {
		return ""
	}
	pub, err := ss58.DecodeToPub(address)

	if err != nil {
		return ""
	}
	if len(pub) != 32 {
		return ""
	}
	pubHex := hex.EncodeToString(pub)
	return pubHex
}

func Remove0X(hexData string) string {
	if strings.HasPrefix(hexData, "0x") {
		return hexData[2:]
	}
	return hexData
}

func (tx *Transaction) GetSignTransaction(signature string) (string, error) {
	signed := make([]byte, 0)

	signed = append(signed, SigningBitV4)

	if AccounntIDFollow {
		signed = append(signed, 0xff)
	}

	from, err := hex.DecodeString(tx.SenderPubkey)
	if err != nil || len(from) != 32 {
		return "", nil
	}

	signed = append(signed, from...)

	signed = append(signed, 0x01) // ed25519: 0x00 sr25519: 0x01
	signature = Remove0X(signature)
	sig, err := hex.DecodeString(signature)
	if err != nil || len(sig) != 64 {
		return "", nil
	}
	signed = append(signed, sig...)

	if tx.BlockHeight == 0 {
		return "", errors.New("invalid block height")
	}

	signed = append(signed, GetEra(tx.BlockHeight)...)

	if tx.Nonce == 0 {
		signed = append(signed, 0)
	} else {
		nonce, err := codec.Encode(Compact_U32, uint64(tx.Nonce))
		if err != nil {
			return "", err
		}

		nonceBytes, _ := hex.DecodeString(nonce)
		signed = append(signed, nonceBytes...)
		//fmt.Println("nonce",nonce)
		//uNonce:=types.UCompact(tx.Nonce)
		//var buf = bytes.Buffer{}
		//s:=scale.NewEncoder(&buf)
		//errA:=uNonce.Encode(*s)
		//if errA != nil {
		//	return "", fmt.Errorf("encode ucompact nonce error,Err=[%v]",errA)
		//}
		//signed = append(signed,buf.Bytes()...)
	}

	if tx.Fee == 0 {
		//return "", errors.New("a none zero fee must be payed")
		signed = append(signed, []byte{0}...)
	} else {
		fee, err := codec.Encode(Compact_U32, uint64(tx.Fee))
		if err != nil {
			return "", err
		}
		feeBytes, _ := hex.DecodeString(fee)
		signed = append(signed, feeBytes...)
		//uTip:=types.UCompact(tx.Fee)
		//var buf = bytes.Buffer{}
		//s:=scale.NewEncoder(&buf)
		//errA:=uTip.Encode(*s)
		//if errA != nil {
		//	return "", fmt.Errorf("encode ucompact nonce error,Err=[%v]",errA)
		//}
		//signed = append(signed, buf.Bytes()...)
	}

	method, err := NewMethodTransfer(tx.RecipientPubkey, tx.Amount)
	if err != nil {
		return "", err
	}

	methodBytes, err := method.ToBytes(tx.CallId)
	if err != nil {
		return "", err
	}
	signed = append(signed, methodBytes...)

	length, err := codec.Encode(Compact_U32, uint64(len(signed)))

	if err != nil {
		return "", err
	}
	lengthBytes, _ := hex.DecodeString(length)

	lengthBytes[0] += 1
	return "0x" + hex.EncodeToString(lengthBytes) + hex.EncodeToString(signed), nil
}
