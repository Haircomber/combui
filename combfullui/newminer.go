package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/http2"
	"hash"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type Parsed_Block struct {
	height int
	hash   string
	pbh    string
	txes   []*Parsed_TX
}

type Parsed_TX struct {
	commit string
	tag    utxotag
}

// JSON Structs
type Block_Info struct {
	//	Bits string `json:"bits"`
	//	Chainwork string `json:"chainwork"`
	//	Conf int `json:"confirmations"`
	//	Diff float64 `json:"difficulty"`
	Hash   string `json:"hash"`
	Height int    `json:"height"`
	//	MTime float64 `json:"mediantime"`
	//	MRoot string `json:"merkleroot"`
	//	NTX int `json:"nTx"`
	//	NBH string `json:"nextblockhash"`
	//	Nonce float64 `json:"nonce"`
	PBH string `json:"previousblockhash"`
	//	Size int `json:"size"`
	//	SSize int `json:"strippedsize"`
	//	Time float64 `json:"time"`
	TX []TX_Info `json:"tx"`
	//	Version int `json:"version"`
	//	VHex string `json:"versionHex"`
	//	Weight int `json:"weight"`
}

type TX_Info struct {
	VOut []map[string]interface{} `json:"vout"`
}

type WaitForBlock struct {
	Hash   string `json:"hash"`
	Height uint64 `json:"height"`
}

type Chaintip struct {
	Height uint64 `json:"height"`
	Hash   string `json:"hash"`
	Status string `json:"status"`
}

var last_known_btc_height = -1
var last_known_btc_height_mutex sync.RWMutex

var run = true
var run_mutex sync.RWMutex

var mut sync.Mutex // locks both the count ints and the map
var results_map = make(map[int]*Parsed_Block)
var next_mine int
var next_download int
var max_block int
var http_client *http.Client
var wg sync.WaitGroup

// TODO: remove this exotic bullshit when debian does have a new golang (i.e. golang 1.21)
func DialTLSContext[T any](a *func(ctx context.Context, network, addr string, _ *T) (net.Conn, error)) {
	*a = DialTLSContext2[T]
}

// It's used like this: http2.DialTLSContext = DialTLSContext2
// But there is an error: implicitly instantiated function in assignment requires go1.21 or later
// So we workaround using the other function
func DialTLSContext2[T any](ctx context.Context, network, addr string, _ *T) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, network, addr)
}
func make_client2[T any](use_http2 bool) *http.Client {
	if use_http2 {
		var http2 *http2.Transport = &http2.Transport{
			AllowHTTP: true,
		}
		DialTLSContext(&http2.DialTLSContext)
		client := &http.Client{
			Transport: http2,
		}
		return client
	}
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 130,
		},
		Timeout: 24 * time.Hour,
	}
	return client
}
func make_client(use_http2 bool) *http.Client {
	return make_client2[interface{}](use_http2)
}

func check_run() bool {
	run_mutex.RLock()
	defer run_mutex.RUnlock()
	return run
}

func stop_run() {
	run_mutex.Lock()
	defer run_mutex.Unlock()
	run = false
}

func set_last_known_btc_height(h int) {
	last_known_btc_height_mutex.Lock()
	defer last_known_btc_height_mutex.Unlock()
	last_known_btc_height = h
}

func check_last_known_btc_height() int {
	last_known_btc_height_mutex.RLock()
	defer last_known_btc_height_mutex.RUnlock()
	return last_known_btc_height
}

// ~~~GENERAL FUNC~~~

