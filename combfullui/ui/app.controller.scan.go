package main

import "encoding/binary"
import "crypto/rand"
import "fmt"
import "strings"

const FrontierCheckInitialSize = 1000
const FrontierCheckBloomFuncs = 13
const FrontierCheckBloomPrefixes = 251
const FrontierCheckBloomPrefixLen = 7

const InternalCheckInitialSize = 1490000
const InternalCheckBloomFuncs = 2
const InternalCheckBloomPrefixes = 138

const tx_final_confirmations = 6

func service_mempool_frontier_scan(progressFunc func(int)) {
	var progress = 0
	var total = len(txmempool)

	for txid, tx := range txmempool {

		progressFunc(50000 * progress / total)

		progress++

		func(txid string, tx [23]string) {
			var funcs = uint32(FrontierCheckBloomFuncs)
			var prefxlen = uint32(FrontierCheckBloomPrefixLen)

			var tweakbuf [4]byte
			rand.Read(tweakbuf[:])

			var tweak = binary.LittleEndian.Uint32(tweakbuf[:])
			var buf2 []byte

			var inactive = make(map[string]struct{})

			for i := 2; i < 23; i++ {

				var txi = strings.ToLower(commitment(tx[i]))

				buf2 = append(buf2, []byte(txi[0:2*FrontierCheckBloomPrefixLen])...)

				inactive[txi] = struct{}{}
			}

			//var filter = BloomSerialize(buf)

			var maxPrefix UtxoTag

			var recursive func(str string)
			recursive = func(str string) {

				data := Parse(str)
				if data == nil {
					return
				}
				if !data.Success {
					return
				}
				if data.Count == 0 {
					return
				}

				if data.Commitments != nil {
					var best UtxoTag
					var found byte

					for _, v := range data.Commitments {
						if utag_cmp(&v.UtxoTag, &maxPrefix) >= 0 {
							maxPrefix = v.UtxoTag
						}

						if _, ok1 := inactive[v.Commit]; ok1 {
							if found == 0 {
								best = v.UtxoTag
							} else {
								if utag_cmp(&v.UtxoTag, &best) >= 0 {
									best = v.UtxoTag
								}
							}
							found++
							delete(inactive, v.Commit)
						}
					}

					if found == 21 {

						var confirmed = uint64(best.Height)+tx_final_confirmations < uint64(data.Height)

						if !confirmed {
							return
						}

						var result = mempool_internal_frontier_scan(txid, best)
						if result {

							txactive[txid] = txmempool[txid]
							delete(txmempool, txid)
							delete(possibleSpend, tx[0])

							TryActivateTx(tx[0], tx[1])

						}
					}
					return
				}

				if data.Bloom != nil {
					var buf2 []byte

					var tweakbuf [4]byte
					rand.Read(tweakbuf[:])
					var tweak_new = binary.LittleEndian.Uint32(tweakbuf[:])
					var continueIterate bool

					for ctxi := range inactive {
						if LibbloomGet(funcs, tweak, ctxi, data.Bloom) {
							continueIterate = true
							buf2 = append(buf2, []byte(ctxi[0:2*FrontierCheckBloomPrefixLen])...)
						} else {
							delete(inactive, ctxi)
						}
					}
					tweak = tweak_new

					if !continueIterate {
						return
					}

					Call(backend, fmt.Sprintf("%08d.%016d.%08d.9999999999999999.%s.%016d.%016d.js", funcs, tweak, prefxlen, buf2, maxPrefix.Height, maxPrefix.Number), recursive)
				}
			}

			Call(backend, fmt.Sprintf("%08d.%016d.%08d.9999999999999999.%s.%016d.%016d.js", funcs, tweak, prefxlen, buf2, maxPrefix.Height, maxPrefix.Number), recursive)
		}(txid, tx)
	}
}

