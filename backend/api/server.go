package api

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Alter-Sitanshu/campaignHub/internals/auth"
	"github.com/Alter-Sitanshu/campaignHub/internals/cache"
	"github.com/Alter-Sitanshu/campaignHub/internals/chats"
	"github.com/Alter-Sitanshu/campaignHub/internals/db"
	"github.com/Alter-Sitanshu/campaignHub/internals/mailer"
	"github.com/Alter-Sitanshu/campaignHub/internals/workers"
	"github.com/Alter-Sitanshu/campaignHub/services/b2"
	"github.com/Alter-Sitanshu/campaignHub/services/platform"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	_ "github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type Application struct {
	store  *db.Store
	server *http.Server
	wg     sync.WaitGroup
	// Authentication services
	jwtMaker    auth.TokenMaker
	pasetoMaker auth.TokenMaker

	mailer  *mailer.MailService
	cache   *cache.Service
	cfg     Config
	factory *platform.Factory
	msgHub  *chats.Hub
	workers *workers.AppWorkers
	s3Store *b2.B2Storage
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
	B2Cfg      B2Config
}

type B2Config struct {
	KeyID    string
	AppKey   string
	Endpoint string
	Region   string
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

func (app *Application) AddRoutes(addr string, router *gin.Engine) {
	// Public routes
	base := router.Group("/api/v1")
	base.GET("/", app.HealthCheck)
	base.GET("/ws", app.AuthMiddleware(), app.WebSocketHandler)
	// query parameter: token: Example /verify/?token =
	base.GET("/verify", app.Verification)
	base.POST("/login", app.Login)
	base.POST("/users/signup", app.CreateUser)
	base.POST("/brands/signup", app.CreateBrand)
	// entity should be in ["users", "brands"]
	base.POST("/forgot_password/request/:entity", app.ForgotPassword)
	base.POST("/forgot_password/confirm/:entity", app.ResetPassword) // query parameter token

	// Users routes
	users := base.Group("/users", app.AuthMiddleware())
	{
		users.GET("/me", app.GetCurrentUser)
		users.GET("/:id", app.GetUserById)
		users.GET("/email/:email", app.GetUserByEmail)
		users.DELETE("/:id", app.DeleteUser)
		users.PATCH("/:id", app.UpdateUser)
		users.GET("/campaigns/:id", app.GetUserCampaigns) // parameter: id
		// query paramater ext (supported: jpeg, jpg, png)
		users.GET("/profile_picture/", app.GetProfilePicUpdateURL)
		// request must contain json{objectKey: ""}
		users.POST("/profile_picture/confirm", app.ConfirmProfilePicUpload)
		// query parameter id(user id)
		users.GET("/profile_picture/download/", app.GetUserProfilePic)
	}

	// Brand routes
	brands := base.Group("/brands", app.AuthMiddleware())
	{
		brands.GET("/:brand_id", app.GetBrand)
		brands.DELETE("/:brand_id", app.DeleteBrand)
		brands.PATCH("/:brand_id", app.UpdateBrand)
		brands.GET("/campaigns/:brand_id", app.GetBrandCampaigns) // parameter: brandid
	}

	// campaign routes
	campaigns := base.Group("/campaigns", app.AuthMiddleware())
	{
		campaigns.GET("/feed", app.GetCampaignFeed) // query parametes: cursor
		campaigns.GET("/:campaign_id", app.GetCampaign)
		campaigns.GET("/user/:userid", app.GetUserCampaigns)    // query parameters: cursor
		campaigns.GET("/brand/:brandid", app.GetBrandCampaigns) // query parameters: cursor
		campaigns.POST("", app.CreateCampaign)
		campaigns.PUT("/stop/:campaign_id", app.StopCampaign)
		campaigns.PUT("/activate/:campaign_id", app.ActivateCampaign)
		campaigns.DELETE("/:campaign_id", app.DeleteCampaign)
		campaigns.PATCH("/:campaign_id", app.UpdateCampaign)
	}

	applications := base.Group("/applications", app.AuthMiddleware())
	{
		applications.GET(":application_id", app.GetApplication)
		applications.GET("/campaigns/:campaign_id", app.GetCampaignApplications)
		applications.GET("/my-applications", app.GetCreatorApplications)        // query: offset, limit
		applications.PATCH("/status/:application_id", app.SetApplicationStatus) // query: status
		applications.DELETE("/delete/:application_id", app.DeleteApplication)
		applications.POST("/:campaign_id", app.CreateApplication)
	}

	// tickets routes
	tickets := base.Group("/tickets", app.AuthMiddleware())
	{
		tickets.GET("", app.GetRecentTickets) // query: status("open"/"close"), limit, offset
		tickets.GET("/:ticket_id", app.GetTicket)
		tickets.POST("", app.RaiseTicket)
		tickets.PUT("/:ticket_id", app.CloseTicket)
		tickets.DELETE("/:ticket_id", app.DeleteTicket)
	}

	// submissions routes
	submission := base.Group("/submissions", app.AuthMiddleware())
	{
		submission.GET("", app.FilterSubmissions)               // query: creator_id, campaign_id, time
		submission.GET("/my-submissions", app.GetMySubmissions) // query: time
		submission.GET("/:sub_id", app.GetSubmission)
		submission.POST("", app.CreateSubmission)
		submission.DELETE("/:sub_id", app.DeleteSubmission)
		submission.PATCH("/:sub_id", app.UpdateSubmission)
	}

	// accounts routes
	accounts := base.Group("/accounts", app.AuthMiddleware())
	{
		accounts.GET("", app.GetAllAccounts)
		accounts.GET("/:acc_id", app.GetUserAccount)
		accounts.POST("", app.CreateAccount)
		accounts.DELETE("/accounts/:acc_id", app.DeleteUserAccount)
		accounts.PUT("/accounts/:acc_id", app.DisableUserAccount)
	}

	// messaging routes
	conversations := base.Group("/private/conversations", app.AuthMiddleware())
	{
		conversations.GET("", app.GetEntityConversations)
		// query parameters timestamp and cursor
		conversations.GET(":conversation/messages", app.GetConversationMessages)
	}

	app.server = &http.Server{
		Addr:    addr,             // address the server listens on
		Handler: router.Handler(), // HTTP handler to invoke, in this case, the Gin router
	}
}

func NewApplication(addr string, store *db.Store, cfg *Config, JWT, PASETO auth.TokenMaker,
	mailer *mailer.MailService, cacheService *cache.Service, factory *platform.Factory,
	appHub *chats.Hub, workers *workers.AppWorkers, s3Store *b2.B2Storage,
) *Application {
	gin.SetMode(gin.TestMode) // TODO: change to release mode in production
	router := gin.Default()

	app := Application{
		store:       store,
		cfg:         *cfg,
		jwtMaker:    JWT,
		pasetoMaker: PASETO,
		mailer:      mailer,
		cache:       cacheService,
		factory:     factory,
		msgHub:      appHub,
		workers:     workers,
		s3Store:     s3Store,
	}

	// rate limiter
	limiter := rate.NewLimiter(rate.Limit(100), 50)
	// Attach the limiter to the router
	router.Use(app.RateLimitter(limiter))

	// Adding CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"https://campaignhub.com",
			"https://www.campaignhub.com",
			"http://localhost:5173", // for dev
			"http://localhost:5173", // for dev
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Authorization",
			"X-Requested-With", "Accept", "X-Request-ID",
		},
		ExposeHeaders: []string{
			"Content-Length", "Content-Disposition", // for file downloads
		},
		AllowCredentials: true,
		AllowWebSockets:  true,
		MaxAge:           12 * time.Hour,
	}))

	// Attaching Security headers
	// router.Use(secure.New(secure.Config{
	// 	FrameDeny:          true, // Prevent clickjacking
	// 	ContentTypeNosniff: true, // Stop MIME-type sniffing
	// 	BrowserXssFilter:   true, // Basic browser XSS protection
	// router.Use(secure.New(secure.Config{
	// 	FrameDeny:          true, // Prevent clickjacking
	// 	ContentTypeNosniff: true, // Stop MIME-type sniffing
	// 	BrowserXssFilter:   true, // Basic browser XSS protection

	// 	// Forcing HTTPS
	// 	SSLRedirect:          true,
	// 	SSLProxyHeaders:      map[string]string{"X-Forwarded-Proto": "https"},
	// 	STSSeconds:           31536000, // 1 year
	// 	STSIncludeSubdomains: true,
	// 	STSPreload:           true,
	// 	// Forcing HTTPS
	// 	SSLRedirect:          true,
	// 	SSLProxyHeaders:      map[string]string{"X-Forwarded-Proto": "https"},
	// 	STSSeconds:           31536000, // 1 year
	// 	STSIncludeSubdomains: true,
	// 	STSPreload:           true,

	// 	// Content Security Policy (CSP) For handling the Media/Chats
	// 	// TODO: change values to match the CDN, API, and frontend origins.
	// 	ContentSecurityPolicy: "default-src 'self'; " +
	// 		"img-src 'self' data: blob: https://cdn.campaignhub.com; " +
	// 		"media-src 'self' blob: https://cdn.campaignhub.com; " +
	// 		"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
	// 		"connect-src 'self' wss://api.campaignhub.com https://api.campaignhub.com; " +
	// 		"style-src 'self' 'unsafe-inline'; " +
	// 		"font-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com;",
	// 	// Content Security Policy (CSP) For handling the Media/Chats
	// 	// TODO: change values to match the CDN, API, and frontend origins.
	// 	ContentSecurityPolicy: "default-src 'self'; " +
	// 		"img-src 'self' data: blob: https://cdn.campaignhub.com; " +
	// 		"media-src 'self' blob: https://cdn.campaignhub.com; " +
	// 		"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
	// 		"connect-src 'self' wss://api.campaignhub.com https://api.campaignhub.com; " +
	// 		"style-src 'self' 'unsafe-inline'; " +
	// 		"font-src 'self' https://fonts.googleapis.com https://fonts.gstatic.com;",

	// 	ReferrerPolicy: "strict-origin-when-cross-origin", // prevent leaking full URLs
	// }))
	// 	ReferrerPolicy: "strict-origin-when-cross-origin", // prevent leaking full URLs
	// }))

	// Attaching request id headers
	router.Use(requestid.New())

	// adds the routes to the gin.Engine and attaches the handler to the application
	app.AddRoutes(addr, router)

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
	// context of the workers
	ctx, cancel := context.WithCancel(context.Background())
	app.workers.SetCancel(cancel)
	app.wg.Add(3) // One for each service running
	// Start the batch workers go routines
	go func() {
		defer app.wg.Done()
		app.workers.Batch.Start(ctx)
	}()
	go func() {
		defer app.wg.Done()
		app.workers.Poll.Start(ctx)
	}()

	// Start the Sockets Hub in a go routine
	go func() {
		defer app.wg.Done()
		app.msgHub.Run()
	}()
	return app.server.ListenAndServe()
}

func (app *Application) Shutdown(ctx context.Context) error {
	// closing the workers routine
	app.workers.Batch.Stop()
	app.workers.Poll.Stop()

	// closing the sockets routine
	app.msgHub.Stop()
	app.wg.Wait()
	return app.server.Shutdown(ctx)
}
