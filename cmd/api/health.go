package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"message": "API is healthy",
		"version": version,
		"env":     app.config.env,
	}

	if err := app.jsonResponse(w, http.StatusOK, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
