package main

import (
	"crypto/aes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"hash"
	"log"
	"sync"
)

type Commit_Block_Metadata struct {
	// Put header info here later
	Height      uint64 // the block height
	Hash        string // the block hash
	Fingerprint string // the finger print of all the commits CAT together and run through SHA256
}

const COMMITS_LVLDB_PATH = "commits"
const COMMITS_LVLDB_PATH_REGTEST = "commits_regtest"
const COMMITS_LVLDB_PATH_TESTNET = "commits_testnet"
const COMMITS_LVLDB_PATH_TESTNET4 = "commits_testnet4"

var commitsdb_mutex sync.Mutex
var commitsdb *leveldb.DB
var commitsbatch *leveldb.Batch
var commits_format_v1 bool
var commits_format_v2 bool
var commits_format_v2b bool // format V2 B (extended header)
var commits_format_v2b_is_unprune bool
var commits_format_v2b_is_duplex_encap bool
var commits_pruned_height uint64 // equal to u_config.prune_disk-1 (when the storage was pruned), equal to v2_hdr[32:40]
var commits_pruned_deleted uint64
var commits_pruned_combbases uint64
var prefixkernels = byte(1) // prefix kernels 9 by default
var prefixlinks = byte(3)   // prefix links, 3 by default
var commits_db_backend byte

type CommitLvlDbIterator struct {
	LvlDbIter    iterator.Iterator
	PebbleDbIter PebbleIterator
}

func (c *CommitLvlDbIterator) First() {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		c.PebbleDbIter.First()
		return
	}
	c.LvlDbIter.First()
	return
}

func (c *CommitLvlDbIterator) Valid() bool {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		return c.PebbleDbIter.Valid()
	}
	return c.LvlDbIter.Valid()
}

func (c *CommitLvlDbIterator) Next() {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		c.PebbleDbIter.Next()
		return
	}
	c.LvlDbIter.Next()
	return
}

func (c *CommitLvlDbIterator) Key() []byte {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		return c.PebbleDbIter.Key()
	}
	return c.LvlDbIter.Key()
}

func (c *CommitLvlDbIterator) Value() []byte {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		return c.PebbleDbIter.Value()
	}
	return c.LvlDbIter.Value()
}

func (c *CommitLvlDbIterator) Release() error {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		return c.PebbleDbIter.Close()
	}
	c.LvlDbIter.Release()
	return c.LvlDbIter.Error()
}

func CommitLvlDbNewIterator(begin, end []byte) (*CommitLvlDbIterator, error) {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		return CommitPebbleDbNewIterator(begin, end)
	}
	iter := commitsdb.NewIterator(&util.Range{Start: begin, Limit: end}, nil)
	return &CommitLvlDbIterator{
		LvlDbIter: iter,
	}, nil
}

func CommitLvlDbNewBatch() {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		CommitPebbleDbNewBatch()
		return
	}
	commitsbatch = new(leveldb.Batch)
}

func CommitLvlDbWriteBatch() {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		CommitPebbleDbWriteBatch()
		return
	}
	err := commitsdb.Write(commitsbatch, &opt.WriteOptions{
		Sync: true,
	})
	if err != nil {
		log.Fatal("DB write batch error: ", err)
	}
	commitsbatch = nil
}

func CommitLvlDbOpen() {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		CommitPebbleDbOpen()
		return
	}
	path := COMMITS_LVLDB_PATH
	if u_config.testnet {
		path = COMMITS_LVLDB_PATH_TESTNET
	}
	if u_config.testnet4 {
		path = COMMITS_LVLDB_PATH_TESTNET4
	}
	if u_config.regtest {
		path = COMMITS_LVLDB_PATH_REGTEST
	}
	db, err := leveldb.OpenFile(u_config.db_dir+path, &opt.Options{
		Compression: opt.NoCompression,
		Filter:      filter.NewBloomFilter(10),
	})
	if err != nil {
		log.Fatal("DB open error: ", err)
	}
	commitsdb_mutex.Lock()
	commitsdb = db
}

func CommitLvlDbGet(key []byte, _ *struct{}) (out []byte, err error) {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		return CommitPebbleDbGet(key, nil)
	}
	out, err = commitsdb.Get(key, nil)
	if err != nil {
		return nil, err
	}
	return
}
func CommitLvlDbPut(key, value []byte, _ *struct{}) (err error) {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		return CommitPebbleDbPut(key, value, nil)
	}
	err = commitsdb.Put(key, value, &opt.WriteOptions{
		Sync: true,
	})
	return
}
func CommitLvlBatchPut(key, value []byte) (err error) {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		return CommitPebbleBatchPut(key, value)
	}
	commitsbatch.Put(key, value)
	return
}
func CommitLvlBatchDelete(key []byte) (err error) {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		return CommitPebbleBatchDelete(key)
	}
	commitsbatch.Delete(key)
	return
}

func CommitLvlDbClose() {
	switch commits_db_backend {
	case CommitsDbBackendPebble:
		CommitPebbleDbClose()
		return
	}
	if commitsdb == nil {
		return
	}
	commitsdb.Close()
	commitsdb = nil
	commitsdb_mutex.Unlock()
}

func keyLimit(b []byte) []byte {
	end := make([]byte, len(b))
	copy(end, b)
	for i := len(end) - 1; i >= 0; i-- {
		end[i] = end[i] + 1
		if end[i] != 0 {
			return end[:i+1]
		}
	}
	return nil // no upper-bound
}

