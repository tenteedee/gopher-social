package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/tenteedee/gopher-social/internal/db"
	"github.com/tenteedee/gopher-social/internal/env"
	"github.com/tenteedee/gopher-social/internal/mailer"
	"github.com/tenteedee/gopher-social/internal/store"
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

	storage := store.NewStorage(db)

	mailer := mailer.NewSendGridMailer(cfg.mail.sendgrid.apikey, cfg.mail.fromEmail)
	// mailtrap, err := mailer.NewMailTrapClient(cfg.mail.mailTrap.apikey, cfg.mail.fromEmail)
	// if err != nil {
	// 	logger.Fatal(err)
	// }

	app := &application{
		config: cfg,
		store:  storage,
		logger: logger,
		mailer: mailer,
	}
	mux := app.mount()

	logger.Fatal(app.serve(mux))
}
