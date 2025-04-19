package main

import (
	"net/http"

	"github.com/tenteedee/gopher-social/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	// user := app.getCurrentUser(r)
	// userID := user.ID

	fq := store.PaginationFeedQuery{
		Limit:  10,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.BadRequest(w, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	posts, err := app.store.Post.GetByUserId(r.Context(), int64(20), fq)
	if err != nil {
		switch err {
		case store.ErrorNotFound:
			app.NotFound(w, r)
			return
		default:
			app.InternalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, posts); err != nil {
		app.InternalServerError(w, r, err)
		return
	}
}
