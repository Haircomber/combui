package main

import "net/http"
import "embed"
import "encoding/json"
import "fmt"
import "encoding/hex"
import "bitbucket.org/watashi564/accelerator/libbloom"
import "sort"
import "io"
import "strings"
import "crypto/sha256"
import "encoding/base64"
import "github.com/golangplus/container/heap"

//go:embed ui/index*.html
//go:embed ui/doc*.html
//go:embed ui/swagger*.yaml
//go:embed ui/robots*.txt
//go:embed ui/sitemap*.xml
//go:embed ui/style*.css
//go:embed ui/w3*.css
//go:embed ui/wasm_exec*.js
//go:embed ui/combfullui.wasm
var content embed.FS

var FAIL = []byte(`e({"Success":false});`)

const MaxCommitmentsArray = 256

// cuturl cuts the url into up to 8 (max_sep) parts separated by dots after the last slash
func cuturl(path string) (paths []string) {
	const max_sep = 8
	var seps = make([]int, 1, max_sep)
	for i, c := range path {
		if c == '/' {
			seps = seps[0:1]
			seps[0] = i
		} else if c == '.' {
			seps = append(seps, i)
			if len(seps) == max_sep {
				break
			}
		}
	}
	for i := 1; i < len(seps); i++ {
		var min = 1 + seps[i-1]
		var max = seps[i]
		if min < max {
			paths = append(paths, path[min:max])
		} else {
			paths = append(paths, "")
		}
	}
	return
}

// example call http://127.0.0.1:21212/0000000000481824.js
func controller_getblock(height string, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/javascript")

	if checkDEC8(height) != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	var h uint64

	for _, v := range height {
		h *= 10
		h += uint64(v) - '0'
	}

	var data = struct {
		Testnet bool
		DbData  map[string]interface{}
		Commits map[string]utxotagJson
		Success bool
	}{
		Testnet(),
		nil,
		nil,
		true,
	}

	data.DbData, data.Commits = CommitLvlDbDumpHeight(h)

	_, _ = w.Write([]byte("b("))
	_ = json.NewEncoder(w).Encode(data)
	_, _ = w.Write([]byte(");"))
}

// since we might cache in controller_getrange, we need to uncache when new blocks get mined
func uncache_getrange(height uint64) {
	{
		h := (height / 125000) * 125000
		// uncache
		r_url_path := fmt.Sprintf("/%016d.%016d", h, h + 125000)
		//println("-", r_url_path)
		ph := sha256.Sum256([]byte(r_url_path))
		GzipUnstorePathHash(ph)
	}
	if height % 125000 == 0 && height > 0 { // edge case
		// uncache
		r_url_path := fmt.Sprintf("/%016d.%016d", height - 125000, height)
		//println("-", r_url_path)
		ph := sha256.Sum256([]byte(r_url_path))
		GzipUnstorePathHash(ph)
	}
}

