package main

import (
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/tenteedee/gopher-social/internal/db"
	"github.com/tenteedee/gopher-social/internal/env"
	"github.com/tenteedee/gopher-social/internal/store"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

const version = "1.0.0"

func main() {
	env.Init()

	cfg := config{
		address: env.ApiPort,
		db: dbConfig{
			dsn:          fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", env.DbUser, env.DbPassword, env.DbHost, env.DbPort, env.DbName),
			maxOpenConns: env.DB_MAX_OPEN_CONNS,
			maxIdleConns: env.DB_MAX_IDLE_CONNS,
			maxIdleTime:  env.DB_MAX_IDLE_TIME,
		},
		env: env.ApiEnv,
	}

	db, err := db.New(
		cfg.db.dsn,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		log.Panicf("failed to connect to database: %v", err)
	}

	defer db.Close()
	log.Println("Connected to database")

	storage := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  storage,
	}
	mux := app.mount()

	log.Fatal(app.serve(mux))
}
