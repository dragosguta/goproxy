package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func getEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	panic(fmt.Sprintf("unable to read environment key: %s", key))
}

func stringifyAndLog(item interface{}) string {
	jsoned, err := json.Marshal(item)

	if err != nil {
		log.Println(err)
		log.Println("unable to JSON stringify item")
		return ""
	}

	stringified := string(jsoned)
	log.Println(stringified)

	return stringified
}

func convertJSONKeys(j json.RawMessage, t string) json.RawMessage {
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
		m[fixed] = convertJSONKeys(v, t)
	}

	b, err := json.Marshal(m)
	if err != nil {
		return j
	}

	return json.RawMessage(b)
}
