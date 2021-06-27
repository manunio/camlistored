package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/manunio/camlistored/blobserver/go/util"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var listen *string = flag.String("listen", "0.0.0.0:3179", "host:port to listen on")
var storageRoot *string = flag.String("root", "/tmp/camliroot", "Root directory to store files")
var stealthMode *bool = flag.Bool("stealth", true, "Run in stealth mode.")

var accessPassword string

var basicAuthPattern = regexp.MustCompile(`^Basic ([a-zA-Z0-9+/=]+)`)

func isAuthorized(req *http.Request) bool {
	auth := req.Header.Get("Authorization")
	if auth == "" {
		return false
	}

	matches := basicAuthPattern.FindAllStringSubmatch(auth, -1)
	if len(matches) != 1 || len(matches[0]) != 2 {
		return false
	}

	encoded := matches[0][1]
	enc := base64.StdEncoding
	decBuff := make([]byte, enc.DecodedLen(len(encoded)))

	n, err := enc.Decode(decBuff, []byte(encoded))
	if err != nil {
		return false
	}

	userpass := strings.Split(string(decBuff[0:n]), ":")
	if len(userpass) != 2 {
		fmt.Println("didn't get two pieces")
		return false
	}
	password := userpass[1]
	return password != "" && password == accessPassword
}

func requireAuth(handler func(conn http.ResponseWriter, req *http.Request)) func(conn http.ResponseWriter,
	req *http.Request) {
	return func(conn http.ResponseWriter, req *http.Request) {
		if !isAuthorized(req) {
			conn.Header().Set("WWW-Authenticate", "Basic realm=\"camlistored\"")
			conn.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintf(conn, "Authentication required.\n")
			return
		}
		handler(conn, req)
	}
}

func badRequestError(conn http.ResponseWriter, errorMessage string) {
	conn.WriteHeader(http.StatusBadRequest)
	_, _ = fmt.Fprintf(conn, "%s\n", errorMessage)
}

func serverError(conn http.ResponseWriter, err error) {
	conn.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintf(conn, "Server error: %s\n", err)
}

func putAllowed(req *http.Request) bool {
	return isAuthorized(req)
}

func getAllowed(req *http.Request) bool {
	// for now
	return putAllowed(req)
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

func returnJSON(conn http.ResponseWriter, data interface{}) {
	bytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		badRequestError(conn, fmt.Sprintf("JSON serialization error: %v", err))
		return
	}
	if _, err = conn.Write(bytes); err != nil {
		serverError(conn, err)
		return
	}
	if _, err = conn.Write([]byte("\n")); err != nil {
		serverError(conn, err)
		return
	}
}

func handleCamli(conn http.ResponseWriter, req *http.Request) {
	handler := func(conn http.ResponseWriter, req *http.Request) {
		badRequestError(conn, "Unsupported path or method.")
	}
	switch req.Method {
	case "GET":
		handler = requireAuth(handleGet)
	case "POST":
		switch req.URL.Path {
		case "/camli/preupload":
			handler = requireAuth(handlePreUpload)
		case "/camli/upload":
			handler = requireAuth(handleMultiPartUpload)
		case "/camli/testform": // debug only
			handler = handleTestForm
		case "/camli/form": // debug only
			handler = handleCamliForm
		}
	case "PUT": // no longer part of spec
		handler = requireAuth(handlePut)
	}
	handler(conn, req)
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

func handlePreUpload(conn http.ResponseWriter, req *http.Request) {
	if !(req.Method == "POST" && req.URL.Path == "/camli/preupload") {
		badRequestError(conn, "Inconfigured handler.")
		return
	}
	if err := req.ParseForm(); err != nil {
		serverError(conn, err)
		return
	}
	camliVersion := req.FormValue("camliversion")
	if camliVersion == "" {
		badRequestError(conn, "No camliversion")
		return
	}
	n := 0
	var haveVector []*map[string]interface{}
	haveChan := make(chan *map[string]interface{})
	for {
		key := fmt.Sprintf("blob%v", n+1)
		value := req.FormValue(key)
		if value == "" {
			break
		}
		ref := ParseBlobRef(value)
		if ref == nil {
			badRequestError(conn, "Bogus blobref for key"+key)
			return
		}
		if !ref.IsSupported() {
			badRequestError(conn, "Unsupported or bogus blobref "+key)
			return
		}
		n++

		// parallel stat all the files
		go func() {
			fi, err := os.Stat(ref.FileName())
			if err == nil && fi.Mode().IsRegular() {
				info := make(map[string]interface{})
				info["blobRef"] = ref.String()
				info["size"] = fi.Size()
				haveChan <- &info
			} else {
				haveChan <- nil
			}
		}()
	}

	if n > 0 {
		for have := range haveChan {
			if have != nil {
				haveVector = append(haveVector, have)
			}
			n--
			if n == 0 {
				break
			}
		}
	}

	tmp := make([]*map[string]interface{}, len(haveVector))
	copy(tmp, haveVector)

	ret := make(map[string]interface{})
	ret["maxUploadSize"] = 2147483647 // 2GB.. *shrug* :p
	ret["alreadyHave"] = tmp
	ret["uploadUrl"] = "http://localhost:3179/camli/upload"
	ret["uploadUrlExpirationSeconds"] = 86400
	returnJSON(conn, ret)

}

func handleMultiPartUpload(conn http.ResponseWriter, req *http.Request) {
	if !(req.Method == "POST" && req.URL.Path == "/camli/upload") {
		badRequestError(conn, "In-configured handler.")
	}

	if !putAllowed(req) {
		conn.Header().Set("WWW-Authenticate", "Basic realm=\"camlistored\"")
		conn.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprintf(conn, "Authentication required.\n")
		return
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

	if !putAllowed(req) {
		conn.Header().Set("WWW-Authentication", "Basic realm=\"camlistored\"")
		conn.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(conn, "Authentication required.")
		return

	}

	// TODO(manun): auth/authz checks here
	if _, err := receiveBlob(blobRef, req.Body); err != nil {
		serverError(conn, err)
		return
	}
	_, _ = fmt.Fprintf(conn, "OK")
}

// HandleRoot ...
func HandleRoot(conn http.ResponseWriter, req *http.Request) {
	if *stealthMode {
		_, _ = fmt.Fprintf(conn, "Hi.\n")
	} else {
		_, _ = fmt.Fprintf(conn, `This is camlistored, a Camlistore storage daemon`)
	}
}

func main() {
	flag.Parse()

	accessPassword = os.Getenv("CAMLI_PASSWORD")
	if len(accessPassword) == 0 {
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