func handle_reorg_direct(our_height, inc_target_height int) {

	log.Println("reorg handle begun;", inc_target_height)

	// Setup
	height := uint32(our_height)
	target_height := uint32(inc_target_height)

	commit_cache_mutex.Lock()
	commits_mutex.Lock()

	commit_currently_loaded = utxotag{height: target_height, txnum: 0, outnum: 0, commitnum: 0}

	log.Println("reorg: height, target_height:", height, target_height)

	type temp_commit struct {
		add [32]byte
		tag utxotag
	}
	var temp_commits_map = make(map[uint32][]temp_commit)

	CommitLvlDbBatchCleanupHeight(uint64(target_height)+1, uint64(height+1), func(key, val []byte) {
		if len(key) == 16 {
			tag := new_utxotag_from_leveldb(key)
			if tag.height > target_height {
				var add [32]byte
				copy(add[:], val)
				temp_commits_map[tag.height] = append(temp_commits_map[tag.height], temp_commit{add: add, tag: tag})
			}
		}
	})

	for h_height := height; h_height > target_height; h_height-- {

		// Remove from live map
		pop_block(uint64(h_height))

	}

	log.Println("reorg: total blocks to remove commits:", len(temp_commits_map))

	// Remove live commits
	for height > target_height {

		var specificCommits [][32]byte

		// Sort the commits in the arrays
		sort.Slice(temp_commits_map[height], func(i, j int) (less bool) {
			return utag_cmp(
				&temp_commits_map[height][i].tag,
				&temp_commits_map[height][j].tag) > 0
		})
		for _, temp_commit := range temp_commits_map[height] {
			add := temp_commit.add

			specificCommits = append(specificCommits, add)
		}

		reorgSpecificCommits(uint64(height), specificCommits)

		unwritten := false
		for _, temp_commit := range temp_commits_map[height] {
			add := temp_commit.add

			// If combase has...
			if _, ok := combbases[add]; ok {
				// ...unmine
				segments_coinbase_unmine(add, uint64(height))
				// Delete
				delete(combbases, add)
			}

			// Commits delete
			delete(commits, add)

			// decrease ram pruned counter
			if uint64(height) < u_config.prune_ram && commits_ram_pruned_count > 0 {
				commits_ram_pruned_count--
			}

			// uncompute checksum
			//CommitLvlDbUnCompute(add)

			// Used Keys?
			if enable_used_key_feature {
				used_key_commit_reorg(add, uint64(height))
			}

			// Merkle unmine
			merkle_unmine(add)

			// TX Leg Rollback
			tx_leg_rollback(add)

			// Unwritten
			unwritten = true

		}
		if unwritten && enable_used_key_feature {
			used_key_height_reorg(uint64(height))
		}

		height--
	}

	reorgCommits(uint64(target_height) + 1) // +1 is correct, TESTED
	log.Println("finished removing commits")
	resetgraph()

	temp_commits_map = nil
	commits_mutex.Unlock()
	commit_cache_mutex.Unlock()

	// reorg accelerator
	accelerator_reorg(uint64(target_height) + 0) // +0 is correct, TESTED


	for h_height := height; h_height >= target_height; h_height-=62500 {
		uncache_getrange(uint64(h_height))
	}
}

func get_block_info_for_height(height int, client *http.Client) (*Block_Info, error) {

	var hash string
	log.Println("Getting info for height:", height)

	// Only return an error when loop starts and run is false
	if !check_run() {
		log.Println("get_block_info: run = false")
		return nil, errors.New("run is false")
	}

	// Get hash and remove \n
	var result_json, call_err = make_bitcoin_call(client, "getblockhash", fmt.Sprint(height))
	if call_err != nil {
		log.Println("get_block_info: block hash call error:", call_err)
		return nil, call_err
	}
	hash = string(result_json)

	//log.Println("pulled block hash, attempting conv")

	//log.Println("hash conv success:", hash)

	// Get Block
	block_json, err := make_bitcoin_call(client, "getblock", hash+", "+"2")
	if err != nil {
		log.Println("get_block_info: block call error:", err)
		return nil, err
	}

	var block_info Block_Info
	err2 := json.Unmarshal(block_json, &block_info)
	if err2 != nil {
		log.Println(err2)
	}

	if `"`+block_info.Hash+`"` == hash && block_info.Height == height {
		return &block_info, nil
	} else {
		return nil, errRetriedNoBlock
	}
}

func parse_block(block_info *Block_Info) *Parsed_Block {
	output := &Parsed_Block{}

	// Metadata
	output.height = block_info.Height
	output.hash = block_info.Hash
	output.pbh = block_info.PBH

	// Commits

	var j = 0

	for x, tx := range block_info.TX {
		// Check all outputs for P2WSH
		for i, vout := range tx.VOut {

			// If it has a scriptPubKey
			if vout["scriptPubKey"] != nil {
				scriptPubKey := vout["scriptPubKey"].(map[string]interface{})

				// If it has type
				if scriptPubKey["type"] != nil {
					my_type := scriptPubKey["type"].(string)

					// If type is "witness_v0_scripthash"
					if my_type == "witness_v0_scripthash" {

						// Pull the hex
						if scriptPubKey["hex"] != nil {

							hex := fmt.Sprintf("%v", scriptPubKey["hex"])

							output.txes = append(output.txes, &Parsed_TX{
								commit: strings.ToUpper(hex[4:]),
								tag:    new_utxotag(output.height, j, x, i),
							})

							j++
						}
					}
				}
			}
		}
	}

	return output
}