func mempool_internal_frontier_scan(txid string, utxo UtxoTag) (success bool) {

	success = true

	for l := 0; success && l < 21; l++ {

		var tx = txmempool[txid]
		var funcs = uint32(InternalCheckBloomFuncs)
		var prefxlen = uint32(FrontierCheckBloomPrefixLen)

		var tweakbuf [4]byte
		rand.Read(tweakbuf[:])

		var tweak = binary.LittleEndian.Uint32(tweakbuf[:])
		var buf2 []byte

		var inactive = make(map[string]struct{})

		var i = 2 + l

		var all = manyhashall(tx[i], 59213)

		for _, txi := range all {

			txi = strings.ToLower(txi)

			buf2 = append(buf2, []byte(txi[0:2*FrontierCheckBloomPrefixLen])...)

			inactive[txi] = struct{}{}

		}

		//var filter = BloomSerialize(buf)

		var maxPrefix UtxoTag

		var recursive func(str string)
		recursive = func(str string) {

			data := Parse(str)
			if data == nil {
				success = false
				return
			}
			if !data.Success {
				success = false
				return
			}
			if data.Commitments != nil {

				for _, v := range data.Commitments {
					if utag_cmp(&v.UtxoTag, &maxPrefix) >= 0 {
						maxPrefix = v.UtxoTag
					}
				}

				for _, current := range data.Commitments {
					if _, ok := inactive[current.Commit]; ok {
						if utag_cmp(&utxo, &current.UtxoTag) >= 0 {
							success = false
							return
						}
						delete(inactive, current.Commit)
					}
				}

				//success = true
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
						buf2 = append(buf2, []byte(ctxi[0:2*FrontierCheckBloomPrefixLen])...)
					} else {
						delete(inactive, ctxi)
					}
				}

				tweak = tweak_new
				funcs = funcs_new
				if !continueIterate {
					return
				}
				Call(backend, fmt.Sprintf("%08d.%016d.%08d.%016d.%s.%016d.%016d.js", funcs, tweak, prefxlen, utxo.Height+1, buf2, maxPrefix.Height, maxPrefix.Number), recursive)
			}

		}

		Call(backend, fmt.Sprintf("%08d.%016d.%08d.%016d.%s.%016d.%016d.js", funcs, tweak, prefxlen, utxo.Height+1, buf2, maxPrefix.Height, maxPrefix.Number), recursive)
	}

	return success
}

func service_mempool_leg_scan(decider_hardfork bool, progressFunc func(int)) {

	var progress = 0
	var total = len(merklemempool)

	for txid, tx := range merklemempool {
		progressFunc(50000 + 50000*progress/total)

		progress++

		func(txid string, tx [6]string) {
			var funcs = uint32(FrontierCheckBloomFuncs)
			var prefxlen = uint32(FrontierCheckBloomPrefixLen)

			var tweakbuf [4]byte
			rand.Read(tweakbuf[:])

			var tweak = binary.LittleEndian.Uint32(tweakbuf[:])
			var buf2 []byte

			var inactive = make(map[string]struct{})

			for i := 2; i < 4; i++ {

				var txi = strings.ToLower(commitment(tx[i]))

				buf2 = append(buf2, []byte(txi[0:2*FrontierCheckBloomPrefixLen])...)

				inactive[txi] = struct{}{}

			}

			//var filter = BloomSerialize(buf)

			var maxPrefix UtxoTag

			var recursive func(str string)
			recursive = func(str string) {

				data := Parse(str)
				if data == nil {
					return
				}
				if !data.Success {
					return
				}
				if data.Count == 0 {
					return
				}

				if data.Commitments != nil {
					var best UtxoTag
					var found byte

					for _, v := range data.Commitments {
						if utag_cmp(&v.UtxoTag, &maxPrefix) >= 0 {
							maxPrefix = v.UtxoTag
						}
						if _, ok1 := inactive[v.Commit]; ok1 {
							if found == 0 {
								best = v.UtxoTag
							} else {
								if utag_cmp(&v.UtxoTag, &best) >= 0 {
									best = v.UtxoTag
								}
							}
							found++
							delete(inactive, v.Commit)
						}
					}

					if found == 2 {
						var result bool
						if decider_hardfork {
							var sig uint16
							for _, digit := range tx[4] {
								sig *= 10
								sig += uint16(digit - '0')
							}
							result = mempool_internal_leg_scan_hardforked(txid, best, sig)
						} else {
							result = mempool_internal_leg_scan(txid, best)
						}
						if result {

							merkleactive[tx[0]] = merklemempool[txid]
							delete(merklemempool, txid)
							delete(possibleSpend, tx[0])

							TryActivateTx(tx[0], tx[1])

						}
					}
					return
				}

				if data.Bloom != nil {
					var buf2 []byte

					var tweakbuf [4]byte
					rand.Read(tweakbuf[:])
					var tweak_new = binary.LittleEndian.Uint32(tweakbuf[:])

					var continueIterate bool

					for ctxi := range inactive {
						if LibbloomGet(funcs, tweak, ctxi, data.Bloom) {
							continueIterate = true
							buf2 = append(buf2, []byte(ctxi[0:2*FrontierCheckBloomPrefixLen])...)
						} else {
							delete(inactive, ctxi)
						}
					}
					tweak = tweak_new

					if !continueIterate {
						return
					}

					Call(backend, fmt.Sprintf("%08d.%016d.%08d.9999999999999999.%s.%016d.%016d.js", funcs, tweak, prefxlen, buf2, maxPrefix.Height, maxPrefix.Number), recursive)
				}
			}

			Call(backend, fmt.Sprintf("%08d.%016d.%08d.9999999999999999.%s.%016d.%016d.js", funcs, tweak, prefxlen, buf2, maxPrefix.Height, maxPrefix.Number), recursive)
		}(txid, tx)
	}
}

