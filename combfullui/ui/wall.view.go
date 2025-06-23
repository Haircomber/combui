package main

import "fmt"

func (a *Wall) loadCss() {
	Undisplay(jQuery("#loading"))
	DisplayBlock(jQuery("#paywall"))
}

func (a *Wall) ViewStealth(active_page uint64) (pending, rejected, deposit uint64) {

	// pending map construction
	var pendingMap = make(map[string]struct{})
	for _, k := range a.tx {
		if k == nil {
			continue
		}
		for _, comm := range k.TxOut {
			if comm == nil {
				continue
			}
			bechaddr := bech32get(comm.Commitment)
			println(bechaddr)
			pendingMap[bechaddr] = struct{}{}
		}
	}

	//

	var strings []string
	for k := range stealthbalances {
		strings = append(strings, k)
	}
	sortStrings(strings)

	a.stealthtable.SetInnerHTML("")
	for _, k := range strings {
		_, ok := stealthused[k]
		_, ok2 := stealthclaimed[k]
		var commitstr = commitment(k)
		var str2 = bech32get(commitstr)

		_, ok0 := pendingMap[str2]
		var v0 uint64
		if ok0 && !(ok && !ok2) && !ok2 {
			v0 = 1
		}
		var v1 uint64
		if ok && !ok2 {
			v1 = 1
		}
		var v2 uint64
		if ok2 {
			v2 = 1
		}

		if !ok {
			AppendChild(a.stealthtable, "tr", row_slash(str2, v0, v1, v2))

			rejected += v1
			deposit += v2
		}
	}
	for _, k := range strings {
		_, ok := stealthused[k]
		_, ok2 := stealthclaimed[k]
		var commitstr = commitment(k)
		var str2 = bech32get(commitstr)
		_, ok0 := pendingMap[str2]
		var v0 uint64
		if ok0 && !(ok && !ok2) && !ok2 {
			v0 = 1
		}
		var v1 uint64
		if ok && !ok2 {
			v1 = 1
		}
		var v2 uint64
		if ok2 {
			v2 = 1
		}

		if ok {
			AppendChild(a.stealthtable, "tr", row_slash(str2, v0, v1, v2))

			rejected += v1
			deposit += v2
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

	return
}
