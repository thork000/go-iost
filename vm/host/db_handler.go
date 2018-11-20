package host

import (
	"github.com/iost-official/go-iost/core/contract"
	"github.com/iost-official/go-iost/vm/database"
	"strings"
)

// DBHandler is an application layer abstraction of our base basic_handler and map_handler.
// it offers interface which has an interface{} type value and ramPayer semantic
// it also handles the Marshal and Unmarshal work and determine the cost of each operation
type DBHandler struct {
	h *Host
}

// NewDBHandler ...
func NewDBHandler(h *Host) DBHandler {
	return DBHandler{
		h: h,
	}
}

// Put put kv to db
func (h *DBHandler) Put(key string, value interface{}, ramPayer ...string) contract.Cost {
	mk := h.modifyKey(key)
	sv := h.modifyValue(value, ramPayer...)
	h.payRAM(mk, sv, ramPayer...)
	h.h.db.Put(mk, sv)
	return PutCost
}

// Get get value of key from db
func (h *DBHandler) Get(key string) (value interface{}, cost contract.Cost) {
	mk := h.modifyKey(key)
	rtn := h.parseValue(h.h.db.Get(mk))
	return rtn, GetCost
}

// Del delete key
func (h *DBHandler) Del(key string) contract.Cost {
	mk := h.modifyKey(key)
	h.releaseRAM(mk)
	h.h.db.Del(mk)
	return DelCost
}

// Has if db has key
func (h *DBHandler) Has(key string) (bool, contract.Cost) {
	mk := h.modifyKey(key)
	return h.h.db.Has(mk), GetCost
}

// MapPut put kfv to db
func (h *DBHandler) MapPut(key, field string, value interface{}, ramPayer ...string) contract.Cost {
	mk := h.modifyKey(key)
	sv := h.modifyValue(value, ramPayer...)

	h.payRAMForMap(mk, field, sv, ramPayer...)
	h.h.db.MPut(mk, field, sv)
	return PutCost
}

// MapGet get value by kf from db
func (h *DBHandler) MapGet(key, field string) (value interface{}, cost contract.Cost) {
	mk := h.modifyKey(key)
	rtn := h.parseValue(h.h.db.MGet(mk, field))
	return rtn, GetCost
}

// MapKeys list keys
func (h *DBHandler) MapKeys(key string) (fields []string, cost contract.Cost) {
	mk := h.modifyKey(key)
	return h.h.db.MKeys(mk), KeysCost
}

// MapDel delete field
func (h *DBHandler) MapDel(key, field string) contract.Cost {
	mk := h.modifyKey(key)
	h.releaseRAMForMap(mk, field)
	h.h.db.MDel(mk, field)
	return DelCost
}

// MapHas if has field
func (h *DBHandler) MapHas(key, field string) (bool, contract.Cost) {
	mk := h.modifyKey(key)
	return h.h.db.MHas(mk, field), GetCost
}

// MapLen get length of map
func (h *DBHandler) MapLen(key string) (int, contract.Cost) {
	keys, cost := h.MapKeys(key)
	return len(keys), cost
}

// GlobalHas if another contract's db has key
func (h *DBHandler) GlobalHas(con, key string) (bool, contract.Cost) {
	mk := h.modifyGlobalKey(con, key)
	return h.h.db.Has(mk), GetCost
}

// GlobalGet get another contract's data
func (h *DBHandler) GlobalGet(con, key string) (value interface{}, cost contract.Cost) {
	mk := h.modifyGlobalKey(con, key)
	rtn := h.parseValue(h.h.db.Get(mk))
	return rtn, GetCost
}

// GlobalMapHas if another contract's map has field
func (h *DBHandler) GlobalMapHas(con, key, field string) (bool, contract.Cost) {
	mk := h.modifyGlobalKey(con, key)
	return h.h.db.MHas(mk, field), GetCost
}

// GlobalMapGet get another contract's map data
func (h *DBHandler) GlobalMapGet(con, key, field string) (value interface{}, cost contract.Cost) {
	mk := h.modifyGlobalKey(con, key)
	rtn := h.parseValue(h.h.db.MGet(mk, field))
	return rtn, GetCost
}

// GlobalMapKeys get another contract's map keys
func (h *DBHandler) GlobalMapKeys(con, key string) (keys []string, cost contract.Cost) {
	mk := h.modifyGlobalKey(con, key)
	return h.h.db.MKeys(mk), GetCost
}

// GlobalMapLen get another contract's map length
func (h *DBHandler) GlobalMapLen(con, key string) (length int, cost contract.Cost) {
	k, cost := h.GlobalMapKeys(con, key)
	return len(k), cost
}

func (h *DBHandler) modifyKey(key string) string {
	contractName, ok := h.h.ctx.Value("contract_name").(string)
	if !ok {
		return ""
	}
	return h.modifyGlobalKey(contractName, key)
}

func (h *DBHandler) modifyGlobalKey(contractName, key string) string {
	return contractName + database.Separator + key
}

func (h *DBHandler) modifyValue(value interface{}, ramPayer...string) string {
	payer := ""
	if len(ramPayer) > 0 {
		payer = ramPayer[0]
	}
	return database.MustMarshal(value) + database.ApplicationSeparator + payer
}

func (h *DBHandler) parseValue(value string) interface{} {
	idx := strings.LastIndex(value, database.ApplicationSeparator)
	if idx == -1 {
		return value
	}
	return database.MustUnmarshal(value[0:idx])
}

func (h *DBHandler) parseValuePayer(value string) string {
	idx := strings.LastIndex(value, database.ApplicationSeparator)
	if idx == -1 {
		return ""
	}
	return value[idx + 1:]
}

func (h *DBHandler) payRAM(k, v string, who ...string) {
	oldV := h.h.db.Get(k)
	oLen := int64(len(oldV) + len(k))
	nLen := int64(len(v) + len(k))
	h.payRAMInner(oldV, oLen, nLen, who...)
}

func (h *DBHandler) payRAMForMap(k, f, v string, who ...string) {
	oldV := h.h.db.MGet(k, f)
	oLen := int64(len(oldV) + len(k) + 2 * len(f))
	nLen := int64(len(v) + len(k) + 2 * len(f))
	h.payRAMInner(oldV, oLen, nLen, who...)
}

func (h *DBHandler) payRAMInner(oldV string, oLen int64, nLen int64, who ...string) {
	var payer string
	if len(who) > 0 {
		payer = who[0]
	} else {
		payer, _ = h.h.ctx.Value("contract_name").(string)
	}

	if oldV == "n" {
		h.h.PayCost(contract.Cost{
			Data:nLen,
		}, payer)
	} else {
		oldPayer := h.parseValuePayer(oldV)
		if oldPayer == "" {
			oldPayer = h.h.ctx.Value("contract_name").(string)
		}
		if oldPayer == payer {
			h.h.PayCost(contract.Cost{
				Data: nLen - oLen,
			}, payer)
		} else {
			h.h.PayCost(contract.Cost{
				Data: -oLen,
			}, oldPayer)
			h.h.PayCost(contract.Cost{
				Data: nLen,
			}, payer)
		}
	}
}

func (h *DBHandler) releaseRAM(k string, who ...string) {
	v := h.h.db.Get(k)
	oLen := int64(len(k) + len(v))
	h.payRAMInner(v, oLen, 0, who...)
}

func (h *DBHandler) releaseRAMForMap(k, f string, who ...string) {
	v := h.h.db.MGet(k, f)
	oLen := int64(len(k) + 2 * len(f) + len(v))
	h.payRAMInner(v, oLen, 0, who...)
}
