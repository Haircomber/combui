package main

import "fmt"
import "crypto/rand"
import "encoding/binary"
import "strings"

const InternalUsedToothCheckBloomFuncs = 13
const InternalUsedToothCheckPrefixLen = 9

func (a *AppWallet) controller_used() {
	a.controller_used_keys(keys)
}
func (a *AppWallet) controller_used_memkeys_net(entropy string, testnet bool) {
	var memkeys = make(map[string][21]string)
	var runner = controller_keygen_init(entropy, testnet)
	for {

		var addr, key21 = controller_keygen_one(&runner, false, testnet)
		if _, ok := keysbalances[addr]; !ok {
			break
		} else {
			memkeys[addr] = key21
		}
	}
	a.controller_used_keys(memkeys)
}

func (a *AppWallet) controller_used_keys(keys map[string][21]string) {
	for i := 0; i < 21; i++ {
		for k, v := range keys {
			var w = v[i]
			// private key is upercase everywhere, fix it for net
			if !AddrsCompatible(k, w) {
				w = strings.ToLower(w)
			}
			if controller_used_tooth(w) {
				if _, ok := possibleSpend[strings.ToLower(k)]; !ok {
					possibleSpend[strings.ToLower(k)] = nil
				}
				a.ViewKeys(Value(a.key))
				continue
			}
		}
	}
	return
}
func controller_used_tooth(privkey string) (success bool) {
	var funcs = uint32(InternalUsedToothCheckBloomFuncs)
	var prefxlen = uint32(InternalUsedToothCheckPrefixLen)

	var tweakbuf [4]byte
	rand.Read(tweakbuf[:])

	var tweak = binary.LittleEndian.Uint32(tweakbuf[:])
	var buf2 []byte

	var inactive = make(map[string]struct{})

	var all = manyhashall(privkey, 59213)

	all = all[1:] // don't leak the private key part

	for _, txi := range all {

		buf2 = append(buf2, []byte(txi[0:2*prefxlen])...)

		inactive[txi] = struct{}{}

	}

	//var filter = BloomSerialize(buf)

	var recursive func(str string)
	recursive = func(str string) {

		data := Parse(str)
		if data == nil {
			return
		}
		if !data.Success {
			return
		}
		if data.Commitments != nil {

			for _, c := range data.Commitments {
				if _, ok := inactive[c.Commit]; ok {
					success = true
					return
				}
			}

			success = false
			return
		}

		if data.Bloom != nil {
			var buf2 []byte

			var tweakbuf [4]byte
			rand.Read(tweakbuf[:])
			var tweak_new = binary.LittleEndian.Uint32(tweakbuf[:])
			var funcs_new = funcs + 1
			var continueIterate bool
			for ctxi := range inactive {

				if LibbloomGet(funcs, tweak, ctxi, data.Bloom) {
					continueIterate = true
					buf2 = append(buf2, []byte(ctxi[0:2*prefxlen])...)
				} else {
					delete(inactive, ctxi)
				}
			}

			tweak = tweak_new
			funcs = funcs_new
			if !continueIterate {
				return
			}
			Call(backend, fmt.Sprintf("%08d.%016d.%08d.9999999999999999.%s.0000000000000000.0000000000000000.js", funcs, tweak, prefxlen, buf2), recursive)
		}

	}

	Call(backend, fmt.Sprintf("%08d.%016d.%08d.9999999999999999.%s.0000000000000000.0000000000000000.js", funcs, tweak, prefxlen, string(buf2)), recursive)
	return
}

func controller_used_page(stealth_unknown map[string]string) {
	var funcs = uint32(InternalUsedToothCheckBloomFuncs)
	var prefxlen = uint32(InternalUsedToothCheckPrefixLen)

	var tweakbuf [4]byte
	rand.Read(tweakbuf[:])

	var tweak = binary.LittleEndian.Uint32(tweakbuf[:])
	var buf2 []byte

	for txi := range stealth_unknown {

		buf2 = append(buf2, []byte(txi[0:2*prefxlen])...)

	}

	//var filter = BloomSerialize(buf)

	var recursive func(str string)
	recursive = func(str string) {

		data := Parse(str)
		if data == nil {
			return
		}
		if !data.Success {
			return
		}
		if data.Commitments != nil {

			for _, txd := range data.Commitments {
				if val, ok := stealth_unknown[txd.Commit]; ok {
					// XXX: here might be a bug
					stealthused[val] = txd.Commit
					if txd.Combbase {
						stealthclaimed[val] = uint64(txd.Height)
					}
				}
			}

			return
		}

		if data.Bloom != nil {
			var buf2 []byte

			var tweakbuf [4]byte
			rand.Read(tweakbuf[:])
			var tweak_new = binary.LittleEndian.Uint32(tweakbuf[:])
			var funcs_new = funcs + 1
			if prefxlen < 16 {
				prefxlen++
			}
			var continueIterate bool
			for ctxi := range stealth_unknown {

				if LibbloomGet(funcs, tweak, ctxi, data.Bloom) {
					continueIterate = true
					buf2 = append(buf2, []byte(ctxi[0:2*prefxlen])...)
				} else {
					delete(stealth_unknown, ctxi)
				}
			}

			tweak = tweak_new
			funcs = funcs_new
			if !continueIterate {
				return
			}
			Call(backend, fmt.Sprintf("%08d.%016d.%08d.9999999999999999.%s.0000000000000000.0000000000000000.js", funcs, tweak, prefxlen, buf2), recursive)
		}

	}

	Call(backend, fmt.Sprintf("%08d.%016d.%08d.9999999999999999.%s.0000000000000000.0000000000000000.js", funcs, tweak, prefxlen, string(buf2)), recursive)
	return
}
