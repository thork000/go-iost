package native

import (
	"errors"

	"fmt"

	"github.com/bitly/go-simplejson"
	"github.com/iost-official/Go-IOS-Protocol/core/contract"
	"github.com/iost-official/Go-IOS-Protocol/vm/host"
)

var domainABIs map[string]*abi

func init() {
	domainABIs = make(map[string]*abi)
	register(&domainABIs, link)
	register(&domainABIs, transferURL)
	register(&domainABIs, domainInit)
}

// const table names
const (
	DHCPTable      = "dhcp_table"
	DHCPRTable     = "dhcp_revert_table"
	DHCPOwnerTable = "dhcp_owner_table"
)

var (
	link = &abi{
		name: "Link",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			url := args[0].(string)
			cid := args[1].(string)

			txInfo, c := h.TxInfo()
			cost.AddAssign(c)
			tij, err := simplejson.NewJson(txInfo)
			if err != nil {
				panic(err)
			}

			applicant := tij.Get("publisher").MustString()

			ownerR, c := h.MapGet(DHCPOwnerTable, url)
			cost.AddAssign(c)
			owner, ok := ownerR.(string)

			if ok && owner != applicant {
				cost.AddAssign(host.CommonErrorCost(1))
				return nil, cost, fmt.Errorf("no privilege of claimed url: %v", owner)
			}

			h.MapPut(DHCPTable, url, cid)
			h.MapPut(DHCPRTable, cid, url)
			h.MapPut(DHCPOwnerTable, url, owner)
			cost.AddAssign(host.PutCost)
			cost.AddAssign(host.PutCost)
			cost.AddAssign(host.PutCost)

			return []interface{}{}, cost, nil
		},
	}
	transferURL = &abi{
		name: "Transfer",
		args: []string{"string", "string"},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			cost = contract.Cost0()
			url := args[0].(string)
			to := args[1].(string)

			txInfo, c := h.TxInfo()
			cost.AddAssign(c)
			tij, err := simplejson.NewJson(txInfo)
			if err != nil {
				panic(err)
			}

			applicant := tij.Get("publisher").MustString()

			ownerR, c := h.MapGet(DHCPOwnerTable, url)
			cost.AddAssign(c)
			owner, ok := ownerR.(string)

			if !ok {
				cost.AddAssign(host.CommonErrorCost(1))
				return nil, cost, errors.New("transfer unclaimed url")
			}

			if owner != applicant {
				cost.AddAssign(host.CommonErrorCost(1))
				return nil, cost, errors.New("no privilege of claimed url")
			}

			h.MapPut(DHCPOwnerTable, url, to)
			cost.AddAssign(host.PutCost)

			return []interface{}{}, cost, nil

		},
	}
	domainInit = &abi{
		name: "init",
		args: []string{},
		do: func(h *host.Host, args ...interface{}) (rtn []interface{}, cost *contract.Cost, err error) {
			return []interface{}{}, host.CommonErrorCost(1), nil
		},
	}
)
