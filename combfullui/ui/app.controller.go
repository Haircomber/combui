package main

import "strings"
import "strconv"
import "fmt"
import "time"

func (a *App) bindEvents() {
	a.menu.bindEvents(a.main)
	a.home.bindEvents(a.export, a.wallet)
	a.coins.bindEvents(a.wallet)
	a.export.bindEvents(a.wallet)
	a.wallet.bindEvents(a.pay, a.main)
	a.pay.bindEvents(a.export)
	a.chart.bindEvents()
}

func (a *AppMain) bindEvents(s *AppMain) {

	AddEventListener(a.home, "click", "target", func(this, target JQuery) {
		DisplayBlock(s.home)
		Undisplay(s.wallet)
		Undisplay(s.pay)
		Undisplay(s.imports)
		Undisplay(s.coins)
		Undisplay(s.docs)
		Undisplay(s.chart)

	})

	AddEventListener(a.wallet, "click", "target", func(this, target JQuery) {
		Undisplay(s.home)
		DisplayBlock(s.wallet)
		Undisplay(s.pay)
		Undisplay(s.imports)
		Undisplay(s.coins)
		Undisplay(s.docs)
		Undisplay(s.chart)

	})

	AddEventListener(a.pay, "click", "target", func(this, target JQuery) {
		Undisplay(s.home)

		Undisplay(s.wallet)
		DisplayBlock(s.pay)
		Undisplay(s.imports)
		Undisplay(s.coins)
		Undisplay(s.docs)
		Undisplay(s.chart)

	})

	AddEventListener(a.imports, "click", "target", func(this, target JQuery) {
		Undisplay(s.home)

		Undisplay(s.wallet)
		Undisplay(s.pay)
		DisplayBlock(s.imports)
		Undisplay(s.coins)
		Undisplay(s.docs)
		Undisplay(s.chart)

	})

	AddEventListener(a.coins, "click", "target", func(this, target JQuery) {
		Undisplay(s.home)

		Undisplay(s.wallet)
		Undisplay(s.pay)
		Undisplay(s.imports)
		DisplayBlock(s.coins)
		Undisplay(s.docs)
		Undisplay(s.chart)

	})
	AddEventListener(a.docs, "click", "target", func(this, target JQuery) {
		Undisplay(s.home)

		Undisplay(s.wallet)
		Undisplay(s.pay)
		Undisplay(s.imports)
		Undisplay(s.coins)
		DisplayBlock(s.docs)
		Undisplay(s.chart)

	})

	AddEventListener(a.chart, "click", "target", func(this, target JQuery) {
		Undisplay(s.home)

		Undisplay(s.wallet)
		Undisplay(s.pay)
		Undisplay(s.imports)
		Undisplay(s.coins)
		Undisplay(s.docs)
		DisplayBlock(s.chart)

	})
}

var epoch uint64

func (a *AppHome) bindEvents(i *AppImport, w *AppWallet) {

	// even if main.js load occurs after the wasm load, we want still redraw the mainpage
	AddEventListener(a.mainLoad, "click", "target", func(this, target JQuery) {
		a.redrawJsMainPageData(i, w)
	})

	AddEventListener(a.off, "touchstart", "target", func(this, target JQuery) {
		a.taptime = time.Now().UnixMilli()
		ShowX(a.offspin)
	})
	AddEventListener(a.off, "mousedown", "target", func(this, target JQuery) {
		ShowX(a.offspin)
	})
	AddEventListener(a.off, "touchend", "target", func(this, target JQuery) {
		if !a.click {
			now := time.Now().UnixMilli()
			if now-a.taptime > 100 {
				HideX(a.offspin)
			}
		}
	})
	AddEventListener(a.off, "mouseout", "target", func(this, target JQuery) {
		if !a.click {
			HideX(a.offspin)
		}
	})
	AddEventListener(a.off, "click", "target", func(this, target JQuery) {
		var timeout = 10
		ShowX(a.offspin)
		go Call("http://127.0.0.1:2121", "shutdown", func(_ string) {
			var call func()
			call = func() {
				go Call("http://127.0.0.1:2121", "version.json", func(s2 string) {
					data := Parse(s2)
					if data == nil || data.Success == false {
						SetTimeout(func(JQuery) {
							call()
						}, timeout)
						timeout *= 2
					} else {
						HideX(a.offspin)
						a.click = false
						Window.Call("close")
					}
				})
			}
			call()
		})
	})

	AddEventListener(a.refresh, "click", "target", func(this, target JQuery) {
		backend = Value(a.backend)
		var str = a.refreshFreq.InnerHTML()
		epoch++
		if strings.Contains(str, `e="1"`) {
			a.controller_refresh(i, w, 1)
		} else if strings.Contains(str, `e="10"`) {
			var again func(uint64)
			again = func(my_epoch uint64) {
				if epoch != my_epoch {
					return
				}
				a.controller_refresh(i, w, 10)
				my_epoch = epoch
				SetTimeout(func(JQuery) {
					again(my_epoch)
				}, 10000)
			}
			again(epoch)
		} else if strings.Contains(str, `e="60"`) {
			var again func(uint64)
			again = func(my_epoch uint64) {
				if epoch != my_epoch {
					return
				}
				a.controller_refresh(i, w, 60)
				my_epoch = epoch
				SetTimeout(func(JQuery) {
					again(my_epoch)
				}, 60000)
			}
			again(epoch)
		}
	})
}

