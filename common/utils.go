package common

import (
	"errors"
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

func GetEnv(key string) (string, error) {
	value, exists := os.LookupEnv(key)
	if !exists {
		return "", errors.New("Environment Variable Not Set: " + key)
	}
	return value, nil
}

func FormatClientKey(status, id string) string {
	return fmt.Sprintf("client:%s:%s", status, id)
}
