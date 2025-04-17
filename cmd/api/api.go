package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tenteedee/gopher-social/internal/store"
)

type application struct {
	config config
	store  *store.Storage
}

type config struct {
	address string
	db      dbConfig
	env     string
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

		r.Route("/posts", func(r chi.Router) {
			// r.Get("/", app.getPostsHandler)
			r.Post("/", app.createPostHandler)

			r.Route("/{id}", func(r chi.Router) {
				r.Use(app.postContextMiddleware)

				r.Get("/", app.getPostByIdHandler)
				r.Patch("/", app.updatePostHandler)
				r.Delete("/", app.deletePostHandler)
			})
		})

		// r.Route("/users", func(r chi.Router) {
		// 	r.Get("/", app.getUsersHandler)
		// 	r.Post("/", app.createUserHandler)
		// 	r.Get("/{id}", app.getUserHandler)
		// 	r.Put("/{id}", app.updateUserHandler)
		// 	r.Delete("/{id}", app.deleteUserHandler)
		// })
	})

	return r
}

func (app *application) serve(mux *chi.Mux) error {

	server := &http.Server{
		Addr:         app.config.address,
		Handler:      mux,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting server on %s", app.config.address)

	return server.ListenAndServe()
}