func (a *AppCoins) bindEvents(w *AppWallet) {

	AddEventListener(a.combbasesrefresh, "touchstart", "target", func(this, target JQuery) {
		a.taptime = time.Now().UnixMilli()
		ShowX(a.combbasesrefreshspin)
	})
	AddEventListener(a.combbasesrefresh, "mousedown", "target", func(this, target JQuery) {
		ShowX(a.combbasesrefreshspin)
	})
	AddEventListener(a.combbasesrefresh, "touchend", "target", func(this, target JQuery) {
		if !a.click {
			now := time.Now().UnixMilli()
			if now-a.taptime > 100 {
				HideX(a.combbasesrefreshspin)
			}
		}
	})
	AddEventListener(a.combbasesrefresh, "mouseout", "target", func(this, target JQuery) {
		if !a.click {
			HideX(a.combbasesrefreshspin)
		}
	})
	AddEventListener(a.combbasesrefresh, "click", "target", func(this, target JQuery) {

		a.click = true

		ShowX(a.combbasesrefreshspin)

		go func() {
			defer func() {
				HideX(a.combbasesrefreshspin)
				a.click = false
			}()
			h, err := strconv.ParseInt(a.heightholder.InnerHTML(), 10, 64)
			if err != nil {
				return
			}
			var str = a.combbasesrefreshdepth.InnerHTML()
			if strings.Contains(str, "100000") {
				h -= 100000
				if h <= 0 {
					h = 0
				}
			} else if strings.Contains(str, "10000") {
				h -= 10000
				if h <= 0 {
					h = 0
				}
			} else if strings.Contains(str, "1000") {
				h -= 1000
				if h <= 0 {
					h = 0
				}
			} else {
				h = 0
			}
			for a.controller_refresh(uint64(h), uint64(h)+125000) {
				h += 125000
			}
			for s := range stack {
				CheckBalance(s)
			}
			a.set_combbases_content_paginator_max()
			w.rebalance_wallet()
		}()
	})

	AddEventListener(a.balancestablepaginator, "click", "target", func(this, target JQuery) {
		var str = target.InnerHTML()
		if str == PaginatorLeftHTML {
			var prefix = ""
			for i := AllBalancesPrefixLen(); i > 0; i-- {
				prefix += "0"
			}
			a.ViewBalances(prefix)
		} else if str == PaginatorRightHTML {
			var prefix = ""
			for i := AllBalancesPrefixLen(); i > 0; i-- {
				prefix += "f"
			}
			a.ViewBalances(prefix)
		} else if len(str) <= 16 {
			a.ViewBalances(str)
		}
	})

	AddEventListener(a.combbasestablepaginator, "click", "target", func(this, target JQuery) {
		var str = target.InnerHTML()
		if str == PaginatorLeftHTML {
			a.set_combbases_content_paginator_max()
		} else if str == PaginatorRightHTML {
			a.set_combbases_content_paginator_min()
		} else {
			n, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return
			}
			a.set_combbases_content_paginator(uint64(n))
		}
	})
}

