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
	connFourMeme     net.Conn
	connOnceFourMeme sync.Once
	connMuFourMeme   sync.Mutex

	connPancakeV2     net.Conn
	connOncePancakeV2 sync.Once
	connMuPancakeV2   sync.Mutex
)

func initTCPFourMeme() {
	connOnceFourMeme.Do(func() {
		go func() {
			for {
				c, err := net.Dial("tcp", "0.0.0.0:8999")
				if err != nil {
					time.Sleep(time.Second)
					continue
				}
				connMuFourMeme.Lock()
				connFourMeme = c
				connMuFourMeme.Unlock()
				return
			}
		}()
	})
}

func initTCPPancakeV2() {
	connOncePancakeV2.Do(func() {
		go func() {
			for {
				c, err := net.Dial("tcp", "0.0.0.0:8999")
				if err != nil {
					time.Sleep(time.Second)
					continue
				}
				connMuPancakeV2.Lock()
				connPancakeV2 = c
				connMuPancakeV2.Unlock()
				return
			}
		}()
	})
}

func writeJSONFourMeme(b []byte) {
	connMuFourMeme.Lock()
	defer connMuFourMeme.Unlock()

	if connFourMeme == nil {
		return
	}

	if _, err := connFourMeme.Write(b); err != nil {
		_ = connFourMeme.Close()
		connFourMeme = nil
		connOnceFourMeme = sync.Once{}
	}
}

func writeJSONPancakeV2(b []byte) {
	connMuPancakeV2.Lock()
	defer connMuPancakeV2.Unlock()

	if connPancakeV2 == nil {
		return
	}

	if _, err := connPancakeV2.Write(b); err != nil {
		_ = connPancakeV2.Close()
		connPancakeV2 = nil
		connOncePancakeV2 = sync.Once{}
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

	var isFourMeme = false
	var isPancakeV2 = false
	if PassFilterFourMeme(tx) {
		initTCPFourMeme()
		isFourMeme = true
	} else if PassFilterPancakeV2(tx) {
		initTCPPancakeV2()
		isPancakeV2 = true
	} else {
		return
	}

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
	if isFourMeme {
		writeJSONFourMeme(b)
	} else if isPancakeV2 {
		writeJSONPancakeV2(b)
	}
}
