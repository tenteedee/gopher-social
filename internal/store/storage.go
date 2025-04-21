package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrorNotFound          = errors.New("resource not found")
	ErrConflict            = errors.New("resource already exists")
	QueryTimeoutDuration   = time.Second * 5
	ErrorDuplicateEmail    = errors.New("email already used")
	ErrorDuplicateUsername = errors.New("username already exists")
)

type Storage struct {
	Post interface {
		Create(context.Context, *Post) (*CreatePostResponse, error)
		GetByID(context.Context, int64) (*Post, error)
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
		GetByUserId(context.Context, int64, PaginationFeedQuery) ([]*PostWithMetadata, error)
	}

	User interface {
		Create(context.Context, *sql.Tx, *User) error
		GetById(context.Context, int64) (*User, error)
		CreateAndInvite(context.Context, *User, string, time.Duration) error
		Activate(context.Context, string) error
		Delete(context.Context, int64) error
	}

	Comment interface {
		GetCommentByPostId(context.Context, int64) ([]Comment, error)
		Create(context.Context, *Comment) error
	}

	Follow interface {
		Follow(context.Context, int64, int64) error
		Unfollow(context.Context, int64, int64) error
	}
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		Post:    &PostStore{db: db},
		User:    &UserStore{db: db},
		Comment: &CommentStore{db: db},
		Follow:  &FollowStore{db: db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}
	return tx.Commit()
}