func CommitLvlDbBatchCleanupHeight(h1, h2 uint64, cb func([]byte, []byte)) {

	// delete blocks in the batch
	CommitLvlDbNewBatch()

	CommitLvlDbReorgV2Header(h1)

	iter, err := CommitLvlDbNewIterator(new_height_tag(h1), new_height_tag(h2))
	if err != nil {
		log.Fatal("DB iter error: ", err)
	}
	for iter.First(); iter.Valid(); iter.Next() {
		key, value := iter.Key(), iter.Value()
		cb(key, value)
		if len(key) == 8 || len(key) == 16 {
			CommitLvlBatchDelete(key)
		}
	}
	CommitLvlDbWriteBatch()
	err = iter.Release()
	if err != nil {
		log.Fatal("Hash Erase error: ", err)
	}
}

func CommitLvlDbReadDuplex(hash [32]byte, maxHeight uint64) (tag utxotag, ok bool) {
	if !commits_format_v2 {
		return
	}
	if commits_pruned_height <= 1 && u_config.prune_ram == 0 {
		return
	}
	var value, err = CommitLvlDbGet(hash[:], nil)
	if (err == leveldb.ErrNotFound) || (err == ErrPebbleNotFound) {
		return
	}
	if err != nil {
		log.Fatal("DB read error: ", err)
	}
	if len(value) < 48 {
		return
	}
	var utag = new_utxotag_from_leveldb(value)
	if uint64(utag.height) >= maxHeight {
		return
	}
	var blockhash [32]byte
	copy(blockhash[:], value[16:48])

	if !CommitLvlDbCheckHeight(uint64(utag.height), blockhash) {
		return
	}
	return utag, true
}

func CommitLvlDbReadPrefix(height uint64, ordernum uint32) (hash [32]byte, utag [16]byte, ok bool) {
	var buf [12]byte

	binary.BigEndian.PutUint64(buf[0:8], height)
	binary.BigEndian.PutUint32(buf[8:12], ordernum)

	iter, err := CommitLvlDbNewIterator(
		buf[:],
		keyLimit(buf[:]),
	)
	if err != nil {
		log.Fatal("DB iter error: ", err)
	}
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if len(key) != 16 {
			continue
		}
		value := iter.Value()

		var net byte
		if len(value) >= 33 {
			net = value[32]
		}

		copy(utag[:], key)
		hash = midhash(value, net)
		ok = true
		break
	}
	err = iter.Release()
	if err != nil {
		log.Fatal("DB read error: ", err)
	}
	return hash, utag, ok
}

func CommitLvlDbReadBlock(height uint64) (hash [64]byte, p2wshcount [16]byte) {
	var buf [8]byte

	binary.BigEndian.PutUint64(buf[0:8], height)

	val, err := CommitLvlDbGet(buf[0:8], nil)
	if (err == leveldb.ErrNotFound) || (err == ErrPebbleNotFound) {
		return
	}
	if err != nil {
		log.Fatal("DB read error: ", err)
	}
	if commits_format_v2 {
		var hsh = fmt.Sprintf("%x", val[0:32])
		copy(hash[0:64], []byte(hsh))
		copy(p2wshcount[0:16], val[128:144])
	} else {
		copy(hash[0:64], val[0:64])
	}
	return
}

