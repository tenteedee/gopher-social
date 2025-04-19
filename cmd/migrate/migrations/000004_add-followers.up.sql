CREATE TABLE IF NOT EXISTS followers (
  user_id BIGSERIAL NOT NULL,
  follower_id BIGSERIAL NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),

  PRIMARY KEY (user_id, follower_id), -- Composite primary key to ensure unique pairs
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  FOREIGN KEY (follower_id) REFERENCES users (id) ON DELETE CASCADE
)