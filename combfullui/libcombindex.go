package main

import "bitbucket.org/watashi564/accelerator/libcombindex"
import "crypto/rand"

//import "fmt"

var index_online_min_prefix uint64 = 9

var index_online_padding [32]byte

func indexOnlinePadIfNeeded(buf *[32]byte) {
	if index_online_min_prefix > 0 {
		copy(buf[index_online_min_prefix:], index_online_padding[index_online_min_prefix:])
	}
}

func init() {
	rand.Read(index_online_padding[:])
}

var commitment_height_table = libcombindex.Table{
	Bytes:   0,
	MaxLoad: 33,
	Hops:    500,
	Tweak:   10000,
}
var commitment_index_config = libcombindex.IndexConfig{
	Min: 10,
	Max: 22,
}
var index_scaling_factor byte = 8

var commitment_ordernum_index = libcombindex.MakeIndex()
var lenCommits uint64

func commitsCount() uint64 {
	if !commitment_height_table.ConfiguredCorrectly() {
		return uint64(len(commits)) + commits_ram_pruned_count + commits_pruned_deleted - commits_pruned_combbases
	}
	return lenCommits + commits_ram_pruned_count + commits_pruned_deleted - commits_pruned_combbases
}

func commitsCheckNoMaxHeight(hash [32]byte) (tag utxotag, ok bool) {
	return commitsCheck(hash, ^uint64(0))
}

func commitsCheck(hash [32]byte, maxHeight uint64) (tag utxotag, ok bool) {
	tag, ok = commits[hash]
	if !ok {
		tag, ok = getCommitMaxHeight(hash, maxHeight)
	}
	if !ok {
		tag, ok = CommitLvlDbReadDuplex(hash, maxHeight)
	}
	return tag, ok
}

func index_scaling_height(height uint64, hashb0 byte) uint64 {
	return height*uint64(index_scaling_factor) + uint64(uint64(hashb0)%uint64(index_scaling_factor))
}

func getCommitMaxHeight(hash [32]byte, maxHeight uint64) (tag utxotag, ok bool) {
	if !commitment_height_table.ConfiguredCorrectly() {
		return tag, false
	}
	var possiblyShortHash = hash
	indexOnlinePadIfNeeded(&possiblyShortHash)
	ok = commitment_height_table.Get(possiblyShortHash, maxHeight, func(height uint64) bool {

		ordernum := commitment_ordernum_index.Get(index_scaling_height(height, possiblyShortHash[0]), possiblyShortHash)

		if ordernum == 0 {
			return false
		}
		ordernum--

		h, utag, ok := CommitLvlDbReadPrefix(height, uint32(ordernum))
		if !ok {
			//println("not ok", height, ordernum)
			return false
		}
		if h != hash {
			//fmt.Printf("%x <> %x", h, hash)
			return false
		}
		tag = new_utxotag_from_leveldb(utag[:])
		//fmt.Printf("%x %x %d %d", hash, h, tag.height, tag.commitnum)
		ok = true
		return ok
	})
	return tag, ok
}
func getCommitMaxHeights(hash [16]byte, maxHeight uint64, cb func(hash [32]byte, tag utxotag)) {
	if !commitment_height_table.ConfiguredCorrectly() {
		return
	}
	var possiblyShortHash [32]byte
	copy(possiblyShortHash[:], hash[:])
	indexOnlinePadIfNeeded(&possiblyShortHash)
	commitment_height_table.Get(possiblyShortHash, maxHeight, func(height uint64) bool {

		ordernum := commitment_ordernum_index.Get(index_scaling_height(height, possiblyShortHash[0]), possiblyShortHash)

		if ordernum == 0 {
			return false
		}
		ordernum--

		h, utag, ok := CommitLvlDbReadPrefix(height, uint32(ordernum))
		if !ok {
			//println("not ok", height, ordernum)
			return false
		}
		var possiblyShortFoundHash = h
		indexOnlinePadIfNeeded(&possiblyShortFoundHash)
		if possiblyShortFoundHash != possiblyShortHash {
			//fmt.Printf("%x <> %x", h, hash)
			return false
		}
		//fmt.Printf("%x %x %d %d", hash, h, tag.height, tag.commitnum)
		cb(h, new_utxotag_from_leveldb(utag[:]))
		return true
	})
	return
}

func writeCommits(height uint64) {
	if !commitment_height_table.ConfiguredCorrectly() {
		return
	}
	if height < u_config.prune_ram {
		return
	}
	var currentCommits = make([]map[[32]byte]uint64, index_scaling_factor)
	for i := 0; i < int(index_scaling_factor); i++ {
		currentCommits[i] = make(map[[32]byte]uint64)
	}
	for commit, tag := range commits {
		indexOnlinePadIfNeeded(&commit)
		if uint64(tag.height) == height {
			commitment_height_table.Insert(commit, height)
			currentCommits[uint64(commit[0])%uint64(index_scaling_factor)][commit] = uint64(tag.commitnum) + 1
			//fmt.Printf("adding %x %d\n", commit, tag.commitnum+1)
			lenCommits++
		}
	}
	//println("Height is:", height)
	for i := 0; i < int(index_scaling_factor); i++ {
		if !commitment_ordernum_index.InsertCfg(index_scaling_height(height, byte(i)), currentCommits[i], commitment_index_config) {
			panic("cannot insert to index, this cannot occur")
		}
	}
	commits = make(map[[32]byte]utxotag)
	//println(commitment_height_table.Len(), commitment_height_table.Cap())
}

func reorgSpecificCommits(height uint64, readedBlock [][32]byte) {
	if !commitment_height_table.ConfiguredCorrectly() {
		return
	}
	if height < u_config.prune_ram {
		return
	}
	for i := range readedBlock {
		indexOnlinePadIfNeeded(&readedBlock[i])
	}
	lenCommits -= uint64(len(readedBlock))
	for i := 0; i < int(index_scaling_factor); i++ {
		commitment_ordernum_index.Delete(index_scaling_height(height, byte(i)))
	}
	commitment_height_table.Delete(readedBlock, height)
	//println("height", height, "reorged", lenCommits, "==", commitment_height_table.Len(), commitment_height_table.Cap())
}

func reorgCommits(height uint64) {
	if !commitment_height_table.ConfiguredCorrectly() {
		return
	}
	if height < u_config.prune_ram {
		return
	}
	var readedBlock = CommitLvlDbReadHeight(height)
	for i := range readedBlock {
		indexOnlinePadIfNeeded(&readedBlock[i])
	}
	lenCommits -= uint64(len(readedBlock))
	for i := 0; i < int(index_scaling_factor); i++ {
		commitment_ordernum_index.Delete(index_scaling_height(height, byte(i)))
	}
	commitment_height_table.Delete(readedBlock, height)
	//println("height", height, "reorged", lenCommits, "==", commitment_height_table.Len(), commitment_height_table.Cap())
}
