CREATE TABLE IF NOT EXISTS roles (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  level INT NOT NULL DEFAULT 0,
  description TEXT
);

INSERT INTO roles (name, description, level)
VALUES
  ('user', 'create posts, comments, follow', 1),
  ('moderator', 'update other users posts', 2),
  ('admin', 'update and delete users posts', 3)
  ;