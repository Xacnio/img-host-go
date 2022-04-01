package configs

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	HASH1 = "dfgdff-dfhgfd"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		// do nothing
	}
}

func Get(key string) string {
	return os.Getenv(key)
}

func GetInt(key string, defaultValue int) int {
	if os.Getenv(key) == ""  {
		return defaultValue
	}
	i, err := strconv.ParseInt(os.Getenv(key), 10, 32)
	if err != nil {
		return defaultValue
	}
	return int(i)
}

func GetInt64(key string) int64 {
	i, _ := strconv.ParseInt(os.Getenv(key), 10, 64)
	return i
}