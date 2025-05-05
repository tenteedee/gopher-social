package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	"github.com/go-chi/cors"
	docs "github.com/tenteedee/gopher-social/docs" // required for swagger to work
	"github.com/tenteedee/gopher-social/internal/auth"
	"github.com/tenteedee/gopher-social/internal/mailer"
	ratelimiter "github.com/tenteedee/gopher-social/internal/rate-limiter"
	"github.com/tenteedee/gopher-social/internal/store"
	"github.com/tenteedee/gopher-social/internal/store/cache"
)

type application struct {
	config        config
	store         *store.Storage
	cacheStorage  cache.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	rateLimiter   ratelimiter.Limiter
}

type config struct {
	address     string
	db          dbConfig
	env         string
	apiURL      string
	mail        mailConfig
	frontendURL string
	auth        authConfig
	redisCfg    redisConfig
	rateLimiter ratelimiter.Config
}

type mailConfig struct {
	exp       time.Duration
	fromEmail string
	sendgrid  sendgridConfig
	mailTrap  mailTrapConfig
}

type sendgridConfig struct {
	apikey string
}

type mailTrapConfig struct {
	apikey string
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type basicConfig struct {
	user     string
	password string
}

type tokenConfig struct {
	secret   string
	audience string
	issuer   string
	exp      time.Duration
}

type redisConfig struct {
	addr    string
	pw      string
	db      int
	enabled bool
}

type dbConfig struct {
	dsn          string // Data Source Name
	maxOpenConns int    // set an upper limit on the number of open connections to the database
	maxIdleConns int    // set an upper limit on the number of idle connections in the pool
	maxIdleTime  string // set the maximum amount of time a connection may be idle before being closed
}

func (app *application) mount() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	if app.config.rateLimiter.Enabled {
		r.Use(app.RateLimiterMiddleware)
	}

	// set request timeout on context to signal ctx.Done() if the request has timeout
	// and further process would be stopped
	r.Use(middleware.Timeout(60 * time.Second))

	// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("welcome"))
	// })
	r.Route("/v1", func(r chi.Router) {
		// r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheckHandler)
		r.Get("/health", app.healthCheckHandler)

		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("http://localhost:8080/v1/swagger/doc.json"), // The url pointing to API definition
		))

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware)
			// r.Get("/", app.getPostsHandler)
			r.Post("/", app.createPostHandler)

			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.postContextMiddleware)

				r.Get("/", app.getPostByIdHandler)
				r.Patch("/", app.CheckPostOwnership("moderator", app.updatePostHandler))
				r.Delete("/", app.CheckPostOwnership("admin", app.deletePostHandler))
				r.Post("/comments", app.createCommentHandler)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)

			r.Route(("/{id}"), func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				// r.Use(app.userContextMiddleware)

				r.Get("/", app.getUserByIdHandler)
				// r.Get("/me", app.getUserProfileHandler)

				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserhandler)
			r.Post("/token", app.createTokenHandler)
		})
	})

	return r
}

func (app *application) serve(mux *chi.Mux) error {
	// Docs
	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = app.config.apiURL
	docs.SwaggerInfo.BasePath = "/v1"

	server := &http.Server{
		Addr:         app.config.address,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		app.logger.Infow("signal caught", "signal", s)

		shutdown <- server.Shutdown(ctx)
	}()

	app.logger.Infow("server started",
		"address", app.config.address,
		"env", app.config.env,
	)

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		app.logger.Errorw("server error", "error", err)
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.logger.Infow("server has stopped",
		"address", app.config.address,
		"env", app.config.env,
	)

	return nil
}
