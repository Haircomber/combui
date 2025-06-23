package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"sync"
)

var merkle_txid_to_tx_mutex sync.RWMutex
var merkle_txid_to_tx map[[32]byte][2][32]byte

func init() {
	merkle_txid_to_tx = make(map[[32]byte][2][32]byte)
}

// has to be xored to the other end where hash_seq_next increments
func hash_encode_uint16(hash [32]byte, sig uint16) [32]byte {
	if u_config.deciders_fork {
		hash[30] ^= byte(sig >> 8)
		hash[31] ^= byte(sig)
	}
	return hash
}

func merkle_mine(c [32]byte) {

	segments_merkle_mutex.RLock()
	var data, ok = segments_merkle_uncommit[c]
	segments_merkle_mutex.RUnlock()

	if !ok {
		return
	}

	var txkeys = make(map[[32]byte]struct{})

	merkleleg_mutex.RLock()

	merkle_each_otherleg(c, func(o *[32]byte, sig uint16) bool {

		var entrenched1 = merkle_scan_one_leg_activity(*o, data, 65535-sig)
		var entrenched2 = merkle_scan_one_leg_activity(data, *o, sig)

		if entrenched1 && entrenched2 {

			var tx1 = merkle(data[0:], (*o)[0:])
			var tx2 = merkle((*o)[0:], data[0:])

			tx1 = hash_encode_uint16(tx1, sig)
			tx2 = hash_encode_uint16(tx2, 65535-sig)

			txkeys[tx1] = struct{}{}
			txkeys[tx2] = struct{}{}
		}
		return true
	})

	merkleleg_mutex.RUnlock()

	for txkey := range txkeys {

		merkle_each_legs_transactions(txkey, func(tx *[32]byte) bool {

			//commits_mutex.Lock()
			reactivate_txid(false, true, *tx)
			//commits_mutex.Unlock()

			return true
		})

	}
}

func merkle_unmine(c [32]byte) {

	segments_merkle_mutex.RLock()
	var data, ok = segments_merkle_uncommit[c]
	segments_merkle_mutex.RUnlock()

	if !ok {
		return
	}

	var txkeys = make(map[[32]byte]struct{})

	merkleleg_mutex.RLock()

	merkle_each_otherleg(c, func(o *[32]byte, sig uint16) bool {

		var tx1 = merkle(data[0:], (*o)[0:])
		var tx2 = merkle((*o)[0:], data[0:])

		tx1 = hash_encode_uint16(tx1, sig)
		tx2 = hash_encode_uint16(tx2, 65535-sig)

		txkeys[tx1] = struct{}{}
		txkeys[tx2] = struct{}{}

		return true
	})

	merkleleg_mutex.RUnlock()

	for txkey := range txkeys {

		merkle_each_legs_transactions(txkey, func(tx *[32]byte) bool {

			//commits_mutex.Lock()
			reactivate_txid(true, false, *tx)
			//commits_mutex.Unlock()

			return true
		})

	}
}

func merkle_scan_one_leg_activity(data, otherdata [32]byte, signature uint16) (activity bool) {

	var rawroottag, ok2 = commitsCheckNoMaxHeight(commit(data[0:]))

	if !ok2 {
		return false
	}

	var otherroottag, ok3 = commitsCheckNoMaxHeight(commit(otherdata[0:]))

	if !ok3 {
		return false
	}

	var roottag = rawroottag

	// pick the later commited leg's utxo tag
	if utag_cmp(&rawroottag, &otherroottag) <= 0 {
		roottag = otherroottag
	}

	var hash = data

	var sig = 65536
	if u_config.deciders_fork {
		sig = int(signature)
	}

	for j := 0; j < sig; j++ {
		hash = nethash(hash[0:])

		var candidaterawtag, ok3 = commitsCheck(commit(hash[0:]), uint64(roottag.height)+1)

		if !ok3 {
			continue
		}
		var candidatetag = candidaterawtag

		if utag_cmp(&roottag, &candidatetag) >= 0 {

			return false
		}
	}

	return true
}

