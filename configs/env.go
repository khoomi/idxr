package configs

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

func LoadEnvFor(v string) (x string) {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Unable to load .env file")
	}

	x = os.Getenv(v)
	return
}