func CommitLvlDbDumpHeight(height uint64) (result map[string]interface{}, commitments map[string]utxotagJson) {

	var is_commit, should_be_commit bool
	var block [32]byte
	var base [32]byte

	var buf [8]byte
	result = make(map[string]interface{})
	commitments = make(map[string]utxotagJson)

	binary.BigEndian.PutUint64(buf[0:8], height)

	iter, err := CommitLvlDbNewIterator(
		buf[:],
		keyLimit(buf[:]),
	)
	if err != nil {
		log.Fatal("DB iter error: ", err)
	}
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()
		if len(key) == 8 && len(value) == 160 {
			copy(block[:], value[0:32])
			copy(base[:], value[32:64])
			should_be_commit = base != [32]byte{}
			result["BlockHash"] = fmt.Sprintf("%x", value[0:32])
			result["CombBase"] = fmt.Sprintf("%x", value[32:64])
			result["BlockShaSum"] = fmt.Sprintf("%x", value[64:96])
			result["Fingerprint"] = fmt.Sprintf("%x", value[96:128])
			var p2wshcount = hex2byte8(value[128:144])
			var commits_cummulative = hex2byte8(value[144:160])
			result["P2wshCount"] = binary.BigEndian.Uint64(p2wshcount[:])
			result["CommitsCummulative"] = binary.BigEndian.Uint64(commits_cummulative[:])
		} else if len(key) == 16 && len(value) == 32 {
			is_commit = true
			commitments[fmt.Sprintf("%x", value)] = utxotag_to_json(new_utxotag_from_leveldb(key))
		} else if len(key) == 32 && len(value) >= 48 {
			is_commit = true
			commitments[fmt.Sprintf("%x", key)] = utxotag_to_json(new_utxotag_from_leveldb(value[0:16]))
		} else {
			result[fmt.Sprintf("%x", key)] = fmt.Sprintf("%x", value)
		}
	}
	err = iter.Release()
	if err != nil {
		log.Fatal("DB read error: ", err)
	}

	if !is_commit && should_be_commit {
		CommitLvlDbIteratePrunedBlock(base, block, height, func(comm [32]byte, tag utxotag) {
			commitments[fmt.Sprintf("%x", comm)] = utxotag_to_json(tag)
		})
	}

	return result, commitments
}
func CommitLvlDbPruned() bool {
	if !commits_format_v2 {
		return false
	}
	if commits_pruned_height <= 1 && u_config.prune_ram == 0 {
		return false
	}
	return true
}
func CommitLvlDbReadDuplexPrefix(prefix [16]byte, bytes, reduce byte, hashes map[[32]byte]utxotag) {
	if !commits_format_v2 {
		return
	}
	if commits_pruned_height <= 1 && u_config.prune_ram == 0 {
		return
	}
	if bytes > 16 {
		bytes = 16
	}
	bytes -= reduce
	var sli = prefix[:bytes]

	if prefixkernels != 0 {
		for i := 0; i < 7; i++ {
			if (prefixkernels & (1 << i)) == 0 {
				continue
			}
			if bytes < byte(9+i) {
				continue
			}

			_, err := CommitLvlDbGet(prefix[:9+i], nil)
			if err != nil {
				if (err == leveldb.ErrNotFound) || (err == ErrPebbleNotFound) {
					return
				}
				log.Fatal("DB read error: ", err)
			}
		}
	}

	iter, err := CommitLvlDbNewIterator(
		sli,
		keyLimit(sli),
	)
	if err != nil {
		log.Fatal("DB iter error: ", err)
	}
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()
		if len(key) != 32 {
			continue
		}
		if len(value) < 48 {
			continue
		}
		var utag = new_utxotag_from_leveldb(value)

		if uint64(utag.height) >= u_config.prune_ram {
			continue
		}
		var hash [32]byte
		copy(hash[:], key[:])

		var blockhash [32]byte
		copy(blockhash[:], value[16:48])
		if CommitLvlDbCheckHeight(uint64(utag.height), blockhash) {
			hashes[hash] = utag
		}
	}
	err = iter.Release()
	if err != nil {
		log.Fatal("DB read error: ", err)
	}
}

func CommitLvlDbCheckHeight(height uint64, blockhash [32]byte) (is_ok bool) {
	var buf [8]byte

	binary.BigEndian.PutUint64(buf[0:8], height)

	var val, err = CommitLvlDbGet(buf[:], nil)
	if err != nil {
		if (err == leveldb.ErrNotFound) || (err == ErrPebbleNotFound) {
			return false
		}
		log.Fatal("DB read error: ", err)
	}
	var blkhash [32]byte
	copy(blkhash[:], val)

	return blkhash == blockhash
}

func CommitLvlDbReadHeight(height uint64) (hashes [][32]byte) {
	var buf [8]byte

	binary.BigEndian.PutUint64(buf[0:8], height)

	iter, err := CommitLvlDbNewIterator(
		buf[:],
		keyLimit(buf[:]),
	)
	if err != nil {
		log.Fatal("DB iter error: ", err)
	}
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if len(key) != 16 {
			continue
		}
		value := iter.Value()

		var net byte
		if len(value) >= 33 {
			net = value[32]
		}

		hashes = append(hashes, midhash(value, net))
	}
	err = iter.Release()
	if err != nil {
		log.Fatal("DB read error: ", err)
	}
	return hashes
}

func CommitLvlDbWrite(address, blockhash, previous [32]byte, height uint64, tag []byte, again bool) {
	if !initial_writeback_over {
		return
	}

	//if height < u_config.prune_disk (we can't do this now because of the quirk - the toppest block cannot be pruned)
	// Convert to general bytes([:]) and store
	CommitLvlBatchPut(tag[:], address[:])

	// I need to write duplex when pruned ram underwater, so I can dedup commits on disk
	if height < u_config.prune_ram {
		CommitLvlDbWriteDuplex(address, blockhash, previous, new_utxotag_from_leveldb(tag))
	}

	if again {
		return
	}

	// begin aes

	var aestmp [32]byte

	var aes, err5 = aes.NewCipher(address[0:])
	if err5 != nil {
		log.Fatal(err5)
	}

	aes.Encrypt(aestmp[0:16], database_aes[0:16])
	aes.Decrypt(aestmp[16:32], database_aes[16:32])

	for i := 8; i < 16; i++ {
		aestmp[i], aestmp[8+i] = aestmp[8+i], aestmp[i]
	}

	aes.Encrypt(database_aes[0:16], aestmp[0:16])
	aes.Decrypt(database_aes[16:32], aestmp[16:32])

	// end aes
}

func deserialize_metadata(height int, inc_val []byte) *Commit_Block_Metadata {

	//Hash([:64]) CAT Fingerprint[64:]
	val := string(inc_val)

	if commits_format_v2 {
		return &Commit_Block_Metadata{
			Height:      uint64(height),
			Hash:        fmt.Sprintf("%x", val[0:32]),
			Fingerprint: fmt.Sprintf("%x", val[64:96]),
		}
	}

	return &Commit_Block_Metadata{
		Height:      uint64(height),
		Hash:        val[0:64],
		Fingerprint: val[64:128],
	}

}

type commitLvlOngoingLoad struct {
	deletor bool
	meta    *Commit_Block_Metadata
	iscomm  bool
}

