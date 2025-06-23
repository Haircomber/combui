package main

type App struct {
	loader JQuery

	menu   *AppMain
	main   *AppMain
	home   *AppHome
	coins  *AppCoins
	export *AppImport
	wallet *AppWallet
	pay    *AppPay
	chart  *AppChart
}

func NewApp() *App {
	return &App{jQuery("#loader"),
		&AppMain{
			jQuery("#mainbutton-home"),
			jQuery("#mainbutton-wallet"),
			jQuery("#mainbutton-pay"),
			jQuery("#mainbutton-imports"),
			jQuery("#mainbutton-coins"),
			jQuery("#mainbutton-docs"),
			jQuery("#mainbutton-chart"),
		},
		&AppMain{
			jQuery("#body-home"),
			jQuery("#body-wallet"),
			jQuery("#body-pay"),
			jQuery("#body-imports"),
			jQuery("#body-coins"),
			jQuery("#body-docs"),
			jQuery("#body-chart"),
		},
		&AppHome{
			jQuery("#home-refresh"),
			jQuery("#home-refresh-freq"),
			jQuery("#home-fingerprint"),
			jQuery("#home-commitments"),
			jQuery("#home-p2wsh"),
			jQuery("#home-accelerator"),
			jQuery("#home-existing"),
			jQuery("#home-remaining"),
			jQuery("#home-blockhash"),
			jQuery("#home-blockheight"),
			jQuery("#home-outofsync"),
			jQuery("#home-progress"),
			jQuery("#home-backend"),
			jQuery("#home-main-load"),
			jQuery("#home-shutdown"),
			jQuery("#home-shutdown-spinner"),

			false,
			0,
		},
		&AppCoins{
			jQuery("#coins-combbases-body"),
			jQuery("#coins-combbases-paginator"),
			jQuery("#coins-combbases-refresh"),
			jQuery("#coins-combbases-refresh-spinner"),
			jQuery("#coins-balances-body"),
			jQuery("#coins-balances-paginator"),
			jQuery("#home-blockheight"),
			jQuery("#coins-refresh-depth"),
			false,
			0,
		},
		&AppImport{
			jQuery("#import-check"),
			jQuery("#export-submit"),
			jQuery("#export-data"),
			jQuery("#export2-submit"),
			jQuery("#export2-data"),
			jQuery("#import-scan-mempool"),
			jQuery("#import-scan-mempool-spinner"),
			jQuery("#import-progress"),
			jQuery("#import-file"),
			jQuery("#import-file-progress"),
			jQuery("#import-file-load"),
			jQuery("#import-file-bug"),
			jQuery("#import-file-load-bug"),
			jQuery("#export-deciders-hardfork"),
			jQuery("#export-deciders-hardfork-on-sorry"),
			jQuery("#export-deciders-hardfork-off-sorry"),
			jQuery("#export-deciders-hardfork-hider"),
			false,
			0,
			false,
		},
		&AppWallet{
			jQuery("#wallet-keys"),
			jQuery("#wallet-key"),
			jQuery("#wallet-key-change"),
			jQuery("#wallet-key-claim"),
			jQuery("#wallet-form"),
			jQuery("#wallet-gen-main"),
			jQuery("#wallet-gen-test"),
			jQuery("#wallet-used-check"),
			jQuery("#wallet-gen-main-spinner"),
			jQuery("#wallet-gen-test-spinner"),
			jQuery("#wallet-used-check-spinner"),
			jQuery("#wallet-keycount"),
			jQuery("#wallet-password"),
			jQuery("#wallet-stealths"),
			jQuery("#wallet-stealth-clicked"),
			jQuery("#wallet-key-clicked"),
			jQuery("#wallet-image"),
			jQuery("#wallet-image2"),

			jQuery("#wallet-stealths-paginator"),

			jQuery("#wallet-stealth-image"),
			jQuery("#wallet-stealth-image2"),
			jQuery("#wallet-stealth-key"),
			jQuery("#wallet-stealth-claim-addr"),
			jQuery("#wallet-stealth-claim-count"),
			jQuery("#wallet-stealth-claim-name"),
			jQuery("#wallet-stealth-claim-url"),
			jQuery("#wallet-stealth-claim-copy"),

			jQuery("#wallet-change"),
			jQuery("#wallet-sweep"),
			jQuery("#wallet-spend"),

			jQuery("#wallet-stealths-paginator"),
			jQuery("#wallet-stealths-paginator-page"),
			jQuery("#wallet-stealths-paginator-goto"),

			jQuery("#wallet-stealth-used-check"),
			jQuery("#wallet-stealth-used-check-spinner"),

			jQuery("#wallet-stealth-base"),
			jQuery("#wallet-stealth-stealth"),
			jQuery("#wallet-stealth-sweep"),

			jQuery("#wallet-stealth-half"),

			jQuery("#wallet-hint"),

			jQuery("#wallet-stealths-claimings"),
			jQuery("#wallet-stealth-256-btn"),
			jQuery("#wallet-stealth-16-btn"),
			jQuery("#wallet-claim-256-btn"),
			jQuery("#wallet-claim-16-btn"),
			jQuery("#wallet-claiming-visible"),

			jQuery("#wallet-nochange-visible"),

			false,
			0,
		},
		&AppPay{
			jQuery("#pay-key-source"),
			jQuery("#pay-key-change"),
			jQuery("#pay-stack-top"),
			jQuery("#pay-dest-body"),
			jQuery("#pay-destinations"),
			jQuery("#pay-destination-pop"),
			jQuery("#pay-destination-add"),
			jQuery("#pay-destination-addr"),
			jQuery("#pay-destination-amount"),
			jQuery("#pay-use"),
			jQuery("#pay-keycount"),
			jQuery("#pay-password"),
		},
		&AppChart{
			jQuery("#chart-bids-reload-btn"),
			jQuery("#chart-bids-table"),
			jQuery("#chart-bids-filter-btn"),
			jQuery("#chart-chart-img"),
			jQuery("#chart-chart-div"),
			jQuery("#chart-refresh-btn"),
			jQuery("#chart-reload-btn"),
			jQuery("#chart-in-btn"),
			jQuery("#chart-out-btn"),
			jQuery("#chart-left-btn"),
			jQuery("#chart-right-btn"),
			nil,
			[5]int64{481824, 1000000, 0, 1500000, 0},
			nil,
		},
	}
}

