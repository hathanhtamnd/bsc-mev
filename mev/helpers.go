package mev

import (
	"encoding/hex"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

func readUint256(b []byte, offset int) int {
	return int(new(big.Int).SetBytes(b[offset : offset+32]).Int64())
}

func readAddress(b []byte, offset int) string {
	if offset+32 > len(b) {
		return ""
	}
	return "0x" + hex.EncodeToString(b[offset+12:offset+32])
}

func readUint256Big(b []byte, offset int) *big.Int {
	if offset+32 > len(b) {
		return nil
	}
	return new(big.Int).SetBytes(b[offset : offset+32])
}

func hasSwapInMulticall(data []byte) bool {
	var sel [4]byte
	copy(sel[:], data[:4])

	var offset int
	if sel == [4]byte{0xac, 0x96, 0x50, 0xd8} {
		offset = readUint256(data, 4)
	} else {
		offset = readUint256(data, 4+32)
	}

	base := 4 + offset
	n := readUint256(data, base)
	pos := base + 32

	for i := 0; i < n; i++ {
		innerOffset := readUint256(data, pos)
		pos += 32
		start := base + innerOffset

		if start+4 > len(data) {
			continue
		}

		var s [4]byte
		copy(s[:], data[start:start+4])
		if _, ok := innerSwapSelectors[s]; ok {
			return true
		}
	}
	return false
}

func hasSwapInUniversalRouter(data []byte) bool {
	offset := readUint256(data, 4+32)
	base := 4 + offset

	n := readUint256(data, base)
	pos := base + 32

	for i := 0; i < n; i++ {
		innerOffset := readUint256(data, pos)
		pos += 32
		start := base + innerOffset

		if start+4 > len(data) {
			continue
		}

		var s [4]byte
		copy(s[:], data[start:start+4])
		if _, ok := innerSwapSelectors[s]; ok {
			return true
		}
	}
	return false
}

func HasRealSwapAfterDecode(tx *types.Transaction) bool {
	data := tx.Data()
	if len(data) < 4 {
		return false
	}

	var sel [4]byte
	copy(sel[:], data[:4])

	if _, ok := innerSwapSelectors[sel]; ok {
		return true
	}
	if _, ok := multicallSelectors[sel]; ok {
		return hasSwapInMulticall(data)
	}
	if sel == universalRouterSelector {
		return hasSwapInUniversalRouter(data)
	}
	return false
}

func extractV2Router(data []byte) *SwapExtract {
	amountIn := readUint256Big(data, 4)
	pathOffset := int(readUint256Big(data, 4+64).Int64())

	base := 4 + pathOffset
	if base+32 > len(data) {
		return nil
	}

	n := int(readUint256Big(data, base).Int64())
	if n < 2 {
		return nil
	}

	tokenIn := readAddress(data, base+32)
	tokenOut := readAddress(data, base+32+(n-1)*32)

	return &SwapExtract{
		SwapType: "v2_router",
		TokenIn:  tokenIn,
		TokenOut: tokenOut,
		AmountIn: amountIn,
	}
}

func extractV3ExactInputSingle(data []byte) *SwapExtract {
	tokenIn := readAddress(data, 4)
	tokenOut := readAddress(data, 4+32)
	amountIn := readUint256Big(data, 4+128)

	return &SwapExtract{
		SwapType: "v3_router",
		TokenIn:  tokenIn,
		TokenOut: tokenOut,
		AmountIn: amountIn,
	}
}

func extractV2PoolSwap(tx *types.Transaction, data []byte) *SwapExtract {
	amount0Out := readUint256Big(data, 4)
	amount1Out := readUint256Big(data, 4+32)

	var tokenInSide string

	switch {
	case amount0Out.Sign() > 0 && amount1Out.Sign() == 0:
		tokenInSide = "token1"

	case amount1Out.Sign() > 0 && amount0Out.Sign() == 0:
		tokenInSide = "token0"

	default:
		return nil
	}

	return &SwapExtract{
		SwapType: "v2_pool",
		PoolHint: tx.To().Hex(),
		TokenIn:  tokenInSide,
		AmountIn: nil,
	}
}

func extractFromMulticall(data []byte) *SwapExtract {
	// reuse hasSwapInMulticall logic
	// khi thấy innerSel là swap:
	// → gọi extractV2Router / extractV3ExactInputSingle tương ứng
	return nil
}

func ExtractSwapInfo(tx *types.Transaction) *SwapExtract {
	data := tx.Data()
	var sel [4]byte
	copy(sel[:], data[:4])

	switch sel {

	// V2 router
	case [4]byte{0x38, 0xed, 0x17, 0x39},
		[4]byte{0x7f, 0xf3, 0x6a, 0xb5},
		[4]byte{0x18, 0xcb, 0xaf, 0xe5}:
		return extractV2Router(data)

	// V3 exactInputSingle
	case [4]byte{0x04, 0xe4, 0x5a, 0xaf}:
		return extractV3ExactInputSingle(data)

	// V2 pool
	case [4]byte{0x02, 0x2c, 0x0d, 0x9f}:
		return extractV2PoolSwap(tx, data)

	// multicall
	case [4]byte{0xac, 0x96, 0x50, 0xd8},
		[4]byte{0x5a, 0xe4, 0x01, 0xdc}:
		return extractFromMulticall(data)
	}

	return nil
}