func (a *AppImport) bindEvents(w *AppWallet) {

	AddEventListener(a.file, "change", "target", func(this, target JQuery) {
		var count = GetFileCount(target)
		var current int
		var total = 1
		var oldProgress int
		var done int
		var totalCount = count
		for i := 0; i < totalCount; i++ {
			GetFileString(target, fmt.Sprintf("%d", i), "data", func(this JQuery, str string) {
				a.controller_import(str, &current, &total, &oldProgress, func(something bool) {
					done++
					if done == count {
						for i := 10; i < 10000; i *= 10 {
							SetTimeout(func(this JQuery) {
								w.rebalance_wallet()
							}, i)
						}
					}
				})
			})
		}

	})
	AddEventListener(a.fileloadbug, "click", "target", func(this, target JQuery) {
		a.filebug.Call("click")
	})
	AddEventListener(a.filebug, "change", "target", func(this, target JQuery) {
		var count = GetFileCount(target)
		var current int
		var total = 1
		var oldProgress int
		var done int
		var totalCount = count
		for i := 0; i < totalCount; i++ {
			GetFileBase64(target, fmt.Sprintf("%d", i), "data", func(this JQuery, str string) {

				res := strings.Index(str, ";base64,")
				str = str[res+8:]

				for _, fixer := range wallet_fix_strings {
					str = strings.ReplaceAll(str, fixer[0], fixer[1])
				}

				a.controller_import(str, &current, &total, &oldProgress, func(something bool) {
					done++
					if done == count {
						for i := 10; i < 10000; i *= 10 {
							SetTimeout(func(this JQuery) {
								w.rebalance_wallet()
							}, i)
						}
					}
				})
			})
		}

	})

	AddEventListener(a.fileload, "click", "target", func(this, target JQuery) {
		a.file.Call("click")
	})
	AddEventListener(Document, "paste", "clipboardData", func(this, target JQuery) {
		if !Checked(a.importcheck) {
			return
		}
		var count = GetAsCount(target)
		var current int
		var total = 1
		var oldProgress int
		var done int
		var totalCount = count
		for i := 0; i < totalCount; i++ {
			GetAsString(target, fmt.Sprintf("%d", i), "data", func(this JQuery, str string) {
				a.controller_import(str, &current, &total, &oldProgress, func(something bool) {
					done++
					if done == count {
						if Checked(a.importcheck) {
							for i := 10; i < 10000; i *= 10 {
								SetTimeout(func(this JQuery) {
									w.rebalance_wallet()
								}, i)
							}
							a.importcheck.Call("click")
						}
					}
				})
			})
		}

	})
	AddEventListener(a.exportsubmit, "click", "target", func(this, target JQuery) {
		var name = Value(a.exportdata)

		DownloadBase64(a.controller_export(nil), name+".dat")

	})
	AddEventListener(a.export2submit, "click", "target", func(this, target JQuery) {
		var keystr = Value(a.export2data)

		if !AddrsCompatible(keystr) {
			return
		}

		var export_cover_set = AnonMinize(keystr)

		DownloadBase64(a.controller_export(export_cover_set), "public"+keystr+".dat")

	})
	AddEventListener(a.scanmempool, "touchstart", "target", func(this, target JQuery) {
		a.taptime = time.Now().UnixMilli()
		ShowX(a.scanmempoolspin)
	})
	AddEventListener(a.scanmempool, "mousedown", "target", func(this, target JQuery) {
		ShowX(a.scanmempoolspin)
	})
	AddEventListener(a.scanmempool, "mouseout", "target", func(this, target JQuery) {
		if !a.click {
			HideX(a.scanmempoolspin)
		}
	})
	AddEventListener(a.scanmempool, "touchend", "target", func(this, target JQuery) {
		if !a.click {
			now := time.Now().UnixMilli()
			if now-a.taptime > 100 {
				HideX(a.scanmempoolspin)
			}
		}
	})
	AddEventListener(a.scanmempool, "click", "target", func(this, target JQuery) {

		ShowX(a.scanmempoolspin)
		a.click = true

		progressFunc := func(percent int) {
			a.ViewProgress(percent)
		}

		var is_decider_hardfork = Checked(a.decidershardfork)
		if !a.decider_hardfork_ignore_admin_default && len(merklemempool) > 0 {
			a.decider_hardfork_ignore_admin_default = true
			Undisplay(a.decidershardforkhider)
			if is_decider_hardfork {
				DisplayBlock(a.decidershardforkonsorry)
			} else {
				DisplayBlock(a.decidershardforkoffsorry)
			}
			a.decider_hardfork_ignore_admin_default = true
		}

		go func() {
			service_mempool_frontier_scan(progressFunc)
			service_mempool_leg_scan(is_decider_hardfork, progressFunc)

			for k := range keysbalances {
				keysbalances[k] = CheckBalance(k)
			}

			progressFunc(100000)

			w.rebalance_wallet()

			HideX(a.scanmempoolspin)

			a.click = false
		}()
	})
}

