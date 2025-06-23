package main

import (
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"github.com/andybalholm/brotli"
	"io"
	"net/http"
	"strings"
	"sync"
)

var gzip_cache_mutex sync.RWMutex
var brotli_cache = make(map[string][]byte)
var gzip_cache = make(map[string][]byte)

var gzip_url_cache_mutex sync.RWMutex
var gzip_url_path_hash = make(map[[32]byte][3]string)

func GzipUnstorePathHash(hash [32]byte) {
	gzip_url_cache_mutex.Lock()
	var values = gzip_url_path_hash[hash]
	delete(gzip_url_path_hash, hash)
	gzip_url_cache_mutex.Unlock()

	gzip_cache_mutex.Lock()
	delete(brotli_cache, values[0])
	delete(gzip_cache, values[0])
	gzip_cache_mutex.Unlock()
}

func GzipStoreUrlPathHashDigest(hash [32]byte, digest, content_type string, size int) {
	gzip_url_cache_mutex.Lock()
	gzip_url_path_hash[hash] = [3]string{digest, content_type, fmt.Sprint(size)}
	gzip_url_cache_mutex.Unlock()
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

type compressWrapperResponseWriter struct {
	io.Writer
	gzip_key   string
	brotli_key string
}

func (w *compressWrapperResponseWriter) Write(b []byte) (i int, err error) {
	if len(w.gzip_key) > 0 {
		// the caller holds the cache write mutex
		gzip_cache[w.gzip_key] = append(gzip_cache[w.gzip_key], b...)
	}
	if len(w.brotli_key) > 0 {
		// the caller holds the cache write mutex
		brotli_cache[w.brotli_key] = append(brotli_cache[w.brotli_key], b...)
	}
	i, err = w.Writer.Write(b)
	if err != nil && err != io.EOF {
		// clean up half processed cache records on user cancellations
		// we will do a complete cache record the next time
		if len(w.gzip_key) > 0 {
			delete(gzip_cache, w.gzip_key)
		}
		if len(w.brotli_key) > 0 {
			delete(brotli_cache, w.brotli_key)
		}
	}
	return
}

// source: https://gist.github.com/the42/1956518
// but modded :)
func makeGzipHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/chartmempool.js") {
			fn(w, r)
			return
		}

		can_provide_gzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		can_provide_brotli_brotli := strings.Contains(r.Header.Get("Accept-Encoding"), "brotli")
		can_provide_brotli_br := strings.Contains(", "+r.Header.Get("Accept-Encoding")+", ", ", br, ")
		can_provide_brotli := can_provide_brotli_brotli || can_provide_brotli_br

		if !can_provide_gzip && !can_provide_brotli {
			fn(w, r)
			return
		}
		if can_provide_brotli_brotli {
			w.Header().Set("Content-Encoding", "brotli")
			can_provide_gzip = false
		}
		if can_provide_brotli_br {
			w.Header().Set("Content-Encoding", "br")
			can_provide_gzip = false
		}
		if can_provide_gzip {
			w.Header().Set("Content-Encoding", "gzip")
			can_provide_brotli = false
		}
		// at this point client wants either brotli (preferred) or gzip (backup option)

		r_url_path := r.URL.Path

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

		pathHash := sha256.Sum256([]byte(r_url_path))

		gzip_url_cache_mutex.RLock()
		digest_and_ctype_and_size, is_digest_and_ctype := gzip_url_path_hash[pathHash]
		gzip_url_cache_mutex.RUnlock()

		digest := digest_and_ctype_and_size[0]
		ctype := digest_and_ctype_and_size[1]
		sizelog := len(digest_and_ctype_and_size[2])

		// this if enables the caching path, we cache after doing gzip/brotli with the highest compression
		if is_digest_and_ctype {

			w.Header().Set("Repr-Digest", digest)
			w.Header().Set("Digest", digest)
			w.Header().Set("Content-Type", ctype)
			w.Header().Set("Access-Control-Allow-Origin", "*")

			if can_provide_brotli {
				gzip_cache_mutex.RLock()
				data, is_cached := brotli_cache[digest]
				gzip_cache_mutex.RUnlock()
				if is_cached {
					w.Write(data)
					return
				}
				gz := brotli.NewWriterLevel(
					&compressWrapperResponseWriter{
						w, "", digest,
					},
					brotli.BestCompression-(sizelog/3),
				)
				gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
				gzip_cache_mutex.Lock()
				data, is_cached = brotli_cache[digest]
				if is_cached {
					gz.Close()
					gzip_cache_mutex.Unlock()

					w.Write(data)
					return
				}
				fn(&gzr, r)
				gz.Close()
				gzip_cache_mutex.Unlock()
				return
			}
			if can_provide_gzip {
				gzip_cache_mutex.RLock()
				data, is_cached := gzip_cache[digest]
				gzip_cache_mutex.RUnlock()
				if is_cached {
					w.Write(data)
					return
				}
				gz, err := gzip.NewWriterLevel(
					&compressWrapperResponseWriter{
						w, digest, "",
					},
					gzip.BestCompression-(sizelog/3),
				)
				if err != nil {
					println(err.Error())
					return
				}
				gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
				gzip_cache_mutex.Lock()
				data, is_cached = gzip_cache[digest]
				if is_cached {
					gz.Close()
					gzip_cache_mutex.Unlock()

					w.Write(data)
					return
				}
				fn(&gzr, r)
				gz.Close()
				gzip_cache_mutex.Unlock()
				return
			}
			return
		}
		// non cached path, gzip
		if can_provide_gzip {
			gz := gzip.NewWriter(w)
			defer gz.Close()
			gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
			fn(&gzr, r)
			return
		}
		// non cached path, brotli
		if can_provide_brotli {
			gz := brotli.NewWriter(w)
			defer gz.Close()
			gzr := gzipResponseWriter{Writer: gz, ResponseWriter: w}
			fn(&gzr, r)
			return
		}
	}
}