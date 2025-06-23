package main

import "strings"

func (a *Wall) controller_resort_str(str string) []*AppModelResponseBid {
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

func (a *Wall) controller_block_str(str string) string {
	data := Parse(str)
	if data == nil {
		return ""
	}
	if !data.Success {
		return ""
	}
	return data.BlockHash
}

func (a *Wall) controller_refresh(source string, done_action func(string)) {
	go Call(backend, "chart"+source+".js", func(str string) {
		if source == "mempool" {
			a.tx = append(a.tx, a.controller_resort_str(str)...)
		}

		done_action(a.controller_block_str(str))
	})
}
