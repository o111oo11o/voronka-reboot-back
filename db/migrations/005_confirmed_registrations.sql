-- +goose Up
CREATE TABLE IF NOT EXISTS confirmed_registrations (
    token         TEXT        PRIMARY KEY,
    access_token  TEXT        NOT NULL,
    refresh_token TEXT        NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS confirmed_registrations;
