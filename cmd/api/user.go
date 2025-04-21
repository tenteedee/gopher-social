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

// Get User by ID godoc
//
//	@Summary		Fetch a user by ID
//	@Description	Fetch a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *application) getUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getCurrentUser(r)

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.InternalServerError(w, r, err)
		return
	}
}

// Follow User godoc
//
//	@Summary		Follow a user
//	@Description	Follow a user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		204 {object}	nil
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id}/follow [put]
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
			app.NotFound(w, r, err)
			return
		case store.ErrConflict:
			app.conflictResponse(w, r, err)
			return
		default:
			app.InternalServerError(w, r, err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// Unfollow User godoc
//
//	@Summary		Unfollow a user
//	@Description	Unfollow a user
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id}/unfollow [put]
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
			app.NotFound(w, r, err)
			return
		default:
			app.InternalServerError(w, r, err)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// Activate User godoc
//
//	@Summary		Activates/Register a user
//	@Description	Activates/Register a user by invitation token
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.User.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrorNotFound:
			app.NotFound(w, r, err)
			return
		default:
			app.InternalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusNoContent, "User activated"); err != nil {
		app.InternalServerError(w, r, err)
		return
	}
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
				app.NotFound(w, r, err)
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
