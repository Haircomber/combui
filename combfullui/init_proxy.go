package main

import "sync"
import "math/rand"
import "fmt"

// mutexed boolean
var init_proxy_mutex sync.RWMutex

// are we init proxying or the initial db load is over
var init_proxy_is_over bool

// init proxy nonce - don't proxy requests that we proxied already
var init_proxy_nonce = string(fmt.Sprint((rand.Uint64())))

func get_init_proxy_is_over() (ret bool) {
	init_proxy_mutex.RLock()
	ret = init_proxy_is_over
	init_proxy_mutex.RUnlock()
	return
}
func set_init_proxy_is_over() {
	init_proxy_mutex.Lock()
	init_proxy_is_over = true
	init_proxy_mutex.Unlock()
}
