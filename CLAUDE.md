# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./...

# Vet
go vet ./...

# Run
go run ./cmd/app

# Sync dependencies
go mod tidy

# Apply DB migrations (requires goose)
go run github.com/pressly/goose/v3/cmd/goose@latest postgres "$POSTGRES_DSN" up -dir db/migrations

# Regenerate sqlc query wrappers (run after editing db/queries/*.sql)
sqlc generate
```

There are no tests yet. When tests are added, run them with `go test ./...` and a single package with `go test ./internal/service/...`.

## Architecture

Single binary (`cmd/app/main.go`) that starts three transports concurrently: a Gin HTTP server (REST + WebSocket), and a Telegram bot polling loop. All share the same service layer.

### Layer isolation

```
handler/{rest,ws,telegram}  →  service (interfaces)  →  repository (interfaces)  →  postgres
```

- **Handlers** depend only on `service.XxxService` interfaces — never on concrete types or repos.
- **Services** depend only on `repository.XxxRepository` interfaces.
- **`internal/app/wire.go`** is the single place that names concrete types and wires everything together. The init order matters: the Telegram bot must be created first (to get its username for deeplinks), then `AuthService` is created with the bot as its `MessageSender`, then `bot.SetAuth(authSvc)` is called before the bot starts polling.
- Adding a new domain means: `domain/` structs → append to `repository/interfaces.go` and `service/interfaces.go` → postgres impl → service impl → handler → register in `router.go` → wire in `wire.go`.

### Key packages

- `internal/domain/` — pure structs, request/filter types, and sentinel errors (`ErrNotFound`, `ErrConflict`, `ErrBadRequest`, `ErrForbidden`). No DB tags, no framework imports.
- `internal/repository/interfaces.go` — all repository interfaces in one file.
- `internal/service/interfaces.go` — all service interfaces in one file. Also defines `MessageSender` (implemented by the bot to break the auth↔bot import cycle).
- `internal/repository/postgres/` — pgx-backed implementations. Error translation: `pgx.ErrNoRows` → `domain.ErrNotFound`. `merch_item.go` uses an atomic `UPDATE … WHERE stock >= $2` to prevent overselling. `order.go` runs a pgx transaction with `SELECT FOR UPDATE` to lock items before decrement.
- `internal/handler/rest/` — Gin handlers. `writeError()` in `user.go` maps domain errors to HTTP status codes and is used by all REST handlers. Admin routes live under `/api/v1/admin/` and are protected by `middleware.Admin` (constant-time `X-Admin-Token` check). Protected user routes use `middleware.Auth(authSvc)` which validates Bearer JWTs.
- `internal/handler/ws/` — gorilla/websocket. `Hub.Run()` must be started in a goroutine before accepting connections.
- `internal/config/config.go` — viper-backed; reads env vars (`HTTP_PORT`, `POSTGRES_DSN`, `TELEGRAM_TOKEN`, `ADMIN_TOKEN`, `JWT_SECRET`). Copy `.env.example` → `.env` to run locally.

### Auth flow

Auth is Telegram-based. Users are identified by `tg_id` and `tg_username` (no email/password).

**New user registration:**
1. `POST /api/v1/auth/register` `{"name": "...", "tg_username": "..."}` → `{"deeplink": "https://t.me/<bot>?start=<token>"}`
2. User clicks deeplink, presses Start in Telegram → bot receives `/start <token>`
3. Bot calls `AuthService.ConfirmTelegram` → verifies token, creates user, issues JWT pair, stores it in `confirmed_registrations` (10-min TTL); bot replies "Return to the browser"
4. Browser polls `POST /api/v1/auth/confirm` `{"token": "<token>"}` → `TokenPair` (single-use claim; returns 404 until confirmed)

**Existing user login:**
1. `POST /api/v1/auth/register` with a known `tg_username` → `{"code_sent": true}`
2. Bot delivers a 4-digit code (5-min TTL) to the user's Telegram
3. `POST /api/v1/auth/verify` `{"tg_username": "...", "code": "1234"}` → `TokenPair`

**Token refresh:** `POST /api/v1/auth/refresh` `{"refresh_token": "..."}` — rotates the refresh token (old one deleted from DB).

Access tokens are short-lived JWTs (15 min). Refresh tokens are JWTs (30 days) whose SHA-256 hash is stored in the `refresh_tokens` table for revocation.

### Database

Migrations use goose format (`-- +goose Up` / `-- +goose Down`) in `db/migrations/`. The bot name is **not** configured — it is read live from the Telegram API at startup via `bot.Username()`.

- `merch_items.image_urls` — Postgres `TEXT[]`, scanned to `[]string` via pgx array support
- `orders.status` — Postgres enum (`order_status`); serialised as a plain string in Go (`domain.OrderStatus`)
- `pending_registrations` — short-lived nonces (15 min) for the deeplink registration flow
- `confirmed_registrations` — stores issued token pair after bot confirmation (10-min TTL); claimed once by the browser via `POST /auth/confirm`, then deleted
- `login_codes` — one row per user (upserted), 4-digit code with 5-min TTL for returning-user login
- `refresh_tokens` — hashed refresh JWTs, deleted on use (rotation) or expiry

### Transports

| Transport | Entry point | Notes |
|-----------|-------------|-------|
| REST | `handler/rest/router.go` | Gin, `/api/v1/` prefix |
| WebSocket | `handler/ws/chat.go` | `GET /ws/chat`, broadcasts to all connected clients |
| Telegram | `handler/telegram/bot.go` | Long-polling; also implements `service.MessageSender` |
