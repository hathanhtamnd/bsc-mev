package mev

import "math/big"

type TxEvent struct {
	Hash   string `json:"hash"`
	From   string `json:"from"`
	To     string `json:"to"`
	Value  string `json:"value"`
	Input  string `json:"input"`
	RawTx  string `json:"rawTx"`
	Peer   string `json:"peer"`
	TsNano int64  `json:"tsNano"`

	Swap *SwapExtract `json:"swap,omitempty"`
}

type SwapExtract struct {
	SwapType string
	TokenIn  string
	TokenOut string
	AmountIn *big.Int
	PoolHint string
}
