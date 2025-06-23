package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	//"log"
)

func tx_mine(key [32]byte, tagval utxotag) {

	txleg_mutex.RLock()

	txlegs_each_leg_target(key, func(tx *[32]byte) bool {
		tx_scan_leg_activity(*tx)
		return true
	})

	txleg_mutex.RUnlock()
}

func tx_scan_leg_activity(tx [32]byte) {

	segments_transaction_mutex.RLock()

	var val = segments_transaction_data[tx]
	_, active := segments_transaction_next[val[21]]

	segments_transaction_mutex.RUnlock()

	if !active {

		var txdata = val

		var maxtag utxotag

		for i := uint(0); i < 21; i++ {
			var val = txdata[i]

			var candidaterawtag, ok3 = commitsCheckNoMaxHeight(commit(val[0:]))
			if !ok3 {
				return
			}
			if i == 0 {
				maxtag = candidaterawtag
			} else if utag_cmp(&candidaterawtag, &maxtag) >= 0 {
				maxtag = candidaterawtag
			}
		}

		//fmt.Printf("maxtag %v\n", maxtag)

		for i := uint(0); i < 21; i++ {
			var hash = txdata[i]

			for j := 0; j < 65536; j++ {
				hash = nethash(hash[0:])

				var candidatetag, ok3 = commitsCheck(commit(hash[0:]), uint64(maxtag.height)+1)

				if !ok3 {
					continue
				}

				if utag_cmp(&maxtag, &candidatetag) >= 0 {
					return
				}
			}

		}

		var actuallyfrom = txdata[21]
		//fmt.Printf("block confirms transaction %X \n", tx)

		txdoublespends_each_doublespend_target(actuallyfrom, func(txidto *[2][32]byte) bool {
			if tx == (*txidto)[0] {
				segments_transaction_mutex.Lock()
				segments_transaction_next[actuallyfrom] = *txidto
				segments_transaction_mutex.Unlock()
				return false
			}
			return true
		})

		segments_transaction_mutex.RLock()
		segments_merkle_mutex.RLock()

		var maybecoinbase = commit(actuallyfrom[0:])
		if _, ok1 := combbases[maybecoinbase]; ok1 {
			segments_coinbase_trickle_auto(maybecoinbase, actuallyfrom)
		}

		segments_transaction_trickle(make(map[[32]byte]struct{}), actuallyfrom)

		segments_merkle_mutex.RUnlock()
		segments_transaction_mutex.RUnlock()
	}
}

func tx_receive_transaction(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var txn = ps.ByName("txn")

	fmt.Fprint(w, testnetColorBody())
	defer fmt.Fprintf(w, `</body></html>`)

	var back = tx_receive_transaction_internal(w, txn)
	fmt.Fprintf(w, `<a href="/sign/pay/%s/%s">&larr; Back to payment</a><br />`, CombAddr(back[0]), CombAddr(back[1]))

}

func tx_receive_transaction_internal(w http.ResponseWriter, txn string) [2][32]byte {

	err1 := checkHEX736upper(txn)
	if err1 != nil {
		fmt.Fprintf(w, "error decoding transaction from hex: %s", err1)
		return [2][32]byte{}
	}

	transaction := hex2byte736([]byte(txn))

	var txcommitsandfrom [22][32]byte
	var txidandto [2][32]byte

	copy(txcommitsandfrom[21][0:], transaction[0:32])

	for j := 2; j < 23; j++ {

		copy(txcommitsandfrom[j-2][0:], transaction[j*32:j*32+32])
	}

	copy(txidandto[1][0:], transaction[32:64])

	if txcommitsandfrom[21] == txidandto[1] {
		fmt.Fprintf(w, "warning: transaction forms a trivial loop from to: %X", txidandto[1])
	}

	txidandto[0] = nethash(transaction[0:64])

	var teeth_lengths = CutCombWhere(txidandto[0][0:])

	for i := 0; i < 21; i++ {
		var hashchain = txcommitsandfrom[i]

		for j := uint16(0); j < teeth_lengths[i]; j++ {
			hashchain = nethash(hashchain[0:])
		}

		copy(transaction[64+i*32:i*32+32+64], hashchain[0:])
	}

	var actuallyfrom = nethash(transaction[64:])

	if actuallyfrom != txcommitsandfrom[21] {
		fmt.Fprintf(w, "error: transaction is from user: %X", actuallyfrom)
		return [2][32]byte{txcommitsandfrom[21], txidandto[1]}
	}
	commits_mutex.RLock()
	txleg_mutex.Lock()
	segments_transaction_mutex.Lock()
	segments_merkle_mutex.Lock()

	segments_transaction_uncommit[commit(actuallyfrom[0:])] = actuallyfrom

	for i := 0; i < 21; i++ {
		txlegs_store_leg(commit(txcommitsandfrom[i][0:]), txidandto[0])
	}

	if _, ok := segments_transaction_data[txidandto[0]]; !ok {
		txdoublespends_store_doublespend(actuallyfrom, txidandto)
	}

	segments_transaction_data[txidandto[0]] = txcommitsandfrom

	fmt.Fprintf(w, "<pre>%X</pre>\n", actuallyfrom)

	segments_merkle_mutex.Unlock()
	segments_transaction_mutex.Unlock()
	tx_scan_leg_activity(txidandto[0])

	txleg_mutex.Unlock()
	commits_mutex.RUnlock()
	return [2][32]byte{txcommitsandfrom[21], txidandto[1]}
}
