package main

import (
	"encoding/json"
	"html"
	"log"
	"net/http"
	"time"
)

type incomingRequestLogItem struct {
	Address   string            `json:"address"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	Timestamp int64             `json:"timestamp"`
}

func transformHeaders(headers http.Header) map[string]string {
	parsedHeaders := make(map[string]string, 10)
	for name, values := range headers {
		// Don't log out the token
		if name == "Authorization" {
			parsedHeaders[name] = "true"
			continue
		}
		parsedHeaders[name] = values[0]
	}

	// If no auth header is present, set to false
	if _, ok := parsedHeaders["Authorization"]; !ok {
		parsedHeaders["Authorization"] = "false"
	}

	return parsedHeaders
}

func logIncoming(req *http.Request) (string, error) {
	parsedHeaders := transformHeaders(req.Header)
	item := incomingRequestLogItem{
		Address:   req.RemoteAddr,
		Method:    req.Method,
		Path:      html.EscapeString(req.URL.Path),
		Headers:   parsedHeaders,
		Timestamp: time.Now().Unix(),
	}

	jsoned, err := json.Marshal(item)

	if err != nil {
		return "", err
	}

	stringified := string(jsoned)
	log.Println(stringified)

	return stringified, nil
}
