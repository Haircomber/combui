package main

import "crypto/rand"
import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)
import "sync"

var wallet_mutex sync.RWMutex

var wallet map[[32]byte][21][32]byte
var wallet_commitments map[[32]byte][32]byte
var wallet_saved int

func key_load_data_internal(w http.ResponseWriter, keykey string) {
	err1 := checkHEX672upper(keykey)
	if err1 != nil {
		fmt.Fprintf(w, "error decoding key from hex: %s", err1)
		return
	}

	rawkey := hex2byte672([]byte(keykey))
	var key [21][32]byte
	var buf [672]byte
	var sli []byte
	sli = buf[0:0]

	var mintip [32]byte
	var minheight = ^uint64(0)

	for i := 0; i < 21; i++ {
		copy(key[i][0:], rawkey[32*i:32*i+32])
		tip := key[i]
		for j := 0; j < 59213; j++ {
			if enable_used_key_feature {
				minheight = used_key_try_add(tip, &mintip, minheight)
			}
			tip = nethash(tip[0:])
		}
		sli = append(sli, tip[0:]...)
	}
	pub := nethash(sli)

	if enable_used_key_feature && minheight != ^uint64(0) {
		used_key_add_new_minimal_commit_height(pub, mintip, minheight)
	}

	cpub := commit(pub[0:])

	wallet_mutex.Lock()

	if wallet == nil {
		wallet = make(map[[32]byte][21][32]byte)
	}
	if wallet_commitments == nil {
		wallet_commitments = make(map[[32]byte][32]byte)
	}

	var newkeys = 0

	if _, ok := wallet[pub]; !ok {
		newkeys++
	}

	wallet[pub] = key
	wallet_commitments[cpub] = pub

	wallet_saved += newkeys

	wallet_mutex.Unlock()

	commit_cache_mutex.Lock()
	commits_mutex.Lock()

	if _, ok1 := combbases[cpub]; ok1 {
		segments_coinbase_trickle_auto(cpub, pub)
	}
	commits_mutex.Unlock()
	commit_cache_mutex.Unlock()
}

