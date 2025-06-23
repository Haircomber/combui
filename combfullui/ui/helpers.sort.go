package main

import (
	"sort"
)

func sortStrings(strings []string) {
	sort.Slice(strings, func(i, j int) bool {
		for k := 0; k < 32; k++ {
			if strings[i][k] == strings[j][k] {
				continue
			}
			return strings[i][k] < strings[j][k]
		}
		return true
	})
}