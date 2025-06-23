package main

func LoopDetect(addr, target string) bool {
	return loopDetectInternal(addr, target, make(map[string]struct{}))
}

func loopDetectInternal(addr, target string, visited map[string]struct{}) bool {

	if addr == target {
		return true
	}

	if _, ok := visited[addr]; ok {
		return false
	}

	visited[addr] = struct{}{}

	// take active transactions

	if dst, ok := tx[addr]; ok {
		if loopDetectInternal(dst, target, visited) {
			return true
		}
	}

	// take stacks changes

	if dst, ok := stack[addr]; ok {
		if loopDetectInternal(dst[Change], target, visited) {
			return true
		}
	}

	// take inactive transactions

	for _, dst := range possibleSpend[addr] {
		if loopDetectInternal(dst, target, visited) {
			return true
		}
	}

	return false
}
