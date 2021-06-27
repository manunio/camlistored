package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
)

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
