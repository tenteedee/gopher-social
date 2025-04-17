package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrorNotFound = errors.New("resource not found")

	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Post interface {
		Create(context.Context, *Post) (*CreatePostResponse, error)
		GetByID(context.Context, int64) (*Post, error)
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
	}

	User interface {
		Create(context.Context, *User) error
	}

	Comment interface {
		GetCommentByPostId(context.Context, int64) ([]Comment, error)
	}
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		Post:    &PostStore{db: db},
		User:    &UserStore{db: db},
		Comment: &CommentStore{db: db},
	}
}
