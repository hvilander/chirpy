package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/google/uuid"
	"github.com/hvilander/chirpy/internal/auth"
	"github.com/hvilander/chirpy/internal/database"
)

type postChirpBody struct {
	Body   string `json:"body"`
	UserId string `json:"user_id"`
	Token  string `json:"token"`
}

type chirpCreatedRes struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Body      string `json:"body"`
	UserId    string `json:"user_id"`
}

type chirpGetReq struct {
	AuthorID string `json:"author_id"`
}

func getNullUUID(id uuid.UUID) uuid.NullUUID {
	return uuid.NullUUID{
		UUID:  id,
		Valid: true,
	}
}

func getNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, req *http.Request) {

	chirpId := req.PathValue("chirpID")
	if chirpId == "" {
		fmt.Println("error getting chipr: no id provided")
		w.WriteHeader(400)
		return
	}

	chirpUUID, err := uuid.Parse(chirpId)
	if err != nil {
		fmt.Println("error getting chirp:", err)
		w.WriteHeader(400)
		return
	}

	c, err := cfg.db.GetChirpById(req.Context(), chirpUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("chirp not found")
			w.WriteHeader(404)
			return
		}
		fmt.Println("chirpUUID", chirpUUID)
		fmt.Println("error querying chirp:", err)
		w.WriteHeader(400)
		return
	}

	resChirp := chirpCreatedRes{
		ID:        c.ID.String(),
		CreatedAt: c.CreatedAt.Time.String(),
		UpdatedAt: c.UpdatedAt.Time.String(),
		Body:      c.Body,
		UserId:    c.UserID.UUID.String(),
	}

	respondWithJson(w, 200, resChirp)
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, req *http.Request) {
	var chirps []database.Chirp
	var err error

	authorID := req.URL.Query().Get("author_id")
	sortDirection := req.URL.Query().Get("sort")

	if authorID != "" {
		userId, err := uuid.Parse(authorID)
		if err != nil {
			fmt.Println("error parsing author id:", err)
			w.WriteHeader(500)
			return
		}
		chirps, err = cfg.db.GetAllChripsByUserId(req.Context(), getNullUUID(userId))
	} else {
		chirps, err = cfg.db.GetAllChirps(req.Context())
	}

	if err != nil {
		fmt.Println("error getting chiprs:", err)
		w.WriteHeader(500)
		return
	}
	//default is asc

	if sortDirection == "desc" {
		fmt.Println("sort direction desc")
		slices.SortFunc(chirps, func(a, b database.Chirp) int {
			return b.CreatedAt.Time.Compare(a.CreatedAt.Time)
		})
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

	bearer, err := auth.GetBearerToken(req.Header)
	if err != nil {
		fmt.Println("error getting BearerToken from header:", err)
		w.WriteHeader(401)
		return
	}

	userID, err := auth.ValidateJWT(bearer, cfg.secret)
	if err != nil {
		fmt.Println("invalid token:", err)
		w.WriteHeader(401)
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
		return
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, req *http.Request) {
	// make sure access token is valid
	bearer, err := auth.GetBearerToken(req.Header)
	if err != nil {
		fmt.Println("error getting BearerToken from header:", err)
		w.WriteHeader(401)
		return
	}

	userID, err := auth.ValidateJWT(bearer, cfg.secret)
	if err != nil {
		fmt.Println("invalid token:", err)
		w.WriteHeader(401)
		return
	}

	// need to get chrip by id
	chirpId := req.PathValue("chirpID")
	if chirpId == "" {
		fmt.Println("error getting chirp: no id provided")
		w.WriteHeader(400)
		return
	}

	chirpUUID, err := uuid.Parse(chirpId)
	if err != nil {
		fmt.Println("error getting chirp:", err)
		w.WriteHeader(400)
		return
	}

	c, err := cfg.db.GetChirpById(req.Context(), chirpUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(404)
			return
		}
		fmt.Println("chirpUUID", chirpUUID)
		fmt.Println("error querying chirp:", err)
		w.WriteHeader(400)
		return
	}

	chirpOwnerID := c.UserID.UUID

	// make sure the user owns the chirp
	if chirpOwnerID != userID {
		fmt.Println("user is not owner of chrip they are trying to delete")
		fmt.Println("user id: ", userID, "chrip owner:", chirpOwnerID)
		w.WriteHeader(403)
		return
	}

	err = cfg.db.DeleteChirpById(req.Context(), c.ID)
	if err != nil {
		fmt.Println("error deleting chirp:", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(204)
	return

}
