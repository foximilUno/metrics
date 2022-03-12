package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func GzipDecompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("catch", r.Header.Get("Content-Encoding"))
		if r.Header.Get("Content-Encoding") == "gzip" {
			rb, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			fromBase64Bytes, err := base64.StdEncoding.DecodeString(string(rb))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			gzR, err := gzip.NewReader(bytes.NewReader(fromBase64Bytes))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			b, err := ioutil.ReadAll(gzR)
			if err != nil {
				fmt.Println("read gzr bytes", err)
			}
			fmt.Println(string(b))
			defer func(gzR *gzip.Reader) {
				err := gzR.Close()
				if err != nil {
					log.Println(err)
				}
			}(gzR)
		}
		next.ServeHTTP(w, r)
	})
}
