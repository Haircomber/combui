package main

import (
	"crypto/sha256"
	"fmt"
)

func AddrsCompatible(strs ...string) bool {
	var net bool
	for i, addr := range strs {
		if len(addr) != 64 {
			return false
		}
		var AF bool
		var af bool
		for _, c := range addr {
			if c >= 'A' && c <= 'F' {
				AF = true
			}
			if c >= 'a' && c <= 'f' {
				af = true
			}
			if c < '0' {
				return false
			}
			if c > 'f' {
				return false
			}
			if c > '9' && c < 'A' {
				return false
			}
			if c > 'F' && c < 'a' {
				return false
			}
		}
		if AF && af {
			return false
		}
		if i == 0 {
			net = isMainnetAddr(addr)
		} else {
			if net != isMainnetAddr(addr) {
				return false
			}
		}
	}
	return true
}

func isMainnetAddr(str string) bool {
	for _, c := range str {
		if c >= 'a' {
			return false
		}
	}
	return true
}

func Net(testnet bool) string {
	if testnet {
		return "a"
	}
	return ""
}

func commit(hash []byte, testnet bool) [32]byte {
	var buf [64]byte
	var sli []byte
	sli = buf[0:0]
	var whitepaper = [32]byte{0x6A, 0xFB, 0xAC, 0x59, 0x5C, 0x1D, 0x07,
		0xA3, 0xD4, 0xC5, 0x17, 0x97, 0x58, 0xF5, 0xBC, 0xE4, 0x46,
		0x2A, 0x6C, 0x26, 0x3F, 0x6E, 0x6D, 0xFC, 0xD9, 0x42, 0x01,
		0x14, 0x33, 0xAD, 0xAA, 0xE7}

	sli = append(sli, whitepaper[0:]...)
	sli = append(sli, hash[0:]...)
	return nethash(sli, testnet)
}

func commitment(hex string) string {
	var hash [32]byte
	for i := range hash {
		hash[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}

	return CombAddr(commit(hash[:], !isMainnetAddr(hex)), !isMainnetAddr(hex))
}

func manyhash(hex string, n uint16) string {
	var hash [32]byte
	for i := range hash {
		hash[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	var testnet = !isMainnetAddr(hex)
	for j := uint16(0); j < uint16(LEVELS)-n; j++ {
		hash = nethash(hash[:], testnet)
	}
	return CombAddr(hash, testnet)
}
func hashbranch(hex string, branch [16]string, sig uint16) string {
	var testnet = !isMainnetAddr(hex)
	var b0 = hex[0:64]
	for i := uint16(0); i < 16; i++ {
		if ((sig >> i) & 1) == 1 {
			b0 = merkle(b0 + branch[i] + Net(testnet))
		} else {
			b0 = merkle(branch[i] + b0 + Net(testnet))
		}
	}
	return b0
}

func hashuntil(hex string, target string) (uint16, bool) {
	var testnet = !isMainnetAddr(hex + target)
	var hash [32]byte
	for i := range hash {
		hash[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	var u1 [32]byte
	for i := range u1 {
		u1[i] = (x2b(target[i<<1]) << 4) | x2b(target[i<<1|1])
	}
	for i := 0; i < 65536; i++ {
		if hash == u1 {
			return uint16(i), true
		}
		hash = nethash(hash[0:], testnet)
	}
	return 0, false
}

func manyhashall(hex string, n uint16) (out []string) {
	out = make([]string, n)
	var hash [32]byte
	for i := range hash {
		hash[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	var testnet = !isMainnetAddr(hex)
	for j := uint16(0); j < n; j++ {
		hash = nethash(hash[:], testnet)
		out[j] = CombAddr(commit(hash[:], testnet), testnet)
	}
	return
}

// input is 64 hex bytes, two hashes strings concatenated
func merkle(hex string) string {
	var buf [64]byte
	for i := range buf {
		buf[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}

	return CombAddr(nethash(buf[:], !isMainnetAddr(hex)), !isMainnetAddr(hex))
}

func stackhash(hex string, testnet bool) string {
	var buf [72]byte
	for i := range buf {
		buf[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return CombAddr(nethash(buf[:], !isMainnetAddr(hex)), !isMainnetAddr(hex))
}
func deciderhash(hex string) string {
	var buf [96]byte
	for i := range buf {
		buf[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return CombAddr(nethash(buf[:], !isMainnetAddr(hex)), !isMainnetAddr(hex))
}

func combhash(hex string, testnet bool) string {
	var buf [21 * 32]byte
	for i := range buf {
		buf[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return CombAddr(nethash(buf[:], !isMainnetAddr(hex)), !isMainnetAddr(hex))
}

func nethash(sli []byte, testnet bool) [32]byte {
	if testnet {

		var whitepaper = [32]byte{0x2e, 0x38, 0x41, 0xb6, 0xe7, 0x5e,
			0x97, 0x17, 0xab, 0x7d, 0x2a, 0x8b, 0x57, 0x24, 0x8b,
			0x7f, 0x61, 0x1a, 0x54, 0x73, 0x38, 0x1b, 0x5e, 0x43,
			0x2a, 0xaf, 0x8f, 0xe8, 0x88, 0x74, 0xfb, 0xfe}

		sli = append(whitepaper[:], sli...)
		sli = append(whitepaper[:], sli...)
	}
	return sha256.Sum256(sli)
}

// comb address is uppercase on main-net, and lower-case on testnet
func CombAddr(x [32]byte, testnet bool) string {
	if testnet {
		return fmt.Sprintf("%x", x)
	}
	return fmt.Sprintf("%X", x)
}
