package main

import "fmt"
import "sort"
import "strings"
import "crypto/rand"
import "encoding/binary"

func (a *AppHome) controller_repaint_str(i *AppImport, w *AppWallet, str string) bool {
	data := Parse(str)
	if data == nil {
		return false
	}
	if !data.Success {
		return false
	}
	if data.OutOfSync {
		DisplayBlock(a.outofsync)
	} else {
		Undisplay(a.outofsync)
	}
	if data.LastBtcHeight == -1 {
		a.progress.SetInnerHTML("100%")
		Width(a.progress, "100%")
	} else {
		var kpercent = 100000 * data.BlockHeight / uint64(data.LastBtcHeight)
		var progress = fmt.Sprintf("%d.%03d%%", kpercent/1000, kpercent%1000)
		a.progress.SetInnerHTML(progress)
		Width(a.progress, progress)
	}
	if data.Testnet {
		Document.Get("body").Set("style", "background-color:#AAFFAA")
		Undisplay(w.genmain)
		DisplayInline(w.gentest)
	} else {
		Document.Get("body").Set("style", "background-color:white")
		Undisplay(w.gentest)
		DisplayInline(w.genmain)
	}
	if data.ShutdownButton {
		DisplayBlock(a.off)
	} else {
		Undisplay(a.off)
	}
	if !i.decider_hardfork_ignore_admin_default {
		if data.DecidersHardfork != Checked(i.decidershardfork) {
			i.decidershardfork.Call("click")
		}
	}
	a.fingerprint.SetInnerHTML(data.CryptoFingerPrint)
	a.commitments.SetInnerHTML(fmt.Sprintf("%d", data.CommitmentsCount))
	a.p2wsh.SetInnerHTML(fmt.Sprintf("%d", data.P2WSHCount))
	a.accelerator.SetInnerHTML(data.Accelerator)
	a.existing.SetInnerHTML(fmt.Sprintf("%d.%08d", combs(data.SumExistence), nats(data.SumExistence)))
	a.remaining.SetInnerHTML(fmt.Sprintf("%d.%08d", combs(data.SumRemaining), nats(data.SumRemaining)))
	a.blockhash.SetInnerHTML(data.BlockHash)
	a.blockheight.SetInnerHTML(fmt.Sprintf("%d", data.BlockHeight))

	return data.OutOfSync
}

func (a *AppHome) controller_refresh(i *AppImport, w *AppWallet, freq int) {
	go Call(backend, "main.js", func(str string) {
		if a.controller_repaint_str(i, w, str) && freq == 1 &&
			(strings.Contains(backend, "localhost") ||
			strings.Contains(backend, "127.0.0.1")) {
			SetTimeout(func(JQuery) {
				a.controller_refresh(i, w, freq)
			}, 100)
		}
	})
}

func (a *AppCoins) controller_refresh(min, max uint64) (ret bool) {
	if min > max {
		return false
	}
	Call(backend, fmt.Sprintf("%016d.%016d", min, max)+".js", func(str string) {
		data := Parse(str)
		if data == nil {
			return
		}
		if !data.Success {
			return
		}
		const ExistsForSureBlock = 750000
		ret = max <= ExistsForSureBlock || len(data.Combbases) > 0

		if !sort.SliceIsSorted(data.Combbases, func(i, j int) bool {
			return data.Combbases[i].Height > data.Combbases[j].Height
		}) {
			sort.Slice(data.Combbases, func(i, j int) bool {
				return data.Combbases[i].Height > data.Combbases[j].Height
			})
		}

		for i := range data.Combbases {
			if data.Testnet {
				data.Combbases[i].Combbase = strings.ToLower(data.Combbases[i].Combbase)
			} else {
				data.Combbases[i].Combbase = strings.ToUpper(data.Combbases[i].Combbase)
			}
		}

		for claim, height := range combbases {
			if height >= min && height <= max {
				pos := sort.Search(len(data.Combbases), func(i int) bool { return data.Combbases[i].Height >= height })

				if pos < len(data.Combbases) && data.Combbases[pos].Height == height {
					// found, check for different
					if data.Combbases[pos].Combbase != claim {
						// claim moved to another block or went away
						TryUnGenerateCoin(tx["$"+claim])
						delete(combbases, claim)
					}
				} else {
					// not found
					// claim moved to another block or went away
					TryUnGenerateCoin(tx["$"+claim])
					delete(combbases, claim)
				}
			}
		}

		for _, cb := range data.Combbases {

			var claim = cb.Combbase

			if h, ok := combbases[claim]; ok && h != cb.Height && cb.Height >= min {
				// claim moved to another block or went away
				TryUnGenerateCoin(tx["$"+claim])
			}
			combbases[claim] = cb.Height

			// sweep stealth stacks
			if addr, ok := stackReverse[claim]; ok {
				println("checking stealth balance: ", addr, " for claim", claim)
				CheckBalance(addr)
			}

		}
	})
	return
}

func (a *AppCoins) set_combbases_content_paginator_min() {
	var minheight = ^uint64(0)
	for _, h := range combbases {
		if minheight > h {
			minheight = h
		}
	}
	a.set_combbases_content_paginator(minheight)
}

