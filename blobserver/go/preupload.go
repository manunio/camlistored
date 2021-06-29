package main

import (
	"fmt"
	"net/http"
	"os"
)

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
	ret["uploadUrlExpirationSeconds"] = 86400
	if len(req.Host) > 0 {
		scheme := "http" // TODO: https
		ret["uploadUrl"] = fmt.Sprintf("%s://%s/camli/upload", scheme, req.Host)
	} else {
		ret["uploadUrl"] = "/camli/upload"
	}

	returnJSON(conn, ret)

}
