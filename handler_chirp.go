package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/hvilander/chirpy/internal/database"
)

type postChirpBody struct {
	Body   string `json:"body"`
	UserId string `json:"user_id"`
}

type chirpCreatedRes struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body      string `json:"body"`
	UserId    string `json:"user_id"`
}

func getNullUUID(id uuid.UUID) uuid.NullUUID {
	return uuid.NullUUID{
		UUID:  id,
		Valid: true,
	}
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.db.GetAllChirps(req.Context())
	if err != nil {
		fmt.Println("error getting chiprs:", err)
		w.WriteHeader(500)
		return
	}

	resChirps := make([]chirpCreatedRes, len(chirps))
	for i, c := range chirps {
		resChirps[i] = chirpCreatedRes{
			ID:        c.ID.String(),
			CreatedAt: c.CreatedAt.Time.String(),
			UpdatedAt: c.UpdatedAt.Time.String(),
			Body:      c.Body,
			UserId:    c.UserID.UUID.String(),
		}
	}

	respondWithJson(w, 200, resChirps)

}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := postChirpBody{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Println("error decoding:", err)
		w.WriteHeader(500)
		return
	}

	userID, err := uuid.Parse(params.UserId)
	if err != nil {
		fmt.Println("error parsing user id:", err)
		w.WriteHeader(500)
		return
	}

	args := database.CreateChirpParams{
		UserID: getNullUUID(userID),
		Body:   params.Body,
	}

	chirp, err := cfg.db.CreateChirp(req.Context(), args)
	if err != nil {
		fmt.Println(err)
		respondWithError(w, 500, "error creating chirp")
	}

	responseBody := chirpCreatedRes{
		ID:        chirp.ID.String(),
		CreatedAt: chirp.CreatedAt.Time.String(),
		UpdatedAt: chirp.UpdatedAt.Time.String(),
		Body:      chirp.Body,
		UserId:    chirp.UserID.UUID.String(),
	}

	respondWithJson(w, 201, responseBody)

}
