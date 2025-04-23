package main

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/tenteedee/gopher-social/internal/auth"
	"github.com/tenteedee/gopher-social/internal/db"
	"github.com/tenteedee/gopher-social/internal/env"
	"github.com/tenteedee/gopher-social/internal/mailer"
	ratelimiter "github.com/tenteedee/gopher-social/internal/rate-limiter"
	"github.com/tenteedee/gopher-social/internal/store"
	"github.com/tenteedee/gopher-social/internal/store/cache"
	"go.uber.org/zap"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

const version = "2.0.0"

//	@title			Gopher Social API
//	@description	API for Gopher Social
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v1
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				API Key Authorization header
func main() {
	env.Init()

	cfg := config{
		address: env.ApiPort,
		apiURL:  env.ApiURL,
		db: dbConfig{
			dsn:          fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", env.DbUser, env.DbPassword, env.DbHost, env.DbPort, env.DbName),
			maxOpenConns: env.DB_MAX_OPEN_CONNS,
			maxIdleConns: env.DB_MAX_IDLE_CONNS,
			maxIdleTime:  env.DB_MAX_IDLE_TIME,
		},
		env: env.ApiEnv,
		mail: mailConfig{
			exp:       env.MailExp,
			fromEmail: env.FromEmail,
			sendgrid: sendgridConfig{
				apikey: env.SendgridAPIKey,
			},
			mailTrap: mailTrapConfig{
				apikey: env.MailTrapAPIKey,
			},
		},
		frontendURL: env.FrontendURL,
		auth: authConfig{
			basic: basicConfig{
				user:     env.AuthBasicUser,
				password: env.AuthBasicPassword,
			},
			token: tokenConfig{
				secret:   env.AuthTokenSecret,
				audience: env.AuthTokenAudience,
				issuer:   env.AuthTokenIssuer,
				exp:      time.Hour * 24 * 3,
			},
		},
		redisCfg: redisConfig{
			addr:    env.RedisAddress,
			pw:      env.RedisPassword,
			db:      env.RedisDB,
			enabled: env.RedisEnabled,
		},
		rateLimiter: ratelimiter.Config{
			RequestsPerTimeFrame: env.RateLimiterRequestCount,
			TimeFrame:            env.RateLimiterTimeFrame,
			Enabled:              env.RateLimiterEnabled,
		},
	}

	// Logger
	config := zap.NewProductionConfig()
	config.EncoderConfig.StacktraceKey = ""

	logger := zap.Must(config.Build()).Sugar()
	defer logger.Sync()

	db, err := db.New(
		cfg.db.dsn,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		logger.Fatal("failed to connect to database: %v", err)
	}

	defer db.Close()
	logger.Info("Connected to database")

	// Cache
	var redisDB *redis.Client
	if cfg.redisCfg.enabled {
		redisDB = cache.NewRedisClient(cfg.redisCfg.addr, cfg.redisCfg.pw, cfg.redisCfg.db)
		logger.Info("Connected to redis")
	}

	storage := store.NewStorage(db)
	cacheStorage := cache.NewRedisStorage(redisDB)

	mailer := mailer.NewSendGridMailer(cfg.mail.sendgrid.apikey, cfg.mail.fromEmail)
	// mailtrap, err := mailer.NewMailTrapClient(cfg.mail.mailTrap.apikey, cfg.mail.fromEmail)
	// if err != nil {
	// 	logger.Fatal(err)
	// }

	JwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.audience, cfg.auth.token.issuer)

	app := &application{
		config:        cfg,
		store:         storage,
		cacheStorage:  cacheStorage,
		logger:        logger,
		mailer:        mailer,
		authenticator: JwtAuthenticator,
	}
	mux := app.mount()

	logger.Fatal(app.serve(mux))
}
