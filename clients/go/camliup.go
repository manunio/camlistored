package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var flagFile = flag.String("file", "", "file to upload")
var flagServer = flag.String("server", "http://localhost:3179", "camlistore server")

type UploadHandler struct {
	blobRef  string
	contents io.ReadSeeker
}

// Upload agent
type Agent struct {
	server string
}

func NewAgent(server string) *Agent {
	return &Agent{server}
}

func (a *Agent) Upload(h *UploadHandler) {
	// TODO:
	url := fmt.Sprintf("%s/camli/preupload", a.server)
	fmt.Println("Need to upload: ", h, "to", url)

	e := func(msg string, e error) {
		_, _ = fmt.Fprintf(os.Stderr, "%s on %v: %v\n", msg, h.blobRef, e)
		return
	}

	resp, err := http.Post(
		url,
		"application/x-www-form-urlencoded",
		strings.NewReader("camliversion=1&blob1="+h.blobRef))
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Upload error for %v: %v\n",
			h.blobRef, err)
	}
	fmt.Println("Got response:", resp)
	buf := new(bytes.Buffer)
	_, _ = io.Copy(buf, resp.Body)
	_ = resp.Body.Close()

	pur := make(map[string]interface{})
	jerr := json.Unmarshal(buf.Bytes(), &pur)
	if jerr != nil {
		e("preupload parse error", jerr)
		return
	}
	uploadUrl, ok := pur["uploadUrl"].(string)
	if uploadUrl == "" {
		e("no uploadUrl in preupload response", nil)
		return
	}
	alreadyHave, ok := pur["alreadyHave"].([]interface{})
	if !(ok) {
		e("no alreadyHave array in preupload response", nil)
	}

	for _, haveObj := range alreadyHave {
		haveObj := haveObj.(map[string]interface{})
		if haveObj["blobRef"].(string) == h.blobRef {
			fmt.Println("already have it!")
			return
		}
	}
	fmt.Println("preupload done:", pur, alreadyHave)
}

func (a *Agent) Wait() int {
	// 	TODO:
	return 0
}

func blobName(contents io.ReadSeeker) (name string, err error) {
	s1 := sha1.New()
	if _, err := contents.Seek(0, 0); err != nil {
		return
	}
	if _, err := io.Copy(s1, contents); err != nil {
		return
	}
	return fmt.Sprintf("sha1-%x", s1.Sum(nil)), nil
}

func uploadFile(agent *Agent, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	blobRef, err := blobName(file)
	if err != nil {
		return err
	}
	fmt.Println("blob is: ", blobRef)
	handle := &UploadHandler{blobRef, file}
	agent.Upload(handle)
	return nil
}

func main() {
	flag.Parse()

	// remove trailing slash if provided
	if strings.HasSuffix(*flagServer, "/") {
		*flagServer = (*flagServer)[0 : len(*flagServer)-1]
	}

	agent := NewAgent(*flagServer)
	if *flagFile != "" {
		if err := uploadFile(agent, *flagFile); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "[ERROR] camliup: %v\n", err)
			os.Exit(1)
		}

	}

	stats := agent.Wait()
	fmt.Println("Done uploading stats: ", stats)
}
