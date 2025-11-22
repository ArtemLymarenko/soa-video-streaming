-- Drop triggers
DROP TRIGGER IF EXISTS update_media_content_updated_at ON media_content;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_media_content_categories_category_id;
DROP INDEX IF EXISTS idx_media_content_categories_media_id;
DROP INDEX IF EXISTS idx_media_content_name;
DROP INDEX IF EXISTS idx_media_content_type;
DROP INDEX IF EXISTS idx_categories_updated_at;
DROP INDEX IF EXISTS idx_categories_name;

-- Drop tables (in reverse order due to foreign key constraints)
DROP TABLE IF EXISTS media_content_categories;
DROP TABLE IF EXISTS media_content;
DROP TABLE IF EXISTS categories;
