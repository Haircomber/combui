// This file implements Haircomb Graph Balance flow and reorg.

// The reorg algorithm is a throwback algorithm. (the money is thrown back
// into the respective combbases and then retrickled back into the economy)

// The Balances data structure keeps track of source combbases for each coin.

// When a transaction is reorganized, all balances it affected are returned
// back into the stacks/combbases they were from, recursively.

// Stacks which need to be untriggered are untriggered.

// The returned balances are then trickled (flown) back into their destinations.

// Flowing back the money then retriggers stacks which need to be triggered.

// The code was fuzzed extensively and is not known to contain a flaw.

// In other words, for every graph G doing full reflow (after reorging one tx)
// is deemed equal to reorging (removing) the one tx using the reorg method.

package main

const Change = 0
const Target = 1
const ChangeTarget = 2

type Graph struct {
	Stack    map[string][ChangeTarget]string
	StackAmt map[string]uint64
	Tx       map[string]string
}

type Balances struct {
	Throwback map[string]map[string]uint64
	Loops     map[string]map[string]uint64
	StackUsed map[string]map[string]uint64
}

func (b *Balances) Delete(addr string) {
	if _, ok := b.Loops[addr]; !ok {
		if _, ok2 := b.Throwback[addr]; ok2 {
			b.Loops[addr] = b.Throwback[addr]
		}
	} else {
		for k, v := range b.Throwback[addr] {
			b.Loops[addr][k] += v
		}
	}
	delete(b.Throwback, addr)
}
func (b *Balances) Create(addr string, bal uint64) {
	if bal == 0 {
		return
	}
	if _, ok := b.Throwback[addr]; !ok {
		var newbal = make(map[string]uint64)
		newbal[addr] = uint64(bal)
		b.Throwback[addr] = newbal
	} else {
		b.Throwback[addr][addr] += uint64(bal)
	}
}
func (b *Balances) ThrowbackStack(addr string, bases map[string]struct{}) {
	for k, v := range b.StackUsed[addr] {
		bases[k] = struct{}{}
		if b.Throwback[k] == nil {
			b.Throwback[k] = make(map[string]uint64)
		}
		b.Throwback[k][k] += v
	}
}

func (b *Balances) PostThrowback(src string, addr string, stackused map[string]uint64, bases map[string]struct{}) bool {
	for k, v := range stackused {

		var val = b.Throwback[addr][k]

		if val == v {
			delete(b.Throwback[addr], k)
			if len(b.Throwback[addr]) == 0 {
				delete(b.Throwback, addr)
			}
			delete(stackused, k)
		} else if val > v {
			b.Throwback[addr][k] -= v
			if b.Throwback[addr][k] == 0 {
				delete(b.Throwback[addr], k)
				if len(b.Throwback[addr]) == 0 {
					delete(b.Throwback, addr)
				}
			}
			delete(stackused, k)
		} else if val < v {

			val = b.Throwback[addr][addr]

			if val == v {
				delete(b.Throwback[addr], addr)
				if len(b.Throwback[addr]) == 0 {
					delete(b.Throwback, addr)
				}
				delete(stackused, k)
			} else if val > v {
				b.Throwback[addr][addr] -= v
				if b.Throwback[addr][addr] == 0 {
					delete(b.Throwback[addr], addr)
					if len(b.Throwback[addr]) == 0 {
						delete(b.Throwback, addr)
					}
				}
				delete(stackused, k)
			} else {

				val = b.Throwback[k][k]

				if val == v {
					delete(b.Throwback[k], k)
					if len(b.Throwback[k]) == 0 {
						delete(b.Throwback, k)
					}
					delete(stackused, k)
				} else if val > v {
					b.Throwback[k][k] -= v
					if b.Throwback[k][k] == 0 {
						delete(b.Throwback[k], k)
						if len(b.Throwback[k]) == 0 {
							delete(b.Throwback, k)
						}
					}
					delete(stackused, k)
				} else {

					panic("this never happens")
				}
			}

		}
	}
	if len(stackused) == 0 {
		if len(b.Throwback[addr]) == 0 {
			delete(b.Throwback, addr)
		}
		delete(b.StackUsed, src)
		return true
	}
	return false
}
func (b *Balances) Unloopback(addr string) {
	for k, v := range b.Loops[addr] {

		if b.Throwback[addr] == nil {
			b.Throwback[addr] = make(map[string]uint64)
		}
		b.Throwback[addr][k] += v
		delete(b.Loops[addr], k)
	}
}
func (b *Balances) ThrowBack(addr string, bases map[string]struct{}) {
	var same bool
	for k, v := range b.Loops[addr] {
		if k == addr {
			same = true
		} else {
			bases[k] = struct{}{}
		}
		delete(b.Loops[addr], k)

		if b.Throwback[k] == nil {
			b.Throwback[k] = make(map[string]uint64)
		}
		println(v)
		b.Throwback[k][k] += v

	}
	for k, v := range b.Throwback[addr] {
		if k == addr {
			same = true
		} else {
			bases[k] = struct{}{}
		}
		delete(b.Throwback[addr], k)

		if b.Throwback[k] == nil {
			b.Throwback[k] = make(map[string]uint64)
		}
		println(v)
		b.Throwback[k][k] += v
	}

	if !same {
		delete(b.Throwback, addr)
	} else {
		bases[addr] = struct{}{}
	}
}

