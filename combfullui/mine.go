package main

import (
	"crypto/sha256"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"hash"
	"log"
	"net/http"
	"sync"
)

// This file oversees the portion of mining process that is processing incoming commitments before they are stored to the DB or Memory

var commits_mutex sync.RWMutex

var commits map[[32]byte]utxotag
var combbases map[[32]byte]uint64

var commit_cache_mutex sync.Mutex

var commit_currently_loaded utxotag
var commits_ram_pruned_count uint64
var commit_cache [][32]byte
var commit_tag_cache []utxotag

func init() {
	commits = make(map[[32]byte]utxotag)
	combbases = make(map[[32]byte]uint64)
}

func miner_unmine_pending_commits() {
	commit_cache_mutex.Lock()
	commit_cache = nil
	commit_tag_cache = nil
	commit_cache_mutex.Unlock()
}

func height_view(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Used to tell BTC what height to start at. Done after 6 flush validation
	// Replacing this with a 99999999 response to make this code compatible with the modded BTC
	fmt.Fprintf(w, "99999999")

}

func miner_mine_commit_pulled(commitinput string, utxotagval utxotag) hash.Hash {
	// This gets run when new blocks come in.

	// this is not address related, so we check hex is upper on all networks

	err1 := checkHEX32upper(commitinput)
	if err1 != nil {
		log.Println("error mining by using commit: ", err1)
		return nil
	}

	return miner_mine_commit_internal(commitinput, utxotagval)
}

var mine_cycles = 0

func miner_mine_end_of_block(tag utxotag, blkHash string, callback func(sum hash.Hash, winner [32]byte) string) string {

	var sum hash.Hash
	var winner, previous [32]byte
	var winner_levelkey []byte
	var ret string

	commit_cache_mutex.Lock()
	commits_mutex.Lock()

	posttag(&commit_currently_loaded, uint64(tag.height))

	if len(commit_cache) > 0 {

		for i := range commit_cache {

			if _, ok5 := commitsCheck(commit_cache[i], uint64(tag.height)); !ok5 {
				var basetag = commit_tag_cache[i]
				var btag = basetag

				var bheight = uint64(btag.height)

				segments_coinbase_mine(commit_cache[i], bheight)
				combbases[commit_cache[i]] = bheight

				winner = commit_cache[i]
				previous = winner
				winner_levelkey = utxotag_to_leveldb(commit_tag_cache[i])

				break
			}
		}

	} else {
		goto adios
	}

	sum = sha256.New()

	for key, val := range commit_cache {
		// This is the new prexisting commit check.
		if _, ok5 := commitsCheck(val, uint64(tag.height)); ok5 {

		} else {

			//CommitDbWrite(val, hex2byte8(serializeutxotag(commit_tag_cache[key])))
			if initial_writeback_over {

				var levelkey = utxotag_to_leveldb(commit_tag_cache[key])

				sum.Write(levelkey)
				sum.Write(val[:])

				// Write commitment with link to the prev commitment
				CommitLvlDbWrite(val, hex2byte32([]byte(blkHash)), previous, uint64(tag.height), levelkey, false)
				previous = val
			}
			commits[val] = commit_tag_cache[key]
			mine_cycles++
		}
	}

	for iter, key := range commit_cache {
		var tagval = commit_tag_cache[iter]

		merkle_mine(key)
		tx_mine(key, tagval)
	}
	if len(commit_cache) > 0 {

		if winner_levelkey != nil {
			// Write the commitment of the combbase again, with link to the last commitment
			CommitLvlDbWrite(winner, hex2byte32([]byte(blkHash)), previous, uint64(tag.height), winner_levelkey, true)
		}

		// do ram pruning
		if uint64(tag.height) < u_config.prune_ram {
			commits_ram_pruned_count += uint64(len(commits))
			commits = make(map[[32]byte]utxotag)
		}

		// now, fill the index
		writeCommits(uint64(tag.height))
	}

	commit_cache = nil
	commit_tag_cache = nil
adios:

	if callback != nil {
		ret = callback(sum, winner)
	}

	commits_mutex.Unlock()
	commit_cache_mutex.Unlock()

	return ret
}

func miner_mine_commit_internal(commitinput string, tag utxotag) (sum hash.Hash) {

	// This is run to input commits into program memory, either from the DB or when new blocks come in.
	var rawcommit = hex2byte32([]byte(commitinput))

	commit_cache_mutex.Lock()

	commit_cache = append(commit_cache, rawcommit)
	//mine_cycles++
	commit_tag_cache = append(commit_tag_cache, tag)

	commits_mutex.Lock()
	commit_currently_loaded = tag
	commits_mutex.Unlock()
	commit_cache_mutex.Unlock()

	return nil
}

func tx_leg_rollback(key [32]byte) {
	txleg_mutex.RLock()

	txlegs_each_leg_target(key, func(tx *[32]byte) bool {

		segments_transaction_mutex.Lock()
		var txdata = segments_transaction_data[*tx]
		var actuallyfrom = txdata[21]

		var ok = false

		for i := uint(0); i < 21; i++ {
			if commit(txdata[i][0:]) == key {

				ok = true
				break
			}
		}
		if ok {
			segments_transaction_untrickle(nil, actuallyfrom, 0xffffffffffffffff)

			delete(segments_transaction_next, actuallyfrom)

		}

		segments_transaction_mutex.Unlock()

		return true
	})

	txleg_mutex.RUnlock()
}
