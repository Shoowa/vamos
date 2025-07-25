-- name: GetAuthor :one
SELECT * FROM authors WHERE name = $1 LIMIT 1;

-- name: ListAuthors :many
SELECT * FROM authors ORDER BY name;

-- name: CreateAuthor :execresult
INSERT INTO authors (name, bio) VALUES ($1, $2);

-- name: DeleteAuthor :exec
DELETE FROM authors WHERE id = $1;
