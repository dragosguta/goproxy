package main

import "encoding/json"

// IsJSON check for request body
func IsJSON(content []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(content, &js) == nil
}
