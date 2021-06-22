package main

import (
	"flag"
	"fmt"
	"net/http"
	"regexp"
)

var listen *string = flag.String("listen", "0.0.0.0:3179", "host:port to listen on")
var storageRoot *string = flag.String("root", "/tmp/camliroot", "Root directory to store files")

var sharedSecret string

var putPattern *regexp.Regexp = regexp.MustCompile(`^camli/{sha1}-{[a-f0-9]+}$`)

func badRequestError(conn http.ResponseWriter, errorMessage string) {
	conn.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(conn, "%s\n", errorMessage)
}

func serverError(conn http.ResponseWriter, err error) {
	conn.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(conn, "Server error: %s\n", err)
}

func handleCamli(conn http.ResponseWriter, req *http.Request) {
	if req.Method == "PUT" {
		handlePut(conn, req)
		return
	}
}
