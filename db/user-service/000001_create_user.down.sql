-- Drop triggers
DROP TRIGGER IF EXISTS update_user_info_updated_at ON user_info;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_user_preferred_categories_user_id;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS user_preferred_categories;
DROP TABLE IF EXISTS user_info;
DROP TABLE IF EXISTS users;
