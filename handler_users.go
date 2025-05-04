package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/hvilander/chirpy/internal/auth"
	"github.com/hvilander/chirpy/internal/database"
)

type postUserBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userCreatedRes struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := postUserBody{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Println("error decoding params:", err)
		w.WriteHeader(500)
		return
	}
	user, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
	if err != nil {

		fmt.Println("error user:", params.Email, "not found:", err)
		w.WriteHeader(500)
		return
	}

	// for the pw check the error will be nil if the hash and pw match
	err = auth.CheckPasswordHash(user.HashedPassword.String, params.Password)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	resUser := userCreatedRes{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt.Time.String(),
		UpdatedAt: user.UpdatedAt.Time.String(),
		Email:     user.Email,
	}

	respondWithJson(w, 200, resUser)

}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, req *http.Request) {
	platform := os.Getenv("PLATFORM")

	if platform != "dev" {
		respondWithError(w, 401, "DOH")
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := postUserBody{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Println("error decoding params:", err)
		w.WriteHeader(500)
		return
	}
	shhh, err := auth.HashPassword(params.Password)
	if err != nil {
		fmt.Println("error hashing password:", err)
		w.WriteHeader(500)
		return

	}
	args := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: getNullString(shhh),
	}

	user, err := cfg.db.CreateUser(req.Context(), args)
	if err != nil {
		fmt.Println(err)
		respondWithError(w, 500, "error creating user")
	}

	responseBody := userCreatedRes{
		ID:        user.ID.String(),
		CreatedAt: user.CreatedAt.Time.String(),
		UpdatedAt: user.UpdatedAt.Time.String(),
		Email:     user.Email,
	}

	respondWithJson(w, 201, responseBody)
}
