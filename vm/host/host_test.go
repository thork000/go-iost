package host

import (
	"testing"

	"github.com/iost-official/Go-IOS-Protocol/core/state"
	"github.com/iost-official/Go-IOS-Protocol/db"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTransfer(t *testing.T) {
	Convey("Test of transfer", t, func() {
		db, _ := db.DatabaseFactory("redis")
		mdb := state.NewDatabase(db)
		pool := state.NewPool(mdb)
		pool.PutHM("iost", "a", state.MakeVTokenByLiteral(100))
		pool.PutHM("iost", "b", state.MakeVTokenByLiteral(100))

		ans := Transfer(pool, "a", "b", 20)
		So(ans, ShouldBeTrue)
		aa, _ := pool.GetHM("iost", "a")
		So(aa.(*state.VToken).ToInt64(), ShouldEqual, 80000000000)
		bb, _ := pool.GetHM("iost", "b")
		So(bb.(*state.VToken).ToInt64(), ShouldEqual, 120000000000)

	})
}
