CREATE SCHEMA IF NOT EXISTS media_content;

-- Create user_storage_accounts table
CREATE TABLE IF NOT EXISTS media_content.user_storage_accounts (
    user_id UUID PRIMARY KEY,
    bucket_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on bucket_name for lookups
CREATE INDEX IF NOT EXISTS idx_user_storage_accounts_bucket_name ON media_content.user_storage_accounts(bucket_name);
CREATE INDEX IF NOT EXISTS idx_user_storage_accounts_status ON media_content.user_storage_accounts(status);
