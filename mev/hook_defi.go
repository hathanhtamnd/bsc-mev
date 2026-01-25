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
	connMuDefi sync.RWMutex
	connDefi   net.Conn
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
			if connDefi != nil {
				_ = connDefi.Close()
			}
			connDefi = c
			connMuDefi.Unlock()

			return
		}
	}()
}

func writeToTCP(b []byte) {
	connMuDefi.RLock()
	c := connDefi
	connMuDefi.RUnlock()

	if c == nil {
		return
	}

	if _, err := c.Write(b); err != nil {
		connMuDefi.Lock()
		if connDefi == c {
			_ = connDefi.Close()
			connDefi = nil
		}
		connMuDefi.Unlock()

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
	}

	b, err := json.Marshal(&ev)
	if err != nil {
		return
	}
	writeToTCP(append(b, '\n'))
}