func (l *commitLvlOngoingLoad) commitLvlHandleBlock(tag_height uint64, val []byte, load_aes_lvl [32]byte, sum hash.Hash) {
	var chk [32]byte

	if l.meta != nil {

		if commits_format_v2 && tag_height >= commits_pruned_height {
			var aes [32]byte
			copy(aes[:], val[96:128])
			if aes != load_aes_lvl {

				fmt.Printf("%x != %x (%d %d)\n", aes, load_aes_lvl, tag_height, commits_pruned_height)
				panic("aes fingerprint doesnt check out")
			}

		}

		sum.Write([]byte(l.meta.Hash))
		sum.Sum(chk[0:0:32])

		if uint64(l.meta.Height) > commits_pruned_height || !commits_format_v2 {
			if fmt.Sprintf("%x", chk) != l.meta.Fingerprint {

				l.deletor = true

				CommitLvlDbNewBatch()

			}
		}

		//log.Printf("%d, previous sum:%x, next_sum: %s\n", l.meta.Height, chk, l.meta.Fingerprint)

		sum.Reset()

		if !l.deletor {

			// Mine the flush, important, for previous block
			miner_mine_end_of_block(new_flush_utxotag(uint64(l.meta.Height)), l.meta.Hash, nil)
			put_block(uint64(l.meta.Height), l.meta)

		} else {
			miner_unmine_pending_commits()

			log.Println("need deletion:", l.meta.Height)

			iter, err := CommitLvlDbNewIterator(
				new_height_tag(uint64(l.meta.Height)),
				new_height_tag(uint64(l.meta.Height+1)),
			)
			if err != nil {
				log.Fatal("DB iter error: ", err)
			}
			for iter.First(); iter.Valid(); iter.Next() {
				if len(iter.Key()) == 8 || len(iter.Key()) == 16 {
					CommitLvlBatchDelete(iter.Key())
				}
			}
			err = iter.Release()
			if err != nil {
				log.Fatal("Hash Erase error: ", err)
			}

		}
	}

	if val != nil {
		l.meta = deserialize_metadata(int(tag_height), val)

	}
	if commits_pruned_height <= 1 {
		return
	}
	if tag_height == 0 {
		return
	}

	if commits_format_v2 && val != nil && tag_height <= commits_pruned_height {
		var claim [32]byte
		copy(claim[:], val[32:64])

		if claim != [32]byte{} {

			// lookup
			var info, ok = CommitLvlDbReadDuplex(claim, tag_height+1)

			if ok {

				// lookup in normal commits (was it really pruned?)
				var buf = utxotag_to_leveldb(info)
				_, err := CommitLvlDbGet(buf[0:16], nil)

				if (err == leveldb.ErrNotFound) || (err == ErrPebbleNotFound) {

					commits_pruned_combbases++
					// create combbases for pruned blocks
					miner_mine_commit_pulled(fmt.Sprintf("%X", claim), info)

				}

			}
		}
	}
}

func CommitLvlDbReorgV2Header(height uint64) {
	if !commits_format_v2 {
		return
	}
	var keylow [8]byte
	var key [8]byte
	binary.BigEndian.PutUint64(keylow[:], height-1)
	binary.BigEndian.PutUint64(key[:], height)
	var v2_old_hdr, err2 = CommitLvlDbGet([]byte{0}, nil)
	if err2 != nil {
		if !((err2 == leveldb.ErrNotFound) || (err2 == ErrPebbleNotFound)) {
			log.Fatal("Commitdb ReorgV2Header Error: ", err2)
		}
		return
	}
	var pruned [32]byte
	var cummulative uint64

	var v2_block_higher, err3 = CommitLvlDbGet(key[:], nil)
	if err3 != nil {
		if !((err3 == leveldb.ErrNotFound) || (err3 == ErrPebbleNotFound)) {
			log.Fatal("Commitdb ReorgV2Header Error: ", err3)
		}
	} else {
		copy(database_aes[:], v2_block_higher[96:128])
		cummulative = CommitLvlDbBlockGetCummulative(v2_block_higher)
	}

	var v2_block_lower, err4 = CommitLvlDbGet(keylow[:], nil)
	if err4 != nil {
		if !((err4 == leveldb.ErrNotFound) || (err4 == ErrPebbleNotFound)) {
			log.Fatal("Commitdb ReorgV2Header Error: ", err4)
		}
	} else {
		copy(pruned[:], v2_block_lower[96:128])
	}

	if (height + 1) <= commits_pruned_height {
		copy(v2_old_hdr[0:32], pruned[:])
		copy(v2_old_hdr[32:40], keylow[:])
		binary.BigEndian.PutUint64(v2_old_hdr[41:49], cummulative)
		CommitLvlBatchPut([]byte{0}, v2_old_hdr[:])

		commits_pruned_deleted += commitsCount() - cummulative
	}
}

