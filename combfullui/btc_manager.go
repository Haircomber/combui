package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	//"time"
	"errors"
)

type RPC_Result struct {
	Result json.RawMessage `json:"result"`
}

var conn_mut sync.Mutex
var btc_is_connected bool

func get_connected() bool {
	conn_mut.Lock()
	var conn = btc_is_connected
	conn_mut.Unlock()
	return conn
}

func set_connected(onoff bool) {
	conn_mut.Lock()
	defer conn_mut.Unlock()
	if onoff != btc_is_connected {
		switch onoff {
		case true:
			fmt.Println("Connected to BTC")
		case false:
			fmt.Println("Disconnected from BTC")
		}
		btc_is_connected = onoff
	}
}

var errNilJson = errors.New("nil_json")
var errRetriedNoBlock = errors.New("retried_but_no_block")
var errNoBtc = errors.New("no_btc")

func make_bitcoin_call(client *http.Client, method, params string) (json.RawMessage, error) {

	port := "8332"
	if u_config.regtest {
		port = "18443"
	} else if u_config.testnet4 {
		port = "48332"
	} else if u_config.testnet {
		port = "18332"
	}
	body := strings.NewReader("{\"jsonrpc\":\"1.0\",\"id\":\"curltext\",\"method\":\"" + method + "\",\"params\":[" + params + "]}")
	req, err := http.NewRequest("POST", "http://"+u_config.username+":"+u_config.password+"@127.0.0.1:"+port, body)

	if err != nil {
		log.Fatal("phone btc ERROR", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		set_connected(false)
		log.Println("No btc:", err)
		return nil, errNoBtc
	}

	defer resp.Body.Close()
	resp_bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("body_read_error_3")
	}

	// Check for a return value
	if len(resp_bytes) == 0 {
		fmt.Println("BTC Comms Error: Likely an incorrect RPC username or password. Please make sure the username and password in Haircomb's config.txt file match the ones stored in Bitcoin's bitcoin.conf file, then restart both programs.")
		return nil, errors.New("bad_log_info")
	}

	var result RPC_Result

	err = json.Unmarshal(resp_bytes, &result)
	if err != nil {
		return nil, errors.New("bad_json")
	}

	if result.Result == nil || string(result.Result) == "null" {
		return nil, errNilJson
	}

	//set_connected(true)

	return result.Result, nil

}
