package main

var keysbalances = make(map[string]uint64)
var stealthbalances = make(map[string]uint64)
var combbases = make(map[string]uint64)
var keys = make(map[string][21]string)
var txmempool = make(map[string][23]string)
var txactive = make(map[string][23]string)

// merkle internal
// var merkleLegOther = make(map[string][]string)
// var merkleLegTx = make(map[string][]string)
var merklemempool = make(map[string][6]string)
var merkleactive = make(map[string][6]string)
var merkleTx = make(map[string][22]string)

// graph
var stack = make(map[string][ChangeTarget]string)
var stackAmt = make(map[string]uint64)
var stackReverse = make(map[string]string)
var tx = make(map[string]string)

// balances
var throwback = make(map[string]map[string]uint64)
var loops = make(map[string]map[string]uint64)
var stackUsed = make(map[string]map[string]uint64)

// spawned money
var combbasesUncommit = make(map[string]string)

// possible spend, to print double spends, and for loop detector
var possibleSpend = make(map[string][]string)

// used stealth address mapped to a commitment
var stealthused = make(map[string]string)
var stealthclaimed = make(map[string]uint64)

var backend string

func main() {
	if len(DocumentHash()) > 64 {
		wall := NewWall()
		wall.bindEvents()
		Undisplay(wall.img)
		wall.loadbtn.Call("click")
		wall.reloadbtn.Call("click")
		wall.syncbtn.Call("click")
		wall.loadCss()
	} else {
		app := NewApp()
		app.bindEvents()
		app.home.redrawJsMainPageData(app.export)
		app.coins.combbasesrefresh.Call("click")
		app.loadCss()
		backend = DocumentOrigin()
		SetValue(app.home.backend, backend)
	}
	select {}
}