// this function is pretty dangerous as it doesn't check the utxotag heights
// within batch due to leveldb
// and thus it may overwrite
func CommitLvlDbWriteDuplex(hash, block, previous [32]byte, tag utxotag) {
	if !commits_format_v2 {
		return
	}

	value, err := CommitLvlDbGet(hash[:], nil)
	if err != nil {
		if (err == leveldb.ErrNotFound) || (err == ErrPebbleNotFound) {
			// we are the first commit, gogogo, write it
			// fall through

		} else {
			log.Fatal("CommitLvlDbWriteDuplex Error: ", err)
		}
	} else {
		var utag = new_utxotag_from_leveldb(value)
		var blockhash [32]byte
		copy(blockhash[:], value[16:48])
		if !CommitLvlDbCheckHeight(uint64(utag.height), blockhash) {

			// fall through, the blockhash isnt on chain, so we must write
		} else {
			if utag_cmp(&tag, &utag) > 0 { // if tag > utag
				return
			}

			// if tag==utag we may be overwriting to get trailer data to disk

			// owerwrite stale commitment with lower utxo tag value commitment
			// could happen in case buggy (previous) haircomb core left over garbage
		}
	}

	var buf = utxotag_to_leveldb(tag)
	buf = append(buf, block[:]...)
	buf = append(buf, previous[:prefixlinks]...)

	CommitLvlBatchPut(hash[:], buf)

	for i := 0; i < 7; i++ {
		if (prefixkernels & (1 << i)) != 0 {
			CommitLvlBatchPut(hash[:9+i], []byte{})
		}
	}
}

func CommitLvlDbUnPrune(load_aes_lvl *[32]byte, maxkey *[8]byte) {
	if !commits_format_v2b {
		//println("Not unpruning because not V2b")
		return
	}
	if !commits_format_v2b_is_unprune {
		//println("Not unpruning because not told to unprune")
		return
	}
	CommitLvlDbNewBatch()
	var start [9]byte
	var cnt byte
	var progress byte
	iter, err := CommitLvlDbNewIterator(
		start[:],
		nil,
	)
	if err != nil {
		log.Fatal("DB iter error: ", err)
	}
	for iter.First(); iter.Valid(); iter.Next() {
		var key = iter.Key()
		// drop prefix kernels
		if len(key) >= 9 && len(key) <= 15 {
			CommitLvlBatchDelete(key)
		}
		// drop reverse commitments
		if len(key) == 32 {
			prog := (100 * uint16(key[0])) / 255
			if byte(prog) >= progress {
				fmt.Print("Unpruning ", prog, "%\r")
				progress = byte(prog) + 1
			}
			var value = iter.Value()
			if len(value) < 48 {
				CommitLvlBatchDelete(key)
				continue
			}
			var utag = new_utxotag_from_leveldb(value)
			var blockhash [32]byte
			copy(blockhash[:], value[16:48])

			if !CommitLvlDbCheckHeight(uint64(utag.height), blockhash) {
				CommitLvlBatchDelete(key)
				continue
			}
			CommitLvlBatchPut(value[0:16], key)
			CommitLvlBatchDelete(key)
			cnt++
			if cnt == 0 {
				CommitLvlDbWriteBatch()
				CommitLvlDbNewBatch()
			}
		}
	}
	err = iter.Release()
	if err != nil {
		log.Fatal("Hash Erase error: ", err)
	}
	CommitLvlBatchDelete([]byte{1})
	CommitLvlBatchPut([]byte{0}, make([]byte, 49))
	CommitLvlDbWriteBatch()
	fmt.Println("Unpruning complete")
	commits_format_v2b = false
	commits_format_v2b_is_unprune = false
	commits_format_v2b_is_duplex_encap = false

}

