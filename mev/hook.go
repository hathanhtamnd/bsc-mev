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
	conn     net.Conn
	connOnce sync.Once
	connMu   sync.Mutex
)

func initTCP() {
	connOnce.Do(func() {
		go func() {
			for {
				c, err := net.Dial("tcp", "127.0.0.1:8999")
				if err != nil {
					time.Sleep(time.Second)
					continue
				}
				connMu.Lock()
				conn = c
				connMu.Unlock()
				return
			}
		}()
	})
}

func writeJSON(b []byte) {
	connMu.Lock()
	defer connMu.Unlock()

	if conn == nil {
		return
	}

	if _, err := conn.Write(b); err != nil {
		_ = conn.Close()
		conn = nil
		connOnce = sync.Once{}
	}
}

func OnRawTxFromPeer(
	tx *types.Transaction,
	peerID string,
	ts time.Time,
) {
	if tx == nil {
		return
	}

	if !PassFilter(tx) {
		return
	}

	initTCP()

	raw, err := tx.MarshalBinary()
	if err != nil {
		return
	}

	var from string

	if chainID := tx.ChainId(); chainID != nil && chainID.Sign() > 0 {
		if sender, err := types.Sender(
			types.LatestSignerForChainID(chainID),
			tx,
		); err == nil {
			from = sender.Hex()
		}
	} else {
		if sender, err := types.Sender(
			types.HomesteadSigner{},
			tx,
		); err == nil {
			from = sender.Hex()
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
	}

	b, err := json.Marshal(&ev)
	if err != nil {
		return
	}

	b = append(b, '\n')
	writeJSON(b)
}