func mempool_internal_leg_scan(txid string, utxo UtxoTag) (success bool) {

	success = true

	for l := 0; success && l < 2; l++ {

		var tx = merklemempool[txid]
		var funcs = uint32(InternalCheckBloomFuncs)
		var prefxlen = uint32(FrontierCheckBloomPrefixLen)

		var tweakbuf [4]byte
		rand.Read(tweakbuf[:])

		var tweak = binary.LittleEndian.Uint32(tweakbuf[:])
		var buf2 []byte

		var inactive = make(map[string]struct{})

		var i = 2 + l

		var all = manyhashall(tx[i], 65535)

		for _, txi := range all {

			txi = strings.ToLower(txi)

			buf2 = append(buf2, []byte(txi[0:2*FrontierCheckBloomPrefixLen])...)

			inactive[txi] = struct{}{}

		}

		//var filter = BloomSerialize(buf)

		var maxPrefix UtxoTag

		var recursive func(str string)
		recursive = func(str string) {

			data := Parse(str)
			if data == nil {
				success = false
				return
			}
			if !data.Success {
				success = false
				return
			}
			if data.Commitments != nil {

				for _, v := range data.Commitments {
					if utag_cmp(&v.UtxoTag, &maxPrefix) >= 0 {
						maxPrefix = v.UtxoTag
					}
				}

				for _, current := range data.Commitments {
					if _, ok := inactive[current.Commit]; ok {
						if utag_cmp(&utxo, &current.UtxoTag) >= 0 {
							success = false
							return
						}
						delete(inactive, current.Commit)
					}
				}

				//success = true
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
						buf2 = append(buf2, []byte(ctxi[0:2*FrontierCheckBloomPrefixLen])...)
					} else {
						delete(inactive, ctxi)
					}
				}

				tweak = tweak_new
				funcs = funcs_new
				if !continueIterate {
					return
				}
				Call(backend, fmt.Sprintf("%08d.%016d.%08d.%016d.%s.%016d.%016d.js", funcs, tweak, prefxlen, utxo.Height+1, buf2, maxPrefix.Height, maxPrefix.Number), recursive)
			}

		}

		Call(backend, fmt.Sprintf("%08d.%016d.%08d.%016d.%s.%016d.%016d.js", funcs, tweak, prefxlen, utxo.Height+1, buf2, maxPrefix.Height, maxPrefix.Number), recursive)
	}
	return success
}

func mempool_internal_leg_scan_hardforked(txid string, utxo UtxoTag, sig uint16) (success bool) {

	success = true

	var tx = merklemempool[txid]
	var funcs = uint32(InternalCheckBloomFuncs)
	var prefxlen = uint32(FrontierCheckBloomPrefixLen)
	var inactive = make(map[string]struct{})

	var tweakbuf [4]byte
	rand.Read(tweakbuf[:])

	var tweak = binary.LittleEndian.Uint32(tweakbuf[:])

	var buf2 []byte

	for l := 0; l < 2; l++ {

		var i = 2 + l

		var all = manyhashall(tx[i], sig)

		for _, txi := range all {

			txi = strings.ToLower(txi)

			buf2 = append(buf2, []byte(txi[0:2*FrontierCheckBloomPrefixLen])...)

			inactive[txi] = struct{}{}

		}

		sig = 65535 - sig
	}

	//var filter = BloomSerialize(buf)

	var maxPrefix UtxoTag

	var recursive func(str string)
	recursive = func(str string) {

		data := Parse(str)
		if data == nil {
			success = false
			return
		}
		if !data.Success {
			success = false
			return
		}
		if data.Commitments != nil {

			for _, v := range data.Commitments {
				if utag_cmp(&v.UtxoTag, &maxPrefix) >= 0 {
					maxPrefix = v.UtxoTag
				}
			}

			for _, current := range data.Commitments {
				if _, ok := inactive[current.Commit]; ok {
					if utag_cmp(&utxo, &current.UtxoTag) >= 0 {
						success = false
						return
					}
					delete(inactive, current.Commit)
				}
			}

			//success = true
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
					buf2 = append(buf2, []byte(ctxi[0:2*FrontierCheckBloomPrefixLen])...)
				} else {
					delete(inactive, ctxi)
				}
			}

			tweak = tweak_new
			funcs = funcs_new
			if !continueIterate {
				return
			}
			Call(backend, fmt.Sprintf("%08d.%016d.%08d.%016d.%s.%016d.%016d.js", funcs, tweak, prefxlen, utxo.Height+1, buf2, maxPrefix.Height, maxPrefix.Number), recursive)
		}

	}

	Call(backend, fmt.Sprintf("%08d.%016d.%08d.%016d.%s.%016d.%016d.js", funcs, tweak, prefxlen, utxo.Height+1, buf2, maxPrefix.Height, maxPrefix.Number), recursive)

	return success
}
