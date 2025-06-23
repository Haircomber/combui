package main

const WALLET_HEADER_STACK_DATA = 1
const WALLET_HEADER_WALLET_DATA = 2
const WALLET_HEADER_PURSE_DATA = 3
const WALLET_HEADER_TX_RECV = 4
const WALLET_HEADER_MERKLE_DATA = 5

func wallet_header_net(id int, testnet bool) string {
	if testnet {
		switch id {
		case WALLET_HEADER_STACK_DATA:
			return `\stack\data\`
		case WALLET_HEADER_WALLET_DATA:
			return `\wallet\data\`
		case WALLET_HEADER_PURSE_DATA:
			return `\purse\data\`
		case WALLET_HEADER_TX_RECV:
			return `\tx\recv\`
		case WALLET_HEADER_MERKLE_DATA:
			return `\merkle\data\`
		}

	}
	switch id {
	case WALLET_HEADER_STACK_DATA:
		return `/stack/data/`
	case WALLET_HEADER_WALLET_DATA:
		return `/wallet/data/`
	case WALLET_HEADER_PURSE_DATA:
		return `/purse/data/`
	case WALLET_HEADER_TX_RECV:
		return `/tx/recv/`
	case WALLET_HEADER_MERKLE_DATA:
		return `/merkle/data/`
	}
	return ""
}

var wallet_fix_strings = [][2]string{
	{`stackdata`, `\stack\data\`},
	{`walletdata`, `\wallet\data\`},
	{`pursedata`, `\purse\data\`},
	{`txrecv`, `\tx\recv\`},
	{`merkledata`, `\merkle\data\`},
	{`0d`, "\r\n"},
	{`0a`, "\r\n"},
}
