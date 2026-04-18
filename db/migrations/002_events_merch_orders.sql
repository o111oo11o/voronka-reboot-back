-- +goose Up

-- Events: one per calendar day (enforced by UNIQUE on event_date)
CREATE TABLE IF NOT EXISTS events (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    title       TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    event_date  DATE        NOT NULL UNIQUE,
    location    TEXT        NOT NULL DEFAULT '',
    type        TEXT        NOT NULL DEFAULT '',
    image_url   TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS events_type_idx ON events (type);

-- Merch items: stock cannot go negative (CHECK constraint)
CREATE TABLE IF NOT EXISTS merch_items (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    price_cents BIGINT      NOT NULL CHECK (price_cents >= 0),
    stock       INT         NOT NULL DEFAULT 0 CHECK (stock >= 0),
    image_urls  TEXT[]      NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Order status enum
CREATE TYPE order_status AS ENUM ('pending', 'paid', 'fulfilled', 'cancelled');

-- Orders
CREATE TABLE IF NOT EXISTS orders (
    id             UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_name  TEXT         NOT NULL,
    customer_email TEXT         NOT NULL,
    user_id        UUID         REFERENCES users(id) ON DELETE SET NULL,
    status         order_status NOT NULL DEFAULT 'pending',
    total_cents    BIGINT       NOT NULL,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS orders_status_idx        ON orders (status);
CREATE INDEX IF NOT EXISTS orders_customer_email_idx ON orders (customer_email);

-- Order line items (price snapshotted at purchase time)
CREATE TABLE IF NOT EXISTS order_items (
    id                  UUID   PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id            UUID   NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    merch_item_id       UUID   NOT NULL REFERENCES merch_items(id),
    quantity            INT    NOT NULL CHECK (quantity > 0),
    price_at_time_cents BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS order_items_order_id_idx ON order_items (order_id);

-- +goose Down
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TYPE  IF EXISTS order_status;
DROP TABLE IF EXISTS merch_items;
DROP TABLE IF EXISTS events;
