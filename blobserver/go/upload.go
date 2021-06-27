package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/manunio/camlistored/blobserver/go/util"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func handleMultiPartUpload(conn http.ResponseWriter, req *http.Request) {
	if !(req.Method == "POST" && req.URL.Path == "/camli/upload") {
		badRequestError(conn, "In-configured handler.")
	}

	multipart, err := req.MultipartReader()
	if multipart == nil {
		badRequestError(conn,
			fmt.Sprintf("Expected mutltipart/form-data POST POST request: %v", err))
		return
	}

	if err != nil {
		serverError(conn, err)
		return
	}

	for {
		part, err := multipart.NextPart()
		if err != nil {
			fmt.Println("Error reading multipart section: ", err)
			break
		}
		if part == nil {
			break
		}
		formName := part.FormName()
		fmt.Printf("New value [%s], part=%v\n", formName, part)

		ref := ParseBlobRef(formName)
		if ref == nil {
			fmt.Printf("Ignoring form key [%s]\n", formName)
			continue
		}
		ok, err := receiveBlob(ref, part)
		if !ok {
			fmt.Printf("Error receiving blob %v: %v\n", ref, err)
		} else {
			fmt.Printf("Received blob %v\n", ref)
		}
	}
	fmt.Println("Done reading multipart body.")
}

func receiveBlob(blobRef *BlobRef, source io.Reader) (ok bool, err error) {
	hashedDirectory := blobRef.DirectoryName()
	if err = os.MkdirAll(hashedDirectory, 0700); err != nil {
		return
	}

	var tempFile *os.File
	tempFile, err = ioutil.TempFile(hashedDirectory, blobRef.FileBaseName()+".tmp")
	if err != nil {
		return
	}

	success := false // set to true later

	defer func() {
		if !success {
			fmt.Println("Removing temp file: ", tempFile.Name())
			_ = os.Remove(tempFile.Name())
		}
	}()

	sha1Hash := sha1.New()
	var written int64
	written, err = io.Copy(util.NewTee(sha1Hash, tempFile), source)
	if err != nil {
		return
	}
	if err = tempFile.Close(); err != nil {
		return
	}

	fileName := blobRef.FileName()
	if err = os.Rename(tempFile.Name(), fileName); err != nil {
		return
	}

	stat, err := os.Lstat(fileName)
	if err != nil {
		return
	}
	if !stat.Mode().IsRegular() || stat.Size() != written {
		return false, errors.New("written size didn't match")
	}
	success = true
	return true, nil
}

func handlePut(conn http.ResponseWriter, req *http.Request) {
	blobRef := ParsePath(req.URL.Path)
	if blobRef == nil {
		badRequestError(conn, "Malformed PUT URL.")
		return
	}

	if !blobRef.IsSupported() {
		badRequestError(conn, "unsupported object hash function")
		return
	}

	if _, err := receiveBlob(blobRef, req.Body); err != nil {
		serverError(conn, err)
		return
	}
	_, _ = fmt.Fprintf(conn, "OK")
}
