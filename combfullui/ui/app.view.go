package main

import "fmt"
import "strings"
import "encoding/hex"
import "encoding/binary"
import "encoding/base64"
import "sort"
import "strconv"

// redrawJsMainPageData is called when js loads main.js or when this wasm loads and inits
func (a *AppHome) redrawJsMainPageData(i *AppImport, w *AppWallet) {
	data := string(a.mainLoad.InnerHTML())
	if data == "DONE" {
		return
	}
	rawDecodedText, err := base64.StdEncoding.DecodeString(data)
	SetValue(a.mainLoad, "DONE")
	if err != nil {
		a.controller_refresh(i, w, 1)
		return
	}
	a.controller_repaint_str(i, w, string(rawDecodedText))
}

func (a *App) loadCss() {
	Undisplay(jQuery("#loading"))
	DisplayBlock(jQuery("#main"))
}

const PaginatorHTML = `<a class="w3-bar-item w3-button" href="#">`
const PaginatorActiveHTML = `<a class="w3-bar-item w3-button w3-green" href="#">`
const PaginatorLeftHTML = `«`
const PaginatorRightHTML = `»`
const PaginatorCloseHTML = `</a>`

func row_slash(k string, v0, v1, v2 uint64) string {
	if v0 == 0 && v1 == 0 && v2 == 0 {
		return "<td name='key' class='unspent'>" + k +
			"</td><td name='tx' class='pending'></td><td name='tx' class='rejected'></td><td name='bal'></td>"
	}
	return "<td name='key' class='spent'>" + k + "</td><td name='tx' class='pending'>" + fmt.Sprint(v0) + "/1</td><td name='tx' class='rejected'>" + fmt.Sprint(v1) + "/1</td><td name='bal'>" + fmt.Sprint(v2) + "/1</td>"
}

func row_icon(icon, k string, v uint64, icon2, text2, action2 string) string {
	return "<td onclick='this.parentElement.querySelector(\"td.unspent\")?.querySelector(\"a\")?.click();'><i class='fa-solid fa-" + icon + "'></i></td><td name='key' class='unspent'><a name='unspent' href='javascript:void(0)'>" + k + "</a></td><td name='bal'>" + fmt.Sprintf("%d.%08d", combs(v), nats(v)) + 
		"<a class='w3-btn w3-purple w3-ripple' href='javascript:void(0)' onclick='document.getElementById(\""+action2+"\").click(this.parentElement.parentElement.querySelector(\"td.unspent\")?.querySelector(\"a\")?.click())'><i class='fa-solid fa-" + icon2 + "'></i> " + text2 + "</a></td>"
}
func rowspent_icon(icon, k string, v uint64) string {
	return "<td onclick='this.parentElement.querySelector(\"td.spent\")?.querySelector(\"a\")?.click();'><i class='fa-solid fa-" + icon + "'></i></td><td name='key' class='spent'><a name='spent' href='javascript:void(0)'>" + k + "</a></td><td name='bal'>" + fmt.Sprintf("%d.%08d", combs(v), nats(v)) + "</td>"
}

func row(k string, v uint64) string {
	return "<td name='key' class='unspent'>" + k + "</td><td name='bal'>" + fmt.Sprintf("%d.%08d", combs(v), nats(v)) + "</td>"
}
func rowspent(k string, v uint64) string {
	return "<td name='key' class='spent'>" + k + "</td><td name='bal'>" + fmt.Sprintf("%d.%08d", combs(v), nats(v)) + "</td>"
}
func rowbid(k string, tx string, v uint64) string {
	return "<td name='key'>" + k + "</td><td name='tx'>" + tx + "</td>" +
		"<td name='bal'>" + fmt.Sprintf("%d.%03d", v/1000, v%1000) + "</td>" +
		"<td name='bal'>" + fmt.Sprintf("%d.%03d", v/250, v%250) + "</td>"
}

func (a *AppWallet) rebalance_wallet() {

	for addr := range keysbalances {
		keysbalances[addr] = CheckBalance(addr)
	}

	a.ViewKeys(Value(a.key))

	// rebalance stealth if needed

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

	a_key := Value(a.key)

	if len(a_key) > 0 {
		a.controller_stealth(a_key, uint64(n-1), len(Value(a.stealthbase)) != 64)
	}
}

func (a *AppWallet) ViewKeys(opened string) {

	var payText = a.spend.InnerHTML()

	var strs []string
	for k := range keysbalances {
		strs = append(strs, k)
	}
	sortStrings(strs)

	a.keystable.SetInnerHTML("")

	for _, k := range strs {
		var opened_class string
		if k == opened {
			opened_class = "opened"
		}

		v := keysbalances[k]
		if v > 0 {
			AppendChildClass(a.keystable, "tr", row_icon("key", k, v, "shopping-basket", payText, "wallet-spend"), opened_class)
		}
	}
	for _, k := range strs {
		var opened_class string
		if k == opened {
			opened_class = "opened"
		}

		v := keysbalances[k]
		if v == 0 {
			if _, ok := possibleSpend[strings.ToLower(k)]; ok {
				AppendChildClass(a.keystable, "tr", rowspent_icon("key", k, v), opened_class)
			} else if _, ok := tx[k]; ok {
				AppendChildClass(a.keystable, "tr", rowspent_icon("key", k, v), opened_class)
			} else {
				AppendChildClass(a.keystable, "tr", row_icon("key", k, v, "shopping-basket", payText, "wallet-spend"), opened_class)
			}
		}
	}
}

