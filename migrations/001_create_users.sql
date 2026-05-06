CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email                TEXT UNIQUE NOT NULL,
    password_hash        TEXT NOT NULL,
    display_name         TEXT NOT NULL DEFAULT '',
    subscription_tier    TEXT NOT NULL DEFAULT 'free',
    subscription_expires_at TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
