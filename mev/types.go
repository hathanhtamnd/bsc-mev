package mev

type TxEvent struct {
	Hash   string `json:"hash"`
	From   string `json:"from"`
	To     string `json:"to"`
	Value  string `json:"value"`
	Input  string `json:"input"`
	RawTx  string `json:"rawTx"`
	Peer   string `json:"peer"`
	TsNano int64  `json:"tsNano"`
}
