package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hvilander/chirpy/internal/auth"
	"github.com/hvilander/chirpy/internal/database"
)

type postUserBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type response struct {
	ID           string `json:"id"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	IsChirpyRed  bool   `json:"is_chirpy_red"`
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

	expireTime, err := time.ParseDuration("1h")

	token, err := auth.MakeJWT(user.ID, cfg.secret, expireTime)
	if err != nil {
		w.WriteHeader(403)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	expireDuration, err := time.ParseDuration("1440h")
	timeTheTokExpires := time.Now().Add(expireDuration)
	nullTime := sql.NullTime{
		// 1440 hours is 60 days
		Time:  timeTheTokExpires,
		Valid: true,
	}
	args := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    getNullUUID(user.ID),
		ExpiresAt: nullTime,
	}

	savedRefreshToken, err := cfg.db.CreateRefreshToken(req.Context(), args)
	if err != nil {
		fmt.Println("error saving refresh token: ", err)
		w.WriteHeader(500)
		return
	}

	resUser := response{
		ID:           user.ID.String(),
		CreatedAt:    user.CreatedAt.Time.String(),
		UpdatedAt:    user.UpdatedAt.Time.String(),
		Email:        user.Email,
		Token:        token,
		RefreshToken: savedRefreshToken.Token,
		IsChirpyRed:  user.IsChirpyRed.Bool,
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

	responseBody := response{
		ID:          user.ID.String(),
		CreatedAt:   user.CreatedAt.Time.String(),
		UpdatedAt:   user.UpdatedAt.Time.String(),
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
	}

	respondWithJson(w, 201, responseBody)
}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, req *http.Request) {
	//get token off of headers
	bTok, err := auth.GetBearerToken(req.Header)
	if err != nil {
		fmt.Println(err)
		respondWithError(w, 500, "error with refresh header")
		return
	}

	//check for it in the db
	refreshToken, err := cfg.db.GetRefreshTokenById(req.Context(), bTok)
	if err != nil {
		fmt.Println(err)
		//if it doesnt exist or has expired respond with a 401
		respondWithError(w, 401, "getting db ref tok failed")
		return
	}

	if refreshToken.RevokedAt.Valid {
		fmt.Println("token revoked")
		respondWithError(w, 401, "getting db ref tok failed")
		return
	}

	if refreshToken.ExpiresAt.Valid && refreshToken.ExpiresAt.Time.Before(time.Now()) {
		fmt.Println("token revoked")
		respondWithError(w, 401, "getting db ref tok failed")
		return
	}

	//otherwise respond with a 200 and a new access token {token: alskdjflksfl}
	expireTime, err := time.ParseDuration("1h")

	userID := refreshToken.UserID.UUID
	accessToken, err := auth.MakeJWT(userID, cfg.secret, expireTime)
	if err != nil {
		w.WriteHeader(403)
		return
	}

	responseBody := response{
		Token: accessToken,
	}
	respondWithJson(w, 200, responseBody)
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, req *http.Request) {
	bTok, err := auth.GetBearerToken(req.Header)
	if err != nil {
		fmt.Println(err)
		respondWithError(w, 500, "error with refresh header")
	}

	err = cfg.db.RevokeRefreshToken(req.Context(), bTok)
	if err != nil {
		fmt.Println(err)
		respondWithError(w, 500, "error with refresh header")
	}

	w.WriteHeader(204)

}

func (cfg *apiConfig) handlePutUser(w http.ResponseWriter, req *http.Request) {
	//get token off of headers
	bTok, err := auth.GetBearerToken(req.Header)
	if bTok == "" {
		respondWithError(w, 401, "missing auth header")
		return
	}
	if err != nil {
		fmt.Println(err)
		respondWithError(w, 500, "error with refresh header")
		return
	}

	userID, err := auth.ValidateJWT(bTok, cfg.secret)
	if err != nil {
		fmt.Println("invalid token:", err)
		w.WriteHeader(401)
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := postUserBody{}
	err = decoder.Decode(&params)
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

	args := database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: getNullString(shhh),
		ID:             userID,
	}

	result, err := cfg.db.UpdateUser(req.Context(), args)
	if err != nil {
		fmt.Println("error updating user:", err)
		w.WriteHeader(500)
		return
	}

	resUser := response{
		ID:          result.ID.String(),
		CreatedAt:   result.CreatedAt.Time.String(),
		UpdatedAt:   result.UpdatedAt.Time.String(),
		Email:       result.Email,
		IsChirpyRed: result.IsChirpyRed.Bool,
	}

	respondWithJson(w, 200, resUser)

}

type webhookData struct {
	UserId string `json:"user_id"`
}

type redWebhookReq struct {
	Event string      `json:"event"`
	Data  webhookData `json:"data"`
}

func (cfg *apiConfig) handleChripyRedWebHook(w http.ResponseWriter, req *http.Request) {
	apiKey, err := auth.GetAPIKey(req.Header)
	fmt.Println("apikey:", apiKey, "cfg.polkaKey", cfg.polkaKey, "error:", err)
	if err != nil || apiKey != cfg.polkaKey {
		fmt.Println("apikey doesnt match, potential err:", err)
		w.WriteHeader(401)
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := redWebhookReq{}
	err = decoder.Decode(&params)
	if err != nil {
		fmt.Println("error decoding params:", err)
		w.WriteHeader(500)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	//get user id
	userID, err := uuid.Parse(params.Data.UserId)
	if err != nil {
		fmt.Println("error parsing user id:", err)
		w.WriteHeader(500)
		return
	}

	args := database.SetIsRedParams{
		IsChirpyRed: sql.NullBool{Valid: true, Bool: true},
		ID:          userID,
	}

	_, err = cfg.db.SetIsRed(req.Context(), args)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("no row error:", err)
			w.WriteHeader(404)
			return
		}
		fmt.Println("error writing to the db:", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(204)

}
