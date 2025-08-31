package db

import (
	"crypto/rand"
	"database/sql"
	"log"
	"math/big"

	"github.com/Alter-Sitanshu/campaignHub/internals/env"
	"github.com/joho/godotenv"
)

var MockDB *sql.DB

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env", err.Error())
	}
	MockDB, err = Mount(
		env.GetString("DB_ADDR", ""),
		5, // max connections to db
		2, // max idle connections
		1, // 1 min idle time
	)
	if err != nil {
		log.Fatal("Error loading .env", err.Error())
	}
}

func randString(size int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var output string

	for range size {
		// pick a random index
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			log.Printf("error generating random int: %v", err)
			return ""
		}
		output += string(letters[n.Int64()])
	}
	return output
}
