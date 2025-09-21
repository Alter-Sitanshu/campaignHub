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
	OTPCfg   OTPConfig
}

type OTPConfig struct {
}

type MailConfig struct {
	Host        string // e.g., "smtp.gmail.com"
	Port        int    // e.g., 587
	Username    string
	Password    string // The app password for the Gmail
	From        string // sender email address
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
	defaultUserLVL = "LVL1"
	SessionTimeout = time.Hour * 24 * 7 // Timeout of 7 Days
	CookieExp      = 3600 * 24 * 7      // 7 Days
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
	router.GET("/", app.HealthCheck)
	// user verification route
	router.GET("/verify", app.Verification) // query parameter: token
	// Users routes
	router.POST("/users", app.CreateUser)
	router.GET("/users/:id", app.GetUserById)
	router.GET("/users/email/:email", app.GetUserByEmail)
	router.DELETE("/users/:id", app.DeleteUser)
	router.PATCH("/users/:id", app.UpdateUser)
	router.GET("/users/campaigns/:id", app.GetUserCampaigns) // parameter: id

	// Brand routes
	router.GET("/brands/:brand_id", app.GetBrand)
	router.POST("/brands", app.CreateBrand)
	router.DELETE("/brands/:brand_id", app.DeleteBrand)
	router.PATCH("/brands/:brand_id", app.UpdateBrand)
	router.GET("/brands/campaigns/:brand_id", app.GetBrandCampaigns) // parameter: brandid

	// campaign routes
	router.GET("/campaigns", app.GetCampaignFeed)
	router.POST("/campaigns", app.CreateCampaign)
	router.PUT("/campaigns/:campaign_id", app.StopCampaign)
	router.DELETE("/campaigns/:campaign_id", app.DeleteCampaign)
	router.PATCH("/campaigns/:campaign_id", app.UpdateCampaign)

	// tickets routes
	router.GET("/tickets", app.GetRecentTickets) // query: status("open"/"close"), limit, offset
	router.POST("/tickets/", app.RaiseTicket)
	router.PUT("/tickets/:ticket_id", app.CloseTicket)
	router.DELETE("/tickets/:ticket_id", app.DeleteTicket)
	router.GET("/tickets/:ticket_id", app.GetTicket)

	// submissions routes
	router.POST("/submissions", app.CreateSubmission)
	router.GET("/submissions/:sub_id", app.GetSubmission)
	router.GET("/submissions", app.FilterSubmissions) // query: creator_id, campaign_id, time
	router.DELETE("/submissions/:sub_id", app.DeleteSubmission)
	router.PATCH("/submissions/:sub_id", app.UpdateSubmission)

	// accounts routes
	router.GET("/accounts/:acc_id", app.GetUserAccount)
	router.POST("/accounts", app.CreateAccount)
	router.DELETE("/accounts/:acc_id", app.DeleteUserAccount)
	router.PUT("/accounts/:acc_id", app.DisableUserAccount)
	router.GET("/accounts", app.GetAllAccounts)

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
