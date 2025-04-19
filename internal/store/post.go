package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int64     `json:"version"`
	Comments  []Comment `json:"comments"`
	User      *User     `json:"user"`
}

type PostWithMetadata struct {
	Post
	User
	CommentsCount int64 `json:"comments_count"`
}

type CreatePostResponse struct {
	ID        int64  `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type PostStore struct {
	db *sql.DB
}

func NewPostStore(db *sql.DB) *PostStore {
	return &PostStore{db: db}
}

func (store *PostStore) Create(ctx context.Context, post *Post) (*CreatePostResponse, error) {
	query := `
		INSERT INTO posts (content, title, user_id, tags)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
		`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	response := &CreatePostResponse{}

	row := store.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	)

	err := row.Scan(
		&response.ID,
		&response.CreatedAt,
		&response.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return response, nil
}

func (store *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	query := `
		SELECT id, content, title, user_id, tags, created_at, updated_at, version
		FROM posts
		WHERE id = $1
		`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	post := Post{}
	err := store.db.QueryRowContext(
		ctx,
		query,
		id).
		Scan(
			&post.ID,
			&post.Content,
			&post.Title,
			&post.UserID,
			pq.Array(&post.Tags),
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Version,
		)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrorNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (store *PostStore) Delete(ctx context.Context, id int64) error {
	query := `
		DELETE FROM posts
		WHERE id = $1
		`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	result, err := store.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrorNotFound
	}
	return nil
}

func (store *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1,content = $2, tags = $3, version = version + 1
		WHERE id = $4
		AND version = $5
		RETURNING version
		`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := store.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		pq.Array(post.Tags),
		post.ID,
		post.Version,
	).Scan(
		&post.Version,
	)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return ErrorNotFound
		default:
			return err
		}
	}

	return nil

}

func (store *PostStore) GetByUserId(ctx context.Context, userID int64, fq PaginationFeedQuery) ([]*PostWithMetadata, error) {
	query := `
		SELECT 
			p.id, p.title, p."content", p.tags, p."version", p.created_at,
			u."id" AS user_id, u.username, u.email,
			COUNT(c."id") AS comments_count
		FROM posts p
		LEFT JOIN comments c ON p."id" = c.post_id
		LEFT JOIN users u ON p.user_id = u."id"
		JOIN followers f ON p.user_id = f.user_id OR p.user_id = $1
		WHERE 
			f.user_id= $1 
			AND (p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')
			AND (p.tags @> $5 OR $5 = '{}')
			AND (p.created_at >= $6 OR $6 IS NULL)
		GROUP BY p.id, u."id"
		ORDER BY p.created_at ` + fq.Sort + `
		LIMIT $2 OFFSET $3
		`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := store.db.QueryContext(
		ctx,
		query,
		userID,
		fq.Limit,
		fq.Offset,
		fq.Search,
		pq.Array(fq.Tags),
		fq.Since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*PostWithMetadata
	for rows.Next() {
		var post PostWithMetadata
		if err := rows.Scan(
			&post.Post.ID,
			&post.Post.Title,
			&post.Post.Content,
			pq.Array(&post.Post.Tags),
			&post.Post.Version,
			&post.Post.CreatedAt,
			&post.User.ID,
			&post.User.Username,
			&post.User.Email,
			&post.CommentsCount,
		); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	return posts, nil
}
