package db

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if os.Getenv("DOCKER_ENV") != "true" {
		if err := godotenv.Load("../../../.env"); err != nil {
			log.Println("No .env file found — relying on environment variables")
		} else {
			log.Println(".env file loaded successfully")
		}
	}
	Init()
	os.Exit(m.Run())
}