func miner(parsed_block *Parsed_Block) (*Commit_Block_Metadata, error) {

	// Compare inc_block hash to stored previous block hash
	otb := our_top_block()
	cond := no_blocks()

	if !cond && otb.Hash != parsed_block.pbh {
		log.Println("hash mismatch while mining; ourtop.hash != inctop.prevhash:", otb.Hash, parsed_block.pbh)
		return nil, errors.New("hash mismatch")
	}

	var beforeBlockCommitsCummulative = commitsCount()
	var beforeBlockCryptofingerprint = database_aes

	CommitLvlDbNewBatch()

	// Format hash key
	hash_key := new_height_tag(uint64(parsed_block.height))

	// Mine the commits
	for _, tx_data := range parsed_block.txes {
		miner_mine_commit_pulled(tx_data.commit, tx_data.tag)
	}

	// Flush
	var fingerprint = miner_mine_end_of_block(new_flush_utxotag(uint64(parsed_block.height)), parsed_block.hash, func(cfp hash.Hash, winner [32]byte) string {

		var commit_fingerprint = cfp

		if commit_fingerprint == nil {
			commit_fingerprint = sha256.New()
		}

		commit_fingerprint.Write([]byte(parsed_block.hash))

		var sumbuf [32]byte
		var sum = sumbuf[0:0:32]

		sum = commit_fingerprint.Sum(sum)

		// Add the hash to the fingerprint
		fingerprint := fmt.Sprintf("%x", sum)

		// Add the count of the P2WSH
		var commitsCount, commitsCountCumulative string
		if !commits_format_v1 {
			commitsCount = fmt.Sprintf("%016x", len(parsed_block.txes))
		}
		if commits_format_v2 {
			commitsCountCumulative = fmt.Sprintf("%016x", beforeBlockCommitsCummulative)
		}

		// Store the hash last, just in case.
		// Not just hash now, store hash CAT fingerprint CAT p2wsh count
		log.Println("m", parsed_block.hash)

		if commits_format_v2 {
			var pbhash = hex2byte32([]byte(parsed_block.hash))
			var finger = hex2byte32([]byte(fingerprint))
			var block_data []byte
			block_data = append(block_data, pbhash[:]...)
			block_data = append(block_data, winner[:]...)
			block_data = append(block_data, finger[:]...)
			block_data = append(block_data, beforeBlockCryptofingerprint[:]...)
			block_data = append(block_data, []byte(commitsCount)...)
			block_data = append(block_data, []byte(commitsCountCumulative)...)

			CommitLvlBatchPut(hash_key, block_data)
		} else {
			CommitLvlBatchPut(hash_key, []byte(parsed_block.hash+fingerprint+commitsCount))
		}

		CommitLvlDbWriteBatch()

		accelerator_next(parsed_block.hash, uint64(len(parsed_block.txes)))

		return fingerprint
	})
	// Build the block metadata
	metadata := &Commit_Block_Metadata{
		Height:      uint64(parsed_block.height),
		Hash:        parsed_block.hash,
		Fingerprint: fingerprint,
	}

	uncache_getrange(uint64(parsed_block.height))

	log.Println("mined:", parsed_block.height)
	return metadata, nil
}

func downloader(height int, my_client *http.Client) {
	defer wg.Done()

	for check_run() && height != 0 {

		// Download the block info
		block_info, err := get_block_info_for_height(height, my_client)
		if err != nil || block_info == nil {
			if err != errNilJson && err != errRetriedNoBlock && err != errNoBtc {
				fmt.Println("mine_blocks: dl err - get_block_info for height:", err.Error())
			}
			stop_run()
			return
		}

		parsed_block := parse_block(block_info)

		if parsed_block.height != height {
			log.Println("wrong height", parsed_block.height, height)
			stop_run()
			return
		}

		mut.Lock()
		// If my height needs to be mined, mine it then break
		if parsed_block.height == next_mine {
			log.Println("PARSER (MINE ME!):", parsed_block.height, next_mine)

			// Else if the next mine height can be mined, mine it then check again
		} else if blk, ok := results_map[next_mine]; ok {
			log.Println("PARSER (MINE NEXT!):", parsed_block.height, next_mine)
			delete(results_map, next_mine)
			results_map[parsed_block.height] = parsed_block
			parsed_block = blk

			// Else if can't do shit, deposit payload then break
		} else {
			log.Println("PARSER (DROP OFF!):", parsed_block.height, next_mine)
			results_map[parsed_block.height] = parsed_block
			parsed_block = nil
		}
		next_download++
		if next_download <= max_block {
			height = next_download
		} else {
			height = 0
		}

		mut.Unlock()

		for check_run() && parsed_block != nil {

			log.Println("APPLYING ANOTHER:", parsed_block.height)

			block_metadata, err := miner(parsed_block)
			if err != nil {
				log.Println("Miner err:", err)
				stop_run()
				return
			}

			put_block(uint64(parsed_block.height), block_metadata)

			mut.Lock()
			next_mine++
			parsed_block = results_map[next_mine]
			delete(results_map, next_mine)
			mut.Unlock()

			mainPageInvalidate()
		}
	}
}

