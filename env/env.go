package env

import (
	"os"

	"github.com/joho/godotenv"
)

func Load() error {
	return godotenv.Load()
}

func Get(key string) string {
	return os.Getenv(key)
}
