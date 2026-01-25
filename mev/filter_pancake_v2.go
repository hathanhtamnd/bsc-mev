package mev

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	allowedToPancakeV2 = map[common.Address]struct{}{
		common.HexToAddress("0x10ED43C718714eb63d5aA57B78B54704E256024E"): {},
	}

	seenTxsPancakeV2 sync.Map

	txTTLPancakeV2 = int64(60)
)

func PassFilterPancakeV2(tx *types.Transaction) bool {
	to := tx.To()
	if to == nil {
		return false
	}
	if _, ok := allowedToPancakeV2[*to]; !ok {
		return false
	}

	now := time.Now().Unix()
	hash := tx.Hash()

	if v, ok := seenTxsPancakeV2.Load(hash); ok {
		ts := v.(int64)
		if now-ts < txTTLPancakeV2 {
			return false
		}
		seenTxsPancakeV2.Store(hash, now)
		return true
	}

	seenTxsPancakeV2.Store(hash, now)
	return true
}
