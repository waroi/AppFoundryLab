-- Deprecated bootstrap file.
-- Active schema management moved to:
-- backend/services/api-gateway/internal/database/migrations/

CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(120) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO users (name, email)
VALUES
  ('Ada Lovelace', 'ada@example.com'),
  ('Alan Turing', 'alan@example.com')
ON CONFLICT (email) DO NOTHING;
