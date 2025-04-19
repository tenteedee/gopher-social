package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type FollowStore struct {
	db *sql.DB
}

type Follow struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func (store *FollowStore) Follow(ctx context.Context, followedUserID int64, followerID int64) error {
	query := `
		INSERT INTO followers (user_id, follower_id)
		VALUES ($1, $2)
		`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := store.db.ExecContext(ctx, query, followedUserID, followerID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
	}

	return nil
}

func (store *FollowStore) Unfollow(ctx context.Context, followedUserID, followerID int64) error {
	query := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2
		`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := store.db.ExecContext(ctx, query, followedUserID, followerID)
	if err != nil {
		return err
	}

	return nil
}
