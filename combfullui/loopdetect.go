package main

func loopdetect(norecursion, loopkiller map[[32]byte]struct{}, to [32]byte) (result bool) {
	segments_transaction_mutex.RLock()
	segments_merkle_mutex.RLock()

	var type3 = segments_stack_type(to)
	if type3 == SEGMENT_STACK_TRICKLED {
		result = segments_stack_loopdetect(norecursion, loopkiller, to)
	}
	var type2 = segments_merkle_type(to)
	if type2 == SEGMENT_MERKLE_TRICKLED {
		result = segments_merkle_loopdetect(norecursion, loopkiller, to)
	}
	var type1 = segments_transaction_type(to)
	if type1 == SEGMENT_TX_TRICKLED {
		result = segments_transaction_loopdetect(norecursion, loopkiller, to)
	}

	segments_merkle_mutex.RUnlock()
	segments_transaction_mutex.RUnlock()
	return result
}