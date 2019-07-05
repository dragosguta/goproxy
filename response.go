package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type responseLogItem struct {
	StatusCode    int         `json:"statusCode"`
	Headers       http.Header `json:"headers"`
	ResponseBody  interface{} `json:"responseBody"`
	ContentLength int64       `json:"contentLength"`
}

type transport struct {
	http.RoundTripper
}

func errorResponse(status int, message string, res http.ResponseWriter) {
	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.WriteHeader(status)
	json.NewEncoder(res).Encode(map[string]interface{}{"error": message, "data": nil})
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if IsJSON(b) {
		resp.Header.Set("Content-Type", "application/json")
		b = convertJSONKeys(json.RawMessage(b), "snake")
	} else {
		resp.Header.Set("Content-Type", "text/plain")
	}

	resp.Body = ioutil.NopCloser(bytes.NewReader(b))

	// Recalculate the content length
	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))
	resp.ContentLength = int64(len(b))

	// Remove unnecessary headers
	resp.Header.Del("Authorization")
	resp.Header.Del("X-Powered-By")

	// Convert response body to JSON
	var placeholder map[string]interface{}
	err = json.Unmarshal(b, &placeholder)

	if err != nil {
		log.Println(err)
		log.Println("unable to convert response body to JSON")
		return resp, nil
	}

	item := responseLogItem{
		StatusCode:    resp.StatusCode,
		Headers:       resp.Header,
		ContentLength: resp.ContentLength,
		ResponseBody:  placeholder,
	}

	stringifyAndLog(item)

	return resp, nil
}
