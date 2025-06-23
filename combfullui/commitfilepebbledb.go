package main

import (
	"context"
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/bloom"
	"log"
)

const COMMITS_PEBBLEDB_PATH = "commits_pebble"
const COMMITS_PEBBLEDB_PATH_TESTNET = "commits_pebble_testnet"
const COMMITS_PEBBLEDB_PATH_TESTNET4 = "commits_pebble_testnet4"
const COMMITS_PEBBLEDB_PATH_REGTEST = "commits_pebble_regtest"

const CommitsDbBackendPebble byte = 1

// locked by commitsdb_mutex
var commitsdb_pebble *pebble.DB
var commitsbatch_pebble *pebble.Batch
var commitsoptions_pebble = &pebble.WriteOptions{
	Sync: true,
}

var ErrPebbleNotFound = pebble.ErrNotFound

type PebbleIterator = *pebble.Iterator

func CommitPebbleDbNewIterator(begin, end []byte) (*CommitLvlDbIterator, error) {
	iter, err := commitsdb_pebble.NewIterWithContext(context.Background(), &pebble.IterOptions{
		LowerBound: begin,
		UpperBound: end,
	})
	if err != nil {
		return nil, err
	}
	return &CommitLvlDbIterator{
		PebbleDbIter: iter,
	}, nil
}

func CommitPebbleDbNewBatch() {
	commitsbatch_pebble = commitsdb_pebble.NewBatch()
}

func CommitPebbleDbWriteBatch() {
	err := commitsbatch_pebble.Commit(commitsoptions_pebble)
	if err != nil {
		log.Fatal("DB write batch error: ", err)
	}
	commitsbatch_pebble = nil
}

func CommitPebbleDbOpen() {
	path := COMMITS_PEBBLEDB_PATH
	if u_config.testnet {
		path = COMMITS_PEBBLEDB_PATH_TESTNET
	}
	if u_config.testnet4 {
		path = COMMITS_PEBBLEDB_PATH_TESTNET4
	}
	if u_config.regtest {
		path = COMMITS_PEBBLEDB_PATH_REGTEST
	}
	db, err := pebble.Open(u_config.db_dir+path, &pebble.Options{
		Levels: []pebble.LevelOptions{{
			Compression:  pebble.NoCompression,
			FilterPolicy: bloom.FilterPolicy(10),
		}},
	})
	if err != nil {
		log.Fatal("DB open error: ", err)
	}
	commitsdb_mutex.Lock()
	commitsdb_pebble = db
}

func CommitPebbleDbGet(key []byte, _ *struct{}) (out []byte, err error) {
	data, closer, err := commitsdb_pebble.Get(key)
	if err != nil {
		return nil, err
	}
	out = make([]byte, len(data), len(data))
	copy(out, data)
	err = closer.Close()
	if err != nil {
		return nil, err
	}
	return
}
func CommitPebbleDbPut(key, value []byte, _ *struct{}) (err error) {
	err = commitsdb_pebble.Set(key, value, commitsoptions_pebble)
	return
}
func CommitPebbleBatchPut(key, value []byte) (err error) {
	commitsbatch_pebble.Set(key, value, commitsoptions_pebble)
	return
}
func CommitPebbleBatchDelete(key []byte) (err error) {
	commitsbatch_pebble.Delete(key, commitsoptions_pebble)
	return
}

func CommitPebbleDbClose() {
	if commitsdb_pebble == nil {
		return
	}
	commitsdb_pebble.Close()
	commitsdb_pebble = nil
	commitsdb_mutex.Unlock()
}
