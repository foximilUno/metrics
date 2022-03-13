package handlers

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func GzipDecompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			bodyBytes, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			defer r.Body.Close()
			fmt.Println("GzipDecompressHandler", string(bodyBytes))
			gzR, err := gzip.NewReader(bytes.NewReader(bodyBytes))
			gzR.Multistream(false)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			resultBytes, err := ioutil.ReadAll(gzR)
			if err != nil {
				fmt.Print("error at stage 2: " + err.Error())
			}

			fmt.Println("GzipDecompressHandler after decompress", string(resultBytes))
			r.Body = ioutil.NopCloser(bytes.NewReader(resultBytes))
		}
		next.ServeHTTP(w, r)
	})
}
