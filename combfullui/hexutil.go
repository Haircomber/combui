package main

import "fmt"

var ErrBadHex704 error = fmt.Errorf("Not uppercase 704byte hex identifier")
var ErrBadHex72 error = fmt.Errorf("Not uppercase 72byte hex identifier")
var ErrBadHex96 error = fmt.Errorf("Not uppercase 96byte hex identifier")
var ErrBadHex736 error = fmt.Errorf("Not uppercase 736byte hex identifier")
var ErrBadHex672 error = fmt.Errorf("Not uppercase 672byte hex identifier")
var ErrBadHex32 error = fmt.Errorf("Not uppercase 32byte hex identifier")
var ErrBadHex8 error = fmt.Errorf("Not uppercase 8byte hex identifier")
var ErrBadHex2 error = fmt.Errorf("Not uppercase 2byte hex identifier")
var ErrBadDec8 error = fmt.Errorf("Not 8byte decimal hex identifier")
var ErrBadDec4 error = fmt.Errorf("Not 4byte decimal hex identifier")
var ErrBadHexpand32 error = fmt.Errorf("Not uppercase 32byte hexpand identifier")
var ErrBadHexpand66 error = fmt.Errorf("Not uppercase 66byte hexpand identifier")

const NET_LETTER_A = 1
const NET_LETTER_F = 2
const NET_LETTER_G = 3
const NET_LETTER_V = 4

func netLetter(id int) byte {
	if Testnet() {
		switch id {
		case NET_LETTER_A:
			return 'a'
		case NET_LETTER_F:
			return 'f'
		case NET_LETTER_G:
			return 'g'
		case NET_LETTER_V:
			return 'v'
		}
	}
	switch id {
	case NET_LETTER_A:
		return 'A'
	case NET_LETTER_F:
		return 'F'
	case NET_LETTER_G:
		return 'G'
	case NET_LETTER_V:
		return 'V'
	}
	return 0
}

func Hex(x byte) byte {
	return 7*(x/10) + x + '0'
}

func checkDEC8(b string) error {
	if len(b) != 16 {
		return ErrBadDec8
	}
	for i := 0; i < 16; i++ {
		if (b[i] >= '0') && (b[i] <= '9') {
		} else {
			return ErrBadDec8
		}
	}
	return nil
}
func checkDEC4(b string) error {
	if len(b) != 8 {
		return ErrBadDec4
	}
	for i := 0; i < 8; i++ {
		if (b[i] >= '0') && (b[i] <= '9') {
		} else {
			return ErrBadDec4
		}
	}
	return nil
}

func checkHEXupper(b string, length int) bool {
	if len(b) != 2*length {
		return false
	}
	for i := 0; i < 2*length; i++ {
		if ((b[i] >= '0') && (b[i] <= '9')) || ((b[i] >= 'A') && (b[i] <= 'F')) {
		} else {

			return false
		}
	}
	return true
}
func checkHEX(b string, length int) bool {
	if len(b) != 2*length {
		return false
	}
	for i := 0; i < 2*length; i++ {
		if ((b[i] >= '0') && (b[i] <= '9')) || ((b[i] >= netLetter(NET_LETTER_A)) && (b[i] <= netLetter(NET_LETTER_F))) {
		} else {

			return false
		}
	}
	return true
}
func checkHEX96upper(b string) error {
	if checkHEXupper(b, 96) {
		return nil
	}
	return ErrBadHex96
}

func checkHEX72upper(b string) error {
	if checkHEXupper(b, 72) {
		return nil
	}
	return ErrBadHex72
}
func checkHEX704upper(b string) error {
	if checkHEXupper(b, 704) {
		return nil
	}
	return ErrBadHex704
}
func checkHEX736upper(b string) error {
	if checkHEXupper(b, 736) {
		return nil
	}
	return ErrBadHex736
}
func checkHEX672upper(b string) error {
	if checkHEXupper(b, 672) {
		return nil
	}
	return ErrBadHex672
}
func checkHEX32(b string) error {
	if checkHEX(b, 32) {
		return nil
	}
	return ErrBadHex32
}
func checkHEX32upper(b string) error {
	if checkHEXupper(b, 32) {
		return nil
	}
	return ErrBadHex32
}
func checkHEX2upper(b string) error {
	if checkHEXupper(b, 2) {
		return nil
	}
	return ErrBadHex2
}
func checkHEX8upper(b string) error {
	if checkHEXupper(b, 8) {
		return nil
	}
	return ErrBadHex8
}
func hex2byte8(hex []byte) (out [8]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2byte2(hex []byte) (out [2]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2byte32(hex []byte) (out [32]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2uint16(hex []byte) (out uint16) {
	out = uint16(x2b(hex[0]))<<12 | uint16(x2b(hex[1]))<<8 | uint16(x2b(hex[2]))<<4 | uint16(x2b(hex[3]))
	return out
}

func hex2byte96(hex []byte) (out [96]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2byte72(hex []byte) (out [72]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2byte704(hex []byte) (out [704]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2byte736(hex []byte) (out [736]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func hex2byte672(hex []byte) (out [672]byte) {
	for i := range out {
		out[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}
	return out
}
func x2b(hex byte) (lo byte) {
	return (hex & 15) + 9*(hex>>6)
}
func checkHEXPAND(b string, length int) bool {
	if len(b) != 2*length {
		return false
	}
	for i := 0; i < 2*length; i++ {
		if (b[i] >= netLetter(NET_LETTER_G)) && (b[i] <= netLetter(NET_LETTER_V)) {
		} else {
			return false
		}
	}
	return true
}
func checkHEXPAND32(b string) error {
	if checkHEXPAND(b, 32) {
		return nil
	}
	return ErrBadHexpand32
}
func checkHEXPAND66(b string) error {
	if checkHEXPAND(b, 66) {
		return nil
	}
	return ErrBadHexpand66
}
func Hexpand2(v []byte) (out [4]byte) {
	if len(v)*2 != len(out) {
		panic("Hexpand2 not 16bit")
	}
	for i := range out {
		if i&1 == 1 {
			out[i] = (v[i>>1] & 0xF) + netLetter(NET_LETTER_G)
		} else {
			out[i] = (v[i>>1] >> 4) + netLetter(NET_LETTER_G)
		}
	}
	return out
}

func Hexpand32(v []byte) (out [64]byte) {
	if len(v)*2 != len(out) {
		panic("Hexpand32 not 256bit")
	}
	for i := range out {
		if i&1 == 1 {
			out[i] = (v[i>>1] & 0xF) + netLetter(NET_LETTER_G)
		} else {
			out[i] = (v[i>>1] >> 4) + netLetter(NET_LETTER_G)
		}
	}
	return out
}

func hexpand2byte32(hexpand []byte) (out [32]byte) {
	for i := range out {
		out[i] = 16*(hexpand[2*i]-netLetter(NET_LETTER_G)) + (hexpand[2*i+1] - netLetter(NET_LETTER_G))
	}
	return out
}

func hexpand2byte2(hexpand []byte) (out [2]byte) {
	for i := range out {
		out[i] = 16*(hexpand[2*i]-netLetter(NET_LETTER_G)) + (hexpand[2*i+1] - netLetter(NET_LETTER_G))
	}
	return out
}

func Unhexpand(v []byte) {
	for i := range v {
		v[i] -= netLetter(NET_LETTER_G)
		v[i] = Hex(v[i])
	}
}