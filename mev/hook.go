package mev

import (
	"encoding/hex"
	"fmt"
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
					fmt.Println("[MEV] dial 8999 failed, retry in 1s:", err)
					time.Sleep(1 * time.Second)
					continue
				}
				connMu.Lock()
				conn = c
				connMu.Unlock()
				fmt.Println("[MEV] connected to 127.0.0.1:8999")
				return
			}
		}()
	})
}

func writeLine(s string) {
	connMu.Lock()
	defer connMu.Unlock()
	if conn == nil {
		return
	}
	_, err := conn.Write([]byte(s))
	if err != nil {
		fmt.Println("[MEV] write failed:", err)
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

	initTCP()

	if !PassFilter(tx) {
		return
	}

	hash := tx.Hash()

	raw, err := tx.MarshalBinary()
	if err != nil {
		return
	}

	line := fmt.Sprintf(
		"%s hash=%s rawTx=0x%s peer=%s\n",
		ts.Format(time.RFC3339Nano),
		hash.Hex(),
		hex.EncodeToString(raw),
		peerID,
	)

	writeLine(line)
}
