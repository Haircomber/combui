package main

import "crypto/rand"
import "fmt"

func controller_keygen_one(runner *[32]byte, nopass, testnet bool) (addr string, key21 [21]string) {

	var key [21][32]byte
	var tip [21][32]byte
	var buf [672]byte
	var bug [672]byte
	var pub [32]byte
	var sli []byte
	sli = buf[0:0]
	var slj []byte
	slj = bug[0:0]
	if nopass {

		for i := range key {
			_, err := rand.Read(key[i][0:])
			if err != nil {
				return
			}
			slj = append(slj, key[i][:]...)
		}

	} else {

		for i := range key {
			key[i] = commit((*runner)[0:32], testnet)
			*runner = nethash((*runner)[0:], testnet)
		}

	}

	for i := range key {
		tip[i] = key[i]
		for j := 0; j < 59213; j++ {
			tip[i] = nethash(tip[i][:], testnet)
		}
		sli = append(sli, tip[i][:]...)
	}
	pub = nethash(sli, testnet)

	addr = CombAddr(pub, testnet)
	for i := range key21 {
		key21[i] = fmt.Sprintf("%X", key[i])
	}
	return addr, key21
}

func controller_keygen_init(entropy string, testnet bool) (runner [32]byte) {
	return nethash([]byte(entropy), testnet)
}

func (a *AppWallet) controller_keygen(count int64, entropy string, testnet bool) {

	var runner [32]byte

	if len(entropy) != 0 {

		runner = controller_keygen_init(entropy, testnet)

	}

	for ; count > 0; count-- {
		var addr, key21 = controller_keygen_one(&runner, len(entropy) == 0, testnet)
		keysbalances[addr] = CheckBalance(addr)
		if len(entropy) == 0 {
			keys[addr] = key21
		}
	}
	a.ViewKeys(Value(a.key))
}
