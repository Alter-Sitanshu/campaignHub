package main

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/api"
	"github.com/Alter-Sitanshu/campaignHub/env"
	"github.com/Alter-Sitanshu/campaignHub/internals/auth"
	"github.com/Alter-Sitanshu/campaignHub/internals/cache"
	"github.com/Alter-Sitanshu/campaignHub/internals/chats"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/Alter-Sitanshu/campaignHub/internals/mailer"
	"github.com/Alter-Sitanshu/campaignHub/internals/platform"
	"github.com/Alter-Sitanshu/campaignHub/internals/workers.go"
	"github.com/joho/godotenv"
)

const (
	ShutdownDelta time.Duration = 1 * time.Minute
	BatchInterval time.Duration = 10 * time.Minute
	PollInterval  time.Duration = 15 * time.Minute
)

func main() {
	if os.Getenv("DOCKER_ENV") != "true" {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("No .env file found — relying on environment variables")
		} else {
			log.Println(".env file loaded successfully")
		}
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
			Support:     env.GetString("MAIL_SUPPORT", ""),
			MailRetries: 5,
		},
		// TODO: Change in Production
		RedisCfg: api.RedisConfig{
			Addr:     env.GetString("REDIS_ADDR", "localhost:6379"),
			Protocol: env.GetInt("REDIS_PROTOCOL", 3),
			Password: env.GetString("REDIS_PASSWORD", ""),
			DB:       env.GetInt("REDIS_DB", 0),
		},
		FactoryCfg: api.FactoryConfig{
			YouTubeAPIKey: env.GetString("YTAPIKEY", ""),
			MetaToken:     env.GetString("METAKEY", ""),
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
	// initialising the mailer service
	mailer := mailer.NewMailService(
		config.MailCfg.From,
		config.MailCfg.Host,
		config.MailCfg.Username,
		config.MailCfg.Password,
		config.MailCfg.Port,
	)
	// initialising the cache layer
	CacheClient, err := cache.NewClient(
		config.RedisCfg.Addr,
		config.RedisCfg.Password,
		config.RedisCfg.DB,
	)
	if err != nil {
		log.Fatalf("error intialising the cache layer: %v\n", err.Error())
	}
	// initialising the Media Factory
	factory, err := platform.NewFactory(
		config.FactoryCfg.YouTubeAPIKey,
		config.FactoryCfg.MetaToken,
	)
	if err != nil {
		log.Fatalf("error making media factory: %v\n", err.Error())
	}

	// attaching the services to the application
	appStore := db.NewStore(db_)
	appCache := cache.NewService(CacheClient)
	appHub := chats.NewHub(db_, appCache)
	appWorker := workers.NewAppWorker(
		appCache,
		appStore,
		factory,
		BatchInterval,
		PollInterval,
	)

	app := api.NewApplication(
		config.Addr,
		appStore,
		&config,
		jwt_Maker,
		paseto_Maker,
		mailer,
		appCache,
		factory,
		appHub,
		appWorker,
	)

	// Graceful Shutdown
	shutdown := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		sig := <-quit // wait for the interrupt signal

		// log the signal for debug
		log.Printf("captured %v signal, shutting down\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), ShutdownDelta)
		defer cancel()

		// trigger the shutdown
		shutdown <- app.Shutdown(ctx)
	}()

	err = app.Run()
	// check if there is an error while booting up
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("error starting the server: %v\n", err.Error())
	}

	// wait for the shutdown to complete
	err = <-shutdown
	if err != nil {
		log.Printf("error during shutdown: %v\n", err.Error())
	}

	// shutdown successful
	log.Printf("Server Shutdown Successfully\n")
}