func CommitLvlDbPrune(load_aes_lvl *[32]byte, maxkey *[8]byte, was_empty bool) {
	var ramPruneHeight = uint64(u_config.prune_ram)
	var diskPruneHeight = uint64(u_config.prune_disk)

	if !commits_format_v2 {
		if was_empty {
			return
		}
		u_config.prune_ram = 0
		u_config.prune_disk = 0
		return
	}

	var maxPruneHeight = ramPruneHeight
	if maxPruneHeight < diskPruneHeight {
		maxPruneHeight = diskPruneHeight
	}

	if commits_pruned_height >= maxPruneHeight {
		return
	}

	CommitLvlDbNewBatch()
	var block, previous, previous_first [32]byte
	var firsttag *utxotag
	var prunableHeight uint64
	var cnt byte
	var progress byte
	iter, err := CommitLvlDbNewIterator(
		new_height_tag(commits_pruned_height),
		new_height_tag(maxPruneHeight+1),
	)
	if err != nil {
		log.Fatal("DB iter error: ", err)
	}

	var finalKey [8]byte
	var finalValue [32]byte
	writtenDuplex := make(map[[32]byte]struct{})
	deletedForward := make(map[[16]byte]struct{})
	for iter.First(); iter.Valid(); iter.Next() {
		var key = iter.Key()
		if len(key) == 8 {
			height := binary.BigEndian.Uint64(key)
			prog := 100 * (height - commits_pruned_height + 1) / (maxPruneHeight - commits_pruned_height + 1)
			if byte(prog) >= progress {
				fmt.Print("Pruning ", byte(prog), "%\r")
				progress = byte(prog) + 1
			}
			if height > maxPruneHeight {
				continue
			}
			if height > diskPruneHeight {
				continue
			}
			var value = iter.Value()
			if firsttag != nil {
				CommitLvlDbWriteDuplex(previous_first, block, previous, *firsttag)
			}
			copy(previous[:], value[32:64])
			previous_first = previous
			firsttag = nil
			copy(block[:], value[0:32])
			cnt++
			if cnt == 0 {
				var v2_hdr [49]byte
				copy(v2_hdr[0:32], value[96:128])
				copy(v2_hdr[32:40], key)
				v2_hdr[40] = prefixkernels
				binary.BigEndian.PutUint64(v2_hdr[41:49], commits_pruned_deleted)

				CommitLvlBatchPut([]byte{0}, v2_hdr[:])

				for v := range deletedForward {
					CommitLvlBatchDelete(v[:])
				}

				CommitLvlDbWriteBatch()
				writtenDuplex = make(map[[32]byte]struct{})
				deletedForward = make(map[[16]byte]struct{})
				CommitLvlDbNewBatch()
			} else {
				copy(finalValue[0:32], value[96:128])
				copy(finalKey[:], key)
			}
			continue
		}
		if len(key) == 16 {
			newtag := new_utxotag_from_leveldb(key)

			if maxPruneHeight > uint64(newtag.height) {
				var comm [32]byte
				var value = iter.Value()
				copy(comm[:], value[0:32])
				var net byte
				if len(value) >= 33 && value[32] != 0 {
					net = value[32]
				}
				comm = midhash(comm[:], net)
				if _, ok := writtenDuplex[comm]; !ok {
					CommitLvlDbWriteDuplex(comm, block, previous, newtag)
					writtenDuplex[comm] = struct{}{}
					if firsttag == nil && comm == previous {
						firsttag = &newtag
					}
					previous = comm
				}
			}
			if diskPruneHeight > uint64(newtag.height) {

				if uint64(newtag.height) != prunableHeight {
					var wantedKey [8]byte
					// we lookup 1+ higher block
					binary.BigEndian.PutUint64(wantedKey[:], uint64(newtag.height)+1)

					_, err := CommitLvlDbGet(wantedKey[:], nil)
					if (err == leveldb.ErrNotFound) || (err == ErrPebbleNotFound) {
						// don't touch this
						diskPruneHeight = uint64(newtag.height)
						continue
					} else {
						// mark this height as prunable
						prunableHeight = uint64(newtag.height)
					}

				}
				var commkey [16]byte
				copy(commkey[:], key)
				deletedForward[commkey] = struct{}{}
				commits_pruned_deleted++

			}
			continue
		}
	}
	err = iter.Release()
	if err != nil {
		log.Fatal("Hash Erase error: ", err)
	}
	if cnt != 0 {

		diskPruneHeight = binary.BigEndian.Uint64(finalKey[:])

		var v2_hdr [49]byte
		copy(v2_hdr[0:32], finalValue[:])
		copy(v2_hdr[32:40], finalKey[:])

		v2_hdr[40] = prefixkernels
		binary.BigEndian.PutUint64(v2_hdr[41:49], commits_pruned_deleted)

		CommitLvlBatchPut([]byte{0}, v2_hdr[:])

		copy((*load_aes_lvl)[:], finalValue[:])
		copy((*maxkey)[:], new_height_tag(diskPruneHeight))
		commits_pruned_height = diskPruneHeight
	}
	for v := range deletedForward {
		CommitLvlBatchDelete(v[:])
	}
	CommitLvlDbWriteBatch()
	fmt.Println("Pruning complete")
}

func CommitLvlDbBlockGetCummulative(val []byte) uint64 {
	var buf [8]byte
	switch len(val) {
	case 144:
		buf = hex2byte8(val[128:144])
	case 160:
		buf = hex2byte8(val[144:160])
	default:
		panic("CommitLvlDbBlockGetCummulative: unknown format")
	}
	return binary.BigEndian.Uint64(buf[:])
}

