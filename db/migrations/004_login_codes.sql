-- +goose Up

-- One active login code per user at a time (UNIQUE on user_id lets us upsert)
CREATE TABLE IF NOT EXISTS login_codes (
    user_id     UUID        PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    code        TEXT        NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS login_codes;
