CREATE TABLE IF NOT EXISTS onedrive_integrations (
    id SERIAL PRIMARY KEY,
    owner_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,          -- ID of the user who authorized this integration
    encrypted_refresh_token TEXT NOT NULL,  -- Encrypted refresh token (long-lived)
    encrypted_access_token TEXT,            -- Encrypted access token (short-lived)
    access_token_expires_at TIMESTAMPTZ,    -- When the access token expires
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique index on owner_id to ensure one integration per owner
CREATE UNIQUE INDEX idx_onedrive_integrations_owner ON onedrive_integrations(owner_id);

-- Index for finding active integrations
CREATE INDEX idx_onedrive_integrations_active ON onedrive_integrations(is_active);

-- Create a function to automatically update the updated_at timestamp
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add trigger to automatically update updated_at on row update
CREATE TRIGGER set_timestamp
BEFORE UPDATE ON onedrive_integrations
FOR EACH ROW
EXECUTE FUNCTION trigger_set_timestamp();