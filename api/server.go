package api

import (
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals/auth"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/Alter-Sitanshu/campaignHub/internals/mailer"
	"github.com/gin-gonic/gin"
)

type Application struct {
	store       *db.Store
	router      *gin.Engine
	jwtMaker    auth.TokenMaker
	pasetoMaker auth.TokenMaker
	mailer      *mailer.MailService
	cfg         Config
}

type Config struct {
	ISS      string
	AUD      string
	DbCfg    DBConfig
	TokenCfg TokenConfig
	Addr     string
	MailCfg  MailConfig
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
	defaultUserLVL   = "LVL1"
	SessionTimeout   = time.Hour * 24 * 7 // Timeout of 7 Days
	CookieExp        = 3600 * 24 * 7      // 7 Days
	ResetTokenExpiry = time.Minute * 15   // 15 Minutes
)

func NewApplication(store *db.Store, cfg *Config, PASETO, JWT auth.TokenMaker,
	mailer *mailer.MailService,
) *Application {
	router := gin.Default()
	app := Application{
		store:       store,
		router:      router,
		cfg:         *cfg,
		jwtMaker:    JWT,
		pasetoMaker: PASETO,
		mailer:      mailer,
	}

	// Public routes
	router.GET("/", app.HealthCheck)
	// query parameter: token | entity should be in ["users", "brands"]
	router.GET("/:entity/verify", app.Verification)
	router.POST("/login", app.Login)
	router.POST("/users/signup", app.CreateUser)
	router.POST("/brands/signup", app.CreateBrand)
	// entity should be in ["users", "brands"]
	router.POST("/:entity/forgot_password/request", app.ForgotPassword)
	router.POST("/:entity/forgot_password/confirm", app.ResetPassword) // query parameter token
	router.POST("/admin/login", app.AuthMiddleware(), app.AdminLogin)

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
		submission.GET("", app.FilterSubmissions) // query: creator_id, campaign_id, time
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
	return app.router.Run(app.cfg.Addr)
}
