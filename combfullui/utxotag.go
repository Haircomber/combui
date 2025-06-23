package main

import (
	"encoding/binary"
	"fmt"
	"time"
)

type utxotag struct {
	height    uint32
	commitnum uint32
	txnum     uint16
	outnum    uint16
}

func utxotag_to_json(t utxotag) (o utxotagJson) {
	o.Height = t.height
	o.CommitNumber = t.commitnum
	o.TxNumber = t.txnum
	o.OutputNumber = t.outnum
	return o
}

type utxotagJson struct {
	Height       uint32
	CommitNumber uint32
	TxNumber     uint16
	OutputNumber uint16
}

func new_height_tag(height uint64) (h []byte) {
	h = make([]byte, 8)
	binary.BigEndian.PutUint64(h[0:8], uint64(height))
	return h
}

func new_flush_utxotag(height uint64) (t utxotag) {
	t.height = uint32(height)
	return t
}

func new_utxotag(height int, commitnum int, txnum int, outnum int) (t utxotag) {
	t.height = uint32(height)
	t.commitnum = uint32(commitnum)
	t.txnum = uint16(txnum)
	t.outnum = uint16(outnum)
	return t
}

func posttag(t *utxotag, height uint64) {
	t.height = uint32(height)
	t.txnum = 0
	t.outnum = 0
	t.commitnum = 0
}

func makefaketag() (tag utxotag) {
	var t = time.Now().UnixNano()

	var genesis int64 = 1231006505000000000

	if u_config.testnet4 {
		genesis = 1714777860000000000
	}

	var fakeheight = (t - genesis) / 600000000000
	var tremainder = (t - genesis) % 600000000000

	var faketxnnum = tremainder / 60000000
	var ttleftover = tremainder % 60000000

	var outnum = ttleftover / 6000

	tag.height = uint32(fakeheight)
	tag.txnum = uint16(faketxnnum)
	tag.outnum = uint16(outnum)
	tag.commitnum = uint32(faketxnnum)*10000 + uint32(outnum)
	return tag
}

func forcecoinbasefirst(t utxotag) utxotag {
	t.txnum = 0
	t.outnum = 0
	t.commitnum = 0
	return t
}

func utxotag_to_leveldb(t utxotag) (out []byte) {
	out = make([]byte, 16)

	binary.BigEndian.PutUint64(out[0:8], uint64(t.height))
	binary.BigEndian.PutUint32(out[8:12], t.commitnum)
	binary.BigEndian.PutUint16(out[12:14], t.txnum)
	binary.BigEndian.PutUint16(out[14:16], t.outnum)

	return out
}

func new_utxotag_from_leveldb(buf []byte) (t utxotag) {

	t.height = uint32(binary.BigEndian.Uint64(buf[0:8]))
	t.commitnum = binary.BigEndian.Uint32(buf[8:12])
	t.txnum = binary.BigEndian.Uint16(buf[12:14])
	t.outnum = binary.BigEndian.Uint16(buf[14:16])
	return t
}

// serializes to the old commits.db format
func serializeutxotag(t utxotag) []byte {

	// this is where the commits.db format hardfork will live

	if t.height >= strictly_monotonic_vouts_bugfix_fork_height {
		return []byte(fmt.Sprintf("%08d%08d", t.height, t.commitnum))
	} else {
		return []byte(fmt.Sprintf("%08d%04d%04d", t.height, t.txnum, t.outnum))
	}
}

func utag_cmp(l *utxotag, r *utxotag) int {
	if l.height != r.height {
		return int(l.height) - int(r.height)
	}

	// at this point the l and r heights are the same, use Natasha's fork
	// that is at the fork height we start to compare by commitnum
	if l.height >= strictly_monotonic_vouts_bugfix_fork_height {
		return int(l.commitnum) - int(r.commitnum)
	}

	if l.txnum != r.txnum {
		return int(l.txnum) - int(r.txnum)
	}
	return int(l.outnum) - int(r.outnum)
}