// example call http://127.0.0.1:21212/0000000000481824.0000000000482884.js
func controller_getrange(minheight, maxheight string, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/javascript")

	if checkDEC8(minheight) != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}
	if checkDEC8(maxheight) != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	var min, max uint64

	for _, v := range minheight {
		min *= 10
		min += uint64(v) - '0'
	}

	for _, v := range maxheight {
		max *= 10
		max += uint64(v) - '0'
	}

	var is_caching = false
	r_url_path := r.URL.Path
	if min + 125000 == max && min % 125000 == 0 && max % 125000 == 0 {
		is_caching = true
		// caching on

		// we anyway switch only using the part after the last slash and before the last dot
		for i := len(r_url_path) - 1; i >= 0; i-- {
			if r_url_path[i] == '.' {
				r_url_path = r_url_path[:i]
				break
			}
		}
		for i := len(r_url_path) - 1; i >= 0; i-- {
			if r_url_path[i] == '/' {
				r_url_path = r_url_path[i:]
				break
			}
		}
	}

	var data = struct {
		Testnet   bool
		Combbases []struct {
			Combbase string
			Height   uint64
		}
		Success bool
	}{
		Testnet(),
		nil,
		true,
	}

	commits_mutex.RLock()

	for key, height := range combbases {
		if height > max {
			continue
		}
		if height < min {
			continue
		}
		data.Combbases = append(data.Combbases, struct {
			Combbase string
			Height   uint64
		}{
			fmt.Sprintf("%x", key),
			height,
		})
	}

	commits_mutex.RUnlock()

	sort.Slice(data.Combbases, func(i, j int) bool {
		return data.Combbases[i].Height > data.Combbases[j].Height
	})

	if is_caching {
		data, err := json.Marshal(data)
		const ctype = "text/javascript"

		//println("+", r_url_path)

		if err != nil {
			w.WriteHeader(http.StatusNonAuthoritativeInfo)
			_, _ = w.Write(FAIL)
			return
		}
		data = append(data, []byte(");")...)
		data = append([]byte("r("), data...)


		h := sha256.Sum256(data)
		ph := sha256.Sum256([]byte(r_url_path))
		hash_computed := "sha-256=:" + base64.StdEncoding.EncodeToString(h[:]) + ":"
		w.Header().Set("Repr-Digest", hash_computed)
		w.Header().Set("Digest", hash_computed)
		GzipStoreUrlPathHashDigest(ph, hash_computed, ctype, len(data))
		_, _ = w.Write(data)
	} else {
		_, _ = w.Write([]byte("r("))
		_ = json.NewEncoder(w).Encode(data)
		_, _ = w.Write([]byte(");"))
	}
}