func wallet_generate_brain(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var numkeys, pass = ps.ByName("numkeys"), ps.ByName("pass")

	var iterations int = -1
	var n, err = fmt.Sscanf(numkeys, "%d", &iterations)
	if n != 1 || err != nil || iterations < 0 {
		fmt.Fprintf(w, "error generating brainwallet. use number of keys")
		return
	}

	var runner = nethash([]byte(pass))

	var key [21][32]byte
	var tip [21][32]byte
	var buf [672]byte
	var pub [32]byte
	var cpub [32]byte
	var sli []byte

	for ; iterations > 0; iterations-- {

		sli = buf[0:0]

		for i := range key {
			key[i] = commit(runner[0:])
			runner = nethash(runner[0:])
		}
		var mintip [32]byte
		var minheight = ^uint64(0)
		for i := range key {
			tip[i] = key[i]
			for j := 0; j < 59213; j++ {
				if enable_used_key_feature {
					minheight = used_key_try_add(tip[i], &mintip, minheight)
				}
				tip[i] = nethash(tip[i][:])
			}
			sli = append(sli, tip[i][:]...)
		}
		pub = nethash(sli)
		if enable_used_key_feature && minheight != ^uint64(0) {
			used_key_add_new_minimal_commit_height(pub, mintip, minheight)
		}
		cpub = commit(pub[0:])
		wallet_mutex.Lock()

		if wallet == nil {
			wallet = make(map[[32]byte][21][32]byte)
		}
		if wallet_commitments == nil {
			wallet_commitments = make(map[[32]byte][32]byte)
		}

		wallet[pub] = key
		wallet_commitments[commit(pub[0:])] = pub

		wallet_mutex.Unlock()
	}

	commit_cache_mutex.Lock()
	commits_mutex.Lock()

	if _, ok1 := combbases[cpub]; ok1 {
		segments_coinbase_trickle_auto(cpub, pub)
	}

	commits_mutex.Unlock()
	commit_cache_mutex.Unlock()

	// NO LOG PASSWORD IN BROWSER HISTORY
	w.Header().Set("Cache-Control", "no-store, private, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("X-Accel-Expires", "0")
	http.Redirect(w, r, "/wallet/", http.StatusTemporaryRedirect)
}

func wallet_generate_key(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var key [21][32]byte
	var tip [21][32]byte
	var buf [672]byte
	var pub [32]byte
	var sli []byte
	sli = buf[0:0]

	fmt.Fprintf(w, testnetColorBody()+`<a href="/wallet/">&larr; Back to wallet</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	for i := range key {
		_, err := rand.Read(key[i][0:])
		if err != nil {
			fmt.Fprintf(w, "error generating true random key: %s", err)
			return
		}
	}

	for i := range key {
		tip[i] = key[i]
		for j := 0; j < 59213; j++ {
			tip[i] = nethash(tip[i][:])
		}
		sli = append(sli, tip[i][:]...)
	}
	pub = nethash(sli)

	wallet_mutex.Lock()

	if wallet == nil {
		wallet = make(map[[32]byte][21][32]byte)
	}
	if wallet_commitments == nil {
		wallet_commitments = make(map[[32]byte][32]byte)
	}

	wallet[pub] = key
	wallet_commitments[commit(pub[0:])] = pub

	wallet_mutex.Unlock()

	if wallet_selfmining_links {

		fmt.Fprintf(w, `<h1><a href="/mining/mine/%X/%s">%X</a></h1>`, commit(pub[:]), serializeutxotag(forcecoinbasefirst(makefaketag())), pub)

	}
}
func wallet_preview_pay(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var walletkey = ps.ByName("walletkey")

	err1 := checkHEX32(walletkey)
	if err1 != nil {
		fmt.Fprintf(w, "error: paying from input: %s", err1)
		return
	}

	var key = hex2byte32([]byte(walletkey))

	fmt.Fprintf(w, testnetColorBody()+`<a href="/wallet/">&larr; Back to wallet</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	wallet_mutex.Lock()

	balance_mutex.RLock()

	var bal = balance_node[key]

	balance_mutex.RUnlock()

	var exists bool

	for key2 := range wallet {

		if key == key2 {
			continue
		}

		var have_spends bool
		var have_spends2 bool
		commits_mutex.RLock()

		txleg_mutex.RLock()

		txdoublespends_each_doublespend_target(key, func(each *[2][32]byte) bool {
			have_spends = true
			return false
		})
		txdoublespends_each_doublespend_target(key2, func(each *[2][32]byte) bool {
			have_spends2 = true
			return false
		})
		txleg_mutex.RUnlock()

		commits_mutex.RUnlock()

		if !have_spends && !have_spends2 {
			exists = true
			fmt.Fprintf(w, `<ul><li> <a href="/sign/multipay/%s/%s/%s">`+
				`multi-pay %d.%08d COMB with <tt>%s</tt> as change address</a> </li></ul>`,
				CombAddr(key), CombAddr(key2), CombAddr(key2), combs(bal), nats(bal), CombAddr(key2))

		}
	}

	wallet_mutex.Unlock()

	if !exists {
		fmt.Fprintf(w, `<h1>No suitable change address found</h1>`)
		fmt.Fprintf(w, `<p>Go back to your wallet, generate new change address, and try paying again.</p>`)
	}
}
func wallet_view(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, testnetColorBody()+`<a href="/">&larr; Back to home</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	wallet_mutex.RLock()

	for key := range wallet {

		wallet_mutex.RUnlock()

		balance_mutex.RLock()

		var bal = balance_node[key]

		balance_mutex.RUnlock()

		var used_key_reorganized bool
		var used_key_spent_using string
		var used_key_spent_at uint64

		if enable_used_key_feature {

			var used_key_m, minimum_has = used_key_fetch(key)
			if minimum_has {
				var empty_minimum used_key_minimum
				if used_key_m == empty_minimum {
					used_key_reorganized = true
				} else {
					used_key_spent_using = bech32get(used_key_m.commit[0:])
					used_key_spent_at = used_key_m.height
					bal = 0
				}
			}
		}

		fmt.Fprintf(w, `<ul><li> <tt>%s</tt> %d.%08d COMB `, CombAddr(key), combs(bal), nats(bal))

		if bal > 0 && !used_key_reorganized {
			fmt.Fprintf(w, `<a href="/sign/from/%s">pay</a> `, CombAddr(key))
		}

		fmt.Fprintf(w, `<a href="/wallet/stealth/%s/0000000000000000">stealth addresses</a> `, CombAddr(key))

		var have_spends bool
		commits_mutex.RLock()

		txleg_mutex.RLock()

		txdoublespends_each_doublespend_target(key, func(each *[2][32]byte) bool {
			have_spends = true
			return false
		})

		if have_spends {
			fmt.Fprint(w, "<ul>\n")
			txdoublespends_each_doublespend_target(key, func(each *[2][32]byte) bool {
				fmt.Fprintf(w, `<li><a href="/sign/pay/%s/%s">Active spend %X </a></li>`, CombAddr(key), CombAddr(each[1]), each[0])
				return true
			})
			fmt.Fprint(w, "</ul>\n")
		}
		txleg_mutex.RUnlock()

		commits_mutex.RUnlock()

		if used_key_reorganized {
			fmt.Fprint(w, "<ul><li>Reorganized. You can repay to original addresses now.</li></ul>\n")
		}
		if used_key_spent_using != "" {
			fmt.Fprintf(w, "<ul><li>Spent using %s <b>DO NOT PAY BITCOIN HERE</b> at height %d</li></ul>\n", used_key_spent_using, used_key_spent_at)
		}

		fmt.Fprint(w, "</li></ul>\n")

		wallet_mutex.RLock()
	}

	wallet_mutex.RUnlock()

	fmt.Fprintf(w, `
		<a href="/wallet/generator">key generate (always fully save your wallet after pressing this button)</a>
	`)
	fmt.Fprintf(w, `<script type="text/javascript">
			  var pwd = function() {
				if (!document.getElementById("walletkeypass").validity.valid) {
					document.getElementById("walletkeygen").action="brain/0/0";
					document.getElementById("walletok").value="NG";
					return false;
				}
			    document.getElementById("walletkeygen").action="brain/"+
				encodeURIComponent(document.getElementById("walletkeycount").value)+"/"+
				encodeURIComponent(document.getElementById("walletkeypass").value)+"#";
			    document.getElementById("walletok").value="OK";
			    return false;
			  };

			</script>
		<div>
			Key count: <input id="walletkeycount" autocomplete="off" type="text" name="username" oninput="javascript: pwd()" onpropertychange="javascript: pwd()" />
			Brain password: <input id="walletkeypass" autocomplete="off" type="password" required minlength="1" pattern="[^\/]+" name="password" oninput="javascript: pwd()" onpropertychange="javascript: pwd()" /><span class="validity"></span>
			<form method="GET" action="brain/0/0" id="walletkeygen"><input id="walletok" type="submit" onclick="javascript:this.value=''" name="submit" value="OK" style="width: 10em" /></form>
		</div>
	`)
	fmt.Fprintf(w, `
		<div><a href="/stack/">liquidity stacks</a></div>
	`)
	if wallet_selfmining_links {
		fmt.Fprintf(w, `
<a href="/mining/mine/FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF/9999999999999999">flush (finalize) block</a>
		`)
	}
	fmt.Fprintf(w, `
			<hr /><div>
			Address: <input id="addr" style="font-family:monospace;width:45em" />
			<a href="#" id="pay">view derived stealth addresses</a>

			<script type="text/javascript">
			  var foo = function() {
			    document.getElementById("pay").href="/wallet/stealth/"+
				document.getElementById("addr").value+"/0000000000000000";
			    return false;
			  };
			document.getElementById("addr").oninput = foo;
			document.getElementById("addr").onpropertychange = foo;

			</script></div>
	`)
}
