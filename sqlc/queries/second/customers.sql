-- name: GetCustomer :one
SELECT * FROM customers WHERE forename = $1 LIMIT 1;

-- name: ListCustomers :many
SELECT * FROM customers ORDER BY forename;

-- name: CreateCustomer :execresult
INSERT INTO customers (forename, surname, dob, male) VALUES ($1, $2, $3, $4);

-- name: DeleteCustomer :exec
DELETE FROM customers WHERE id = $1;
