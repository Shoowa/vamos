-- name: GetAuthor :one
SELECT * FROM authors WHERE name = $1 LIMIT 1;

-- name: ListAuthors :many
SELECT * FROM authors ORDER BY name;

-- name: CreateAuthor :execresult
INSERT INTO authors (name, bio) VALUES ($1, $2);

-- name: DeleteAuthor :exec
DELETE FROM authors WHERE id = $1;

-- name: MostProductiveAuthor :one
SELECT authors.name
FROM authors
JOIN books ON books.authorID = authors.id
WHERE books.wordcount = (SELECT MAX(books.wordcount) FROM books);

-- name: MostProductiveAuthorAndBook :one
SELECT sqlc.embed(authors), sqlc.embed(books)
FROM authors
JOIN books ON books.authorID = authors.id
WHERE books.wordcount = (SELECT MAX(books.wordcount) FROM books);
