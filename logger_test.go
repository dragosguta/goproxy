package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestTransformHeaders(t *testing.T) {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Authorization", "some_token_123")
	headers.Set("User-Agent", "golang test agent")
	headers.Set("Accept", "*/*")

	parsedHeaders := transformHeaders(headers)

	for name, value := range parsedHeaders {
		if name == "Authorization" && value != "true" {
			t.Fail()
		}
		if name != "Authorization" && headers.Get(name) != parsedHeaders[name] {
			t.Fail()
		}
	}

	headers.Del("Authorization")

	newHeaders := transformHeaders(headers)
	for name, value := range newHeaders {
		if name == "Authorization" && value != "false" {
			t.Fail()
		}
	}
}

func TestLogIncoming(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/server/test", nil)

	req.Header.Set("User-Agent", "test_agent")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "test_token")

	jsonLogItem, err := logIncoming(req)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	item := incomingRequestLogItem{}
	err = json.Unmarshal([]byte(jsonLogItem), &item)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	if item.Method == "" {
		t.Fail()
	}

	if item.Path == "" {
		t.Fail()
	}

	if item.Headers == nil {
		t.Fail()
	}

	if item.Timestamp == 0 {
		t.Fail()
	}
}
