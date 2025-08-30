package db

import (
	"database/sql"
	"log"

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
