-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- Roles lookup table
-- Seeded with the three actor types: rider, driver, admin
-- ============================================================
CREATE TABLE roles (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name) VALUES ('rider'), ('driver'), ('admin');

-- ============================================================
-- Users table
-- Stores identity only — no profile data (that lives in
-- domain-specific services like trip-service, driver-service).
-- ============================================================
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name       VARCHAR(255) NOT NULL,
    phone           TEXT UNIQUE NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    role_id         UUID NOT NULL REFERENCES roles(id),
    email_confirmed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_users_phone_e164
        CHECK (phone ~ '^\+[1-9][0-9]{1,14}$')
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role_id ON users(role_id);

-- ============================================================
-- Email OTP verification codes
-- 6-digit codes with a 10-minute expiry window.
-- ============================================================
CREATE TABLE email_otps (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code       VARCHAR(6) NOT NULL,
    used       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_email_otps_user_id ON email_otps(user_id);
CREATE INDEX idx_email_otps_expires_at ON email_otps(expires_at);

-- ============================================================
-- Refresh tokens
-- Implements rotation with reuse detection via family_id.
-- When a token is refreshed, the old row gets a replaced_by
-- pointer. If a revoked/replaced token is reused, the entire
-- family is revoked (stolen token detection).
-- ============================================================
CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL,
    family_id   UUID NOT NULL,
    replaced_by UUID REFERENCES refresh_tokens(id),
    revoked     BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_family_id ON refresh_tokens(family_id);
