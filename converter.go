package main

import (
	"encoding/json"
)

func convertKeys(j json.RawMessage, t string) json.RawMessage {
	m := make(map[string]json.RawMessage)
	if err := json.Unmarshal([]byte(j), &m); err != nil {
		// Not a JSON object
		return j
	}

	for k, v := range m {
		fixed := k

		if t == "snake" {
			fixed = ToSnake(k)
		} else if t == "camel" {
			fixed = ToLowerCamel(k)
		}

		delete(m, k)
		m[fixed] = convertKeys(v, t)
	}

	b, err := json.Marshal(m)
	if err != nil {
		return j
	}

	return json.RawMessage(b)
}
