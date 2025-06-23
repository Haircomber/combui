package main

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type PanicDisplayerLogHider [2]bool

func (notNewlineNotYear *PanicDisplayerLogHider) Write(buf []byte) (n int, err error) {
	for i, b := range buf {

		if !notNewlineNotYear[0] {
			notNewlineNotYear[1] = (b != '2')
		}
		if notNewlineNotYear[1] {
			var b1 = [1]byte{b}
			_, err := os.Stderr.Write(b1[:])
			if err != nil {
				return i, err
			}
		}

		notNewlineNotYear[0] = (b != '\n' && b != '\r')
	}

	return len(buf), nil
}

func main_pub_server_serve(publn net.Listener) {
	pubr := http.HandlerFunc(makeGzipHandler(public_comb_protocol))
	pubsrv := &http.Server{
		Handler:        pubr,
		WriteTimeout:   60 * time.Second,
		ReadTimeout:    60 * time.Second,
		MaxHeaderBytes: 20000000,
	}
	go func(*http.Server, net.Listener) {
		err := pubsrv.Serve(publn)
		if err != nil {
			return
		}
	}(pubsrv, publn)
}

func main() {

	// Hide logging, but show suspicious lines not starting with timestamp (panics)
	log.SetOutput(&PanicDisplayerLogHider{})

	// Pull the RPC info
	load_config()

	// Setup Listen
	ln, err6 := net.Listen("tcp", "127.0.0.1:"+u_config.listen_port)
	if err6 != nil {
		log.Fatal(err6)
	}
	// Setup Public Listen
	publn, err7 := net.Listen("tcp", "0.0.0.0:"+u_config.public_listen_port)
	if err7 != nil {
		log.Fatal(err7)
	}

	// Setup Logging?
	if u_config.logfile != "" {
		logfile, logerr := os.OpenFile(u_config.logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if logerr != nil {
			log.Fatal(logerr)
		}
		defer logfile.Close()
		log.SetOutput(logfile)
	}

	if u_config.init_proxy != "" {
		main_pub_server_serve(publn)
	}

	// Load commits
	fmt.Println("Loading Commits...")
	start := time.Now()
	CommitLvlDbLoad()
	initial_writeback_over = true
	elapsed := time.Since(start)
	fmt.Println("Commits loaded, time spent: ", elapsed)
	fmt.Println("Welcome to Haircomb Core. To operate, open a web browser and go to 127.0.0.1:" + u_config.listen_port)
	fmt.Println("Haircomb will automatically attempt to connect to BTC and begin mining. Progress may be slow if your BTC chain is not up to date, please be patient.")

	// Open the DB
	CommitLvlDbOpen()

	// Start running the miner
	go new_miner_start()

	r := httprouter.New()

	s0, s1, s2, s4, s5, s6, s7, s8, s9, s10, s11, s12 := r, r, r, r, r, r, r, r, r, r, r, r

	s0.GET("/", main_page)
	s0.GET("/shutdown", shutdown_page)
	s0.GET("/version.json", version_page)

	s1.GET("/wallet/view", wallet_view)
	s1.GET("/wallet/generator", wallet_generate_key)
	s1.GET("/wallet/brain/:numkeys/:pass", wallet_generate_brain)
	s1.GET("/wallet/stealth/:backingkey/:offset", wallet_stealth_view)
	s1.GET("/wallet/index.html", wallet_view)
	s1.GET("/wallet/", wallet_view)

	s2.GET("/sign/decide/:decider/:number", sign_use_decider)
	s2.GET("/sign/pay/:walletkey/:destination", sign_use_key)
	s2.GET("/sign/multipay/:walletkey/:change/:stackbottom", stackbuilder)
	s2.GET("/sign/from/:walletkey", wallet_preview_pay)
	s2.GET("/sign/index.html", sign_gui)
	s2.GET("/sign/", sign_gui)

	s4.GET("/tx/recv/:txn", tx_receive_transaction)

	s5.GET("/utxo/bisect/:cut_off_and_mask", bisect_view)
	s5.GET("/utxo/commit/:hash", commit_view)
	s5.GET("/utxo/index.html", utxo_view)
	s5.GET("/utxo/", utxo_view)

	s6.GET("/merkle/data/:data", merkle_load_data)

	s7.GET("/stack/data/:data", stack_load_data)
	s7.GET("/stack/multipaydata/:wallet/:change/:data", stack_load_multipay_data)
	s7.GET("/stack/stealthdata/:data", stack_load_stealth_data)
	s7.GET("/stack/index.html", stacks_view)
	s7.GET("/stack/", stacks_view)

	s8.GET("/height/get", height_view)

	s9.GET("/export/history/:target", routes_all_export)
	s9.GET("/export/save/:filename", routes_all_save)
	s9.GET("/export/index.html", history_view)

	s10.GET("/import/general", general_purpose_import)
	s10.POST("/import/general", general_purpose_import)
	s10.GET("/import/", gui_import)

	s11.GET("/basiccontract/amtdecidedlater/:decider", gui_adl_contract)
	s11.GET("/basiccontract/trade/:decider", gui_trade_contract)
	s11.GET("/basiccontract/auction/:decider", gui_auction_contract)
	s11.GET("/basiccontract/amtdecidedlatermkl/:decider/:min/:max/:left/:right", gui_adl_contract_merkle)
	s11.GET("/basiccontract/trademkl/:decider/:trade/:forward/:rollback", gui_trade_contract_merkle)
	s11.GET("/basiccontract/auction/:decider/:bidderid/:forward/:rollback", gui_auction_contract_merkle)
	s11.GET("/basiccontract/", gui_contract)

	s12.GET("/purse/view", purse_browse)
	s12.GET("/purse/generator", purse_generate_key)
	s12.GET("/purse/index.html", purse_browse)
	s12.GET("/purse/", purse_browse)

	srv := &http.Server{
		Handler:      r,
		WriteTimeout: 24 * time.Hour,
		ReadTimeout:  24 * time.Hour,
	}

	set_init_proxy_is_over()

	if u_config.init_proxy == "" {
		main_pub_server_serve(publn)
	}

	err := srv.Serve(ln)
	if err != nil {
		return
	}

}
