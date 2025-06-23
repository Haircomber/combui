package main

import "sync"

var segments_merkle_mutex sync.RWMutex

var segments_merkle_uncommit map[[32]byte][32]byte

var e0_to_e1 map[[32]byte][32]byte
var segmets_merkle_userinput map[[32]byte][23][32]byte

const MERKLE_INPUT_A1 = 21
const MERKLE_DATA_E0 = 22

const MERKLE_DATA_U1 = 0
const MERKLE_DATA_U2 = 1
const MERKLE_DATA_Q1 = 2
const MERKLE_DATA_Q2 = 3
const MERKLE_DATA_Z0 = 4

const MERKLE_DATA_B1 = 20

func init() {
	segments_merkle_uncommit = make(map[[32]byte][32]byte)
	e0_to_e1 = make(map[[32]byte][32]byte)
	segmets_merkle_userinput = make(map[[32]byte][23][32]byte)
}

const SEGMENT_MERKLE_TRICKLED byte = 16

func segments_merkle_trickle(loopkiller map[[32]byte]struct{}, commitment [32]byte) {

	if balance_try_increase_loop(commitment) {
		return
	}

	if _, ok2 := loopkiller[commitment]; ok2 {

		balance_create_loop(commitment)
		return
	}
	loopkiller[commitment] = struct{}{}

	//segments_merkle_mutex.RLock()
	var txidandto, ok = e0_to_e1[commitment]
	//segments_merkle_mutex.RUnlock()
	var to = txidandto

	balance_do(commitment, to, 0xffffffffffffffff)

	if !ok {
		println("trickle non existent tx")
	}

	var type3 = segments_stack_type(to)
	if type3 == SEGMENT_STACK_TRICKLED {
		segments_stack_trickle(loopkiller, to)
	}

	var type2 = segments_merkle_type(to)
	if type2 == SEGMENT_MERKLE_TRICKLED {
		segments_merkle_trickle(loopkiller, to)
	}

	var type1 = segments_transaction_type(to)
	if type1 == SEGMENT_TX_TRICKLED {
		segments_transaction_trickle(loopkiller, to)
	}

}
func segments_merkle_untrickle(loopkiller *[32]byte, commitment [32]byte, bal balance) {
	graph_dirty = true
}

func segments_merkle_type(commit [32]byte) segment_type {
	//segments_merkle_mutex.RLock()
	_, ok1 := e0_to_e1[commit]
	//segments_merkle_mutex.RUnlock()

	if ok1 {
		return SEGMENT_MERKLE_TRICKLED
	}

	return SEGMENT_UNKNOWN
}

func segments_merkle_loopdetect(norecursion, loopkiller map[[32]byte]struct{}, commitment [32]byte) bool {
	if _, ok2 := loopkiller[commitment]; ok2 {

		return true
	}
	loopkiller[commitment] = struct{}{}
	//segments_merkle_mutex.RLock()
	var txidandto, ok = e0_to_e1[commitment]
	//segments_merkle_mutex.RUnlock()
	var to = txidandto

	if !ok {
		return false
	}
	if _, ok2 := loopkiller[to]; ok2 {

		return true
	}
	var type3 = segments_stack_type(to)
	if type3 == SEGMENT_STACK_TRICKLED {
		return segments_stack_loopdetect(norecursion, loopkiller, to)
	}
	var type2 = segments_merkle_type(to)
	if type2 == SEGMENT_MERKLE_TRICKLED {
		return segments_merkle_loopdetect(norecursion, loopkiller, to)
	}
	var type1 = segments_transaction_type(to)
	if type1 == SEGMENT_TX_TRICKLED {
		return segments_transaction_loopdetect(norecursion, loopkiller, to)
	}

	return false
}

func segments_merkle_backgraph(backgraph map[[32]byte][][32]byte, norecursion map[[32]byte]struct{}, target, commitment [32]byte) {

	_, is_stack_recursion := norecursion[commitment]

	if is_stack_recursion {
		return
	}

	norecursion[commitment] = struct{}{}

	//segments_merkle_mutex.RLock()
	var txidandto, ok = e0_to_e1[commitment]
	//segments_merkle_mutex.RUnlock()
	var to = txidandto

	if !ok {
		return
	}

	add_to_backgraph(backgraph, to, commitment)

	var type3 = segments_stack_type(to)
	if type3 == SEGMENT_STACK_TRICKLED {
		segments_stack_backgraph(backgraph, norecursion, target, to)
	}
	var type2 = segments_merkle_type(to)
	if type2 == SEGMENT_MERKLE_TRICKLED {
		segments_merkle_backgraph(backgraph, norecursion, target, to)
	}
	var type1 = segments_transaction_type(to)
	if type1 == SEGMENT_TX_TRICKLED {
		segments_transaction_backgraph(backgraph, norecursion, target, to)
	}

}