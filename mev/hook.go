package mev

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

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

	hash := tx.Hash()

	raw, err := tx.MarshalBinary()
	if err != nil {
		return
	}

	fmt.Printf(
		"[MEV] %s hash=%s rawTx=0x%s peer=%s\n",
		ts.Format(time.RFC3339Nano),
		hash.Hex(),
		hex.EncodeToString(raw),
		peerID,
	)
}
