package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

/*
func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrongPassword",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
*/

func TestMakeJWT(t *testing.T) {
	//make takes in a uuid.UUID a token secret and a time.Duration
	uuidOne, err := uuid.NewUUID()
	if err != nil {
		t.Error("could not gen uuid")
	}
	tenSec, err := time.ParseDuration("10s")
	//oneNS, err := time.ParseDuration("1")

	if err != nil {
		t.Error("like what am i even doing here?")
	}

	tests := []struct {
		tUUID       uuid.UUID
		exp         time.Duration
		wantErr     bool
		tokenSecret string
		name        string
	}{
		{
			name:        "test basic jwt",
			tUUID:       uuidOne,
			exp:         tenSec,
			wantErr:     false,
			tokenSecret: "idontkonw?",
		},
		//{ name:        "fail by exp 1ns",
		//	tUUID:       uuidOne,
		//	exp:         oneNS,
		//	wantErr:     true,
		//	tokenSecret: "idontkonw?",
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println(tt.tUUID)
			myJWT, err := MakeJWT(tt.tUUID, tt.tokenSecret, tt.exp)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeJWT() error = %v, wantErr %v", err, tt.wantErr)
			}

			uuidOut, err := ValidateJWT(myJWT, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatejwt() error = %v, wantErr %v", err, tt.wantErr)
			}

			if uuidOut != tt.tUUID {
				t.Error("validatejwt() non matching uuids")
			}

		})
	}

}
