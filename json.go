package main

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	// Default character encoding.
	defaultCharset = "; charset=UTF-8"
	// ContentJSON header value for JSON data.
	ContentJSON = "application/json" + defaultCharset
	// ContentType header constant.
	ContentType = "Content-Type"
)

// Error exposes an error message
type Error struct {
	Error string `json:"error"`
}

func renderJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set(ContentType, ContentJSON)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Can't render to json: %s", err)
	}
}
