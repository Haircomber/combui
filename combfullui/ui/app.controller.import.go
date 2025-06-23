package main

import "strings"
import "encoding/hex"
import "encoding/binary"
import "fmt"

func (a *AppImport) controller_proccess_object(history string, history_buffer []byte, object int, testnet bool) {
	switch object {
	case WALLET_HEADER_STACK_DATA: // stack
		var addr = stackhash(history, testnet)
		var amt = binary.BigEndian.Uint64(history_buffer[64:72])
		LiquidityStack(addr, history[0:64], history[64:128], amt)

		stackReverse[commitment(addr)] = addr

		// try sweep money
		var bal = CheckBalance(addr)

		fmt.Println(bal, addr, amt)

	case WALLET_HEADER_PURSE_DATA: //decider

		// TODO: not implemented

	case WALLET_HEADER_TX_RECV: //tx

		var txid = merkle(history[0:128] + Net(testnet))

		var depths = CutComb(txid)

		var key23 [23]string
		var end21 string
		for i := range key23 {
			key23[i] = history[64*i : 64*i+64]
			if i >= 2 {
				end21 += manyhash(key23[i], 59213-depths[i-2])
			}
		}
		end21 = combhash(end21, testnet)

		if end21 != history[0:64] {
			return
		}
		if _, ok := txactive[txid]; !ok {

			possibleSpend[key23[0]] = append(possibleSpend[key23[0]], key23[1])

			txmempool[txid] = key23

		}

		// try sweep money
		CheckBalance(end21)

		fmt.Println(history[0:64], end21)

	case WALLET_HEADER_MERKLE_DATA: // merkle

		var key22 [22]string
		for i := range key22 {
			key22[i] = history[64*i : 64*i+64]
		}
		var u1 = key22[0]
		var u2 = key22[1]
		var q1 = key22[2]
		var q2 = key22[3]
		var b1 = key22[20]
		var a1 = key22[21]
		var z [16]string
		copy(z[:], key22[4:20])
		var a1_is_zero = a1 == "0000000000000000000000000000000000000000000000000000000000000000"

		sig1, ok1 := hashuntil(q1, u1)
		if !ok1 {
			return
		}
		sig2, ok2 := hashuntil(q2, u2)
		if !ok2 {
			return
		}
		if uint64(sig1)+uint64(sig2) != 65535 {
			return
		}
		var a0 = deciderhash(a1 + u1 + u2 + Net(testnet))
		var b0 = hashbranch(b1+Net(testnet), z, sig1)

		var e = [6]string{merkle(a0 + b0 + Net(testnet)), b1, q1, q2, fmt.Sprint(sig1), "txid"}
		if !a1_is_zero {
			e[1] = merkle(a1 + b1 + Net(testnet))
		}

		//fmt.Println("decider:", a0, "merkleroot:", b0, "source:", e[0], "target:", e[1])

		var txid = merkle(e[0] + e[1] + Net(testnet))

		e[5] = txid

		//var cq1 = commitment(q1)
		//var cq2 = commitment(q2)

		//var mq12 = merkle(q1 + q2 + Net(testnet))
		//var mq21 = merkle(q2 + q1 + Net(testnet))

		//merkleLegOther[cq1] = append(merkleLegOther[cq1], q2)
		//merkleLegOther[cq2] = append(merkleLegOther[cq1], q1)

		//merkleLegTx[mq12] = append(merkleLegTx[mq12], txid)
		//merkleLegTx[mq21] = append(merkleLegTx[mq21], txid)

		if _, ok := merkleactive[e[0]]; !ok {

			possibleSpend[e[0]] = append(possibleSpend[e[0]], e[1])

			merklemempool[txid] = e

			merkleTx[txid] = key22

		}

		//fmt.Println(sig1, sig2, a1_is_zero, txid)

		// try sweep money
		CheckBalance(e[0])

	case WALLET_HEADER_WALLET_DATA: //key

		var key21 [21]string
		var end21 string
		for i := range key21 {
			key21[i] = history[64*i : 64*i+64]
			end21 += manyhash(key21[i]+Net(testnet), 0)
		}
		var addr = combhash(end21, testnet)

		stackReverse[commitment(addr)] = addr

		// try sweep money
		keysbalances[addr] = CheckBalance(addr)
		keys[addr] = key21
	}
}

func (a *AppImport) controller_import_object(history string, object int) {
	for i := 0; i <= 1; i++ {
		var testnet = i == 1
		var prefix = wallet_header_net(object, testnet)
		if strings.HasPrefix(history, prefix) {
			bytes, err := hex.DecodeString(history[len(prefix):])
			if err != nil {
				return
			}
			if testnet {
				a.controller_proccess_object(fmt.Sprintf("%x", bytes), bytes, object, testnet)
			} else {
				a.controller_proccess_object(fmt.Sprintf("%X", bytes), bytes, object, testnet)
			}
		}
	}
}

// controller_import, here current goes from 1 to current == total
func (a *AppImport) controller_import(history string, current, total, oldProgress *int, afterFunc func(bool)) {
	// telegram fix
	history = strings.ReplaceAll(history, " ", "\n")
	// windows newlines
	history = strings.ReplaceAll(history, "\r\n", "\n")
	// done
	var lines = strings.Split(history, "\n")
	// trim trailing newlines, to not affect the progress bar badly (there is usually one)
	for len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {

		prog := 100000 * *current / *total
		if prog >= *oldProgress {
			a.ViewFileProgress(prog)
			*oldProgress = prog
		}
		afterFunc(false)

		return
	} else {
		*total += len(lines)
	}
	var something bool
	var loop func(int)
	loop = func(i int) {
		if i == len(lines) {

			prog := 100000 * (1 + *current) / *total
			if prog >= *oldProgress {
				a.ViewFileProgress(prog)
				*oldProgress = prog
			}
			afterFunc(something)

			return
		}
		go func() {
			line := strings.TrimSpace(lines[i])
			//Alert(fmt.Sprintf("%d %d %d %d\n", i, len(lines), current, total))
			switch len(line) {
			case 156: // stack
				something = true
				a.controller_import_object(line, WALLET_HEADER_STACK_DATA)
			case 204: //decider
				something = true
				a.controller_import_object(line, WALLET_HEADER_PURSE_DATA)
			case 1481: //tx
				something = true
				a.controller_import_object(line, WALLET_HEADER_TX_RECV)
			case 1421: // merkle
				something = true
				a.controller_import_object(line, WALLET_HEADER_MERKLE_DATA)
			case 1357: //key
				something = true
				a.controller_import_object(line, WALLET_HEADER_WALLET_DATA)
			}
			*current++
			i++
			prog := 100000 * *current / *total
			if prog >= *oldProgress {
				a.ViewFileProgress(prog)
				*oldProgress = prog
			}
			loop(i)
		}()
	}

	prog := 100000 * *current / *total
	if prog >= *oldProgress {
		a.ViewFileProgress(prog)
		*oldProgress = prog
	}
	loop(0)

}
