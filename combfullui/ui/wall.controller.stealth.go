package main

import "encoding/hex"
import "strings"

func (a *Wall) controller_used_stealth(backingkey string, off uint64) (pending, rejected, deposit uint64) {
	if !AddrsCompatible(backingkey) {
		return
	}

	var stealth_unknown = make(map[string]string)

	var key, err = hex.DecodeString(backingkey)
	if err != nil {
		return
	}
	var testnet = !isMainnetAddr(backingkey)

	var stack [72]byte
	copy(stack[0:32], key[0:32])
	stack[40] = byte(off >> 56)
	stack[39] = byte(off >> 48)
	stack[38] = byte(off >> 40)
	stack[37] = byte(off >> 32)
	stack[36] = byte(off >> 24)
	stack[35] = byte(off >> 16)
	stack[34] = byte(off >> 8)
	stack[33] = byte(off >> 0)

	for i := uint16(0); i < 256; i++ {

		stack[32] = byte(i)

		var addr = nethash(stack[0:], testnet)

		var combaddr = CombAddr(addr, testnet)

		var commitm = strings.ToLower(commitment(combaddr))

		stealth_unknown[commitm] = combaddr

	}
	controller_used_page(stealth_unknown)

	return a.ViewStealth(off)
}

func (a *Wall) controller_stealth(backingkey string, off uint64) {

	if !AddrsCompatible(backingkey) {
		return
	}

	stealthbalances = make(map[string]uint64)

	var key, err = hex.DecodeString(backingkey)
	if err != nil {
		return
	}
	var testnet = !isMainnetAddr(backingkey)

	var stack [72]byte
	copy(stack[0:32], key[0:32])
	stack[40] = byte(off >> 56)
	stack[39] = byte(off >> 48)
	stack[38] = byte(off >> 40)
	stack[37] = byte(off >> 32)
	stack[36] = byte(off >> 24)
	stack[35] = byte(off >> 16)
	stack[34] = byte(off >> 8)
	stack[33] = byte(off >> 0)

	for i := uint16(0); i < 256; i++ {

		stack[32] = byte(i)

		var addr = nethash(stack[0:], testnet)

		var combaddr = CombAddr(addr, testnet)

		stealthbalances[combaddr] = CheckBalance(combaddr)

	}

	a.ViewStealth(off)
}
