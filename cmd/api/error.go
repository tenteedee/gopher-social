package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	// log.Printf("internal server error: %v path: %v error: %v", r.Method, r.URL.Path, err)

	app.logger.Errorw(
		"internal server error",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err,
	)

	WriteJSON(w, http.StatusInternalServerError, "internal server error")
}

func (app *application) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	//log.Printf("bad request: %v path: %v error: %v", r.Method, r.URL.Path, err)

	app.logger.Warnw(
		"bad request",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)

	WriteJSON(w, http.StatusBadRequest, "bad request")
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request, err error) {
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

func (app *application) unauthorized(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnw(
		"unauthorized",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)

	WriteJSON(w, http.StatusUnauthorized, "unauthorized")
}

func (app *application) unauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf(
		"unauthorized basic error",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	WriteErrorJSON(w, http.StatusUnauthorized, "unauthorized")
}

func (app *application) forbidden(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("forbidden",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)

	WriteErrorJSON(w, http.StatusForbidden, "forbidden")
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Warnw("rate limit exceeded", "method", r.Method, "path", r.URL.Path)

	w.Header().Set("Retry-After", retryAfter)

	WriteErrorJSON(w, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
}
