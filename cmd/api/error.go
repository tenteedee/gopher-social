package main

import (
	"log"
	"net/http"
)

func (app *application) InternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %v path: %v error: %v", r.Method, r.URL.Path, err)

	WriteJSON(w, http.StatusInternalServerError, "internal server error")
}

func (app *application) BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request: %v path: %v error: %v", r.Method, r.URL.Path, err)

	WriteJSON(w, http.StatusBadRequest, "bad request")
}

func (app *application) NotFound(w http.ResponseWriter, r *http.Request) {
	log.Printf("not found: %v path: %v", r.Method, r.URL.Path)

	WriteJSON(w, http.StatusNotFound, "not found")
}
