package main

import "fmt"
import "strings"

func (a *AppImport) controller_export(export_set map[string]struct{}) string {
	var buffer string

	for k, v := range stack {
		if export_set != nil {
			if _, ok := export_set[k]; !ok {
				continue
			}
		}
		var amt = stackAmt[k]
		var chg_dst = v[Change] + v[Target]

		var testnet = !isMainnetAddr(chg_dst)
		var amthex = fmt.Sprintf("%016X", amt)

		buffer += wallet_header_net(WALLET_HEADER_STACK_DATA, testnet) + strings.ToUpper(chg_dst) + amthex + "\r\n"
	}

	for key, value := range merkleactive {
		if export_set != nil {
			if _, ok := export_set[key]; !ok {
				continue
			}
		}
		var testnet = !isMainnetAddr(key)
		var val = wallet_header_net(WALLET_HEADER_MERKLE_DATA, testnet)

		txid := value[5]
		values := merkleTx[txid]

		for i := 0; i < 22; i++ {
			val += strings.ToUpper(values[i])
		}

		buffer += val + "\r\n"
	}

	for key, value := range keys {
		if export_set != nil {
			continue
		}
		var testnet = !isMainnetAddr(key)
		var val = wallet_header_net(WALLET_HEADER_WALLET_DATA, testnet)

		for i := 0; i < 21; i++ {
			val += strings.ToUpper(value[i])
		}

		buffer += val + "\r\n"
	}

	for key, value := range txactive {
		if export_set != nil {
			if _, ok := export_set[value[0]]; !ok {
				continue
			}
		}
		var testnet = !isMainnetAddr(key)
		var val = wallet_header_net(WALLET_HEADER_TX_RECV, testnet)

		for i := 0; i < 23; i++ {
			val += strings.ToUpper(value[i])
		}
		buffer += val + "\r\n"
	}
	for key, value := range txmempool {
		if export_set != nil {
			if _, ok := export_set[value[0]]; !ok {
				continue
			}
		}
		var testnet = !isMainnetAddr(key)
		var val = wallet_header_net(WALLET_HEADER_TX_RECV, testnet)

		for i := 0; i < 23; i++ {
			val += strings.ToUpper(value[i])
		}
		buffer += val + "\r\n"
	}

	return buffer
}
