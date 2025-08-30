package api

import (
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/gin-gonic/gin"
)

type Application struct {
	store  *db.Store
	router *gin.Engine
	cfg    Config
}

type Config struct {
	DbCfg DBConfig
	Addr  string
}

type DBConfig struct {
	ADDR         string
	MaxConns     int
	MaxIdleTime  int
	MaxIdleConns int
}

func NewApplication(store *db.Store, cfg *Config) *Application {
	router := gin.Default()
	app := Application{
		store:  store,
		router: router,
		cfg:    *cfg,
	}
	router.GET("/", app.HealthCheck)
	router.POST("/users/create", app.CreateUser)

	return &app
}

func WriteResponse(obj any) gin.H {
	return gin.H{
		"success": true,
		"data":    obj,
	}
}

func WriteError(err string) gin.H {
	return gin.H{
		"success": false,
		"error":   err,
	}
}

func (app *Application) Run() error {
	return app.router.Run(app.cfg.Addr)
}
