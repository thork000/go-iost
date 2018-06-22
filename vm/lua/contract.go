package lua

import (
	"fmt"

	"sort"

	"github.com/iost-official/prototype/common"
	"github.com/iost-official/prototype/proto"
	"github.com/iost-official/prototype/vm"
)

// contract lua智能合约的实现
type Contract struct {
	info vm.ContractInfo
	code string
	main Method
	apis map[string]Method
}

func (c *Contract) Info() vm.ContractInfo {
	return c.info
}
func (c *Contract) SetPrefix(prefix string) {
	c.info.Prefix = prefix
}
func (c *Contract) SetSender(sender vm.IOSTAccount) {
	c.info.Publisher = sender
}
func (c *Contract) AddSigner(signer vm.IOSTAccount) {
	c.info.Signers = append(c.info.Signers, signer)
}
func (c *Contract) API(apiName string) (vm.Method, error) {
	if apiName == "main" {
		return &c.main, nil
	}
	rtn, ok := c.apis[apiName]
	if !ok {
		return nil, fmt.Errorf("api %v: not found", apiName)
	}
	return &rtn, nil
}
func (c *Contract) Code() string {
	return c.code
}
func (c *Contract) Encode() []byte {
	cp := proto.Contract{}
	cp.Code = c.code
	cp.Lang = "lua"
	cp.Apis = make([]*proto.MethodProto, 0)
	cp.Apis = append(cp.Apis, makeMP(c.main))
	for _, v := range c.apis {
		cp.Apis = append(cp.Apis, makeMP(v))
	}
	sort.Sort(mpSlice(cp.Apis))
	buf, _ := cp.Marshal()
	return buf
}

func makeMP(method Method) *proto.MethodProto {
	mp := proto.MethodProto{}
	mp.Name = method.name
	mp.InCnt = int32(method.inputCount)
	mp.OutCnt = int32(method.outputCount)
	switch method.Privilege() {
	case vm.Public:
		mp.Priv = proto.Privilege_PUBLIC
	case vm.Protected:
		mp.Priv = proto.Privilege_PROTECTED
	default:
		mp.Priv = proto.Privilege_PRIVATE
	}
	return &mp
}

func (c *Contract) Decode(b []byte) error { // depreciated
	var cr contractRaw
	_, err := cr.Unmarshal(b[1:])
	var ci vm.ContractInfo
	err = ci.Decode(cr.info)
	if err != nil {
		return err
	}
	c.info = ci
	c.code = string(cr.code)
	if c.apis == nil {
		c.apis = make(map[string]Method)
	}
	for i := 0; i < len(cr.methods); i++ {
		if cr.methods[i].name == "main" {
			c.main = Method{
				cr.methods[i].name,
				int(cr.methods[i].ic),
				int(cr.methods[i].oc),
				vm.Public,
			}
			continue
		}

		c.apis[cr.methods[i].name] = Method{
			cr.methods[i].name,
			int(cr.methods[i].ic),
			int(cr.methods[i].oc),
			vm.Privilege(cr.methods[i].priv),
		}
	}

	return err
}
func (c *Contract) Hash() []byte {
	return common.Sha256(c.Encode())
}

func NewContract(info vm.ContractInfo, code string, main Method, apis ...Method) Contract {
	c := Contract{
		info: info,
		code: code,
		main: main,
	}
	c.apis = make(map[string]Method)
	for _, api := range apis {
		c.apis[api.name] = api
	}
	return c
}

type mpSlice []*proto.MethodProto

func (m mpSlice) Len() int {
	return len(m)
}
func (m mpSlice) Less(i, j int) bool {
	return m[i].Name < m[j].Name
}
func (m mpSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
