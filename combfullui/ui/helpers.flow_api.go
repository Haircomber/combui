package main

import "fmt"

const flowApiDebug = false

// balances page count base 2 logarithm
// for paginator
func AllBalancesPrefixLen() int {
	var log int
	for cnt := len(throwback); cnt > 0; cnt /= 2 {
		log++
	}
	log -= 8 // 256 per page (8 bits),
	log /= 4 // 8 bits per hex character (4)
	if log < 1 {
		// force at least one hex character for paging
		return 1
	}
	return log
}

// loops over all balances
func AllBalances(f func(string, uint64)) {
	for addr, balances := range throwback {
		var total uint64
		for _, balance := range balances {
			total += balance
		}
		f(addr, total)
	}
}

// called when user views an address balance
// triggers new coin generation once for addressess that claimed successfully
// triggers comb base un generation for addresses that arent claimed (reorg)
func CheckBalance(addr string) uint64 {
	var claim = commitment(addr)
	var height = combbases[claim]
	//if height == 0 {
	//	TryUnGenerateCoin(addr)
	//}
	var amount = Coinbase(height)
	TryGenerateCoin(addr, amount)
	amount = 0
	// throwback is the new balance map
	for _, v := range throwback[addr] {
		amount += v
	}
	return amount
}

// called when the wallet learns a claim exists on the chain
func TryGenerateCoin(addr string, amount uint64) (affected bool) {
	if flowApiDebug {
		fmt.Println("TryGenerateCoin", addr, amount)
	}
	if amount == 0 {
		// Don't generate 0coin coins to prevent garbage in the graph
		return false
	}
	var claim = commitment(addr)
	var b = Balances{
		Throwback: throwback,
		Loops:     loops,
		StackUsed: stackUsed,
	}
	var g = Graph{
		Stack:    stack,
		StackAmt: stackAmt,
		Tx:       tx,
	}
	if combbasesUncommit[claim] == addr {
		return false
	}
	combbasesUncommit[claim] = addr

	// step 1. create coinbase tx,
	// the source is prefixed with $ to not be a real tx
	tx["$"+claim] = addr

	// step 2. generate exactly amount coins
	b.Create("$"+claim, amount)

	// step 3. money enters the economy
	Trickle("$"+claim, g, b)

	return true
}

// called when the wallet learns that a claim no longer exists on the chain
func TryUnGenerateCoin(addr string) (affected bool) {
	if flowApiDebug {
		fmt.Println("TryUnGenerateCoin", addr)
	}
	if len(addr) == 0 {
		return false
	}
	var claim = commitment(addr)
	var b = Balances{
		Throwback: throwback,
		Loops:     loops,
		StackUsed: stackUsed,
	}
	var g = Graph{
		Stack:    stack,
		StackAmt: stackAmt,
		Tx:       tx,
	}
	if combbasesUncommit[claim] != addr {
		// ignored nonexistent claim
		return false
	}
	delete(combbasesUncommit, claim)

	// step 1. delete coinbase tx,
	// the source is prefixed with $ to not be a real tx
	delete(tx, "$"+claim)

	// step 2: Easy implementation, just reorg the combbase tx
	b.Reorg([2]string{"$" + claim, addr}, g)

	// step 3: delete money from combbase tx source
	// this is never a real address due to $ prefix
	b.Delete("$" + claim)

	// affected
	return true
}

// called when wallet validates the tx crypto as active tx
func TryActivateTx(source, dest string) (affected bool) {
	if flowApiDebug {
		fmt.Println("TryActivateTx", source, dest)
	}
	var b = Balances{
		Throwback: throwback,
		Loops:     loops,
		StackUsed: stackUsed,
	}
	var g = Graph{
		Stack:    stack,
		StackAmt: stackAmt,
		Tx:       tx,
	}
	if dst, ok := tx[source]; ok && dst == dest {
		// the tx is already active and the same, do nothing
		return false
	} else if ok && dst != dest {

		// step 1. disconnect tx
		delete(tx, source)

		// step 2. reorg the transaction (from soruce to dst)
		b.Reorg([2]string{source, dst}, g)

		// the following steps are shared:
	}

	// step 3. connect tx, from source to dest
	tx[source] = dest

	// step 4. money enters the economy
	Trickle(source, g, b)

	return true
}

// called when the wallet undoes a transaction during a reorg
func TryUnActivateTx(source, dest string) bool {
	if flowApiDebug {
		fmt.Println("TryUnActivateTx", source, dest)
	}
	var b = Balances{
		Throwback: throwback,
		Loops:     loops,
		StackUsed: stackUsed,
	}
	var g = Graph{
		Stack:    stack,
		StackAmt: stackAmt,
		Tx:       tx,
	}
	if dst, ok := tx[source]; ok && dst == dest {

		// delete the tx edge
		delete(tx, source)

		// the tx is really active, reorg it
		b.Reorg([2]string{source, dst}, g)

		return true
	}
	return false
}

// called when wallet loads and validates a liquidity stack
func LiquidityStack(source, change, target string, amount uint64) {
	stack[source] = [ChangeTarget]string{change, target}
	stackAmt[source] = amount
	var b = Balances{
		Throwback: throwback,
		Loops:     loops,
		StackUsed: stackUsed,
	}
	var g = Graph{
		Stack:    stack,
		StackAmt: stackAmt,
		Tx:       tx,
	}
	// money enters the economy
	Trickle(source, g, b)
}

func AnonMinize(addr string) map[string]struct{} {
	var g = Graph{
		Stack:    stack,
		StackAmt: stackAmt,
		Tx:       tx,
	}
	return Anonminize(addr, g)
}