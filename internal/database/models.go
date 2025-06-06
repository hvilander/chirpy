// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"database/sql"

	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
	Body      string
	UserID    uuid.NullUUID
}

type RefreshToken struct {
	Token     string
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
	UserID    uuid.NullUUID
	ExpiresAt sql.NullTime
	RevokedAt sql.NullTime
}

type User struct {
	ID             uuid.UUID
	CreatedAt      sql.NullTime
	UpdatedAt      sql.NullTime
	Email          string
	HashedPassword sql.NullString
	IsChirpyRed    sql.NullBool
}
