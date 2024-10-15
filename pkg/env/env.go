package env

import (
	"os"
	"strconv"
)

func Get(key string, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

func GetBool(key string, def bool) bool {
	val := os.Getenv(key)
	if val == "TRUE" || val == "1" {
		return true
	}
	if val == "FALSE" || val == "0" {
		return false
	}
	return def
}

func GetInt(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}

	v, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return v
}
