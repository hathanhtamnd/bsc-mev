package mev

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	allowedTo = map[common.Address]struct{}{
		common.HexToAddress("0x1de460f363AF910f51726DEf188F9004276Bf4bc"): {},
		common.HexToAddress("0x5c952063c7fc8610FFDB798152D69F0B9550762b"): {},
	}

	// 0.05 BNB = 50,000,000,000,000,000 wei
	minValueWei = new(big.Int)
)

func init() {
	minValueWei.SetString("49999999999999999", 10)
}

func PassFilter(tx *types.Transaction) bool {
	to := tx.To()
	if to == nil {
		return false
	}
	if _, ok := allowedTo[*to]; !ok {
		return false
	}
	if tx.Value().Cmp(minValueWei) <= 0 {
		return false
	}
	return true
}
