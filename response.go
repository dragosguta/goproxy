package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

type transport struct {
	http.RoundTripper
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
		b = convertKeys(json.RawMessage(b), "snake")
	} else {
		resp.Header.Set("Content-Type", "text/plain")
	}

	body := ioutil.NopCloser(bytes.NewReader(b))

	resp.Body = body
	resp.ContentLength = int64(len(b))

	resp.Header.Set("Content-Length", strconv.Itoa(len(b)))

	resp.Header.Del("Authorization")
	resp.Header.Del("X-Powered-By")

	logResponse(resp)

	return resp, nil
}
