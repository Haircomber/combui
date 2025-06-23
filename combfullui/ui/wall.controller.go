package main

import "fmt"

func (a *Wall) bindEvents() {

	var current_block string

	AddEventListener(a.syncbtn, "click", "target", func(this, target JQuery) {

		addr := DocumentHash()[1 : 64+1]
		amt := byte(DocumentHash()[64+1] - '0')
		name := DocumentHash()[64+1+1:]

		a.addr.SetInnerHTML(addr)
		a.name.SetInnerHTML(name)

		_ = amt

		a.controller_refresh("mempool", func(new_block string) {
			if new_block != current_block {
				current_block = new_block

				Undisplay(a.img)

				_, rejected, deposit := a.controller_used_stealth(addr, 0)

				a.bal.SetInnerHTML(fmt.Sprint(deposit) + "/" + fmt.Sprint(amt))
				a.rej.SetInnerHTML(fmt.Sprint(rejected) + "/" + fmt.Sprint(rejected+deposit))
			} else {
				a.controller_stealth(addr, 0)
			}

			go a.syncbtn.Call("click")
		})
	})

	AddEventListener(a.loadbtn, "click", "target", func(this, target JQuery) {

		addr := DocumentHash()[1 : 64+1]
		amt := byte(DocumentHash()[64+1] - '0')
		name := DocumentHash()[64+1+1:]

		a.addr.SetInnerHTML(addr)
		a.name.SetInnerHTML(name)

		a.controller_stealth(addr, 0)

		a.bal.SetInnerHTML("0" + "/" + fmt.Sprint(amt))
		a.rej.SetInnerHTML("0" + "/" + "0")
	})

	AddEventListener(a.reloadbtn, "click", "target", func(this, target JQuery) {

		addr := DocumentHash()[1 : 64+1]
		amt := byte(DocumentHash()[64+1] - '0')
		name := DocumentHash()[64+1+1:]

		a.addr.SetInnerHTML(addr)
		a.name.SetInnerHTML(name)

		go func() {
			_, rejected, deposit := a.controller_used_stealth(addr, 0)

			a.bal.SetInnerHTML(fmt.Sprint(deposit) + "/" + fmt.Sprint(amt))
			a.rej.SetInnerHTML(fmt.Sprint(rejected) + "/" + fmt.Sprint(rejected+deposit))
		}()
	})
	AddEventListener(a.stealthtable, "click", "target", func(this, target JQuery) {
		var str = target.InnerHTML()

		if len(str) != 62 {
			return
		}

		SetSrc(a.img, QRCODE("bitcoin:"+str+"?amount=0.00000330"))
		DisplayBlock(a.img)
	})
}
