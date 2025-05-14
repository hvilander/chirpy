package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {

	dat, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(dat), nil

}

func CheckPasswordHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {

	registeredClaims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, registeredClaims)

	signed, err := token.SignedString([]byte(tokenSecret))

	return signed, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(token *jwt.Token) (interface{}, error) { return []byte(tokenSecret), nil },
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("For Token: %s \nError parsing with claims: %w\n", tokenString, err)
	}

	strUUID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}

	if issuer != "chirpy" {
		return uuid.Nil, errors.New("invalid issuer")
	}

	id, err := uuid.Parse(strUUID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	bTok := headers.Get("Authorization")
	if bTok == "" {
		return "", fmt.Errorf("no auth header")
	}
	spt := strings.Split(bTok, " ")

	if spt[0] != "Bearer" || len(spt) != 2 {
		return "", fmt.Errorf("token parsing error")
	}
	return spt[1], nil
}

func MakeRefreshToken() (string, error) {
	randN := make([]byte, 32)
	_, err := rand.Read(randN)
	if err != nil {
		return "", fmt.Errorf("rand err: %w", err)

	}
	encoded := hex.EncodeToString(randN)
	return encoded, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	bTok := headers.Get("Authorization")
	if bTok == "" {
		return "", fmt.Errorf("no auth header")
	}
	spt := strings.Split(bTok, " ")

	if spt[0] != "ApiKey" || len(spt) != 2 {
		return "", fmt.Errorf("parsing error")
	}
	return spt[1], nil
}
