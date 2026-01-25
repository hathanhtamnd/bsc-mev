package mev

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

var (
	seenSwapTx sync.Map
	swapTxTTL  = int64(60)

	filterSelectors = map[[4]byte]struct{}{
		// -----------------------------
		// Uniswap / Pancake V2 routers
		// -----------------------------
		{0x38, 0xed, 0x17, 0x39}: {},
		{0x7f, 0xf3, 0x6a, 0xb5}: {},
		{0x18, 0xcb, 0xaf, 0xe5}: {},
		{0x88, 0x03, 0xdb, 0xee}: {},
		{0xfb, 0x3b, 0xdb, 0x41}: {},
		{0x4a, 0x25, 0xd9, 0x4a}: {},

		// fee-on-transfer
		{0x5c, 0x11, 0xd7, 0x95}: {},
		{0xb6, 0xf9, 0xde, 0x95}: {},
		{0x79, 0x1a, 0xc9, 0x47}: {},

		// -----------------------------
		// Uniswap / Pancake V3 routers
		// -----------------------------
		{0x04, 0xe4, 0x5a, 0xaf}: {},
		{0xb8, 0x58, 0x18, 0x3f}: {},
		{0x50, 0x23, 0xb4, 0xdf}: {},
		{0x09, 0xb8, 0x13, 0x46}: {},

		// -----------------------------
		// Direct pool swap
		// -----------------------------
		{0x02, 0x2c, 0x0d, 0x9f}: {},
		{0x12, 0x8a, 0xcb, 0x08}: {},

		// -----------------------------
		// Multicall
		// -----------------------------
		{0xac, 0x96, 0x50, 0xd8}: {},
		{0x5a, 0xe4, 0x01, 0xdc}: {},

		// -----------------------------
		// Universal / Aggregator execute
		// -----------------------------
		{0x35, 0x93, 0x56, 0x4c}: {}, // execute(bytes,bytes[])
		{0x7c, 0x02, 0x52, 0x00}: {}, // execute(bytes)
		{0x09, 0xc5, 0xea, 0xbe}: {},
		{0x1c, 0xff, 0x79, 0xcd}: {},
	}
)

func PassFilterSwapDefi(tx *types.Transaction) bool {
	data := tx.Data()
	if len(data) < 4 {
		return false
	}

	var sel [4]byte
	copy(sel[:], data[:4])

	if _, ok := filterSelectors[sel]; !ok {
		return false
	}

	now := time.Now().Unix()
	h := tx.Hash()

	if v, ok := seenSwapTx.Load(h); ok {
		if now-v.(int64) < swapTxTTL {
			return false
		}
	}
	seenSwapTx.Store(h, now)
	return true
}
