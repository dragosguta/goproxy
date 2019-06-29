package main

import (
	"fmt"
	"os"
)

func getEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	panic(fmt.Sprintf("Unable to read environment key: %s", key))
}
