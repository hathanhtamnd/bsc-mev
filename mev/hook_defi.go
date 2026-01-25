package mev

import (
	"encoding/hex"
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

var (
	connMuDefi sync.Mutex
	connDefi   net.Conn

	// STEP-2: inner swap selectors (real price-moving ops)
	innerSwapSelectors = map[[4]byte]struct{}{
		// V2 routers
		{0x38, 0xed, 0x17, 0x39}: {},
		{0x7f, 0xf3, 0x6a, 0xb5}: {},
		{0x18, 0xcb, 0xaf, 0xe5}: {},
		{0x88, 0x03, 0xdb, 0xee}: {},
		{0xfb, 0x3b, 0xdb, 0x41}: {},
		{0x4a, 0x25, 0xd9, 0x4a}: {},

		// V3 routers
		{0x04, 0xe4, 0x5a, 0xaf}: {},
		{0xb8, 0x58, 0x18, 0x3f}: {},
		{0x50, 0x23, 0xb4, 0xdf}: {},
		{0x09, 0xb8, 0x13, 0x46}: {},

		// direct pools
		{0x02, 0x2c, 0x0d, 0x9f}: {}, // V2 pair
		{0x12, 0x8a, 0xcb, 0x08}: {}, // V3 pool
	}

	multicallSelectors = map[[4]byte]struct{}{
		{0xac, 0x96, 0x50, 0xd8}: {}, // multicall(bytes[])
		{0x5a, 0xe4, 0x01, 0xdc}: {}, // multicall(uint256,bytes[])
	}

	universalRouterSelector = [4]byte{0x35, 0x93, 0x56, 0x4c}
)

func init() {
	go func() {
		for {
			c, err := net.Dial("tcp", "0.0.0.0:8998")
			if err != nil {
				time.Sleep(time.Second)
				continue
			}
			connMuDefi.Lock()
			connDefi = c
			connMuDefi.Unlock()
			return
		}
	}()
}

func writeToTCP(b []byte) {
	connMuDefi.Lock()
	defer connMuDefi.Unlock()

	if connDefi == nil {
		return
	}

	if _, err := connDefi.Write(b); err != nil {
		_ = connDefi.Close()
		connDefi = nil
	}
}

func OnRawTxFromPeer(tx *types.Transaction, peerID string, ts time.Time) {
	if !PassFilterSwapDefi(tx) {
		return
	}

	swap := ExtractSwapInfo(tx)
	if swap == nil {
		return
	}

	raw, err := tx.MarshalBinary()
	if err != nil {
		return
	}

	var from string
	if chainID := tx.ChainId(); chainID != nil && chainID.Sign() > 0 {
		if s, err := types.Sender(types.LatestSignerForChainID(chainID), tx); err == nil {
			from = s.Hex()
		}
	}

	var to string
	if tx.To() != nil {
		to = tx.To().Hex()
	}

	ev := TxEvent{
		Hash:   tx.Hash().Hex(),
		From:   from,
		To:     to,
		Value:  tx.Value().String(),
		Input:  "0x" + hex.EncodeToString(tx.Data()),
		RawTx:  "0x" + hex.EncodeToString(raw),
		Peer:   peerID,
		TsNano: ts.UnixNano(),

		Swap: swap,
	}

	b, _ := json.Marshal(&ev)
	writeToTCP(append(b, '\n'))
}
