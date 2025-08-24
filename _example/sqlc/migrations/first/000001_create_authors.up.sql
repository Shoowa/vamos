CREATE TABLE IF NOT EXISTS authors (
    id UUID DEFAULT uuidv7() PRIMARY KEY,
    name text NOT NULL,
    bio text
);
