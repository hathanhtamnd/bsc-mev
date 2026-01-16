package mev

import (
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	signerMu    sync.RWMutex
	signerCache = map[string]types.Signer{}
)

func cachedSigner(chainID *big.Int) types.Signer {
	key := chainID.String()

	signerMu.RLock()
	s, ok := signerCache[key]
	signerMu.RUnlock()
	if ok {
		return s
	}

	s = types.LatestSignerForChainID(chainID)

	signerMu.Lock()
	signerCache[key] = s
	signerMu.Unlock()

	return s
}

func OnRawTxFromPeer(tx *types.Transaction, peer string, ts time.Time) {
	if !PassFilter(tx) {
		return
	}

	from, err := types.Sender(cachedSigner(tx.ChainId()), tx)
	if err != nil {
		return
	}

	raw, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return
	}

	ev := &TxEvent{
		Hash:   tx.Hash().Hex(),
		From:   from.Hex(),
		To:     tx.To().Hex(),
		Value:  tx.Value().String(),
		Input:  hexutil.Encode(tx.Data()),
		RawTx:  hexutil.Encode(raw),
		Peer:   peer,
		TsNano: ts.UnixNano(),
	}

	send(ev)
}
