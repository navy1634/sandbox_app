-- +goose Up
CREATE TABLE app_accounts (
    id BIGINT PRIMARY KEY,
    email TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL DEFAULT '',
    picture TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX app_accounts_email_idx ON app_accounts (email);

-- +goose Down
DROP TABLE app_accounts;
