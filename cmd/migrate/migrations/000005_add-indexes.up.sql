-- Create the extension and indexes for full text search
-- https://niallburkley.com/blog/index-columns-for-like-in-postgres/

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_comments_content ON comments USING gin (content gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_posts_title ON posts USING gin (title gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_posts_tags ON posts USING gin (tags);

CREATE INDEX IF NOT EXISTS idx_user_username ON users(username);

CREATE INDEX IF NOT EXISTS idx_post_user_id ON posts(user_id);

CREATE INDEX IF NOT EXISTS idx_comment_post_id ON comments(post_id);