func CommitLvlDbLoad() {

	var load_aes_lvl [32]byte
	var maxkey [8]byte // max key (height) that was disk-pruned

	// Open DB
	CommitLvlDbOpen()

	var was_empty = CommitLvlDbIsEmpty()

	// if there is extended header, the format is v2 B
	var v2b_header, v2b_err = CommitLvlDbGet([]byte{1}, nil)
	if v2b_err != nil {
		if !((v2b_err == leveldb.ErrNotFound) || (v2b_err == ErrPebbleNotFound)) {
			log.Fatal("Commitdb V2B Header Error: ", v2b_err)
		}
		if u_config.unprune_disk {
			err := CommitLvlDbPut([]byte{1}, []byte{1}, nil)
			if err != nil {
				log.Fatal("Commitdb V2B Set Header Error: ", v2b_err)
			}
			commits_format_v2b = true
			commits_format_v2b_is_unprune = true
		}
	} else {
		commits_format_v2b = true
		if len(v2b_header) >= 1 {
			// the two V2 Extended features depend on the lowest bit and cannot be both turned on
			commits_format_v2b_is_unprune = (v2b_header[0] & 1) == 1
			commits_format_v2b_is_duplex_encap = !commits_format_v2b_is_unprune
		}
	}

	// Disk UnPrune
	CommitLvlDbUnPrune(&load_aes_lvl, &maxkey)

	// if there is header, the format is v2
	var v2_header, v2_err = CommitLvlDbGet([]byte{0}, nil)
	if v2_err != nil {
		if !((v2_err == leveldb.ErrNotFound) || (v2_err == ErrPebbleNotFound)) {
			log.Fatal("Commitdb V2 Header Error: ", v2_err)
		}
	} else {
		commits_format_v2 = true
		copy(load_aes_lvl[:], v2_header[0:32])
		copy(maxkey[:], v2_header[32:40])
		if len(v2_header) >= 49 {
			commits_pruned_deleted = binary.BigEndian.Uint64(v2_header[41:49])
		}
		if v2_header[40] == 0 {
			// pruned before block 1 (pruning off)
			commits_pruned_height = 1
		} else {
			prefixkernels = v2_header[40]
			var proposed_height = binary.BigEndian.Uint64(maxkey[:])
			commits_pruned_height = proposed_height

			block_val, err := CommitLvlDbGet(maxkey[:], nil)
			if err == nil {
				var good_aes_lvl [32]byte
				copy(good_aes_lvl[:], block_val[96:128])
				if load_aes_lvl != good_aes_lvl {
					fmt.Println("AES fingerprint value was wrong (other code was buggy), fixing...")
					load_aes_lvl = good_aes_lvl
				}
			}
		}
	}

	// RAM Prune
	CommitLvlDbPrune(&load_aes_lvl, &maxkey, was_empty)

	// prefix to skip some commitment keys on initial loop, to search mainly normal (zero prefixed) keys and blocks
	// causes max height to be 2^48
	var pref = []byte{0, 0}

	iter, err := CommitLvlDbNewIterator(
		pref,
		keyLimit(pref),
	)
	if err != nil {
		log.Fatal("DB iter error: ", err)
	}
	var sum = sha256.New()

	var l commitLvlOngoingLoad

	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		val := iter.Value()

		switch len(key) {
		case 8:
			if len(val) < 128 {
				log.Println("short block; val:", val)
				if !l.deletor {
					CommitLvlDbNewBatch()
					l.deletor = true
				}

			}
			if len(val) == 128 {
				commits_format_v1 = true
			}
		case 16:
			if len(val) < 32 {
				log.Println("short commit; val", val)
				if !l.deletor {
					CommitLvlDbNewBatch()
					l.deletor = true
				}
			}
		default:
			continue
		}

		if l.deletor {
			switch len(key) {
			case 16, 8:
				CommitLvlBatchDelete(key)
			}
			continue
		}

		switch len(key) {
		case 8:
			copy(maxkey[:], key)
			tag_height := uint64(binary.BigEndian.Uint64(key[:8]))
			fmt.Print("Loading ", tag_height, "...\r")
			l.commitLvlHandleBlock(tag_height, val, load_aes_lvl, sum)
			if len(val) >= 144 {
				var blk [64]byte
				if commits_format_v2 {
					var str = fmt.Sprintf("%x", val[0:32])
					copy(blk[:], []byte(str))
				} else {
					copy(blk[:], val[0:64])
				}
				var cnt [16]byte
				copy(cnt[:], val[128:144])
				accelerator_apply(blk, cnt)
			}
		case 16:
			tag_height := uint64(binary.BigEndian.Uint64(key[:8]))

			if tag_height < commits_pruned_height {
				// write unprune header, to unprune at the next startup (which will fix this kind of error)
				err := CommitLvlDbPut([]byte{1}, []byte{1}, nil)
				// safely close the db
				CommitLvlDbClose()
				// crash (because continuing will silently mis-sync)
				if err == nil {
					fmt.Println("The recovery will be attempted on the next startup.")
					log.Println("The recovery will be attempted on the next startup.")
				}
				fmt.Println("Exiting: DB loading error: Garbage at height / pruning height: ", tag_height, commits_pruned_height)
				log.Fatalln("Exiting: DB loading error: Garbage at height / pruning height: ", tag_height, commits_pruned_height)
				return
			}

			l.iscomm = true

			var net byte
			if len(val) >= 33 {
				net = val[32]
			}

			address := fmt.Sprintf("%X", midhash(val, net))

			miner_mine_commit_pulled(address, new_utxotag_from_leveldb(key))

			sum.Write(key)
			sum.Write(val)

			var aestmp [32]byte

			var aes, err5 = aes.NewCipher(val)
			if err5 != nil {
				log.Fatal("AES fault")
			}

			aes.Encrypt(aestmp[0:16], load_aes_lvl[0:16])
			aes.Decrypt(aestmp[16:32], load_aes_lvl[16:32])

			for i := 8; i < 16; i++ {
				aestmp[i], aestmp[8+i] = aestmp[8+i], aestmp[i]
			}

			aes.Encrypt(load_aes_lvl[0:16], aestmp[0:16])
			aes.Decrypt(load_aes_lvl[16:32], aestmp[16:32])

		}

	}

	if !commits_format_v1 && !commits_format_v2 && maxkey == [8]byte{} && !l.iscomm {
		// add v2 header if creating new, empty db
		err := CommitLvlDbPut([]byte{0}, make([]byte, 49), nil)
		if err != nil {
			log.Fatal("Commitdb Put Error: ", err)
		}
		commits_format_v2 = true
		// pruned before block 1 (pruning off)
		commits_pruned_height = 1
		fmt.Println("made new v2 db")
	}

	const unused = 0
	l.commitLvlHandleBlock(unused, nil, [32]byte{}, sum)

	// Brake Pull
	err = iter.Release()
	if err != nil {
		log.Fatal("Commitdb Iter Error: ", err)
	}

	if l.deletor {
		CommitLvlDbWriteBatch()
	}

	var upgrade_aes = load_aes_lvl

	// upgrade format to v2 (if v1 not detected and v2 not (yet) detected)
	// 1. compress block hash, compress block sha
	// 2. add claim or 00..00 to block
	// 3. add the AES + maxheight v2 format header
	if !commits_format_v1 && !commits_format_v2 {
		var beforeBlockCommitsCummulative = commitsCount()
		CommitLvlDbNewBatch()
		for i := uint64(binary.BigEndian.Uint64(maxkey[:])); i > 0; i-- {
			var key [8]byte
			binary.BigEndian.PutUint64(key[:], i)
			var val, err = CommitLvlDbGet(key[:], nil)
			if err != nil {
				if (err == leveldb.ErrNotFound) || (err == ErrPebbleNotFound) {
					break
				}
				log.Fatal("Commitdb Iter Error: ", err)
			}
			var hashes = CommitLvlDbReadHeight(i)
			var claim [32]byte
			if len(hashes) > 0 {
				claim = hashes[0]
			}
			// uncalculate total commits to get commits before the block
			beforeBlockCommitsCummulative -= uint64(len(hashes))
			// uncompute commitments aes to get aes before given block
			for i := len(hashes) - 1; i >= 0; i-- {
				var aestmp [32]byte

				var aes, err5 = aes.NewCipher(hashes[i][:])
				if err5 != nil {
					log.Fatal("AES fault")
				}

				aes.Decrypt(aestmp[0:16], upgrade_aes[0:16])
				aes.Encrypt(aestmp[16:32], upgrade_aes[16:32])

				for i := 8; i < 16; i++ {
					aestmp[i], aestmp[8+i] = aestmp[8+i], aestmp[i]
				}

				aes.Decrypt(upgrade_aes[0:16], aestmp[0:16])
				aes.Encrypt(upgrade_aes[16:32], aestmp[16:32])
			}
			var blkhash = hex2byte32(val[0:64])
			var blksha = hex2byte32(val[64:128])
			copy(val[0:32], blkhash[:])
			copy(val[32:64], claim[:])
			copy(val[64:96], blksha[:])
			copy(val[96:128], upgrade_aes[:])

			val = append(val, []byte(fmt.Sprintf("%016x", beforeBlockCommitsCummulative))...)

			CommitLvlBatchPut(key[:], val[:])
			_ = val
		}

		if upgrade_aes != [32]byte{} {
			fmt.Println("Upgade failed, aborting...")
			CommitLvlDbNewBatch()
		} else {
			CommitLvlBatchPut([]byte{0}, make([]byte, 49))
			CommitLvlDbWriteBatch()
			commits_format_v2 = true
		}
	}

	// Close DB
	CommitLvlDbClose()

	fmt.Println("v1:", commits_format_v1, "v2:", commits_format_v2)

	// Fingerprint
	database_aes = load_aes_lvl
}

