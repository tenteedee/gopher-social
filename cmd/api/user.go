package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tenteedee/gopher-social/internal/store"
)

type userKey string

const userContextKey userKey = "user"

type FollowUserPayload struct {
	FollowedUserID int64 `json:"followed_user_id"`
}

func (app *application) getUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getCurrentUser(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.InternalServerError(w, r, err)
		return
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getCurrentUser(r)
	userID := user.ID

	var payload FollowUserPayload

	if err := ReadJSON(w, r, &payload); err != nil {
		app.BadRequest(w, r, err)
		return
	}
	fmt.Print(payload)

	followedUserID := payload.FollowedUserID
	if userID == followedUserID {
		app.BadRequest(w, r, errors.New("cannot follow yourself"))
		return
	}

	err := app.store.Follow.Follow(r.Context(), followedUserID, userID)
	if err != nil {
		switch err {
		case store.ErrorNotFound:
			app.NotFound(w, r)
			return
		case store.ErrConflict:
			app.conflictResponse(w, r)
			return
		default:
			app.InternalServerError(w, r, err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getCurrentUser(r)
	userID := user.ID

	var payload FollowUserPayload
	if err := ReadJSON(w, r, &payload); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	followedUserID := payload.FollowedUserID

	err := app.store.Follow.Unfollow(r.Context(), followedUserID, userID)
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
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			app.BadRequest(w, r, err)
			return
		}

		user, err := app.store.User.GetById(r.Context(), userID)
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

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) getCurrentUser(r *http.Request) *store.User {
	user, ok := r.Context().Value(userContextKey).(*store.User)
	if !ok {
		return nil
	}
	return user
}
