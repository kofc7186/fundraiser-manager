package util

import (
	"fmt"
	"os"
)

func GetEnvOrPanic(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("%s not found", key))
	}
	return val
}