func mine_blocks(start, finish int) {

	run = true
	results_map = make(map[int]*Parsed_Block)

	log.Println("mine_blocks: began:", start, finish)

	next_mine = start
	next_download = start + 20
	max_block = finish

	if next_download > max_block {
		next_download = max_block
	}

	// not making a copy makes the for loop below race with the goroutines that we spawn
	var next_download_copy = next_download

	log.Println("mine_blocks: vars set:", next_mine, next_download, max_block)

	// While i, base start, is less than next download, go and increment
	log.Println(next_mine, next_download)

	// we must use a copy of next_download, because the original
	// could've been modified by the goroutines that we spawn
	for i := next_mine; i <= next_download_copy; i++ {
		wg.Add(1)
		go downloader(i, http_client)
		log.Println("mine_blocks: added downloader", i)
	}

	log.Println("mine_blocks: downloaders added")

	wg.Wait()

	if !check_run() {
		results_map = nil
		return
	}
	if get_connected() == false {
		results_map = nil
		return
	}

	log.Println("final results:", len(results_map))

	log.Println("profile:", next_mine, max_block, start, finish)

	for i := next_mine; i <= max_block; i++ {

		var parsed_block = results_map[i]

		log.Println("APPLYING FINAL:", parsed_block.height)
		block_metadata, err := miner(parsed_block)
		if err != nil {
			break
		}

		put_block(uint64(parsed_block.height), block_metadata)

		mainPageInvalidate()

	}

	results_map = nil
}

// this needs to be synced with leveldb
var blocks_mutex sync.RWMutex
var maximum_block uint64

func put_block(height uint64, blk *Commit_Block_Metadata) {
	blocks_mutex.Lock()
	defer blocks_mutex.Unlock()
	if maximum_block != 0 {
		if height != maximum_block+1 {
			log.Println("must put the next block height (height, max :", height, maximum_block)
			panic("must put the next block height")
		}
	}
	maximum_block = height
}

func pop_block(height uint64) {
	blocks_mutex.Lock()
	defer blocks_mutex.Unlock()

	if height != maximum_block {
		log.Println("Must pop top block")
		panic("must pop top block")
	}

	maximum_block--
}

func our_block_at_height(height uint64) *Commit_Block_Metadata {
	var hash, _ = CommitLvlDbReadBlock(height)

	if hash != [64]byte{} {
		return &Commit_Block_Metadata{
			Hash:   string(hash[:]),
			Height: height,
		}
	}
	return &Commit_Block_Metadata{}
}

func no_blocks() bool {
	blocks_mutex.RLock()
	defer blocks_mutex.RUnlock()

	return maximum_block == 0
}
func our_top_block() *Commit_Block_Metadata {
	blocks_mutex.RLock()
	var maxblock = maximum_block
	blocks_mutex.RUnlock()

	var hash, _ = CommitLvlDbReadBlock(maxblock)

	if hash != [64]byte{} {
		return &Commit_Block_Metadata{
			Hash:   string(hash[:]),
			Height: maxblock,
		}
	}
	return &Commit_Block_Metadata{}
}

func chain_size() (ret uint64) {
	blocks_mutex.RLock()
	defer blocks_mutex.RUnlock()

	return maximum_block
}

func find_reorg2(client *http.Client, hc_height uint64, u_config UserConfig) (uint64, error) {

	log.Println("find_reorg: started")

	// Set a stop limit
	var lowest uint64 = 481823
	if u_config.regtest || u_config.testnet4 {
		lowest = 0
	}

	for hc_height > lowest {

		var ours = our_block_at_height(hc_height).Hash

		result_json, err := make_bitcoin_call(client, "getblockhash", fmt.Sprint(hc_height))
		if err != nil {
			return 0, err
		}
		result := strings.Trim(string(result_json), `"`)

		if result == ours {
			return hc_height, nil
		}

		hc_height--
	}
	return lowest, nil
}

