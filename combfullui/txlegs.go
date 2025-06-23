package main

import "sync"

var txleg_mutex sync.RWMutex
var txleg_to_tx map[[32]byte][32]byte

var merkleleg_mutex sync.RWMutex
var merkle_other_leg map[[32]byte][34]byte
var merkle_legs_to_tx map[[32]byte][32]byte

func init() {
	txleg_to_tx = make(map[[32]byte][32]byte)
	merkle_other_leg = make(map[[32]byte][34]byte)
	merkle_legs_to_tx = make(map[[32]byte][32]byte)
}

func hash_seq_next(h *[32]byte) {

	for i := range *h {

		if (*h)[i] != 255 {
			(*h)[i]++
			break
		}
		(*h)[i] = 0
	}
}

func txlegs_store_leg(leg [32]byte, totx [32]byte) bool {
	var iter = leg

	for {
		hash_seq_next(&iter)

		var maybetx, ok = txleg_to_tx[iter]

		if !ok {
			txleg_to_tx[iter] = totx
			return true
		}
		if ok && maybetx == totx {
			return false
		}
	}
}

func txlegs_each_leg_target(leg [32]byte, eacher func(*[32]byte) bool) {
	var iter = leg

	for {
		hash_seq_next(&iter)
		var maybetx, ok = txleg_to_tx[iter]

		if !ok {
			return
		}

		if !eacher(&maybetx) {
			return
		}
	}
}

func txdoublespends_store_doublespend(source [32]byte, to [2][32]byte) bool {
	var iter = source

	for {
		hash_seq_next(&iter)

		var maybetx, ok = segments_transaction_doublespends[iter]

		if !ok {
			segments_transaction_doublespends[iter] = to
			return true
		}
		if ok && maybetx == to {
			return false
		}
	}
}

func txdoublespends_each_doublespend_target(source [32]byte, eacher func(*[2][32]byte) bool) {
	var iter = source

	for {
		hash_seq_next(&iter)
		var maybetx, ok = segments_transaction_doublespends[iter]

		if !ok {
			return
		}

		if !eacher(&maybetx) {
			return
		}
	}
}

func merkle_store_otherleg(source [32]byte, val [32]byte, sig uint16) bool {
	var to [34]byte
	copy(to[:], val[:])
	to[32] = byte(sig >> 8)
	to[33] = byte(sig)

	var iter = source

	for {
		hash_seq_next(&iter)

		var maybedata, ok = merkle_other_leg[iter]

		if !ok {
			merkle_other_leg[iter] = to
			return true
		}
		if ok && maybedata == to {
			return false
		}
	}
}

func merkle_each_otherleg(source [32]byte, eacher func(*[32]byte, uint16) bool) {
	var iter = source

	for {
		hash_seq_next(&iter)
		var maybedata, ok = merkle_other_leg[iter]

		if !ok {
			return
		}

		var data [32]byte
		copy(data[:], maybedata[:])
		var sig = uint16(maybedata[33]) | uint16(maybedata[32])<<8

		if !eacher(&data, sig) {
			return
		}
	}
}

func merkle_store_legs_transactions(source [32]byte, to [32]byte) bool {
	var iter = source

	for {
		hash_seq_next(&iter)

		var maybedata, ok = merkle_legs_to_tx[iter]

		if !ok {
			merkle_legs_to_tx[iter] = to
			return true
		}
		if ok && maybedata == to {
			return false
		}
	}
}

func merkle_each_legs_transactions(source [32]byte, eacher func(*[32]byte) bool) {
	var iter = source

	for {
		hash_seq_next(&iter)
		var maybedata, ok = merkle_legs_to_tx[iter]

		if !ok {
			return
		}

		if !eacher(&maybedata) {
			return
		}
	}
}
