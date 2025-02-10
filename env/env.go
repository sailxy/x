package env

import (
	"os"

	"github.com/joho/godotenv"
)

func Load(filenames ...string) error {
	return godotenv.Load(filenames...)
}

func Get(key string) string {
	return os.Getenv(key)
}
