package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

// IsJSON check for request body
func IsJSON(content []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(content, &js) == nil
}

func validJSONRequestBody(req *http.Request) (bool, error) {
	if req.Method == "POST" || req.Method == "PUT" {
		body, err := ioutil.ReadAll(req.Body)

		if err != nil {
			log.Println(err)
			log.Println("unable to read request body for JSON validation")
			return false, err
		}

		// Close the body since we read it, and in case it's not valid JSON
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		if !IsJSON(body) {
			log.Println("invalid JSON in request body")
			return false, nil
		}

		body = convertKeys(json.RawMessage(body), "camel")

		// Set content type header since we validated that it is JSON, calculate lengths
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", strconv.Itoa(len(body)))
		req.ContentLength = int64(len(body))

		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

		return true, nil
	}

	return false, nil
}
