package vm

import (
	"github.com/iost-official/prototype/proto"
	"github.com/iost-official/prototype/vm/lua"
)

func (m proto.Contract) Decode(buf []byte) Contract {
	err := m.Unmarshal(buf)
	if err != nil {
		return nil
	}

	switch m.Lang {
	case "lua":
		c := &lua.Contract{}
	default:
		return nil
	}

}
