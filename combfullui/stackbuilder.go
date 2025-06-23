package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func prettyprint_liq_stack_items(w http.ResponseWriter, from, key [32]byte, until *[32]byte) (bool, bool, uint64, int) {
	var val, ok = segments_stack[key]

	if !ok {
		return false, false, 0, 0
	}

	var to, sumto, sum = stack_decode(val[0:])

	//fmt.Fprintf(w, `<h2>%X COMB</h2>`+"\n", key)

	// recognize padding (0 comb output to the original comb)

	var padding = (sum == 0 && sumto == from)

	if padding {
		fmt.Fprintf(w, ` <ul><li>Pay to <tt>%s</tt>: 0 COMB (<b>padding</b>) </li></ul>`+"\n", CombAddr(sumto))
	} else {

		fmt.Fprintf(w, ` <ul><li>Pay to <tt>%s</tt>: %d.%08d COMB </li></ul>`+"\n", CombAddr(sumto), combs(sum), nats(sum))

	}

	if to == *until {
		return padding, true, sum, 1
	}

	var padded, goodstack, totalnats, totalentries = prettyprint_liq_stack_items(w, from, to, until)

	return padded || padding, goodstack, totalnats + sum, totalentries + 1
}

func stack_load_multipay_data(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	var walletdata, changedata, data = ps.ByName("wallet"), ps.ByName("change"), ps.ByName("data")

	fmt.Fprint(w, testnetColorBody())
	defer fmt.Fprintf(w, `</body></html>`)

	err1 := checkHEX32(walletdata)
	if err1 != nil {
		fmt.Fprintf(w, "error stack building from walletdata: %s", err1)
		return
	}
	err2 := checkHEX32(changedata)
	if err2 != nil {
		fmt.Fprintf(w, "error stack building targeting changedata: %s", err2)
		return
	}

	var rawwalletdata = hex2byte32([]byte(walletdata))
	var rawchangedata = hex2byte32([]byte(changedata))

	var stackhash = stack_load_data_internal(w, data)
	fmt.Fprintf(w, `<a href="/sign/multipay/%s/%s/%s">&larr; Back to your multipayment</a><br />`, CombAddr(rawwalletdata), CombAddr(rawchangedata), CombAddr(stackhash))
}

func stackbuilder(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprint(w, testnetColorBody(`onload="javascript:bar();"`)+`<a href="/wallet/">&larr; Back to wallet</a><br />`)
	defer fmt.Fprintf(w, `</body></html>`)

	var walletkey, change, stackbottom = ps.ByName("walletkey"), ps.ByName("change"), ps.ByName("stackbottom")

	_ = walletkey

	err1 := checkHEX32(walletkey)
	if err1 != nil {
		fmt.Fprintf(w, "error stack building from wallet: %s", err1)
		return
	}
	err2 := checkHEX32(change)
	if err2 != nil {
		fmt.Fprintf(w, "error stack building targeting change address: %s", err2)
		return
	}
	err3 := checkHEX32(stackbottom)
	if err3 != nil {
		fmt.Fprintf(w, "error stack building with stack bottom: %s", err3)
		return
	}

	var rawwalletkey = hex2byte32([]byte(walletkey))
	var rawchange = hex2byte32([]byte(change))
	var rawstackbottom = hex2byte32([]byte(stackbottom))
	// print entries below stackbottom until change is reached
	fmt.Fprintf(w, "<h1>From address <tt>%s</tt></h1>\n", CombAddr(rawwalletkey))

	fmt.Fprintf(w, `
		<hr />
		To: <input id="to" style="font-family:monospace;width:45em" />
		Amount (nats): <input pattern="^[0-9]+$" id="amt" style="font-family:monospace;width:10em" />
		<a href="#" id="dostack">Add destination</a>
		<a href="#" id="cleardest" onclick="javascript:bar();">Clear destination</a>
		
		<script type="text/javascript">
		  var bar = function() {
		    document.getElementById("amt").value = "";
		    document.getElementById("to").value = "";
		    document.getElementById("activator").style.visibility = 'visible';
		    document.getElementById("advice").style.visibility = 'hidden';
		  }
		
		  var foo = function() {
		    var len = (document.getElementById("amt").value.length)+(document.getElementById("to").value.length);
		    if (len == 0) {
		        document.getElementById("activator").style.visibility = 'visible';
		        document.getElementById("advice").style.visibility = 'hidden';
		    } else {
		        document.getElementById("activator").style.visibility = 'hidden';
		        document.getElementById("advice").style.visibility = 'visible';
		    }
		  
		    var number = parseInt(document.getElementById("amt").value);
		    
		    document.getElementById("amt").value = (number+"").replace("NaN", "").replace("-", "").replace("+", "");
		    document.getElementById("dostack").href="/stack/multipaydata/%s/%s/%X"+
			document.getElementById("to").value.toUpperCase() +
			""+number.toString(16).toUpperCase().padStart(16, "0");
			
			
			
		    return false;
		  };
		document.getElementById("to").oninput = foo;
		document.getElementById("amt").oninput = foo;
		document.getElementById("to").onpropertychange = foo;
		document.getElementById("amt").onpropertychange = foo;
		</script>
		<hr />

	`, CombAddr(rawwalletkey), CombAddr(rawchange), rawstackbottom)

	var padded, goodstack, totalnats, totalentries = prettyprint_liq_stack_items(w, rawwalletkey, rawstackbottom, &rawchange)

	if !goodstack {
		if rawchange != rawstackbottom {
			fmt.Fprintf(w, `<h1>THE BOTTOM OF THE STACK IS NOT YOUR CHANGE ADDRESS. IT IS NOT SAFE TO CONTINUE!</h1>`)
			return
		}
	}
	fmt.Fprintf(w, "<h2>%d.%08d COMB will be spent to %d destination(s).</h2>\n", combs(totalnats), nats(totalnats), totalentries)
	fmt.Fprintf(w, "<p>The remainder will go to the change address <small><tt>%s</tt></small></p>", CombAddr(rawchange))

	if !padded && rawstackbottom == rawchange {
		fmt.Fprintf(w, "<a href=\"/stack/multipaydata/%s/%s/%X%X0000000000000000\">Pad the stack</a>\n", CombAddr(rawwalletkey), CombAddr(rawchange), rawchange, rawwalletkey)
	}

	fmt.Fprintf(w, "<a href=\"/stack/multipaydata/%s/%s/%X%X%016X\">Fold the stack</a>\n", CombAddr(rawwalletkey), CombAddr(rawchange), rawchange, rawstackbottom, totalnats)
	fmt.Fprintf(w, "<a id=\"activator\" href=\"/sign/pay/%s/%s\">Do payment</a>\n", CombAddr(rawwalletkey), CombAddr(rawstackbottom))
	fmt.Fprintf(w, `<span id="advice" style="visibility:hidden">Please add or clear destination to complete your payment</span>`)

	if !padded && rawstackbottom == rawchange {
		fmt.Fprintf(w, "<h1>Please pad the stack</h1>")
	}
}