func (a *AppWallet) ViewStealth(opened string, active_page uint64, isClaimingVisible bool) {

	var sweepText = a.stealthsweep.InnerHTML()

	var icon = "shield"
	if isClaimingVisible {
		icon = "gem"
	}

	var strings []string
	for k := range stealthbalances {
		strings = append(strings, k)
	}
	sortStrings(strings)

	a.stealthtable.SetInnerHTML("")
	for _, k := range strings {

		v := stealthbalances[k]
		if v > 0 {
			var kk = k
			if isClaimingVisible {
				kk = bech32get(commitment(kk))
			}
			var opened_class string
			if kk == opened {
				opened_class = "opened"
			}
			if len(stealthused[k]) == 0 {
				AppendChildClass(a.stealthtable, "tr", row_icon(icon, kk, v, "hand-sparkles", sweepText, "wallet-sweep"), opened_class)
			} else {
				AppendChildClass(a.stealthtable, "tr", rowspent_icon(icon, kk, v), opened_class)
			}
		}
	}
	for _, k := range strings {
		v := stealthbalances[k]
		if v == 0 {
			var kk = k
			if isClaimingVisible {
				kk = bech32get(commitment(kk))
			}
			var opened_class string
			if kk == opened {
				opened_class = "opened"
			}
			if len(stealthused[k]) == 0 {
				AppendChildClass(a.stealthtable, "tr", row_icon(icon, kk, v, "hand-sparkles", sweepText, "wallet-sweep"), opened_class)
			} else {
				AppendChildClass(a.stealthtable, "tr", rowspent_icon(icon, kk, v), opened_class)
			}
		}
	}

	a.stealthpaginator.SetInnerHTML("")

	for i := uint64(0); i < 5; i++ {
		var pageid = fmt.Sprintf("%d", i+1)
		if i == active_page {
			AppendChild(a.stealthpaginator, "span", PaginatorActiveHTML+pageid+PaginatorCloseHTML)
		} else {
			AppendChild(a.stealthpaginator, "span", PaginatorHTML+pageid+PaginatorCloseHTML)
		}
	}
	AppendChild(a.stealthpaginator, "span", PaginatorHTML+PaginatorRightHTML+PaginatorCloseHTML)
}

func prefix_to_page_num(prefix string) uint64 {
	if len(prefix)&1 == 1 {
		prefix = "0" + prefix
	}

	var buf, err = hex.DecodeString(prefix)
	if err != nil {
		return 0
	}
	var buff [8]byte
	copy(buff[:], buf)

	var num = binary.LittleEndian.Uint64(buff[:])

	if len(prefix)&1 == 1 {
		num /= 16
	}
	return num
}

func (a *AppCoins) ViewBalances(prefix string) {
	var recommendedLength = AllBalancesPrefixLen()
	for len(prefix) < recommendedLength {
		prefix = "0" + prefix
	}
	var pagenum = prefix_to_page_num(prefix)
	var uprefix = strings.ToUpper(prefix)
	a.balancestablepaginator.SetInnerHTML("")
	AppendChild(a.balancestablepaginator, "span", PaginatorHTML+PaginatorLeftHTML+PaginatorCloseHTML)

	var min = pagenum
	var max = pagenum + 7
	if min < 7 {
		max += 7 - min
		min = 0
	} else {
		min -= 7
	}
	for i := min; i <= max; i++ {
		var pageid = fmt.Sprintf("%x", i)
		if len(pageid) > recommendedLength {
			break
		}
		for len(pageid) < recommendedLength {
			pageid = "0" + pageid
		}
		var tag = PaginatorHTML
		if prefix == pageid {
			tag = PaginatorActiveHTML
		}
		AppendChild(a.balancestablepaginator, "span", tag+pageid+PaginatorCloseHTML)
	}

	AppendChild(a.balancestablepaginator, "span", PaginatorHTML+PaginatorRightHTML+PaginatorCloseHTML)

	a.balancestablebody.SetInnerHTML("")

	var strs []string

	AllBalances(func(addr string, bal uint64) {
		if strings.HasPrefix(addr, prefix) || strings.HasPrefix(addr, uprefix) {
			strs = append(strs, row(addr, bal))
		}
	})
	sortStrings(strs)
	for _, str := range strs {
		AppendChild(a.balancestablebody, "tr", str)
	}
}

func (a *AppImport) ViewProgress(percent int) {
	a.progress.SetInnerHTML(fmt.Sprintf("%d.%03d%%", percent/1000, percent%1000))
	Width(a.progress, fmt.Sprintf("%d.%03d%%", percent/1000, percent%1000))
}

func (a *AppImport) ViewFileProgress(percent int) {
	a.fileprogress.SetInnerHTML(fmt.Sprintf("%d.%03d%%", percent/1000, percent%1000))
	Width(a.fileprogress, fmt.Sprintf("%d.%03d%%", percent/1000, percent%1000))
}

func (a *AppChart) ViewTx() {

	var exists = make(map[string]struct{})

	var data [][2]uint64
	for i, k := range a.tx {
		if k == nil {
			continue
		}
		if _, ok := exists[k.TxId]; ok {
			continue
		}
		exists[k.TxId] = struct{}{}
		data = append(data, [2]uint64{uint64(i), k.FeeWeightKb})
	}
	sort.Slice(data, func(i, j int) bool { return data[i][1] > data[j][1] })

	a.bidstable.SetInnerHTML("")
	for _, v := range data {
		k := a.tx[v[0]]
		for _, comm := range k.TxOut {
			if comm == nil {
				continue
			}
			AppendChild(a.bidstable, "tr", rowbid(bech32get(comm.Commitment), k.TxId, k.FeeWeightKb))
			break
		}
	}

}