func notify_transaction(w http.ResponseWriter, a1, a0, u1, u2, q1, q2 [32]byte, z [16][32]byte, b1 [32]byte) (bool, [32]byte, [32]byte) {

	var e [2][32]byte

	var a1_is_zero = a1 == e[0]

	var sig int

	var hash = q1

	for i := 0; i < 65536; i++ {
		if hash == u1 {
			sig = i
			break
		}

		hash = nethash(hash[0:])
	}
	if hash != u1 {
		fmt.Fprintf(w, "error merkle solution sig hash 1 does not match")
		return false, [32]byte{}, [32]byte{}
	}
	hash = q2
	for i := 0; i < 65535-sig; i++ {
		hash = nethash(hash[0:])
	}
	if hash != u2 {
		fmt.Fprintf(w, "error merkle solution sig hash 2 does not match")
		return false, [32]byte{}, [32]byte{}
	}

	var b0 = b1

	for i := byte(0); i < 16; i++ {
		if ((sig >> i) & 1) == 1 {
			b0 = merkle(b0[0:], z[i][0:])
		} else {
			b0 = merkle(z[i][0:], b0[0:])
		}
	}

	var siggy = uint16(sig)

	e[0] = merkle(a0[0:], b0[0:])
	if a1_is_zero {
		e[1] = b1

	} else {
		e[1] = merkle(a1[0:], b1[0:])

	}
	var tx = merkle(e[0][0:], e[1][0:])

	var cq1 = commit(q1[0:])
	var cq2 = commit(q2[0:])

	merkleleg_mutex.Lock()

	merkle_store_otherleg(cq1, q2, siggy)
	merkle_store_otherleg(cq2, q1, 65535-siggy)

	merkle_store_legs_transactions(hash_encode_uint16(merkle(q1[0:], q2[0:]), siggy), tx)
	if !u_config.deciders_fork {
		merkle_store_legs_transactions(merkle(q2[0:], q1[0:]), tx)
	}

	merkleleg_mutex.Unlock()

	segments_merkle_mutex.Lock()

	segments_merkle_uncommit[cq1] = q1
	segments_merkle_uncommit[cq2] = q2

	segments_merkle_mutex.Unlock()

	merkle_txid_to_tx_mutex.Lock()
	merkle_txid_to_tx[tx] = e
	merkle_txid_to_tx_mutex.Unlock()

	commits_mutex.RLock()

	var allright1 = merkle_scan_one_leg_activity(q1, q2, siggy)
	var allright2 = merkle_scan_one_leg_activity(q2, q1, 65535-siggy)

	if allright1 && allright2 {

		reactivate_txid(false, true, tx)

	}

	commits_mutex.RUnlock()

	return true, e[0], tx
}

func reactivate_txid(oldactivity, newactivity bool, tx [32]byte) {
	merkle_txid_to_tx_mutex.RLock()
	var e, ok = merkle_txid_to_tx[tx]
	merkle_txid_to_tx_mutex.RUnlock()

	if !ok {
		return
	}

	if oldactivity != newactivity {
		if oldactivity {
			//var maybecoinbase = commit(e[0][0:])

			segments_merkle_untrickle(nil, e[0], 0xffffffffffffffff)
			//segments_coinbase_untrickle_auto(maybecoinbase, e[0])

			segments_merkle_mutex.Lock()
			delete(e0_to_e1, e[0])
			segments_merkle_mutex.Unlock()
		}
		if newactivity {
			segments_transaction_mutex.Lock()
			segments_merkle_mutex.Lock()
			if old, ok1 := e0_to_e1[e[0]]; ok1 && old != e[1] {

				fmt.Println("Panic: e0 to e1 already have live path")
				panic("")
			}

			e0_to_e1[e[0]] = e[1]
			segments_merkle_mutex.Unlock()
			segments_transaction_mutex.Unlock()

			segments_transaction_mutex.RLock()
			segments_merkle_mutex.RLock()

			var maybecoinbase = commit(e[0][0:])
			if _, ok1 := combbases[maybecoinbase]; ok1 {
				segments_coinbase_trickle_auto(maybecoinbase, e[0])
			}

			segments_merkle_trickle(make(map[[32]byte]struct{}), e[0])

			segments_merkle_mutex.RUnlock()
			segments_transaction_mutex.RUnlock()
		}
	}
}
func merkle_load_data_internal(w http.ResponseWriter, data string) {

	err1 := checkHEX704upper(data)
	if err1 != nil {
		fmt.Fprintf(w, "error decoding transaction from hex: %s", err1)
		return
	}

	var rawdata = hex2byte704([]byte(data))

	var arraydata [23][32]byte

	for i := 0; i < 22; i++ {
		copy(arraydata[i][0:], rawdata[32*i:32+32*i])
	}

	var z [16][32]byte
	for i := range z {
		z[i] = arraydata[MERKLE_DATA_Z0+i]
	}

	var buf3_a0 [96]byte

	copy(buf3_a0[0:32], arraydata[MERKLE_INPUT_A1][0:32])
	copy(buf3_a0[32:64], arraydata[MERKLE_DATA_U1][0:32])
	copy(buf3_a0[64:96], arraydata[MERKLE_DATA_U2][0:32])

	var a0 = nethash(buf3_a0[0:])

	var notified, e0, tx = notify_transaction(w, arraydata[MERKLE_INPUT_A1], a0, arraydata[MERKLE_DATA_U1],
		arraydata[MERKLE_DATA_U2], arraydata[MERKLE_DATA_Q1], arraydata[MERKLE_DATA_Q2], z, arraydata[MERKLE_DATA_B1])

	if notified {

		arraydata[MERKLE_DATA_E0] = e0

		segments_merkle_mutex.Lock()

		segmets_merkle_userinput[tx] = arraydata

		segments_merkle_mutex.Unlock()
	}
}

func merkle_load_data_callable(w http.ResponseWriter, r *http.Request, data string) {
	fmt.Fprint(w, testnetColorBody())
	defer fmt.Fprint(w, `</body></html>`)

	merkle_load_data_internal(w, data)

}

func merkle_load_data(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var data = ps.ByName("data")

	merkle_load_data_callable(w, r, data)
}
