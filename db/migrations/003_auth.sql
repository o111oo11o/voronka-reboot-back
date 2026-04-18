-- +goose Up

-- tg_username stored on users for display / uniqueness checks
ALTER TABLE users ADD COLUMN IF NOT EXISTS tg_username TEXT UNIQUE;

-- Pending registrations: created by POST /auth/register, consumed by bot /start <token>
CREATE TABLE IF NOT EXISTS pending_registrations (
    token       TEXT        PRIMARY KEY,
    name        TEXT        NOT NULL,
    tg_username TEXT        NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Hashed refresh tokens (SHA-256 of the raw JWT)
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT        NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS refresh_tokens_user_id_idx ON refresh_tokens (user_id);

-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS pending_registrations;
ALTER TABLE users DROP COLUMN IF EXISTS tg_username;
