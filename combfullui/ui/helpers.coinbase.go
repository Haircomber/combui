package main

import "math"

func Coinbase(height uint64) uint64 {
	if height >= 21835313 || height == 0 {
		return 0
	}

	const lost_natasha = 0.00000001
	var ll = math.Log2(float64(height) + lost_natasha)
	var l = math.Log2(float64(height))
	var subsidy_proposed = 210000000 - uint64(math.Floor(l*l*l*l*l*ll))
	return subsidy_proposed
}