func CommitLvlDbIteratePrunedBlock(combbase, block [32]byte, height uint64, callback func(comm [32]byte, tag utxotag)) {
	val, err := CommitLvlDbGet(combbase[:], nil)
	if (err == leveldb.ErrNotFound) || (err == ErrPebbleNotFound) {
		return
	}
	if err != nil {
		log.Fatal("DB read error: ", err)
	}
	tag := new_flush_utxotag(height + 1)
	var next *[32]byte
	for {
		if len(val) >= 80 {
			val = val[48:80]
		} else {
			val = val[48:]
		}

		next, val = CommitLvlDbHighestLowerThanByPrefix(val, height, block, tag)
		if next == nil {
			break
		}
		if len(val) < 48+3 {
			break
		}
		tag = new_utxotag_from_leveldb(val)
		callback(*next, tag)
	}
}

func CommitLvlDbHighestLowerThanByPrefix(prefix []byte, height uint64, block [32]byte, lowerThan utxotag) (*[32]byte, []byte) {

	var best_value []byte
	var best_commit *[32]byte

	var height_bytes [8]byte
	copy(height_bytes[:], new_height_tag(height))

	iter, err := CommitLvlDbNewIterator(
		prefix,
		keyLimit(prefix),
	)
	if err != nil {
		log.Fatal("DB iter error: ", err)
	}
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()

		if len(key) != 32 {
			continue
		}
		val := iter.Value()
		var blk [32]byte
		copy(blk[:], val[16:48])
		if blk != block {
			continue
		}
		var hgt [8]byte
		copy(hgt[:], val[0:8])
		if hgt != height_bytes {
			continue
		}
		var tag = new_utxotag_from_leveldb(val[0:16])
		if utag_cmp(&tag, &lowerThan) >= 0 {
			continue
		}
		best_value = make([]byte, len(val), len(val))
		copy(best_value[:], val[:])
		best_commit = new([32]byte)
		copy(best_commit[:], key)
	}
	err = iter.Release()
	if err != nil {
		log.Fatal("DB read error: ", err)
	}
	return best_commit, best_value
}

func CommitLvlDbIsEmpty() (is_empty bool) {
	is_empty = true
	iter, err := CommitLvlDbNewIterator(
		nil,
		nil,
	)
	if err != nil {
		log.Fatal("DB iter error: ", err)
	}
	for iter.First(); iter.Valid(); iter.Next() {
		is_empty = false
		break
	}
	err = iter.Release()
	if err != nil {
		log.Fatal("DB read error: ", err)
	}
	return
}
