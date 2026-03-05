package env

import (
	"fmt"
	"os"
	"strconv"
)

func MustGet(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("missing required env var: %s", key))
	}
	return value
}

func GetWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetIntWithDefault(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return parsed
}
