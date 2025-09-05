package main

import (
	"log"

	"github.com/Alter-Sitanshu/campaignHub/api"
	"github.com/Alter-Sitanshu/campaignHub/env"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, relying on system environment")
	}

	config := api.Config{
		Addr: env.GetString("PORT", ":8080"), // default port 8080
		DbCfg: api.DBConfig{
			ADDR:         env.GetString("DB_ADDR", ""),
			MaxConns:     env.GetInt("MAXCONN", 10),
			MaxIdleConns: env.GetInt("MAXIDLECONN", 5),
			MaxIdleTime:  env.GetInt("MAXIDLETIME", 5), // 5 Minutes -> Converted into an interval internally
		},
	}

	// Making DB connection
	db_, err := db.Mount(
		config.DbCfg.ADDR,
		config.DbCfg.MaxConns,
		config.DbCfg.MaxIdleConns,
		config.DbCfg.MaxIdleTime,
	)
	if err != nil {
		log.Fatalf("Error connecting to DB %v\n", err.Error())
	}

	appStore := db.NewStore(db_)
	app := api.NewApplication(
		appStore,
		&config,
	)

	app.Run()
}
