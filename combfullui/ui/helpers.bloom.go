package main

import "bitbucket.org/watashi564/accelerator/libbloom"
import "encoding/base64"

func LibbloomGet(fun, tweak uint32, hex string, bloom []byte) bool {
	var hash [32]byte
	for i := range hash {
		hash[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}

	return libbloom.Get(fun, tweak, hash, bloom)
}

func LibbloomSet(fun, prefix, tweak uint32, hex string, bloom []byte) {
	var hash [32]byte
	for i := range hash {
		hash[i] = (x2b(hex[i<<1]) << 4) | x2b(hex[i<<1|1])
	}

	libbloom.Set(fun, tweak, hash, bloom)
	libbloom.SetPrefix(prefix, tweak, hash, bloom)
}

func BloomSerialize(bloom []byte) string {
	return base64.URLEncoding.EncodeToString(bloom)
}