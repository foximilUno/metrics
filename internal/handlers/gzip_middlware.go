package handlers

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
)

func GzipDecompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("catch", r.Header.Get("Content-Encoding"))
		if r.Header.Get("Content-Encoding") == "gzip" {

			bb, err := ioutil.ReadAll(r.Body)
			gz_r, _ := gzip.NewReader(bytes.NewReader(bb))
			if err != nil {
				http.Error(w, "error while decompress", http.StatusInternalServerError)
				return
			}
			r.Body = gz_r
		}
		next.ServeHTTP(w, r)
	})
}
