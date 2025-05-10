-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (gen_random_uuid(), now(), now(), $1, $2)

RETURNING *;


-- name: ResetUsers :exec
DELETE from users; 

-- name: GetUserByEmail :one
SELECT * FROM users where email = $1;

-- name: GetUserByID :one
SELECT * FROM users where id = $1;
