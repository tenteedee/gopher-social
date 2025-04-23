package store

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Password    password `json:"-"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
	IsActivated bool     `json:"is_activated"`
	RoleID      int64    `json:"role_id"`
	Role        Role     `json:"role"`
}

type password struct {
	text *string
	hash []byte
}

func (pw *password) Set(text string) error {
	if text == "" {
		return errors.New("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	pw.text = &text
	pw.hash = hash

	return nil
}

func (p *password) Compare(text string) error {
	if p.hash == nil {
		return errors.New("password hash is not set")
	}

	return bcrypt.CompareHashAndPassword(p.hash, []byte(text))
}

type UserStore struct {
	db *sql.DB
}

func (store *UserStore) Create(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		INSERT INTO users (username, email, password_hash, role_id)
		VALUES ($1, $2, $3, (
			SELECT id FROM roles WHERE name = $4
		))
		RETURNING id, created_at
		`

	role := user.Role.Name
	if role == "" {
		role = "user"
	}

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := tx.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password.hash,
		role,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	if err != nil {
		switch err.Error() {
		case "pq: duplicate key value violates unique constraint \"users_email_key\"":
			return ErrorDuplicateEmail
		case "pq: duplicate key value violates unique constraint \"users_username_key\"":
			return ErrorDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (store *UserStore) GetById(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.created_at, u.updated_at, r.id, r.name, r.level 
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
		AND u.is_activated = true
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := store.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Level,
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

func (store *UserStore) CreateAndInvite(ctx context.Context, user *User, hashedToken string, invitationExp time.Duration) error {
	return withTx(store.db, ctx, func(tx *sql.Tx) error {
		// create a user
		if err := store.Create(ctx, tx, user); err != nil {
			return err
		}

		// create user invitation
		if err := store.createUserInvitation(ctx, tx, user.ID, hashedToken, invitationExp); err != nil {
			return err
		}

		return nil
	})
}

func (store *UserStore) Activate(ctx context.Context, token string) error {
	return withTx(store.db, ctx, func(tx *sql.Tx) error {
		// find the user that the token belongs to
		user, err := store.getUserFromInvitation(ctx, tx, token)
		if err != nil {
			log.Print("error get user")
			return err
		}

		// update the user to activated
		user.IsActivated = true
		if err := store.update(ctx, tx, user); err != nil {
			log.Print("error update user")
			return err
		}

		// delete the invitation
		if err := store.deleteUserInvitation(ctx, tx, user.ID); err != nil {
			log.Print("error delete invitation")
			return err
		}

		return nil

	})

}

func (store *UserStore) update(ctx context.Context, tx *sql.Tx, user *User) error {
	query := `
		UPDATE users
		SET username = $2, email = $3, updated_at = $4, is_activated = $5
		WHERE id = $1
		`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, user.ID, user.Username, user.Email, time.Now(), user.IsActivated)
	if err != nil {
		return err
	}

	return nil
}

func (store *UserStore) deleteUserInvitation(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `
		DELETE FROM invitations
		WHERE user_id = $1
		`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) Delete(ctx context.Context, userID int64) error {
	return withTx(s.db, ctx, func(tx *sql.Tx) error {
		if err := s.delete(ctx, tx, userID); err != nil {
			return err
		}

		if err := s.deleteUserInvitations(ctx, tx, userID); err != nil {
			return err
		}

		return nil
	})
}

func (s *UserStore) deleteUserInvitations(ctx context.Context, tx *sql.Tx, userID int64) error {
	query := `DELETE FROM user_invitations WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) delete(ctx context.Context, tx *sql.Tx, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, username, email, password_hash, created_at FROM users
		WHERE email = $1 AND is_activated = true
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	user := &User{}
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
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
