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

-- name: UpdateUser :one
UPDATE users
SET
	updated_at = now(),
	email = $1,
	hashed_password = $2
WHERE id = $3

RETURNING id, created_at, updated_at, email, is_chirpy_red;


-- name: SetIsRed :one
UPDATE users
SET
	updated_at = now(),
	is_chirpy_red = $1
WHERE id = $2

RETURNING id, updated_at, is_chirpy_red;
	



