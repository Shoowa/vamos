CREATE TABLE IF NOT EXISTS books (
    id BIGSERIAL PRIMARY KEY,
    authorID UUID REFERENCES authors(id),
    title TEXT NOT NULL,
    edition SMALLINT,
    volume SMALLINT,
    year DATE,
    wordcount INT
);
