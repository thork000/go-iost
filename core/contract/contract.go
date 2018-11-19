package contract

import (
	"encoding/base64"

	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/gogo/protobuf/proto"
	"github.com/iost-official/go-iost/common"
)

//go:generate protoc --gofast_out=. contract.proto

// VersionCode version of contract
type VersionCode string

// PaymentCode payment mode of contract
type PaymentCode int32

// Payment mode
const (
	SelfPay PaymentCode = iota
	ContractPay
)

// FixedAmount the limit amount of token used by contract
type FixedAmount struct {
	Token string
	Val   *common.Fixed
}

//type ContractInfo struct {
//	Name     string
//	Lang     string
//	Version  VersionCode
//	Payment  PaymentCode
//	Limit    Cost
//	GasPrice uint64
//}
//
//type ContractOld struct {
//	ContractInfo
//	Code string
//}

// Encode contract to string using proto buf
func (m *Contract) Encode() string {
	buf, err := proto.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

// Decode contract from string using proto buf
func (m *Contract) Decode(str string) error {
	err := proto.Unmarshal([]byte(str), m)
	if err != nil {
		return err
	}
	return nil
}

// B64Encode encode contract to base64 string
func (m *Contract) B64Encode() string {
	buf, err := proto.Marshal(m)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(buf)
}

// B64Decode decode contract from base64 string
func (m *Contract) B64Decode(str string) error {
	buf, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return err
	}
	return proto.Unmarshal(buf, m)
}

// DecodeContract static method to decode contract from string
func DecodeContract(str string) *Contract {
	var c Contract
	err := proto.Unmarshal([]byte(str), &c)
	if err != nil {
		panic(err)
	}
	return nil
}

// ABI get abi from contract with specific name
func (m *Contract) ABI(name string) *ABI {
	for _, a := range m.Info.Abi {
		if a.Name == name {
			return a
		}
	}
	return nil
}

// Key get modified key from contract with specific name
func (m *Contract) Key(name string, owner string) (string, error) {
	for _, a := range m.Info.Keys {
		if a.Name == name {
			switch a.Type {
			case Key_Basic:
				return fmt.Sprintf("b-%v-%v", m.ID, a.Name), nil
			case Key_OwnedBasic:
				return fmt.Sprintf("b-%v@%v-%v", m.ID, owner, a.Name), nil
			case Key_Map:
				return fmt.Sprintf("m-%v-%v", m.ID, a.Name), nil
			case Key_OwnedMap:
				return fmt.Sprintf("m-%v@%v-%v", m.ID, owner, a.Name), nil
			}

		}
	}
	return "", fmt.Errorf("key not found")
}

// Compile read src and abi file, generate contract structure
func Compile(id, src, abi string) (*Contract, error) {
	bs, err := ioutil.ReadFile(src)
	if err != nil {
		return nil, err
	}
	code := string(bs)

	as, err := ioutil.ReadFile(abi)
	if err != nil {
		return nil, err
	}

	var info Info
	err = json.Unmarshal(as, &info)
	if err != nil {
		return nil, err
	}
	c := Contract{
		ID:   id,
		Info: &info,
		Code: code,
	}

	return &c, nil
}
