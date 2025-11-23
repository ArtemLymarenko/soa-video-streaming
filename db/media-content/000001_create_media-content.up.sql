CREATE SCHEMA IF NOT EXISTS media_content;

-- Create categories table
CREATE TABLE IF NOT EXISTS media_content.categories (
    id TEXT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on name for faster lookups
CREATE INDEX IF NOT EXISTS idx_categories_name ON media_content.categories(name);

-- Create index on updated_at for timestamp queries
CREATE INDEX IF NOT EXISTS idx_categories_updated_at ON media_content.categories(updated_at);

-- Create media_content table
CREATE TABLE IF NOT EXISTS media_content.media_content (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('movie', 'series')),
    duration INTEGER NOT NULL CHECK (duration > 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for media_content
CREATE INDEX IF NOT EXISTS idx_media_content_type ON media_content.media_content(type);
CREATE INDEX IF NOT EXISTS idx_media_content_name ON media_content.media_content(name);

-- Create media_content_categories junction table
CREATE TABLE IF NOT EXISTS media_content.media_content_categories (
    media_content_id TEXT NOT NULL REFERENCES media_content.media_content(id) ON DELETE CASCADE,
    category_id TEXT NOT NULL REFERENCES media_content.categories(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (media_content_id, category_id)
);

-- Create indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_media_content_categories_media_id
    ON media_content.media_content_categories(media_content_id);
CREATE INDEX IF NOT EXISTS idx_media_content_categories_category_id 
    ON media_content.media_content_categories(category_id);

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
CREATE TRIGGER update_categories_updated_at
    BEFORE UPDATE ON media_content.categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_media_content_updated_at
    BEFORE UPDATE ON media_content.media_content
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
