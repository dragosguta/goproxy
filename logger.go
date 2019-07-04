package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type incomingRequestLogItem struct {
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

type outgoingRequestLogItem struct {
	StatusCode    int               `json:"statusCode"`
	Headers       map[string]string `json:"headers"`
	ResponseBody  interface{}       `json:"responseBody"`
	ContentLength int64             `json:"contentLength"`
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

func readAndParseBody(b io.ReadCloser, t string) (io.ReadCloser, map[string]interface{}, error) {
	body, err := ioutil.ReadAll(b)

	if err != nil {
		log.Println(err)
		log.Println("unexpected error when parsing HTTP body")
		return nil, nil, err
	}

	resetBody := ioutil.NopCloser(bytes.NewBuffer(body))

	if !IsJSON(body) {
		log.Println("response body is not JSON")
		return resetBody, nil, err
	}

	var placeholder map[string]interface{}
	err = json.Unmarshal(body, &placeholder)

	if err != nil {
		log.Println(err)
		log.Printf("unable to parse JSON in HTTP body %s", t)
		return resetBody, nil, err
	}

	return resetBody, placeholder, nil
}

func stringifyAndLog(item interface{}) string {
	jsoned, err := json.Marshal(item)

	if err != nil {
		log.Println(err)
		log.Println("unable to JSON stringify log item")
		return ""
	}

	stringified := string(jsoned)
	log.Println(stringified)

	return stringified
}

func logRequest(req *http.Request) string {
	item := incomingRequestLogItem{
		Host:          req.Host,
		Address:       req.RemoteAddr,
		Headers:       transformHeaders(req.Header),
		Method:        req.Method,
		RequestURI:    req.RequestURI,
		Proto:         req.Proto,
		UserAgent:     req.Header.Get("User-Agent"),
		ContentLength: req.ContentLength,
		Query:         req.URL.Query(),
	}

	if req.Method == "POST" || req.Method == "PUT" {
		if resetBody, requestBody, err := readAndParseBody(req.Body, "request"); err == nil {
			req.Body = resetBody
			item.RequestBody = requestBody
		}
	}

	return stringifyAndLog(item)
}

func logResponse(res *http.Response) string {
	item := outgoingRequestLogItem{
		StatusCode:    res.StatusCode,
		Headers:       transformHeaders(res.Header),
		ContentLength: res.ContentLength,
	}

	if resetBody, responseBody, err := readAndParseBody(res.Body, "response"); err == nil {
		res.Body = resetBody
		item.ResponseBody = responseBody
	}

	return stringifyAndLog(item)
}
