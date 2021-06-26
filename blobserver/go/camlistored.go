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

var listen = flag.String("listen", "0.0.0.0:3179", "host:port to listen on")
var storageRoot = flag.String("root", "/tmp/camliroot", "Root directory to store files")

var putPassword string

var getPutPattern = regexp.MustCompile(`^/camli/(sha1)-([a-f0-9]+)$`)
var basicAuthPattern = regexp.MustCompile(`^Basic ([a-zA-Z0-9+/=]+)`)
var multipartContentPattern = regexp.MustCompile(`^multipart/form-data; boundary="?([^" ]+)"?`)
var blobRefPattern = regexp.MustCompile(`^([a-z0-9]+)-([a-f0-9]+)$`)

type MultiPartReader struct {
	boundary string
	reader   io.Reader
}

type MultiPartBodyPart struct {
	Header map[string]string
	Body   io.Reader
}

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
	var outBuf = make([]byte, base64.StdEncoding.DecodedLen(len(matches[0][1])))
	bytes, err := base64.StdEncoding.Decode(outBuf, []uint8(matches[0][1]))
	if err != nil {
		return false
	}
	password := string(outBuf)
	fmt.Println("Decode bytes:", bytes, " error: ", err)
	fmt.Println("Got userPass:", password)
	return password != "" && password == putPassword
}

func getAllowed(req *http.Request) bool {
	// for now
	return putAllowed(req)
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
	_, _ = fmt.Fprintf(conn, "%s\n", errorMessage)
}

func serverError(conn http.ResponseWriter, err error) {
	conn.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintf(conn, "Server error: %s\n", err)
}

func handleCamli(conn http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" && req.URL.Path == "/camli/upload" {
		handleMultiPartUpload(conn, req)
		return
	}

	if req.Method == "POST" && req.URL.Path == "/camli/testform" {
		handleTestForm(conn, req)
		return
	}

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

func handleCamliForm(conn http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(conn, `
<html>
<body>
<form method='POST' enctype="multipart/form-data" action="/camli/testform">
<input type="hidden" name="имя" value="брэд" />
Text unix: <input type="file" name="file-unix"><br>
Text win: <input type="file" name="file-win"><br>
Text mac: <input type="file" name="file-mac"><br>
Image png: <input type="file" name="image-png"><br>
<input type=submit>
</form>
</body>
</html>
`)
}

func handleTestForm(conn http.ResponseWriter, req *http.Request) {
	if !(req.Method == "POST" && req.URL.Path == "/camli/testform") {
		badRequestError(conn, "Inconfigured handler.")
		return
	}

	multipart, err := req.MultipartReader()
	if multipart == nil {
		badRequestError(conn, fmt.Sprintf("Expected multipart/form-data POST request; %v", err))
		return
	}

	if err != nil {
		serverError(conn, err)
		return
	}

	for {
		part, err := multipart.NextPart()
		if err != nil {
			fmt.Println("Error reading:", err)
			break
		}
		if part == nil {
			break
		}
		formName := part.FormName()
		fmt.Printf("New value [%s], part=%v\n", formName, part)

		s := sha1.New()
		if _, err = io.Copy(s, part); err != nil {
			serverError(conn, err)
			return
		}
		fmt.Printf("Got part digest %x\n", s.Sum(nil))
	}
	fmt.Println("Done reading multipart body.")
}

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

		matches := multipartContentPattern.FindAllStringSubmatch(formName, -1)
		if len(matches) != 1 || len(matches[0]) != 3 {
			fmt.Printf("Ignoring form key [%s]\n", formName)
			continue
		}
		ref := &BlobRef{matches[0][1], matches[0][2]}
		ok, err := receiveBlob(ref, part)
		if !ok {
			fmt.Printf("Error receiving blob %v: %v\n", ref, err)
		} else {
			fmt.Printf("Received blob %v\n", ref)
		}
	}
	fmt.Println("Done reading multipart body.")
}

func handleGet(conn http.ResponseWriter, req *http.Request) {
	if !getAllowed(req) {
		conn.Header().Set("WWW-Authentication", "Basic realm=\"camlistored\"")
		conn.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprintf(conn, "Authentication required.")
		return
	}
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

func receiveBlob(blobRef *BlobRef, source io.Reader) (ok bool, err error) {
	hashedDirectory := blobRef.DirectoryName()

	if err := os.MkdirAll(hashedDirectory, 0700); err != nil {
		return
	}

	tempFile, err := ioutil.TempFile(hashedDirectory, blobRef.FileBaseName()+".tmp")

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

	written, err := io.Copy(tempFile, source)
	if err != nil {
		return
	}
	if _, err = tempFile.Seek(0, 0); err != nil {
		return
	}

	hasher := blobRef.Hash()

	if _, err := io.Copy(hasher, tempFile); err != nil {
		return
	}
	if fmt.Sprintf("%x", hasher.Sum(nil)) != blobRef.Digest {
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
	if !putAllowed(req) {
		conn.Header().Set("WWW-Authentication", "Basic realm=\"camlistored\"")
		conn.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(conn, "Authentication required.")
		return

	}

	blobRef := ParsePath(req.URL.Path)
	if blobRef == nil {
		badRequestError(conn, "Malformed PUT URL.")
		return
	}

	if !blobRef.IsSupported() {
		badRequestError(conn, "unsupported object hash function")
		return
	}

	// TODO(manun): auth/authz checks here
	if _, err := receiveBlob(blobRef, req.Body); err != nil {
		serverError(conn, err)
		return
	}
	_, _ = fmt.Fprintf(conn, "OK")
}

// HandleRoot func
func HandleRoot(conn http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(conn, `This is camlistored, a Camlistore storage daemon`)
}

func main() {
	flag.Parse()

	putPassword = os.Getenv("CAMLI_PASSWORD")
	if len(putPassword) == 0 {
		_, _ = fmt.Fprintf(os.Stderr, "No CAMLI_PASSWORD environment variable set. \n")
		os.Exit(1)
	}
	{
		fi, err := os.Stat(*storageRoot)
		if err != nil || !fi.IsDir() {
			_, _ = fmt.Fprintf(os.Stderr, "Storage root '%s' doesn't exist", *storageRoot)
			os.Exit(1)
		}

	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", HandleRoot)
	mux.HandleFunc("/camli/", handleCamli)

	fmt.Printf("Starting to listen on http://%v\n", *listen)
	if err := http.ListenAndServe(*listen, mux); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error in http server: %v\n", err)
		os.Exit(1)
	}
}