func (b *Balances) Move(src string, dst string) {
	if src == dst {
		return
	}
	if _, ok := b.Throwback[dst]; !ok {
		if _, ok2 := b.Throwback[src]; ok2 {
			b.Throwback[dst] = b.Throwback[src]
		}
	} else {
		for k, v := range b.Throwback[src] {
			b.Throwback[dst][k] += v
		}
	}
	delete(b.Throwback, src)

}
func (b *Balances) MoveAmt(src string, dst string, amt uint64, mapstack map[string]uint64) {
	var tomove = amt
	for k, v := range b.Throwback[src] {
		if tomove == 0 {
			return
		}
		if b.Throwback[dst] == nil {
			b.Throwback[dst] = make(map[string]uint64)
		}
		if v > tomove {
			mapstack[k] += tomove
			b.Throwback[dst][dst] += tomove
			b.Throwback[src][k] -= tomove
			if b.Throwback[src][k] == 0 {
				delete(b.Throwback[src], k)
				if len(b.Throwback[src]) == 0 {
					delete(b.Throwback, src)
				}
			}
			return
		} else if v == tomove {
			mapstack[k] += tomove
			b.Throwback[dst][dst] += tomove
			delete(b.Throwback[src], k)
			if len(b.Throwback[src]) == 0 {
				delete(b.Throwback, src)
			}
			return
		}
		if v > 0 {
			tomove -= v
			mapstack[k] += v
			b.Throwback[dst][dst] += v
		}
		delete(b.Throwback[src], k)
		if len(b.Throwback[src]) == 0 {
			delete(b.Throwback, src)
		}
	}

}
func (b *Balances) Unsufficient(addr string, bal uint64) bool {
	var amt = bal
	for _, v := range b.Throwback[addr] {
		if v >= amt {
			return false
		}
		amt -= v
	}

	return true
}
func Trickle(addr string, g Graph, b Balances) {
	trickleRecursive(addr, make(map[string]struct{}), g, b)
}

