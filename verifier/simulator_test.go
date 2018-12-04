package verifier

import (
	"fmt"
	"testing"

	"github.com/iost-official/go-iost/core/tx"
	"github.com/iost-official/go-iost/ilog"
)

func TestIlog(t *testing.T) {
	ilog.Info(&tx.TxReceipt{}, fmt.Errorf("prepare tx error"))
}
