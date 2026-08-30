-- +goose Up
ALTER TABLE app_accounts ADD COLUMN oidc_subject TEXT;
UPDATE app_accounts SET oidc_subject = 'legacy:' || id::text WHERE oidc_subject IS NULL;
ALTER TABLE app_accounts ALTER COLUMN oidc_subject SET NOT NULL;
ALTER TABLE app_accounts ALTER COLUMN oidc_subject DROP DEFAULT;
CREATE UNIQUE INDEX app_accounts_oidc_subject_key ON app_accounts (oidc_subject);

CREATE SEQUENCE IF NOT EXISTS app_accounts_id_seq;
SELECT setval('app_accounts_id_seq', COALESCE((SELECT MAX(id) FROM app_accounts), 0) + 1, false);
ALTER TABLE app_accounts ALTER COLUMN id SET DEFAULT nextval('app_accounts_id_seq');
ALTER SEQUENCE app_accounts_id_seq OWNED BY app_accounts.id;

-- +goose Down
ALTER TABLE app_accounts ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS app_accounts_id_seq;
DROP INDEX app_accounts_oidc_subject_key;
ALTER TABLE app_accounts DROP COLUMN oidc_subject;