func new_miner_start() {

	// make the http client for main calls
	http_client = make_client(true)

	log.Println("MINER START")
outer:
	for {
		mainPageInvalidate()
		time.Sleep(time.Second)

		chainTipData, err2 := make_bitcoin_call(http_client, "getchaintips", "")
		if err2 != nil {
			set_connected(false)

			continue
		}
		set_connected(true)

		var chainTips []Chaintip
		err3 := json.Unmarshal(chainTipData, &chainTips)
		if err3 != nil {
			set_connected(false)
			continue
		}

		var activeHash string
		var activeHeight, headersOnlyHeight uint64
		for _, tip := range chainTips {
			switch tip.Status {
			case "active":
				activeHeight = tip.Height
				activeHash = tip.Hash
				if headersOnlyHeight < tip.Height {
					headersOnlyHeight = tip.Height
				}
			case "headers-only", "valid-headers":
				if headersOnlyHeight < tip.Height {
					headersOnlyHeight = tip.Height
				}
			}
		}

		if activeHeight != headersOnlyHeight {
			//fmt.Println("activeHeight:", activeHeight, "headersOnlyHeight:", headersOnlyHeight)
			continue
		}

		var hashHeightMutex sync.Mutex
		hashHeightMutex.Lock()
		hash1 := activeHash
		height1 := uint64(activeHeight)
		hashHeightMutex.Unlock()

		// We are synced, wait for block
		if our_top_block().Hash == activeHash {
			if dualWaiters(func(i, j int) error {
				data, err1 := make_bitcoin_call(http_client, "waitfornewblock", fmt.Sprintf("%d", i))
				if err1 != nil {
					log.Println(err1.Error())
					return err1
				}
				var wfb WaitForBlock
				err4 := json.Unmarshal(data, &wfb)
				if err4 != nil {
					log.Println(err4.Error())
					return err4
				}

				if wfb.Hash != activeHash {
					hashHeightMutex.Lock()
					hash1 = wfb.Hash
					height1 = wfb.Height
					hashHeightMutex.Unlock()
					return fmt.Errorf("new_block")
				}
				return nil
			}, 4).Error() != "new_block" {
				continue
			}
		}

		hashHeightMutex.Lock()
		hash2 := hash1
		height2 := height1
		hashHeightMutex.Unlock()

		set_last_known_btc_height(int(height2))
		mainPageInvalidate()

		// Get Haircomb's highest known block
		comb_height := uint64(chain_size())

		// If not regtest, make the current height the first COMB block
		if !(u_config.regtest || u_config.testnet4) && comb_height < 481823 {
			comb_height = 481823
		}

		var ourhash_at_btc_height = our_block_at_height(height2).Hash

		log.Println(hash2, "==", ourhash_at_btc_height)

		if ourhash_at_btc_height == hash2 {
			log.Println("case 1")
			log.Println("reorg type \"direct\"")
			handle_reorg_direct(int(comb_height), int(height2))
			continue outer
		}

		log.Println(height2, "<=", comb_height)

		if height2 <= comb_height {
			target_height, err := find_reorg2(http_client, height2, u_config)
			if err != nil {
				log.Println("mine_loop: find reorg err:", err)
				continue outer
			}
			log.Println("case 2")
			log.Println("reorg type \"direct\"")
			handle_reorg_direct(int(comb_height), int(target_height))
			continue outer
		}

		result_json, err := make_bitcoin_call(http_client, "getblockhash", fmt.Sprint(comb_height))
		if err != nil {
			log.Println("get_block_info: block hash call error")
			continue outer
		}

		var btc_hash_at_our_height = strings.Trim(string(result_json), `"`)

		var our_highest_block = our_top_block()

		log.Println("?", btc_hash_at_our_height, our_highest_block.Hash)

		if btc_hash_at_our_height == our_highest_block.Hash || no_blocks() {

		} else {
			var target_height uint64

			target_height, err = find_reorg2(http_client, comb_height, u_config)
			if err != nil {
				log.Println("mine_loop: find reorg err:", err)
				continue outer
			}
			log.Println("case 3")
			log.Println("reorg type \"direct\"")
			handle_reorg_direct(int(comb_height), int(target_height))
			continue outer
		}
		log.Println("case 4")
		// finally fast forward
		mine_blocks(int(comb_height)+1, int(height2))
		continue outer
	}

}
