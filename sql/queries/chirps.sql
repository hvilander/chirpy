-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (gen_random_uuid(), now(), now(), $1, $2)

RETURNING *;


-- name: GetAllChirps :many
SELECT * from chirps;


-- name: GetChirpById :one
SELECT * from chirps where id = $1;

-- name: DeleteChirpById :exec
DELETE FROM chirps where id = $1;
