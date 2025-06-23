package main

import (
	"bitbucket.org/watashi564/accelerator/sha256"
	"fmt"
)

func commit(hash []byte) [32]byte {
	var buf [64]byte
	var sli []byte
	sli = buf[0:0]

	var whitepaper = hex2byte32([]byte("6AFBAC595C1D07A3D4C5179758F5BCE4462A6C263F6E6DFCD942011433ADAAE7"))

	sli = append(sli, whitepaper[0:]...)
	sli = append(sli, hash[0:]...)

	return nethash(sli)
}

func merkle(a []byte, b []byte) [32]byte {
	var buf [64]byte
	var sli []byte
	sli = buf[0:0]

	sli = append(sli, a[0:]...)
	sli = append(sli, b[0:]...)

	return nethash(sli)
}

func Testnet() bool {
	return u_config.testnet
}

func Regtest() bool {
	return u_config.regtest
}

var Iv = sha256.Iv256
var IvLength uint64

func initTestnet() {
	if Testnet() {
		var whitepaper = hex2byte32([]byte("2e3841b6e75e9717ab7d2a8b57248b7f611a5473381b5e432aaf8fe88874fbfe"))
		var buf [64]byte
		copy(buf[0:32], whitepaper[:])
		copy(buf[32:64], whitepaper[:])
		Iv = sha256.Midstate256(sha256.Iv256, &buf)
		IvLength = 64
	}
}

// for future use
//func midnet() byte {
//	if Testnet() {
//		return 2
//	}
//	return 1
//}

func midhash(midstate []byte, net byte) (commitment [32]byte) {
	var tmp [32]byte
	copy(tmp[:], midstate)
	if net == 0 {
		// net=0 not a midstate, but a commitment
		return tmp
	}

	var data [64]byte
	sha256.Pad256(&data, 64*uint64(net), true)
	return sha256.Midstate256(tmp, &data)
}

func nethash(buf []byte) [32]byte {
	var actual = Iv
	for i := 0; i+64 <= len(buf); i += 64 {
		var data [64]byte
		copy(data[:], buf[i:])
		actual = sha256.Midstate256(actual, &data)
	}
	var data [64]byte
	copy(data[:], buf[(len(buf)/64)*64:])
	if sha256.Pad256(&data, IvLength+uint64(len(buf)), true) {
		actual = sha256.Midstate256(actual, &data)
		sha256.Pad256(&data, IvLength+uint64(len(buf)), false)
	}
	actual = sha256.Midstate256(actual, &data)
	return actual
}

// comb address is uppercase on main-net, and lower-case on testnet
func CombAddr(x [32]byte) string {
	if Testnet() {
		return fmt.Sprintf("%x", x)
	}
	return fmt.Sprintf("%X", x)
}