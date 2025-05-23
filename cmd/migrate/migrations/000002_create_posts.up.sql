CREATE TABLE IF NOT EXISTS posts (
  id BIGSERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  user_id INT NOT NULL,
  tags TEXT[] NOT NULL,
  created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT now(),
  updated_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT now(),
  version INT NOT NULL DEFAULT 0,
  CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
)