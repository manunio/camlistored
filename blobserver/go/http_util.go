package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func badRequestError(conn http.ResponseWriter, errorMessage string) {
	conn.WriteHeader(http.StatusBadRequest)
	_, _ = fmt.Fprintf(conn, "%s\n", errorMessage)
}

func serverError(conn http.ResponseWriter, err error) {
	conn.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintf(conn, "Server error: %s\n", err)
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
