package mev

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	allowedTo = map[common.Address]struct{}{
		common.HexToAddress("0x10ED43C718714eb63d5aA57B78B54704E256024E"): {},
		common.HexToAddress("0x13f4EA83D0bd40E75C8222255bc855a974568Dd4"): {},
		common.HexToAddress("0xEf1c6E67703c7BD7107eed8303Fbe6EC2554BF6B"): {},
	}

	seenTxs sync.Map
	txTTL   = int64(60)
)

func PassFilter(tx *types.Transaction) bool {
	to := tx.To()
	if to == nil {
		return false
	}
	if _, ok := allowedTo[*to]; !ok {
		return false
	}

	now := time.Now().Unix()
	hash := tx.Hash()

	if v, ok := seenTxs.Load(hash); ok {
		ts := v.(int64)
		if now-ts < txTTL {
			return false
		}
		seenTxs.Store(hash, now)
		return true
	}

	seenTxs.Store(hash, now)
	return true
}
