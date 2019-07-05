package main

import (
	"net/http"
	"net/url"
)

type requestLogItem struct {
	Host          string                 `json:"host"`
	Address       string                 `json:"address"`
	Headers       map[string]string      `json:"headers"`
	Method        string                 `json:"method"`
	RequestURI    string                 `json:"requestURI"`
	Proto         string                 `json:"proto"`
	UserAgent     string                 `json:"userAgent"`
	ContentLength int64                  `json:"contentLength"`
	Query         url.Values             `json:"query"`
	RequestBody   map[string]interface{} `json:"requestBody"`
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

	return parsedHeaders
}
