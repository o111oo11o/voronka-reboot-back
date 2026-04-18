-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    tg_id      bigint      NOT NULL UNIQUE,
    email      TEXT        UNIQUE,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS messages (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id    UUID        NOT NULL,
    author_id  UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS messages_room_id_created_at_idx ON messages (room_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS users;
