package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tenteedee/gopher-social/internal/store"
)

type postKey string

const postContextKey postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

type UpdatePostPayload struct {
	Title   *string   `json:"title" validate:"omitempty,max=100"`
	Content *string   `json:"content" validate:"omitempty,max=1000"`
	Tags    *[]string `json:"tags"`
}

type CreateCommentPayload struct {
	UserID  int64  `json:"user_id"`
	Content string `json:"content"`
}

func (app *application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			app.BadRequest(w, r, err)
			return
		}

		post, err := app.store.Post.GetByID(r.Context(), postID)
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
		ctx := context.WithValue(r.Context(), postContextKey, post)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromContext(r *http.Request) (*store.Post, error) {
	post, ok := r.Context().Value(postContextKey).(*store.Post)
	if !ok {
		return nil, store.ErrorNotFound
	}
	return post, nil
}

// Create Post godoc
//
//	@Summary		Create a post
//	@Description	Create a post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreatePostPayload	true	"Post Information"
//	@Success		201		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts [post]
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	if err := ReadJSON(w, r, &payload); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	post := store.Post{
		Title:   payload.Title,
		Content: payload.Content,
		// TODO: change after auth
		UserID: 1,
		Tags:   payload.Tags,
	}

	response, err := app.store.Post.Create(r.Context(), &post)

	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusCreated, response); err != nil {
		app.InternalServerError(w, r, err)
		return
	}

}

// Get Post by ID godoc
//
//	@Summary		Fetch a Post by ID
//	@Description	Fetch a Post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		200	{object}	store.Post
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [get]
func (app *application) getPostByIdHandler(w http.ResponseWriter, r *http.Request) {
	post, err := getPostFromContext(r)
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

	comments, err := app.store.Comment.GetCommentByPostId(r.Context(), post.ID)
	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}
	post.Comments = comments
	// if len(post.Comments) > 0 {
	// 	for i := range post.Comments {
	// 		post.Comments[i].User, err = app.store.User.GetByID(r.Context(), post.Comments[i].UserID)
	// 		if err != nil {
	// 			app.InternalServerError(w, r, err)
	// 			return
	// 		}
	// 	}
	// }

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.InternalServerError(w, r, err)
		return
	}
}

// DeletePost godoc
//
//	@Summary		Deletes a post
//	@Description	Delete a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Post ID"
//	@Success		204	{object}	string
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [delete]
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	post, err := getPostFromContext(r)
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

	if err := app.store.Post.Delete(r.Context(), post.ID); err != nil {
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

// UpdatePost godoc
//
//	@Summary		Updates a post
//	@Description	Updates a post by ID
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int					true	"Post ID"
//	@Param			payload	body		UpdatePostPayload	true	"Post payload"
//	@Success		200		{object}	store.Post
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id} [patch]
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post, err := getPostFromContext(r)
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

	// read the request body into a CreatePostPayload struct
	var payload UpdatePostPayload
	if err := ReadJSON(w, r, &payload); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	// validate the payload
	if err := Validate.Struct(payload); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}
	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Tags != nil {
		post.Tags = *payload.Tags
	}

	// update the post struct with the payload data
	if err := app.store.Post.Update(r.Context(), post); err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.InternalServerError(w, r, err)
		return
	}

}

// Create Comment godoc
//
//	@Summary		Create a comment
//	@Description	Create a comment for a Post
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Post ID"
//	@Param			payload	body		CreateCommentPayload	true	"Comment payload"
//	@Success		200		{object}	store.Comment
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/posts/{id}/comments [post]
func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	post, err := getPostFromContext(r)
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

	var payload CreateCommentPayload
	if err := ReadJSON(w, r, &payload); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	user, err := app.store.User.GetById(r.Context(), payload.UserID)
	if err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	var comment store.Comment
	comment.PostID = post.ID
	comment.UserID = payload.UserID
	comment.Content = payload.Content
	comment.User = *user

	if err := app.store.Comment.Create(r.Context(), &comment); err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, comment); err != nil {
		app.InternalServerError(w, r, err)
		return
	}
}
