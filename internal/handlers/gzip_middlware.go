package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func GzipDecompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		fmt.Println("GzipDecompressHandler", string(bodyBytes))
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {

			textDecoded := make([]byte, len(bodyBytes))
			_, err = base64.RawStdEncoding.Decode(textDecoded, bodyBytes)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			fmt.Println("bytes", string(textDecoded))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			gzR, err := gzip.NewReader(bytes.NewReader(textDecoded))
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
