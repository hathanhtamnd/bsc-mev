package mev

import (
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	allowedToFourMeme = map[common.Address]struct{}{
		common.HexToAddress("0x1de460f363AF910f51726DEf188F9004276Bf4bc"): {},
		common.HexToAddress("0x5c952063c7fc8610FFDB798152D69F0B9550762b"): {},
	}

	minValueWeiFourMeme = new(big.Int)

	seenTxsFourMeme sync.Map

	txTTLFourMeme = int64(60)
)

func init() {
	minValueWeiFourMeme.SetString("99999999999999999", 10)
}

func PassFilterFourMeme(tx *types.Transaction) bool {
	to := tx.To()
	if to == nil {
		return false
	}
	if _, ok := allowedToFourMeme[*to]; !ok {
		return false
	}
	if tx.Value().Cmp(minValueWeiFourMeme) <= 0 {
		return false
	}

	now := time.Now().Unix()
	hash := tx.Hash()

	if v, ok := seenTxsFourMeme.Load(hash); ok {
		ts := v.(int64)
		if now-ts < txTTLFourMeme {
			return false
		}
		seenTxsFourMeme.Store(hash, now)
		return true
	}

	seenTxsFourMeme.Store(hash, now)
	return true
}
