-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS onedrive_integrations (
    id SERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    user_id TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_onedrive_integrations_owner ON onedrive_integrations(owner_id);

-- Create a function to automatically update the updated_at timestamp
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_timestamp
BEFORE UPDATE ON onedrive_integrations
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS set_timestamp ON onedrive_integrations;

DROP FUNCTION IF EXISTS trigger_set_timestamp();

DROP INDEX IF EXISTS idx_onedrive_integrations_owner;

DROP TABLE IF EXISTS onedrive_integrations;
-- +goose StatementEnd
