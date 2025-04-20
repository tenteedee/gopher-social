package main

import (
	"net/http"
)

func (app *application) InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	// log.Printf("internal server error: %v path: %v error: %v", r.Method, r.URL.Path, err)

	app.logger.Errorw(
		"internal server error",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err,
	)

	WriteJSON(w, http.StatusInternalServerError, "internal server error")
}

func (app *application) BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	//log.Printf("bad request: %v path: %v error: %v", r.Method, r.URL.Path, err)

	app.logger.Warnw(
		"bad request",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)

	WriteJSON(w, http.StatusBadRequest, "bad request")
}

func (app *application) NotFound(w http.ResponseWriter, r *http.Request, err error) {
	// log.Printf("not found: %v path: %v", r.Method, r.URL.Path)

	app.logger.Warnw(
		"not found",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)

	WriteJSON(w, http.StatusNotFound, "not found")
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	//log.Printf("conflict: %v path: %v", r.Method, r.URL.Path)

	app.logger.Errorw(
		"conflict",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)

	WriteJSON(w, http.StatusConflict, "conflict")
}
