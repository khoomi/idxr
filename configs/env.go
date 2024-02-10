package configs

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

func LoadEnvFor(v string) (x string) {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Unable to load .env file")
	}

	x = os.Getenv(v)
	return
}
