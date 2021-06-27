package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func handleGet(conn http.ResponseWriter, req *http.Request) {
	objRef := ParsePath(req.URL.Path)
	if objRef == nil {
		badRequestError(conn, "Malformed GET URL.")
		return
	}
	fileName := objRef.FileName()
	stat, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		conn.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(conn, "Object not found")
		return
	}
	if err != nil {
		serverError(conn, err)
		return
	}
	file, err := os.Open(fileName)
	if err != nil {
		serverError(conn, err)
		return
	}
	conn.Header().Set("Content-Type", "application/ocstream")
	bytesCopied, err := io.Copy(conn, file)

	// If there's an error at this point, it's too late to tell the client,
	// as they've already receiving bytes. but they should be smart enough
	// to verify Digest doesn't match. But we close the (chunked) response anyway,
	// to further signal errors.
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error sending file: %v err=%v\n", objRef, err)
		hj, ok := conn.(http.Hijacker)
		if !ok {
			serverError(conn, errors.New("web server doesn't support hijacking"))
			return
		}
		closer, _, err := hj.Hijack()
		if err != nil {
			if closer != nil {
				_ = closer.Close()
			}
		}
		return
	}

	if bytesCopied != stat.Size() {
		_, _ = fmt.Fprintf(os.Stderr, "Error sending file: %v, copied= %d, not %v\n", objRef, bytesCopied, stat.Size())
		hj, ok := conn.(http.Hijacker)
		if !ok {
			serverError(conn, errors.New("web server doesn't support hijacking"))
			return
		}
		closer, _, err := hj.Hijack()
		if err != nil {
			if closer != nil {
				_ = closer.Close()
			}
		}
		return

	}

}
