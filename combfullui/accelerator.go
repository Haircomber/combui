package main

import "encoding/binary"
import "sync"

var accelerator_mut sync.RWMutex
var accelerator [32]byte
var accelerator_bh [32]byte
var acceleratorP2WSH uint64

func accelerator_get() ([32]byte, [32]byte, uint64) {
	accelerator_mut.RLock()
	defer accelerator_mut.RUnlock()
	return accelerator, accelerator_bh, acceleratorP2WSH
}

func accelerator_next(blockhash string, count uint64) {
	if commits_format_v1 {
		return
	}
	var buf [72]byte
	var blkhash = hex2byte32([]byte(blockhash))
	copy(buf[0:32], accelerator[0:32])
	copy(buf[32:32*2], blkhash[0:32])
	binary.BigEndian.PutUint64(buf[64:72], uint64(count))
	accelerator_mut.Lock()
	accelerator = nethash(buf[:])
	acceleratorP2WSH += uint64(count)
	accelerator_bh = blkhash
	accelerator_mut.Unlock()
}
func accelerator_apply(blockhash [64]byte, cnt [16]byte) {
	var buf = hex2byte8(cnt[:])
	var count = binary.BigEndian.Uint64(buf[:])
	accelerator_next(string(blockhash[:]), count)
}
func accelerator_reorg(height uint64) {
	if commits_format_v1 {
		return
	}
	accelerator_mut.Lock()
	accelerator = [32]byte{}
	acceleratorP2WSH = 0
	accelerator_mut.Unlock()
	// Set a stop limit per accelerators
	var lowest uint64 = 481824
	if u_config.regtest || u_config.testnet4 {
		lowest = 1
	}
	for h := lowest; h <= height; h++ {
		blockhash, cnt := CommitLvlDbReadBlock(h)
		if blockhash == [64]byte{} {
			//println("almost ok..")
			return
		}
		accelerator_apply(blockhash, cnt)
	}
}
