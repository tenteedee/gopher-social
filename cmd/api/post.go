package main

import (
	"context"
	"log"
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
				app.NotFound(w, r)
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

func (app *application) getPostByIdHandler(w http.ResponseWriter, r *http.Request) {
	post, err := getPostFromContext(r)
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

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	post, err := getPostFromContext(r)
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

	if err := app.store.Post.Delete(r.Context(), post.ID); err != nil {
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

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post, err := getPostFromContext(r)
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

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	post, err := getPostFromContext(r)
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

	var payload store.Comment
	if err := ReadJSON(w, r, &payload); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	log.Println(payload)

	if err := Validate.Struct(payload); err != nil {
		app.BadRequest(w, r, err)
		return
	}

	payload.PostID = post.ID
	log.Println(payload)

	if err := app.store.Comment.Create(r.Context(), &payload); err != nil {
		app.InternalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, payload); err != nil {
		app.InternalServerError(w, r, err)
		return
	}
}
