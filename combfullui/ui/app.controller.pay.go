package main

import "encoding/hex"
import "encoding/binary"

func (a *AppPay) controller_init_pay(change, source string) string {
	// pad the stack with source and 0
	return a.controller_add_destination(source, change, source, "0")
}

func (a *AppPay) controller_pay(target, source, entropy string, count int64) string {
	var testnet = !isMainnetAddr(target + source)
	var key1, ok = keys[source]
	if !ok && len(entropy) > 0 {

		runner := nethash([]byte(entropy), testnet)
		for ; count > 0; count-- {
			var key [21][32]byte
			var key21 [21]string
			var tip [21][32]byte
			var pub [32]byte
			var sli []byte
			var buf [672]byte
			sli = buf[0:0]
			for i := range key {
				key[i] = commit(runner[0:32], testnet)
				runner = nethash(runner[0:], testnet)
			}
			for i := range key {
				tip[i] = key[i]
				for j := 0; j < 59213; j++ {
					tip[i] = nethash(tip[i][:], testnet)
				}
				sli = append(sli, tip[i][:]...)
			}
			pub = nethash(sli, testnet)

			var addr = CombAddr(pub, testnet)

			if addr == source {

				for i := range key21 {
					key21[i] = CombAddr(key[i], testnet)
				}

				key1 = key21
				ok = true
				break
			}
		}
	}
	if !ok {
		return ""
	}
	var txid = merkle(source + target)
	depths := CutComb(txid)

	var txmem [23]string
	txmem[0] = source
	txmem[1] = target

	Undisplay(a.destinationsblock)

	a.destinationstablebody.SetInnerHTML("")

	for i := range depths {

		txmem[2+i] = manyhash(key1[i]+Net(testnet), depths[i])
		key1[i] = bech32get(commitment(txmem[2+i]))

		AppendChild(a.destinationstablebody, "tr", row(key1[i], 330))

	}
	txmempool[txid] = txmem
	possibleSpend[source] = append(possibleSpend[source], target)

	return txid
}

func (a *AppPay) controller_add_destination(source, top, target string, amount string) string {
	var amt uint64
	for _, digit := range amount {
		if digit >= '0' && digit <= '9' {
			amt *= 10
			amt += uint64(digit - '0')
		} else {
			break
		}
	}

	// simple loops checks
	if source == target && amt != 0 {
		return top
	}
	if top == target && amt != 0 {
		return top
	}

	var testnet = !isMainnetAddr(top + target)
	var key, err = hex.DecodeString(top + target)
	if err != nil {
		return top
	}
	var stack [72]byte
	copy(stack[:], key[:])
	binary.BigEndian.PutUint64(stack[64:72], amt)
	var addr = nethash(stack[0:], testnet)
	var combaddr = CombAddr(addr, testnet)

	LiquidityStack(combaddr, top, target, amt)

	a.set_destinations_content_paginator(combaddr)

	return combaddr
}
func (a *AppPay) controller_pop_destination(combaddr string) string {
	changetarget, ok := stack[combaddr]
	if !ok {
		return combaddr
	}
	//var amt = stackAmt[combaddr]

	a.set_destinations_content_paginator(changetarget[0])

	return changetarget[0]

}
func (a *AppPay) set_destinations_content_paginator(combaddr string) {
	a.destinationstablebody.SetInnerHTML("")
	var total uint64
	for {

		changetarget, ok := stack[combaddr]
		if !ok {
			break
		}
		var amt = stackAmt[combaddr]

		total += amt
		combaddr = changetarget[0]

		AppendChild(a.destinationstablebody, "tr", row(changetarget[1], amt))

	}
	AppendChild(a.destinationstablebody, "tr", row("total", total))
}