func trickleRecursive(addr string, v map[string]struct{}, g Graph, b Balances) {
	if _, ok := v[addr]; ok {
		b.Delete(addr)
		return
	}

	v[addr] = struct{}{}

	if b.Unsufficient(addr, 1) {
		return
	}

	if dsttx, ok := g.Tx[addr]; ok {
		b.Move(addr, dsttx)
		trickleRecursive(dsttx, v, g, b)
		return
	}

	if dststack, ok := g.Stack[addr]; ok {
		if _, ok := b.StackUsed[addr]; ok {
			b.Move(addr, dststack[Change])
			trickleRecursive(dststack[Change], v, g, b)
			return
		}

		stackamount := g.StackAmt[addr]
		if b.Unsufficient(addr, stackamount) {
			return
		}
		if _, ok := b.StackUsed[addr]; !ok && stackamount != 0 {

			var mapstack = make(map[string]uint64)

			b.StackUsed[addr] = mapstack

			b.MoveAmt(addr, dststack[Target], stackamount, mapstack)
			trickleRecursive(dststack[Target], make(map[string]struct{}), g, b)

		}

		b.Move(addr, dststack[Change])
		trickleRecursive(dststack[Change], v, g, b)
		return
	}
}

func (b *Balances) Reorg(sourceTarget [2]string, g Graph) {
	source := sourceTarget[0]
	target := sourceTarget[1]

	var bases = make(map[string]struct{})
	var trickle = make(map[string]struct{})
	var explored = make(map[string]struct{})
	bases[target] = struct{}{}

	if source == target {
		b.Unloopback(target)
	} else {
		for len(bases) > 0 {
			for base := range bases {
				delete(bases, base)
				trickle[base] = struct{}{}
				reorgRecursive(base, explored, bases, g, *b)
			}
		}
	}

	for base := range trickle {
		trickleRecursive(base, make(map[string]struct{}), g, *b)
	}

	for base := range trickle {
		trickleRecursive(base, make(map[string]struct{}), g, *b)
	}
}

func reorgStack(addr string, v, bases map[string]struct{}, g Graph, b Balances) {
	if dststack, ok := g.Stack[addr]; ok {
		dststackamount := g.StackAmt[addr]
		if dststackamount == 0 {
			reorgRecursive(dststack[Change], v, bases, g, b)
			return
		}

		if stackused, ok := b.StackUsed[addr]; ok {

			b.ThrowbackStack(addr, bases)

			reorgRecursive(dststack[Change], v, bases, g, b)

			reorgRecursive(dststack[Target], v, bases, g, b)

			if b.PostThrowback(addr, dststack[Target], stackused, bases) {
				bases[addr] = struct{}{}
			}

			if b.PostThrowback(addr, dststack[Change], stackused, bases) {
				bases[addr] = struct{}{}
			}

		}
	}
}

func reorgRecursive(addr string, v, bases map[string]struct{}, g Graph, b Balances) {
	if _, ok := v[addr]; ok {
		return
	}

	v[addr] = struct{}{}
	b.ThrowBack(addr, bases)

	if dsttx, ok := g.Tx[addr]; ok {
		reorgRecursive(dsttx, v, bases, g, b)
		return
	}

	reorgStack(addr, v, bases, g, b)
}

func anonminize(addr string, set map[string]struct{}, reverse map[string]map[string]struct{}) {

	if _, ok := set[addr]; ok {
		return
	}

	set[addr] = struct{}{}

	for k := range reverse[addr] {
		anonminize(k, set, reverse)
	}
}

func Anonminize(addr string, g Graph) (ret map[string]struct{}) {
	ret = make(map[string]struct{})
	var reverse = make(map[string]map[string]struct{})
	for k1, k2 := range g.Tx {
		if reverse[k2] == nil {
			reverse[k2] = make(map[string]struct{})
		}
		reverse[k2][k1] = struct{}{}
	}
	for k1, s := range g.Stack {
		if reverse[s[Change]] == nil {
			reverse[s[Change]] = make(map[string]struct{})
		}
		reverse[s[Change]][k1] = struct{}{}
		if g.StackAmt[k1] > 0 {
			if reverse[s[Target]] == nil {
				reverse[s[Target]] = make(map[string]struct{})
			}
			reverse[s[Target]][k1] = struct{}{}
		}
	}
	anonminize(addr, ret, reverse)
	return
}