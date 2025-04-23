package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"
)

func (store *UserStore) createUserInvitation(ctx context.Context, tx *sql.Tx, userID int64, hashedToken string, exp time.Duration) error {
	query := `
		INSERT INTO invitations (token, user_id, expiry)
		VALUES ($1, $2, $3)
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, hashedToken, userID, time.Now().Add(exp))
	if err != nil {
		return err
	}

	return nil
}

func (store *UserStore) getUserFromInvitation(ctx context.Context, tx *sql.Tx, token string) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.is_activated
		FROM users u
		JOIN invitations i ON u.id = i.user_id
		WHERE i.token = $1
	`

	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}

	err := tx.QueryRowContext(
		ctx,
		query,
		hashedToken,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.IsActivated,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}

	return user, nil

}
