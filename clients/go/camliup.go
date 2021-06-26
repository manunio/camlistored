package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"os"
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

func (a *Agent) Upload(handle *UploadHandler) {
	// TODO:
	fmt.Println("Need to upload:", handle)
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
	blobname, err := blobName(file)
	if err != nil {
		return err
	}
	fmt.Println("blob is: ", blobname)
	handle := &UploadHandler{blobname, file}
	agent.Upload(handle)
	return nil
}

func main() {
	flag.Parse()
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