func (a *AppWallet) bindEvents(p *AppPay, s *AppMain) {

	AddEventListener(a.keystable, "click", "target", func(this, target JQuery) {
		var str = target.InnerHTML()
		if len(str) != 64 {
			if len(str) == 44+64+4 {
				// remove a href
				str = str[44 : 44+64]
			} else {
				return
			}
		}
		if str == Value(a.key) {
			return
		}

		var outerstr = target.OuterHTML()

		if !strings.Contains(outerstr, `"unspent"`) {
			return
		}
		if strings.Contains(outerstr, `"spent"`) {
			return
		}
		_, is_tx := tx[str]

		var commitstr = commitment(str)
		var str2 = bech32get(commitstr)
		var cstr = strings.ToLower(commitstr)

		_, is_used2 := combbasesUncommit[cstr]

		SetSrc(a.image, QRCODE(str))
		SetSrc(a.image2, QRCODE("bitcoin:"+str2+"?amount=0.00000330"))
		SetValue(a.key, str)

		if is_used2 {
			HideX(a.image2)
			SetValue(a.claimkey, "claimed")
		} else {
			ShowX(a.image2)
			SetValue(a.claimkey, str2)
		}
		DisplayBlock(a.seewallet)
		Undisplay(a.seestealth)

		var chg = Value(a.keychange)

		if str == chg {
			Undisplay(a.spend)
			Undisplay(a.change)
		} else if len(chg) == 0 {
			Undisplay(a.spend)
			if is_tx {
				Undisplay(a.change)
			} else {
				DisplayInline(a.change)
			}
		} else {
			DisplayInline(a.spend)
			if is_tx {
				Undisplay(a.change)
			} else {
				DisplayInline(a.change)
			}
		}

		// At this point, the key has been selected, but the stealth table should collapse if visible
		// Hide it
		Undisplay(a.stealthsclaimings)
		DisplayBlock(a.stealthhalf) // Ensure the stealth column is visible.

		// repaint keys table

		a.ViewKeys(str)
	})

	AddEventListener(a.stealth256btn, "click", "target", func(this, target JQuery) {
		var str = Value(a.key)
		SetValue(a.stealthbase, str)
		DisplayBlock(a.stealthsclaimings)

		// At this point, the key has been selected, so we can now display the stealth table.
		// Let's use the default first page for the stealth table.
		a.controller_stealth(str, 0, false) // This line calls the function to display the stealth address table.

		SetValue(a.stealthkey, "")

		SetValue(a.stealthspaginatorpage, "1") // Ensure he is on page 1

	})

	AddEventListener(a.claim256btn, "click", "target", func(this, target JQuery) {
		DisplayBlock(a.walletclaimingvisible)

		var str = Value(a.key)
		SetValue(a.stealthbase, "")
		DisplayBlock(a.stealthsclaimings)

		// At this point, the key has been selected, so we can now display the stealth table.
		// Let's use the default first page for the stealth table.
		a.controller_stealth(str, 0, true) // This line calls the function to display the stealth address table.

		SetValue(a.stealthkey, "")

		SetValue(a.stealthspaginatorpage, "1") // Ensure he is on page 1

	})

	AddEventListener(a.stealthtable, "click", "target", func(this, target JQuery) {
		var str = target.InnerHTML()
		if len(str) != 64 && len(str) != 62 {
			if len(str) == 44+64+4 {
				// remove a href
				str = str[44 : 44+64]
			} else if len(str) == 44+62+4 {
				// remove a href
				str = str[44 : 44+62]
			} else {
				return
			}
		}
		if str == Value(a.stealthkey) {
			return
		}

		var outerstr = target.OuterHTML()

		if !strings.Contains(outerstr, `"unspent"`) {
			return
		}
		if strings.Contains(outerstr, `"spent"`) {
			return
		}

		var str2 = str
		var is_used2 bool
		if len(str2) != 62 {
			var commitstr = commitment(str)
			str2 = bech32get(commitstr)
			var cstr = strings.ToLower(commitstr)
			_, is_used2 = combbasesUncommit[cstr]
		} else {
			str = ""
		}

		SetSrc(a.stealthimage, QRCODE(str))
		SetSrc(a.stealthimage2, QRCODE("bitcoin:"+str2+"?amount=0.00000330"))
		SetValue(a.stealthkey, str)
		if is_used2 {
			HideX(a.stealthimage2)
			SetValue(a.claimstealth, "claimed")
		} else {
			ShowX(a.stealthimage2)
			SetValue(a.claimstealth, str2)
		}
		DisplayBlock(a.seewallet)
		DisplayBlock(a.seestealth)

		// repaint stealth table

		var str_paginator = Value(a.stealthspaginatorpage)
		if str_paginator == "" {
			str_paginator = "1"
		}
		n, err := strconv.ParseInt(str_paginator, 10, 64)
		if err != nil {
			return
		}
		if n < 1 {
			return
		}
		if len(str) == 64 {
			a.ViewStealth(str, uint64(n), len(Value(a.stealthbase)) != 64)
		} else {
			a.ViewStealth(str2, uint64(n), len(Value(a.stealthbase)) != 64)
		}
	})
	AddEventListener(a.genmain, "touchstart", "target", func(this, target JQuery) {
		a.taptime = time.Now().UnixMilli()
		ShowX(a.genmainspin)
	})
	AddEventListener(a.genmain, "mousedown", "target", func(this, target JQuery) {
		ShowX(a.genmainspin)
	})
	AddEventListener(a.genmain, "mouseout", "target", func(this, target JQuery) {
		if !a.click {
			HideX(a.genmainspin)
		}
	})
	AddEventListener(a.genmain, "touchend", "target", func(this, target JQuery) {
		if !a.click {
			now := time.Now().UnixMilli()
			if now-a.taptime > 100 {
				HideX(a.genmainspin)
			}
		}
	})
	AddEventListener(a.genmain, "click", "target", func(this, target JQuery) {
		n, err := strconv.ParseInt(Value(a.keycount), 10, 64)
		if err != nil {
			return
		}
		a.click = true
		ShowX(a.genmainspin)

		time.Sleep(1 * time.Millisecond)

		a.controller_keygen(n, Value(a.password), false)
		a.form.Call("submit")

		HideX(a.genmainspin)
		a.click = false

		Undisplay(a.nochangevisible)
		DisplayInline(a.hint)
	})
	AddEventListener(a.gentest, "touchstart", "target", func(this, target JQuery) {
		a.taptime = time.Now().UnixMilli()
		ShowX(a.gentestspin)
	})
	AddEventListener(a.gentest, "mousedown", "target", func(this, target JQuery) {
		ShowX(a.gentestspin)
	})
	AddEventListener(a.gentest, "mouseout", "target", func(this, target JQuery) {
		if !a.click {
			HideX(a.gentestspin)
		}
	})
	AddEventListener(a.gentest, "touchend", "target", func(this, target JQuery) {
		if !a.click {
			now := time.Now().UnixMilli()
			if now-a.taptime > 100 {
				HideX(a.gentestspin)
			}
		}
	})
	AddEventListener(a.gentest, "click", "target", func(this, target JQuery) {
		n, err := strconv.ParseInt(Value(a.keycount), 10, 64)
		if err != nil {
			return
		}
		a.click = true
		ShowX(a.gentestspin)

		time.Sleep(1 * time.Millisecond)

		a.controller_keygen(n, Value(a.password), true)
		a.form.Call("submit")

		HideX(a.gentestspin)
		a.click = false

		Undisplay(a.nochangevisible)
		DisplayInline(a.hint)
	})
	AddEventListener(a.usedcheck, "touchstart", "target", func(this, target JQuery) {
		a.taptime = time.Now().UnixMilli()
		ShowX(a.usedcheckspin)
	})
	AddEventListener(a.usedcheck, "mousedown", "target", func(this, target JQuery) {
		ShowX(a.usedcheckspin)
	})
	AddEventListener(a.usedcheck, "mouseout", "target", func(this, target JQuery) {
		if !a.click {
			HideX(a.usedcheckspin)
		}
	})
	AddEventListener(a.usedcheck, "touchend", "target", func(this, target JQuery) {
		if !a.click {
			now := time.Now().UnixMilli()
			if now-a.taptime > 100 {
				HideX(a.usedcheckspin)
			}
		}
	})
	AddEventListener(a.usedcheck, "click", "target", func(this, target JQuery) {
		a.click = true
		ShowX(a.usedcheckspin)

		go func() {
			a.controller_used()

			if Value(a.password) != "" {
				a.controller_used_memkeys_net(Value(a.password), false)
				a.controller_used_memkeys_net(Value(a.password), true)
			}
			a.ViewKeys(Value(a.key))
			HideX(a.usedcheckspin)
			a.click = false
		}()
	})

	AddEventListener(a.usedstealthcheck, "touchstart", "target", func(this, target JQuery) {
		a.taptime = time.Now().UnixMilli()
		ShowX(a.usedstealthcheckspin)
	})
	AddEventListener(a.usedstealthcheck, "mousedown", "target", func(this, target JQuery) {
		ShowX(a.usedstealthcheckspin)
	})
	AddEventListener(a.usedstealthcheck, "mouseout", "target", func(this, target JQuery) {
		if !a.click {
			HideX(a.usedstealthcheckspin)
		}
	})
	AddEventListener(a.usedstealthcheck, "touchend", "target", func(this, target JQuery) {
		if !a.click {
			now := time.Now().UnixMilli()
			if now-a.taptime > 100 {
				HideX(a.usedstealthcheckspin)
			}
		}
	})
	AddEventListener(a.usedstealthcheck, "click", "target", func(this, target JQuery) {

		var str = Value(a.stealthspaginatorpage)
		if str == "" {
			str = "1"
		}
		n, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return
		}
		if n < 1 {
			return
		}

		a.click = true
		ShowX(a.usedstealthcheckspin)

		go func() {

			a.controller_used_stealth(Value(a.stealthbase), uint64(n-1), len(Value(a.stealthbase)) != 64)

			HideX(a.usedstealthcheckspin)
			a.click = false
		}()
	})

	AddEventListener(a.form, "submit", "target", func(this, target JQuery) {
		// fun trick to show the password saving dialog
		Undisplay(a.form)
		ReplaceState(History, map[string]interface{}{
			"success": true,
		}, "title", "../ui/index.html")
		DisplayBlock(a.form)
	})

	AddEventListener(a.stealthsweep, "click", "target", func(this, target JQuery) {
		if Value(a.stealthbase) == "" {
			return
		}

		var str = Value(a.stealthspaginatorpage)
		if str == "" {
			str = "1"
		}
		n, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return
		}
		if n < 1 {
			return
		}
		var key = Value(a.stealthbase)
		if key == "" {
			key = Value(a.key)
		}
		a.controller_stealth_sweep(key, uint64(n-1))
		SetValue(a.stealthspaginatorpage, fmt.Sprintf("%d", n))
	})
	AddEventListener(a.spend, "click", "target", func(this, target JQuery) {

		var change = Value(a.keychange)
		var source = Value(a.key)

		if source == change {
			change = ""
		}

		if len(change) == 0 {
			// try to find different change address
			var unspent string
			for k := range keysbalances {
				if k == source {
					continue
				}
				if _, ok := possibleSpend[strings.ToLower(k)]; !ok {
					if _, ok := tx[k]; !ok {
						unspent = k
						break
					}
				}
			}
			if len(unspent) == 0 {
				// warn user
				DisplayBlock(a.nochangevisible)
				return
			} else {
				if !AddrsCompatible(unspent) {
					return
				}

				SetValue(a.keychange, unspent)

				Undisplay(a.spend)
				Undisplay(a.change)

				change = unspent
			}
		}

		if !AddrsCompatible(change, source) {
			return
		}
		if LoopDetect(change, source) {
			return
		}
		if LoopDetect(source, change) {
			return
		}

		SetValue(p.keychange, change)
		SetValue(p.keysource, source)
		SetValue(p.stacktop, "")

		var stacktop = p.controller_init_pay(change, source)

		SetValue(p.stacktop, stacktop)

		Undisplay(s.wallet)
		DisplayBlock(s.pay)

		SetValue(a.keychange, "")
	})
	AddEventListener(a.change, "click", "target", func(this, target JQuery) {

		var keystr = Value(a.key)
		if !AddrsCompatible(keystr) {
			return
		}

		SetValue(a.keychange, keystr)

		Undisplay(a.spend)
		Undisplay(a.change)

	})

	AddEventListener(a.stealthpaginator, "click", "target", func(this, target JQuery) {
		var key = Value(a.stealthbase)
		var isClaimingVisible = len(key) != 64
		if key == "" {
			key = Value(a.key)
		}
		if key == "" {
			return
		}

		var str = target.InnerHTML()
		n, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return
		}
		if n < 1 {
			return
		}
		a.controller_stealth(key, uint64(n-1), isClaimingVisible)
		SetValue(a.stealthspaginatorpage, str)
	})

	// only when user presses enter, in the box, go to page number,...
	AddEventListener(a.stealthspaginatorpage, "keyup", "key", func(this, target JQuery) {
		var key = Value(a.stealthbase)
		var isClaimingVisible = len(key) != 64
		if key == "" {
			key = Value(a.key)
		}
		if key == "" {
			return
		}

		if target.String() != "Enter" && target.String() != "NumpadEnter" && target.String() != "Return" {
			return
		}
		var str = Value(a.stealthspaginatorpage)
		n, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return
		}
		if n < 1 {
			return
		}
		a.controller_stealth(key, uint64(n-1), isClaimingVisible)
		SetValue(a.stealthspaginatorpage, str)
	})
	AddEventListener(a.stealthspaginatorgoto, "click", "target", func(this, target JQuery) {
		var key = Value(a.stealthbase)
		var isClaimingVisible = len(key) != 64
		if key == "" {
			key = Value(a.key)
		}
		if key == "" {
			return
		}

		var str = Value(a.stealthspaginatorpage)
		n, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return
		}
		if n < 1 {
			return
		}
		a.controller_stealth(key, uint64(n-1), isClaimingVisible)
		SetValue(a.stealthspaginatorpage, str)
	})

	AddEventListener(a.stealthstealth, "click", "target", func(this, target JQuery) {
		SetValue(a.stealthbase, Value(a.stealthkey))

		var str = Value(a.stealthspaginatorpage)
		if str == "" {
			str = "1"
		}
		n, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return
		}
		if n < 1 {
			return
		}
		a.controller_stealth(Value(a.stealthbase), uint64(n-1), len(Value(a.stealthbase)) != 64)
		SetValue(a.stealthspaginatorpage, fmt.Sprintf("%d", n))
	})
	AddEventListener(a.stealthstealthsweep, "click", "target", func(this, target JQuery) {
		var key = Value(a.stealthbase)
		if key == "" {
			key = Value(a.key)
		}
		if key == "" {
			return
		}

		var str = Value(a.stealthspaginatorpage)
		if str == "" {
			str = "1"
		}
		n, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return
		}
		if n < 1 {
			return
		}
		a.controller_stealth_sweep(key, uint64(n-1))
		SetValue(a.stealthspaginatorpage, fmt.Sprintf("%d", n))
	})

	AddEventListener(a.claimstealthcopy, "click", "target", func(this, target JQuery) {
		a.claimstealthurl.Call("select")
		Document.Call("execCommand", "copy")
	})

	var updateClaimWallUrl = func() {
		var count = Value(a.claimstealthcount)
		var name = Value(a.claimstealthname)
		var addr = Value(a.stealthkey)

		SetValue(a.claimstealthurl, DocumentOrigin()+"#"+addr+count+name)
	}

	AddEventListener(a.claimstealthcount, "keydown", "target", func(this, target JQuery) {
		updateClaimWallUrl()
	})

	AddEventListener(a.claimstealthname, "keydown", "target", func(this, target JQuery) {
		updateClaimWallUrl()
	})

	AddEventListener(a.claimstealthcount, "input", "target", func(this, target JQuery) {
		updateClaimWallUrl()
	})

	AddEventListener(a.claimstealthname, "input", "target", func(this, target JQuery) {
		updateClaimWallUrl()
	})

	AddEventListener(a.claimstealthcount, "change", "target", func(this, target JQuery) {
		updateClaimWallUrl()
	})

	AddEventListener(a.claimstealthname, "change", "target", func(this, target JQuery) {
		updateClaimWallUrl()
	})
}