type AppMain struct {
	home    JQuery
	wallet  JQuery
	pay     JQuery
	imports JQuery
	coins   JQuery
	docs    JQuery
	chart   JQuery
}

type AppHome struct {
	refresh     JQuery
	refreshFreq JQuery
	fingerprint JQuery
	commitments JQuery
	p2wsh       JQuery
	accelerator JQuery
	existing    JQuery
	remaining   JQuery
	blockhash   JQuery
	blockheight JQuery
	outofsync   JQuery
	progress    JQuery
	backend     JQuery
	mainLoad    JQuery
	off         JQuery
	offspin     JQuery

	click   bool
	taptime int64
}

type AppCoins struct {
	combbasestablebody      JQuery
	combbasestablepaginator JQuery
	combbasesrefresh        JQuery
	combbasesrefreshspin    JQuery
	balancestablebody       JQuery
	balancestablepaginator  JQuery
	heightholder            JQuery
	combbasesrefreshdepth   JQuery

	click   bool
	taptime int64
}
type AppImport struct {
	importcheck   JQuery
	exportsubmit  JQuery
	exportdata    JQuery
	export2submit JQuery
	export2data   JQuery
	scanmempool   JQuery

	scanmempoolspin JQuery

	progress JQuery

	file         JQuery
	fileprogress JQuery
	fileload     JQuery

	filebug     JQuery
	fileloadbug JQuery

	decidershardfork         JQuery
	decidershardforkonsorry  JQuery
	decidershardforkoffsorry JQuery
	decidershardforkhider    JQuery

	click   bool
	taptime int64

	decider_hardfork_ignore_admin_default bool
}
type AppModelSuccess struct {
	Testnet bool `json:"Testnet"`
	Success bool `json:"Success"`
}

type AppWallet struct {
	keystable     JQuery
	key           JQuery
	keychange     JQuery
	claimkey      JQuery
	form          JQuery
	genmain       JQuery
	gentest       JQuery
	usedcheck     JQuery
	genmainspin   JQuery
	gentestspin   JQuery
	usedcheckspin JQuery
	keycount      JQuery
	password      JQuery

	stealthtable JQuery

	seestealth JQuery
	seewallet  JQuery
	image      JQuery
	image2     JQuery

	stealthpaginator JQuery

	stealthimage  JQuery
	stealthimage2 JQuery
	stealthkey    JQuery
	claimstealth  JQuery

	claimstealthcount JQuery
	claimstealthname  JQuery
	claimstealthurl   JQuery
	claimstealthcopy  JQuery

	change       JQuery
	stealthsweep JQuery

	spend JQuery

	stealthspaginator     JQuery
	stealthspaginatorpage JQuery
	stealthspaginatorgoto JQuery

	usedstealthcheck     JQuery
	usedstealthcheckspin JQuery

	stealthbase         JQuery
	stealthstealth      JQuery
	stealthstealthsweep JQuery

	stealthhalf JQuery

	hint JQuery

	stealthsclaimings     JQuery
	stealth256btn         JQuery
	stealth16btn          JQuery
	claim256btn           JQuery
	claim16btn            JQuery
	walletclaimingvisible JQuery

	nochangevisible JQuery

	click   bool
	taptime int64
}
type AppPay struct {
	keysource JQuery
	keychange JQuery
	stacktop  JQuery

	destinationstablebody JQuery
	destinationsblock     JQuery

	pop    JQuery
	add    JQuery
	addr   JQuery
	amount JQuery

	pay      JQuery
	keycount JQuery
	password JQuery
}

type AppChart struct {
	bidsreload JQuery
	bidstable  JQuery
	bidsfilter JQuery
	chart      JQuery
	chartdiv   JQuery
	refresh    JQuery
	reload     JQuery
	in         JQuery
	out        JQuery
	left       JQuery
	right      JQuery

	data [][2]uint64
	zoom [5]int64
	tx   []*AppModelResponseBid
}
type AppModelResponseUniversal struct {
	AppModelSuccess
	AppModelResponseMain
	AppModelResponseRange
	AppModelResponseGetCommitments
	AppModelResponseChart
	AppModelResponseBids
}
