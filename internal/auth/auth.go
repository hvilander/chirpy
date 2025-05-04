package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {

	dat, err := bcrypt.GenerateFromPassword([]byte(password), 6)
	if err != nil {
		return "", err
	}
	return string(dat), nil

}

func CheckPasswordHash(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