func (a *AppCoins) set_combbases_content_paginator_max() {
	var maxheight uint64
	for _, h := range combbases {
		if maxheight < h {
			maxheight = h
		}
	}
	a.set_combbases_content_paginator(maxheight)
}

func (a *AppCoins) set_combbases_content_paginator(page uint64) {
	var min, max = (page / 100) * 100, (page/100)*100 + 99

	var buf []ResponseCombbase

	var bigger, smaller = false, false

	for cb, height := range combbases {
		if height < min {
			smaller = true
			continue
		}
		if height > max {
			bigger = true
			continue
		}
		buf = append(buf, ResponseCombbase{
			Combbase: cb,
			Height:   height,
		})
	}
	sort.Slice(buf, func(i, j int) bool { return buf[i].Height > buf[j].Height })

	a.combbasestablebody.SetInnerHTML("")
	for _, cb := range buf {
		var amount = Coinbase(cb.Height)
		AppendChild(a.combbasestablebody, "tr", "<td>"+fmt.Sprintf("%d", cb.Height)+"</td><td>"+bech32get(cb.Combbase)+"</td><td>"+fmt.Sprintf("%d.%08d", combs(amount), nats(amount))+"</td>")
	}
	a.combbasestablepaginator.SetInnerHTML("")

	AppendChild(a.combbasestablepaginator, "span", PaginatorHTML+PaginatorLeftHTML+PaginatorCloseHTML)
	var cnt = 3
	var shift uint64
	if !smaller {
		shift = 2
	} else if bigger {
		shift = 1
	}
	for i := (page / 100) + shift; i > 0 && cnt > 0; i-- {
		var tag = PaginatorHTML
		if i*100 == min {
			tag = PaginatorActiveHTML
		}
		AppendChild(a.combbasestablepaginator, "span", tag+fmt.Sprintf("%d", i*100)+PaginatorCloseHTML)
		cnt--
	}
	AppendChild(a.combbasestablepaginator, "span", PaginatorHTML+PaginatorRightHTML+PaginatorCloseHTML)
}

func (a *AppChart) controller_repaint_str(str string) [][2]uint64 {
	data := Parse(str)
	if data == nil {
		return nil
	}
	if !data.Success {
		return nil
	}
	return data.Chart
}
func (a *AppChart) controller_resort_str(str string) []*AppModelResponseBid {
	data := Parse(str)
	if data == nil {
		return nil
	}
	if !data.Success {
		return nil
	}
	for _, k := range data.Tx {
		for _, comm := range k.TxOut {
			if comm == nil {
				continue
			}
			if k.Testnet {
				comm.Commitment = strings.ToLower(comm.Commitment)
			} else {
				comm.Commitment = strings.ToUpper(comm.Commitment)
			}
		}
	}
	return data.Tx
}
func (a *AppChart) controller_filter(done_action func()) {
	var commits = make(map[string]struct{})
	for _, tx := range a.tx {
		if tx == nil {
			continue
		}
		for _, commit := range tx.TxOut {
			if commit == nil {
				continue
			}
			commits[commit.Commitment] = struct{}{}
		}
	}

	var finish_action = func() {
		for i, tx := range a.tx {
			if tx == nil {
				continue
			}
			var is_missing = true
			for j, commit := range tx.TxOut {
				if commit == nil {
					continue
				}
				if _, ok := commits[commit.Commitment]; ok {
					is_missing = false
					break
				} else {
					tx.TxOut[j] = nil
				}
			}
			if is_missing {
				a.tx[i] = nil
			}
		}
	}
	const TxCheckBloomFuncs = 2
	const TxCheckBloomPrefixLen = 9

	var funcs = uint32(TxCheckBloomFuncs)
	var prefxlen = uint32(TxCheckBloomPrefixLen)

	var tweakbuf [4]byte
	rand.Read(tweakbuf[:])
	var tweak = binary.LittleEndian.Uint32(tweakbuf[:])

	var buf2 []byte
	for commit := range commits {
		buf2 = append(buf2, []byte(commit[0:2*prefxlen])...)
	}

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

			for _, txi := range data.Commitments {
				c := txi.Commit

				if data.Testnet {
					c = strings.ToLower(c)
				} else {
					c = strings.ToUpper(c)
				}

				if _, ok := commits[c]; ok {
					delete(commits, c)
				}
			}

			finish_action()
			done_action()
			return
		}
		if prefxlen < 16 {
			prefxlen++
		}
		buf2 = nil
		for commit := range commits {
			buf2 = append(buf2, []byte(commit[0:2*prefxlen])...)
		}

		Call(backend, fmt.Sprintf("%08d.%016d.%08d.9999999999999999.%s.0000000000000000.0000000000000000.js", funcs, tweak, prefxlen, buf2), recursive)
	}
	Call(backend, fmt.Sprintf("%08d.%016d.%08d.9999999999999999.%s.0000000000000000.0000000000000000.js", funcs, tweak, prefxlen, buf2), recursive)
}
func (a *AppChart) controller_refresh(source string, done_action func()) {
	go Call(backend, "chart"+source+".js", func(str string) {
		if source == "disk" {
			a.data = a.controller_repaint_str(str)
		} else if source == "mempool" {
			a.tx = append(a.tx, a.controller_resort_str(str)...)
		}

		done_action()
	})
}