func (a *AppPay) bindEvents(i *AppImport) {
	AddEventListener(a.add, "click", "target", func(this, target JQuery) {

		var source = Value(a.keysource)
		var change = Value(a.stacktop)
		var out = Value(a.addr)
		var amount = Value(a.amount)

		if !AddrsCompatible(change, source, out) {
			return
		}

		var stacktop = a.controller_add_destination(source, change, out, amount)

		SetValue(a.stacktop, stacktop)
	})
	AddEventListener(a.pop, "click", "target", func(this, target JQuery) {

		var change = Value(a.stacktop)

		var stacktop = a.controller_pop_destination(change)

		SetValue(a.stacktop, stacktop)
	})
	AddEventListener(a.pay, "click", "target", func(this, target JQuery) {
		n, err := strconv.ParseInt(Value(a.keycount), 10, 64)
		if err != nil {
			n = 0
		}
		var source = Value(a.keysource)
		var dest = Value(a.stacktop)
		if !AddrsCompatible(dest, source) {
			return
		}
		if LoopDetect(dest, source) {
			return
		}
		if LoopDetect(source, dest) {
			return
		}

		var password = Value(a.password)
		var txid = a.controller_pay(dest, source, password, n)

		if len(txid) > 0 {
			DownloadBase64(i.controller_export(nil), "wallet"+txid+".dat")
			DisplayBlock(a.destinationsblock)
		}
	})
}

