package state

import (
	"github.com/iost-official/Go-IOS-Protocol/iostdb"
	"sync"
)

var PatchDb *iostdb.LDBDatabase

var o sync.Once

//
//func init() {
//	o.Do(func() {
//		var err error
//		PatchDb, err = iostdb.NewLDBDatabase("", 0, 0)
//		if err != nil {
//			panic(err)
//		}
//	})
//}
