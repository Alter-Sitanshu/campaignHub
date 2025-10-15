package api

import (
	"context"
	"net/http"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals/auth"
	"github.com/Alter-Sitanshu/campaignHub/internals/cache"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/Alter-Sitanshu/campaignHub/internals/mailer"
	"github.com/Alter-Sitanshu/campaignHub/internals/platform"
	"github.com/gin-gonic/gin"
)

type Application struct {
	store       *db.Store
	server      *http.Server
	jwtMaker    auth.TokenMaker
	pasetoMaker auth.TokenMaker
	mailer      *mailer.MailService
	cache       *cache.Service
	cfg         Config
	factory     *platform.Factory
}

type Config struct {
	ISS        string
	AUD        string
	DbCfg      DBConfig
	TokenCfg   TokenConfig
	Addr       string
	MailCfg    MailConfig
	RedisCfg   RedisConfig
	FactoryCfg FactoryConfig
}

type FactoryConfig struct {
	YouTubeAPIKey string
	MetaToken     string
}

type RedisConfig struct {
	Addr     string //localhost:6379
	Protocol int    // default: 3 RESPv3
	DB       int    // default: 0 In-Memory Database
	Password string // default: ""
}

type MailConfig struct {
	Host        string // e.g., "smtp.gmail.com"
	Port        int    // e.g., 587
	Username    string
	Password    string // The app password for the Gmail
	From        string // sender email address
	Support     string // tickets are sent to this address
	Expiry      time.Duration
	MailRetries int
}

type DBConfig struct {
	ADDR         string
	MaxConns     int
	MaxIdleTime  int
	MaxIdleConns int
}

type TokenConfig struct {
	JWTSecret        string
	PASETO_SYM       string
	PASETO_ASYM_PRIV string
	PASETO_ASYM_PUB  string
}

const (
	defaultUserLVL       = "LVL1"
	SessionTimeout       = time.Hour * 24 * 7 // Timeout of 7 Days
	CookieExp            = 3600 * 24 * 7      // 7 Days
	ResetTokenExpiry     = time.Minute * 15   // 15 Minutes
	DefaultSyncFrequency = 5                  // in minutes
)

func NewApplication(addr string, store *db.Store, cfg *Config, PASETO, JWT auth.TokenMaker,
	mailer *mailer.MailService, cacheService *cache.Service, factory *platform.Factory,
) *Application {
	router := gin.Default()
	app := Application{
		store: store,
		server: &http.Server{
			Addr:    addr,             // address the server listens on
			Handler: router.Handler(), // HTTP handler to invoke, in this case, the Gin router
		},
		cfg:         *cfg,
		jwtMaker:    JWT,
		pasetoMaker: PASETO,
		mailer:      mailer,
		cache:       cacheService,
		factory:     factory,
	}

	// Public routes
	router.GET("/", app.HealthCheck)
	// query parameter: token | entity should be in ["users", "brands"]
	router.GET("/verify/:entity", app.Verification)
	router.POST("/login", app.Login)
	router.POST("/users/signup", app.CreateUser)
	router.POST("/brands/signup", app.CreateBrand)
	// entity should be in ["users", "brands"]
	router.POST("/forgot_password/request/:entity", app.ForgotPassword)
	router.POST("/forgot_password/confirm/:entity", app.ResetPassword) // query parameter token

	// Users routes
	users := router.Group("/users", app.AuthMiddleware())
	{
		users.GET("/:id", app.GetUserById)
		users.GET("/email/:email", app.GetUserByEmail)
		users.DELETE("/:id", app.DeleteUser)
		users.PATCH("/:id", app.UpdateUser)
		users.GET("/campaigns/:id", app.GetUserCampaigns) // parameter: id
	}

	// Brand routes
	brands := router.Group("/brands", app.AuthMiddleware())
	{
		brands.GET("/:brand_id", app.GetBrand)
		brands.DELETE("/:brand_id", app.DeleteBrand)
		brands.PATCH("/:brand_id", app.UpdateBrand)
		brands.GET("/campaigns/:brand_id", app.GetBrandCampaigns) // parameter: brandid
	}

	// campaign routes
	campaigns := router.Group("/campaigns", app.AuthMiddleware())
	{
		campaigns.GET("", app.GetCampaignFeed) // query parametes: limit, offset
		campaigns.GET("/:campaign_id", app.GetCampaign)
		campaigns.GET("/user/:userid", app.GetUserCampaigns) // parameter: user_id
		campaigns.GET("/brand/:brandid", app.GetBrandCampaigns)
		campaigns.POST("", app.CreateCampaign)
		campaigns.PUT("/:campaign_id", app.StopCampaign)
		campaigns.DELETE("/:campaign_id", app.DeleteCampaign)
		campaigns.PATCH("/:campaign_id", app.UpdateCampaign)
	}

	applications := router.Group("/applications", app.AuthMiddleware())
	{
		applications.GET(":application_id", app.GetApplication)
		applications.GET("/campaigns/:campaign_id", app.GetCampaignApplications)
		applications.GET("/my-applications", app.GetCreatorApplications)        // query: offset, limit
		applications.PATCH("/status/:application_id", app.SetApplicationStatus) // query: status
		applications.DELETE("/delete/:application_id", app.DeleteApplication)
		applications.POST("", app.CreateApplication)
	}

	// tickets routes
	tickets := router.Group("/tickets", app.AuthMiddleware())
	{
		tickets.GET("", app.GetRecentTickets) // query: status("open"/"close"), limit, offset
		tickets.GET("/:ticket_id", app.GetTicket)
		tickets.POST("", app.RaiseTicket)
		tickets.PUT("/:ticket_id", app.CloseTicket)
		tickets.DELETE("/:ticket_id", app.DeleteTicket)
	}

	// submissions routes
	submission := router.Group("/submissions", app.AuthMiddleware())
	{
		submission.GET("", app.FilterSubmissions)               // query: creator_id, campaign_id, time
		submission.GET("/my-submissions", app.GetMySubmissions) // query: time
		submission.GET("/:sub_id", app.GetSubmission)
		submission.POST("", app.CreateSubmission)
		submission.DELETE("/:sub_id", app.DeleteSubmission)
		submission.PATCH("/:sub_id", app.UpdateSubmission)
	}

	// accounts routes
	accounts := router.Group("/accounts", app.AuthMiddleware())
	{
		accounts.GET("", app.GetAllAccounts)
		accounts.GET("/:acc_id", app.GetUserAccount)
		accounts.POST("", app.CreateAccount)
		accounts.DELETE("/accounts/:acc_id", app.DeleteUserAccount)
		accounts.PUT("/accounts/:acc_id", app.DisableUserAccount)
	}

	return &app
}

func WriteResponse[T any](obj T) gin.H {
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
	return app.server.ListenAndServe()
}

func (app *Application) Shutdown(ctx context.Context) error {
	return app.server.Shutdown(ctx)
}
