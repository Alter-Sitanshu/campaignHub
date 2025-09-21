package main

import (
	"encoding/base64"
	"log"

	"github.com/Alter-Sitanshu/campaignHub/api"
	"github.com/Alter-Sitanshu/campaignHub/env"
	"github.com/Alter-Sitanshu/campaignHub/internals/auth"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/Alter-Sitanshu/campaignHub/internals/mailer"
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
		TokenCfg: api.TokenConfig{
			JWTSecret:        env.GetString("JWT_KEY", ""),
			PASETO_SYM:       env.GetString("PASETO_SYM", ""),
			PASETO_ASYM_PRIV: env.GetString("PASETO_ASYM_PRIV", ""),
			PASETO_ASYM_PUB:  env.GetString("PASETO_ASYM,_PUB", ""),
		},
		ISS: env.GetString("ISS", "admin"),
		AUD: env.GetString("AUD", "admin"),
		MailCfg: api.MailConfig{
			Host:        "smtp.gmail.com",
			Port:        587,
			Username:    env.GetString("MAIL_USERNAME", "admin"),
			Password:    env.GetString("MAIL_PASS", ""),
			From:        env.GetString("FROM_ACC", ""),
			MailRetries: 5,
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
	// Setting the Token Makers
	jwt_Maker, err := auth.NewJWTMaker(config.TokenCfg.JWTSecret)
	if err != nil {
		log.Fatalf("error making jwt maker: %v\n", err.Error())
	}
	paseto_key, _ := base64.StdEncoding.DecodeString(config.TokenCfg.PASETO_SYM)
	paseto_Maker, err := auth.NewPASETOMaker(paseto_key)
	if err != nil {
		log.Fatalf("error making paseto maker: %v\n", err.Error())
	}
	mailer := mailer.NewMailService(
		config.MailCfg.From,
		config.MailCfg.Host,
		config.MailCfg.Username,
		config.MailCfg.Password,
		config.MailCfg.Port,
	)
	log.Printf("%s: %s\n", mailer.Username, mailer.Password)
	appStore := db.NewStore(db_)
	app := api.NewApplication(
		appStore,
		&config,
		jwt_Maker,
		paseto_Maker,
		mailer,
	)
	app.Run()
}
