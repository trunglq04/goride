/*
Package env provides a simple way to get environment variables.
*/
package env

import (
	"os"
	"strconv"
)

// Callback is an injected function called when an environment variable is unset or empty.
type Callback func(key string)

// GetString returns the value of the environment variable named by key.
// If the variable is not set or is empty (""), fallback is returned and any
// provided onEmpty callbacks are executed.
func GetString(key, fallback string, onEmpty ...Callback) string {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		for _, fn := range onEmpty {
			if fn != nil {
				fn(key)
			}
		}
		return fallback
	}

	return val
}

func GetInt(key string, fallback int, onEmpty ...Callback) int {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		for _, fn := range onEmpty {
			if fn != nil {
				fn(key)
			}
		}
		return fallback
	}

	valAsInt, err := strconv.Atoi(val)
	if err != nil {
		for _, fn := range onEmpty {
			if fn != nil {
				fn(key)
			}
		}
		return fallback
	}

	return valAsInt
}

func GetBool(key string, fallback bool, onEmpty ...Callback) bool {
	val, ok := os.LookupEnv(key)
	if !ok || val == "" {
		for _, fn := range onEmpty {
			if fn != nil {
				fn(key)
			}
		}
		return fallback
	}

	boolVal, err := strconv.ParseBool(val)
	if err != nil {
		for _, fn := range onEmpty {
			if fn != nil {
				fn(key)
			}
		}
		return fallback
	}

	return boolVal
}
