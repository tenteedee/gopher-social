package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"

	docs "github.com/tenteedee/gopher-social/docs" // required for swagger to work
	"github.com/tenteedee/gopher-social/internal/store"
)

type application struct {
	config config
	store  *store.Storage
	logger *zap.SugaredLogger
}

type config struct {
	address string
	db      dbConfig
	env     string
	apiURL  string
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

	r.Use(middleware.Timeout(60 * time.Second))

	// r.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("welcome"))
	// })
	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheckHandler)

		r.Get("/swagger/*", httpSwagger.Handler(
			httpSwagger.URL("http://localhost:8080/v1/swagger/doc.json"), // The url pointing to API definition
		))

		r.Route("/posts", func(r chi.Router) {
			// r.Get("/", app.getPostsHandler)
			r.Post("/", app.createPostHandler)

			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.postContextMiddleware)

				r.Get("/", app.getPostByIdHandler)
				r.Patch("/", app.updatePostHandler)
				r.Delete("/", app.deletePostHandler)
				r.Post("/comments", app.createCommentHandler)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Route(("/{id}"), func(r chi.Router) {
				r.Use(app.userContextMiddleware)

				r.Get("/", app.getUserByIdHandler)

				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getUserFeedHandler)
			})
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

	app.logger.Infow("server started",
		"address", app.config.address,
		"env", app.config.env,
	)

	return server.ListenAndServe()
}
