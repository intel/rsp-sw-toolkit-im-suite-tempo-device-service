package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"io"
	"log"
	"net/http"
)

func main() {
	port := flag.String("p", "80", "port to serve on")
	flag.Parse()

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		log.Printf("health check from %q", request.UserAgent())
	})
	http.Handle("/hcidump", hciHandler{})

	log.Printf("Serving on HTTP port: %s\n", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal(err)
	}
}

type hciHandler struct{}

func (hciHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	defer Drain(request.Body)

	buff := bytes.Buffer{}
	if err := IgnoreEOF(io.CopyN(&buff, request.Body, 200)); err != nil {
		log.Printf("error: %+v\n", err)
		writer.WriteHeader(500)
	}

	decoder := hex.NewDecoder(NewSpaceSkipReader(&buff))
	data := make([]byte, 100)
	n, err := decoder.Read(data)
	if err != nil {
		log.Printf("error: %+v\n", err)
		writer.WriteHeader(500)
	}

	tcd := TempoDiskCurrent{}
	if err := tcd.UnmarshalBinary(data[:n]); err != nil {
		return
	}

	log.Printf("%+v %02X\n", tcd, data[:n])
}
