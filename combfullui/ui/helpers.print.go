package main

import (
	"github.com/sipa/bech32/ref/go/src/bech32"
)

func x2b(hex byte) (lo byte) {
	return (hex & 15) + 9*(hex>>6)
}
func bech32get(hex string) string {
	var buf [32]int
	for i := range buf {
		buf[i] = int((x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1]))
	}
	var prefix = "bc"
	if !isMainnetAddr(hex) {
		prefix = "tb"
	}

	encoded, err := bech32.SegwitAddrEncode(prefix, 0, buf[0:])
	if err != nil {
		return ""
	}
	return encoded
}

func nats(b uint64) uint32 {
	return uint32(b % 100000000)
}

func combs(b uint64) uint32 {
	return uint32(b / 100000000)
}