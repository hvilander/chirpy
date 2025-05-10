-- name: CreateRefreshToken :one 
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES ($1, now(), now(), $2, $3, NULL) 

RETURNING *;

-- name: GetRefreshTokenById :one
SELECT * FROM refresh_tokens where token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET updated_at = now(), revoked_at = now() where token = $1;