// example call http://127.0.0.1:21212/00000013.0000002491084390.00000007.9999999999999999.000000000000000000.0000000000000000.0000000000000000.js
func controller_getcommitsbyprefixesmin(funcs, tweak, bytes, maxheight, prefixes, minheight, mincommitnum string, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/javascript")

	if checkDEC4(bytes) != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	if checkDEC8(maxheight) != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}
	if checkDEC8(minheight) != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}
	if checkDEC8(mincommitnum) != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	if checkDEC4(funcs) != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}
	if checkDEC8(tweak) != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	var fun, prefx, tweak64 uint64

	for _, v := range funcs {
		fun *= 10
		fun += uint64(v) - '0'
	}
	for _, v := range tweak {
		tweak64 *= 10
		tweak64 += uint64(v) - '0'
	}

	for _, v := range bytes {
		prefx *= 10
		prefx += uint64(v) - '0'
	}

	if prefx > 16 || prefx < 7 {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	var maxh, minh uint64
	var minnum uint32

	for _, v := range maxheight {
		maxh *= 10
		maxh += uint64(v) - '0'
	}

	for _, v := range minheight {
		minh *= 10
		minh += uint64(v) - '0'
	}

	for _, v := range mincommitnum {
		minnum *= 10
		minnum += uint32(v) - '0'
	}

	rawprefixes, err := hex.DecodeString(prefixes)
	if err != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}
	if len(rawprefixes) == 0 {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	if uint64(len(rawprefixes))%prefx != 0 {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	var hashprefixes = make(map[[16]byte]struct{})

	for i := uint64(0); i < uint64(len(rawprefixes)); i += prefx {
		var prefix [16]byte
		copy(prefix[:], rawprefixes[i:i+prefx])
		hashprefixes[prefix] = struct{}{}
	}

	var respbloom = make([]byte, len(rawprefixes))

	var data = struct {
		Testnet bool
		Commitments []struct {
			commit    [32]byte
			Commit    string
			Height    uint32
			CommitNum uint32
			TxNum     uint16
			OutNum    uint16
			Combbase  bool
		}
		Bloom   []byte
		Count   uint64
		Success bool
		Height  uint64
	}{
		Testnet(),
		nil,
		respbloom,
		0,
		true,
		0,
	}

	heapifier := func() {
		heap.InitF(len(data.Commitments), func(i, j int) bool {
			if data.Commitments[i].Height == data.Commitments[j].Height {
				return data.Commitments[i].CommitNum > data.Commitments[j].CommitNum
			}
			return data.Commitments[i].Height > data.Commitments[j].Height
		}, func (i, j int) {
			data.Commitments[i], data.Commitments[j] = data.Commitments[j], data.Commitments[i]
		})
	}
	fixer0 := func() {
		heap.FixF(len(data.Commitments), func(i, j int) bool {
			if data.Commitments[i].Height == data.Commitments[j].Height {
				return data.Commitments[i].CommitNum > data.Commitments[j].CommitNum
			}
			return data.Commitments[i].Height > data.Commitments[j].Height
		}, func (i, j int) {
			data.Commitments[i], data.Commitments[j] = data.Commitments[j], data.Commitments[i]
		}, 0)
	}
	better := func(h, num uint32) bool {
		if data.Commitments[0].Height == h {
			return data.Commitments[0].CommitNum > num
		}
		return data.Commitments[0].Height > h
	}
	filter := func(h, num uint32) bool {
		if uint64(h) > maxh {
			return false
		}
		if minh == uint64(h) {
			return minnum < num
		}
		return minh < uint64(h)
	}
	attacher := func(utag utxotag, key [32]byte) {
		data.Count++

		var _, ok = combbases[key]

		var item = struct {
			commit    [32]byte
			Commit    string
			Height    uint32
			CommitNum uint32
			TxNum     uint16
			OutNum    uint16
			Combbase  bool
		}{
			commit:    key,
			Commit:    fmt.Sprintf("%x", key),
			Height:    utag.height,
			CommitNum: utag.commitnum,
			TxNum:     utag.txnum,
			OutNum:    utag.outnum,
			Combbase:  ok,
		}

		if len(data.Commitments) < MaxCommitmentsArray {



			data.Commitments = append(data.Commitments, item)

			if len(data.Commitments) == MaxCommitmentsArray {
				heapifier()
				//sorter()
			}
		} else if better(item.Height, item.CommitNum) {

			libbloom.Set(uint32(fun), uint32(tweak64), data.Commitments[0].commit, respbloom)

			data.Commitments[0] = item
			fixer0()
			//sorter()
		} else {
			libbloom.Set(uint32(fun), uint32(tweak64), item.commit, respbloom)
		}
	}

	commits_mutex.RLock()

	if len(commits) == 0 {

		if prefx < index_online_min_prefix {

			if commitment_height_table.ConfiguredCorrectly() {

				commits_mutex.RUnlock()

				w.WriteHeader(http.StatusNonAuthoritativeInfo)
				_, _ = w.Write(FAIL)
				return

			}

		}
	}

	data.Height = uint64(commit_currently_loaded.height)

	for key, utag := range commits {

		var keyprefix [16]byte
		copy(keyprefix[0:prefx], key[0:prefx])

		if _, ok := hashprefixes[keyprefix]; ok && filter(utag.height, utag.commitnum) {
			attacher(utag, key)

		}
	}
	for pre := range hashprefixes {
		getCommitMaxHeights(pre, minh, func(key [32]byte, utag utxotag) {

			if !filter(utag.height, utag.commitnum) {
				return
			}

			attacher(utag, key)
		})
	}
	commits_mutex.RUnlock()


	if CommitLvlDbPruned() && prefx > 0 {
		// make map once
		var commits = make(map[[32]byte]utxotag)
		// search disk

		//println("map size", len(hashprefixes))

		for pre := range hashprefixes {

			//fmt.Printf("prefix:%x\n", pre)

			CommitLvlDbReadDuplexPrefix(pre, byte(prefx), 0, commits)
			//fmt.Printf("prefix:%x count %d\n", pre, len(commits))
			// same iteration as above, except the prefix check isn't needed
			for key, utag := range commits {

				var keyprefix [16]byte
				copy(keyprefix[0:prefx], key[0:prefx])

				if _, ok := hashprefixes[keyprefix]; ok && filter(utag.height, utag.commitnum) {

					attacher(utag, key)

				}
			}

			// clean map
			if len(commits) > 0 {
				commits = make(map[[32]byte]utxotag)
			}

		}
	}

	if data.Count <= MaxCommitmentsArray {
		data.Bloom = nil
	}

	_, _ = w.Write([]byte("p("))
	_ = json.NewEncoder(w).Encode(data)
	_, _ = w.Write([]byte(");"))
}

// example call http://127.0.0.1:21212/00000013.0000002491084390.00000009.9999999999999999.000000000000000000.js
func controller_getcommitsbyprefixes(funcs, tweak, bytes, maxheight, prefixes string, w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/javascript")

	if checkDEC4(bytes) != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	if checkDEC8(maxheight) != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	if checkDEC4(funcs) != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}
	if checkDEC8(tweak) != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	var fun, prefx, tweak64 uint64

	for _, v := range funcs {
		fun *= 10
		fun += uint64(v) - '0'
	}
	for _, v := range tweak {
		tweak64 *= 10
		tweak64 += uint64(v) - '0'
	}

	for _, v := range bytes {
		prefx *= 10
		prefx += uint64(v) - '0'
	}

	if prefx > 16 || prefx < 3 {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	var h uint64

	for _, v := range maxheight {
		h *= 10
		h += uint64(v) - '0'
	}

	rawprefixes, err := hex.DecodeString(prefixes)
	if err != nil {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}
	if len(rawprefixes) == 0 {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	if uint64(len(rawprefixes))%prefx != 0 {
		w.WriteHeader(http.StatusNonAuthoritativeInfo)
		_, _ = w.Write(FAIL)
		return
	}

	var hashprefixes = make(map[[16]byte]struct{})

	for i := uint64(0); i < uint64(len(rawprefixes)); i += prefx {
		var prefix [16]byte
		copy(prefix[:], rawprefixes[i:i+prefx])
		hashprefixes[prefix] = struct{}{}
	}

	var respbloom = make([]byte, len(rawprefixes))

	var data = struct {
		Testnet bool
		Commits map[string]struct {
			Height    uint32
			CommitNum uint32
			TxNum     uint16
			OutNum    uint16
			Combbase  bool
		}
		Bloom   []byte
		Count   uint64
		Success bool
		Height  uint64
	}{
		Testnet(),
		make(map[string]struct {
			Height    uint32
			CommitNum uint32
			TxNum     uint16
			OutNum    uint16
			Combbase  bool
		}),
		respbloom,
		0,
		true,
		0,
	}

	commits_mutex.RLock()

	if len(commits) == 0 {

		if prefx < index_online_min_prefix {

			if commitment_height_table.ConfiguredCorrectly() {

				commits_mutex.RUnlock()

				w.WriteHeader(http.StatusNonAuthoritativeInfo)
				_, _ = w.Write(FAIL)
				return

			}

		}
	}

	data.Height = uint64(commit_currently_loaded.height)

	for key, utag := range commits {

		var keyprefix [16]byte
		copy(keyprefix[0:prefx], key[0:prefx])

		if _, ok := hashprefixes[keyprefix]; ok && uint64(utag.height) <= h {
			data.Count++

			libbloom.Set(uint32(fun), uint32(tweak64), key, respbloom)

			if data.Commits != nil {

				var _, ok = combbases[key]

				data.Commits[fmt.Sprintf("%x", key)] = struct {
					Height    uint32
					CommitNum uint32
					TxNum     uint16
					OutNum    uint16
					Combbase  bool
				}{
					Height:    utag.height,
					CommitNum: utag.commitnum,
					TxNum:     utag.txnum,
					OutNum:    utag.outnum,
					Combbase:  ok,
				}
				if len(data.Commits) > MaxCommitmentsArray {
					data.Commits = nil
				}
			}

		}
	}
	for pre := range hashprefixes {
		getCommitMaxHeights(pre, h, func(key [32]byte, utag utxotag) {
			data.Count++

			libbloom.Set(uint32(fun), uint32(tweak64), key, respbloom)

			if data.Commits != nil {
				var _, ok = combbases[key]

				data.Commits[fmt.Sprintf("%x", key)] = struct {
					Height    uint32
					CommitNum uint32
					TxNum     uint16
					OutNum    uint16
					Combbase  bool
				}{
					Height:    utag.height,
					CommitNum: utag.commitnum,
					TxNum:     utag.txnum,
					OutNum:    utag.outnum,
					Combbase:  ok,
				}
				if len(data.Commits) > MaxCommitmentsArray {
					data.Commits = nil
				}
			}
		})
	}
	commits_mutex.RUnlock()

	if CommitLvlDbPruned() && prefx > 0 {
		// make map once
		var commits = make(map[[32]byte]utxotag)
		// search disk

		//println("map size", len(hashprefixes))

		for pre := range hashprefixes {

			//fmt.Printf("prefix:%x\n", pre)

			CommitLvlDbReadDuplexPrefix(pre, byte(prefx), 0, commits)
			//fmt.Printf("prefix:%x count %d\n", pre, len(commits))
			// same iteration as above, except the prefix check isn't needed
			for key, utag := range commits {

				var keyprefix [16]byte
				copy(keyprefix[0:prefx], key[0:prefx])

				if _, ok := hashprefixes[keyprefix]; ok && uint64(utag.height) <= h {

					data.Count++

					libbloom.Set(uint32(fun), uint32(tweak64), key, respbloom)

					if data.Commits != nil {

						commits_mutex.RLock()

						var _, ok = combbases[key]

						commits_mutex.RUnlock()

						data.Commits[fmt.Sprintf("%x", key)] = struct {
							Height    uint32
							CommitNum uint32
							TxNum     uint16
							OutNum    uint16
							Combbase  bool
						}{
							Height:    utag.height,
							CommitNum: utag.commitnum,
							TxNum:     utag.txnum,
							OutNum:    utag.outnum,
							Combbase:  ok,
						}
						if len(data.Commits) > MaxCommitmentsArray {
							data.Commits = nil
						}
					}

				}
			}

			// clean map
			if len(commits) > 0 {
				commits = make(map[[32]byte]utxotag)
			}

		}
	}

	if data.Commits != nil {
		data.Bloom = nil
	}

	_, _ = w.Write([]byte("p("))
	_ = json.NewEncoder(w).Encode(data)
	_, _ = w.Write([]byte(");"))
}

func controller_main_page_read(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	http.Redirect(w, r, "ui/index.html", 301)
	_, _ = w.Write([]byte(`<meta http-equiv="refresh" content="0; url=ui/index.html">`))
}

func controller_send_file(w http.ResponseWriter, name, ctype string) {
	data, _ := content.ReadFile(name)
	w.Header().Set("Content-Type", ctype)
	_, _ = w.Write(data)
}

// send file cached can be only used to send files (or data) which literally never changes during the wallet runtime
func controller_send_file_cached(w http.ResponseWriter, name, ctype, r_url_path string) {
	// we anyway switch only using the part after the last slash and before the last dot
	for i := len(r_url_path) - 1; i >= 0; i-- {
		if r_url_path[i] == '.' {
			r_url_path = r_url_path[:i]
			break
		}
	}
	for i := len(r_url_path) - 1; i >= 0; i-- {
		if r_url_path[i] == '/' {
			r_url_path = r_url_path[i:]
			break
		}
	}

	w.Header().Set("Content-Type", ctype)
	data, _ := content.ReadFile(name)
	h := sha256.Sum256(data)
	ph := sha256.Sum256([]byte(r_url_path))
	hash_computed := "sha-256=:" + base64.StdEncoding.EncodeToString(h[:]) + ":"
	w.Header().Set("Repr-Digest", hash_computed)
	w.Header().Set("Digest", hash_computed)
	GzipStoreUrlPathHashDigest(ph, hash_computed, ctype, len(data))
	_, _ = w.Write(data)
}

/*
	func controller_send_disk_file(w http.ResponseWriter, name, ctype string) {
		data, _ := os.ReadFile(name)
		w.Header().Set("Content-Type", ctype)
		_, _ = w.Write(data)
	}
*/
func controller_send_proxy_file(w http.ResponseWriter, name, ctype string) {
	resp, err := http.Get(name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", ctype)
		// Write the response back with all headers
		for key, value := range resp.Header {
			w.Header()[key] = value
		}
		w.WriteHeader(resp.StatusCode)
		_, err = io.Copy(w, resp.Body)
	}

}

func public_comb_protocol(w http.ResponseWriter, r *http.Request) {
	if !get_init_proxy_is_over() && strings.HasSuffix(r.URL.Path, init_proxy_nonce) {
		return
	}

	var paths = cuturl(r.URL.Path)

	w.Header().Set("Access-Control-Allow-Origin", "*")

	switch len(paths) {
	case 0:
		controller_main_page_read(w, r)
		return
	case 1: // get height call

		switch paths[0] {
		case "main":
			if get_init_proxy_is_over() {
				main_page_json(w, r)
			} else {
				controller_send_proxy_file(w, u_config.init_proxy+
					"main.js?nonce="+init_proxy_nonce, "text/javascript")
			}
			return
		case "wasm_exec":
			controller_send_file_cached(w, "ui/wasm_exec.js", "text/javascript", r.URL.Path)
			return
		case "combfullui":
			controller_send_file_cached(w, "ui/combfullui.wasm", "application/wasm", r.URL.Path)
			return
		case "index":
			controller_send_file_cached(w, "ui/index.html", "text/html", r.URL.Path)
			return
		case "style":
			controller_send_file_cached(w, "ui/style.css", "text/css", r.URL.Path)
			return
		case "w3":
			controller_send_file_cached(w, "ui/w3.css", "text/css", r.URL.Path)
			return
		case "doc":
			controller_send_file_cached(w, "ui/doc.html", "text/html", r.URL.Path)
			return
		case "swagger":
			controller_send_file_cached(w, "ui/swagger.yaml", "application/x-yaml", r.URL.Path)
			return
		case "robots":
			controller_send_file_cached(w, "ui/robots.txt", "text/plain", r.URL.Path)
			return
		case "sitemap":
			controller_send_file_cached(w, "ui/sitemap.xml", "application/xml", r.URL.Path)
			return
		case "chartdisk":
			controller_send_proxy_file(w, u_config.chart_proxy+"chartdisk.js", "text/javascript")
			return
		case "chartmempool":
			controller_send_proxy_file(w, u_config.chart_proxy+"chartmempool.js", "text/javascript")
			return
		default:
			if get_init_proxy_is_over() {
				controller_getblock(paths[0], w, r)
			} else {
				controller_send_proxy_file(w, u_config.init_proxy+
					paths[0]+".js?nonce="+init_proxy_nonce, "text/javascript")
			}
			return
		}

	case 2: // get base range call

		switch paths[0] {
		case "wasm_exec":
			controller_send_file(w, "ui/wasm_exec."+paths[1]+".js", "text/javascript")
			return
		case "index":
			controller_send_file(w, "ui/index."+paths[1]+".html", "text/html")
			return
		case "style":
			controller_send_file(w, "ui/style."+paths[1]+".css", "text/css")
			return
		case "w3":
			controller_send_file(w, "ui/w3."+paths[1]+".css", "text/css")
			return
		default:
			if get_init_proxy_is_over() {
				controller_getrange(paths[0], paths[1], w, r)
			} else {
				controller_send_proxy_file(w, u_config.init_proxy+
					paths[0]+"."+paths[1]+".js?nonce="+init_proxy_nonce, "text/javascript")
			}

			return
		}

	case 5: // get commits by verbatim prefixes bloom-less call
		if get_init_proxy_is_over() {
			controller_getcommitsbyprefixes(paths[0], paths[1], paths[2], paths[3], paths[4], w, r)
		} else {
			controller_send_proxy_file(w, u_config.init_proxy+
				paths[0]+"."+paths[1]+"."+paths[2]+"."+paths[3]+"."+paths[4]+".js?nonce="+
				init_proxy_nonce, "text/javascript")
		}
		return
	case 7: // get commits by verbatim prefixes bloom-less min call
		if get_init_proxy_is_over() {
			controller_getcommitsbyprefixesmin(paths[0], paths[1], paths[2], paths[3], paths[4], paths[5], paths[6], w, r)
		} else {
			controller_send_proxy_file(w, u_config.init_proxy+
				paths[0]+"."+paths[1]+"."+paths[2]+"."+paths[3]+"."+paths[4]+"."+paths[5]+"."+paths[6]+".js?nonce="+
				init_proxy_nonce, "text/javascript")
		}
		return
	default: // bug/unimplemented call
		return
	}
}
