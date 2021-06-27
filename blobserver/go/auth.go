package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

var basicAuthPattern = regexp.MustCompile(`^Basic ([a-zA-Z0-9+/=]+)`)

var accessPassword string

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

// requireAuth wraps a function to another function that encforces
// HTTP Basic Auth.
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
