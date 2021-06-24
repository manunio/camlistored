package main

import (
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

var listen *string = flag.String("listen", "0.0.0.0:3179", "host:port to listen on")
var storageRoot *string = flag.String("root", "/tmp/camliroot", "Root directory to store files")

var putPassword string

var getPutPattern *regexp.Regexp = regexp.MustCompile(`^/camli/(sha1)-([a-f0-9]+)$`)
var basicAuthPattern *regexp.Regexp = regexp.MustCompile(`^Basic ([a-zA-Z0-9\+/=]+)`)

// BlobRef ...
type BlobRef struct {
	HashName string
	Digest   string
}

func putAllowed(req *http.Request) bool {
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return false
	}
	matches := basicAuthPattern.FindAllStringSubmatch(auth, -1)
	if len(matches) != 1 || len(matches[0]) != 2 {
		return false
	}
	var outBuf []byte = make([]byte, base64.StdEncoding.DecodedLen(len(matches[0][1])))
	bytes, err := base64.StdEncoding.Decode(outBuf, []uint8(matches[0][1]))
	if err != nil {
		return false
	}
	password := string(outBuf)
	fmt.Println("Decode bytes:", bytes, " error: ", err)
	fmt.Println("Got userPass:", password)
	return password != "" && password == putPassword
}

// ParsePath ...
func ParsePath(path string) *BlobRef {
	groups := getPutPattern.FindAllStringSubmatch(path, -1)
	if len(groups) != 1 || len(groups[0]) != 3 {
		return nil
	}
	obj := &BlobRef{groups[0][1], groups[0][2]}
	if obj.HashName == "sha1" && len(obj.Digest) != 40 {
		return nil
	}
	return obj
}

// IsSupported ...
func (o *BlobRef) IsSupported() bool {
	return o.HashName == "sha1"
}

// Hash ...
func (o *BlobRef) Hash() hash.Hash {
	if o.HashName == "sha1" {
		return sha1.New()
	}
	return nil
}

// FileBaseName ...
func (o *BlobRef) FileBaseName() string {
	return fmt.Sprintf("%s-%s.dat", o.HashName, o.Digest)
}

// DirectoryName ...
func (o *BlobRef) DirectoryName() string {
	return fmt.Sprintf("%s/%s/%s", *storageRoot, o.Digest[0:3], o.Digest[3:6])
}

// FileName ...
func (o *BlobRef) FileName() string {
	return fmt.Sprintf("%s/%s-%s.dat", o.DirectoryName(), o.HashName, o.Digest)
}

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

	if req.Method == "GET" {
		handleGet(conn, req)
		return
	}

	badRequestError(conn, "unsupported method.")
}

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
		fmt.Fprintf(conn, "Object not found")
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
		fmt.Fprintf(os.Stderr, "Error sending file: %v err=%v\n", objRef, err)
		hj, ok := conn.(http.Hijacker)
		if !ok {
			serverError(conn, errors.New("webserver doesn't support hijacking"))
			return
		}
		closer, _, err := hj.Hijack()
		if err != nil {
			closer.Close()
		}
		return
	}

	if bytesCopied != stat.Size() {
		fmt.Fprintf(os.Stderr, "Error sending file: %v, copied= %d, not %v\n", objRef, bytesCopied, stat.Size())
		hj, ok := conn.(http.Hijacker)
		if !ok {
			serverError(conn, errors.New("webserver doesn't support hijacking"))
			return
		}
		closer, _, err := hj.Hijack()
		if err != nil {
			closer.Close()
		}
		return

	}

}

func handlePut(conn http.ResponseWriter, req *http.Request) {
	objRef := ParsePath(req.URL.Path)
	if objRef == nil {
		badRequestError(conn, "Malformed PUT URL.")
		return
	}

	if !objRef.IsSupported() {
		badRequestError(conn, "unsupported object hash function")
		return
	}

	if !putAllowed(req) {
		conn.Header().Set("WWW-Authentication", "Basic realm=\"camlistored\"")
		conn.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(conn, "Authentication required.")
		return

	}
	// TODO(manunio): auth/authz checks here

	hashedDirectory := objRef.DirectoryName()

	if err := os.MkdirAll(hashedDirectory, 0700); err != nil {
		serverError(conn, err)
		return
	}

	tempFile, err := ioutil.TempFile(hashedDirectory, objRef.FileBaseName()+".tmp")

	if err != nil {
		serverError(conn, err)
		return
	}

	success := false // set to true later

	defer func() {
		if !success {
			fmt.Println("Removing temp file: ", tempFile.Name())
			os.Remove(tempFile.Name())
		}
	}()

	written, err := io.Copy(tempFile, req.Body)
	if err != nil {
		serverError(conn, err)
		return
	}
	if _, err = tempFile.Seek(0, 0); err != nil {
		serverError(conn, err)
		return
	}

	hasher := objRef.Hash()

	if _, err := io.Copy(hasher, tempFile); err != nil {
		serverError(conn, err)
		return
	}
	if fmt.Sprintf("%x", hasher.Sum(nil)) != objRef.Digest {
		badRequestError(conn, "Digest didn't match as declared")
		return
	}
	if err = tempFile.Close(); err != nil {
		serverError(conn, err)
		return
	}

	fileName := objRef.FileName()
	if err = os.Rename(tempFile.Name(), fileName); err != nil {
		serverError(conn, err)
		return
	}

	stat, err := os.Lstat(fileName)
	if err != nil {
		serverError(conn, err)
		return
	}
	if !stat.Mode().IsRegular() || stat.Size() != written {
		serverError(conn, errors.New("written size didn't match"))
		// unlink it? Bogus? Naah, better to not lose data.
		// we can clean it up later in GC phase.
	}
	success = true
	fmt.Fprintf(conn, "OK")
}

// HandleRoot func
func HandleRoot(conn http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(conn, `This is camlistored, a Camlistore storage daemon`)
}

func main() {
	flag.Parse()

	putPassword = os.Getenv("CAMLI_PASSWORD")
	if len(putPassword) == 0 {
		fmt.Fprintf(os.Stderr, "No CAMLI_PASSWORD environment variable set. \n")
		os.Exit(1)
	}
	{
		fi, err := os.Stat(*storageRoot)
		if err != nil || !fi.IsDir() {
			fmt.Fprintf(os.Stderr, "Storage root '%s' doesn't exist", *storageRoot)
			os.Exit(1)
		}

	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", HandleRoot)
	mux.HandleFunc("/camli/", handleCamli)

	fmt.Printf("Starting to listen on http://%v\n", *listen)
	if err := http.ListenAndServe(*listen, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Error in http server: %v\n", err)
		os.Exit(1)
	}
}