func (a *AppChart) bindEvents() {

	var zooming = func(n int) {
		var width = a.zoom[1] - a.zoom[0]
		if n < 0 {
			if a.zoom[4] > 0 {
				a.zoom[4]--
				a.zoom[1] += width / 2
				a.zoom[0] -= width / 2
			}
		} else {
			if a.zoom[4] < 16 {
				a.zoom[4]++
				a.zoom[0] += width / 4
				a.zoom[1] -= width / 4
			}
		}
	}
	AddWindowEventListener("resize", "target", func(this, target JQuery) {
		SetSrc(a.chart, CandlestickChart(a.data, uint64(ClientWidth(a.chartdiv)), a.zoom[0], a.zoom[1], a.zoom[2], a.zoom[3], a.zoom[4]))
	})

	AddEventListener(a.in, "click", "target", func(this, target JQuery) {
		zooming(1)
		SetSrc(a.chart, CandlestickChart(a.data, uint64(ClientWidth(a.chartdiv)), a.zoom[0], a.zoom[1], a.zoom[2], a.zoom[3], a.zoom[4]))
	})
	AddEventListener(a.out, "click", "target", func(this, target JQuery) {
		zooming(-1)
		SetSrc(a.chart, CandlestickChart(a.data, uint64(ClientWidth(a.chartdiv)), a.zoom[0], a.zoom[1], a.zoom[2], a.zoom[3], a.zoom[4]))
	})
	AddEventListener(a.chart, "wheel", "wheelDeltaY", func(this, target JQuery) {
		zooming(target.Int())
		SetSrc(a.chart, CandlestickChart(a.data, uint64(ClientWidth(a.chartdiv)), a.zoom[0], a.zoom[1], a.zoom[2], a.zoom[3], a.zoom[4]))
	})
	AddEventListener(a.left, "click", "target", func(this, target JQuery) {
		var width = (a.zoom[1] - a.zoom[0]) / 3
		a.zoom[0] -= width
		a.zoom[1] -= width

		SetSrc(a.chart, CandlestickChart(a.data, uint64(ClientWidth(a.chartdiv)), a.zoom[0], a.zoom[1], a.zoom[2], a.zoom[3], a.zoom[4]))
	})
	AddEventListener(a.right, "click", "target", func(this, target JQuery) {
		var width = (a.zoom[1] - a.zoom[0]) / 3
		a.zoom[0] += width
		a.zoom[1] += width

		SetSrc(a.chart, CandlestickChart(a.data, uint64(ClientWidth(a.chartdiv)), a.zoom[0], a.zoom[1], a.zoom[2], a.zoom[3], a.zoom[4]))
	})
	AddEventListener(a.reload, "click", "target", func(this, target JQuery) {

		a.controller_refresh("disk", func() {
			SetSrc(a.chart, CandlestickChart(a.data, uint64(ClientWidth(a.chartdiv)), a.zoom[0], a.zoom[1], a.zoom[2], a.zoom[3], a.zoom[4]))
		})

	})
	AddEventListener(a.bidsreload, "click", "target", func(this, target JQuery) {
		a.controller_refresh("mempool", func() {
			a.ViewTx()
		})
	})
	AddEventListener(a.bidsfilter, "click", "target", func(this, target JQuery) {
		go a.controller_filter(func() {
			a.ViewTx()
		})
	})
}
