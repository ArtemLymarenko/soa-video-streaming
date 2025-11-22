-- Create users table
CREATE TABLE IF NOT EXISTS user_service.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON user_service.users(email);

-- Create user_info table
CREATE TABLE IF NOT EXISTS user_service.user_info (
    user_id UUID PRIMARY KEY REFERENCES user_service.users(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create user_preferred_categories table
CREATE TABLE IF NOT EXISTS user_service.user_preferred_categories (
    user_id UUID NOT NULL REFERENCES user_service.users(id) ON DELETE CASCADE,
    category_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, category_id)
);

-- Create index on user_id for faster category lookups
CREATE INDEX IF NOT EXISTS idx_user_preferred_categories_user_id ON user_service.user_preferred_categories(user_id);

-- Create trigger function to update updated_at timestamp
-- Note: PostgreSQL doesn't support ON UPDATE CURRENT_TIMESTAMP like MySQL
-- We need to use triggers for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for automatic updated_at updates
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON user_service.users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_info_updated_at
    BEFORE UPDATE ON user_service.user_info
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
