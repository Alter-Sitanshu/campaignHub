package db

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found for tests")
	}
	log.Printf("loaded env files for tests.\n")
	Init()
	os.Exit(m.Run())
}
