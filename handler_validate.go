package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

func handlerValidateChirp(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := jsonRes{}
	err := decoder.Decode(&params)
	if err != nil {
		fmt.Println("error decoding params:", err)
		w.WriteHeader(500)
		return
	}

	chirp := params.Body

	if len(chirp) > 140 {
		respondWithError(w, 400, "Chirp is too long")
	}

	profane := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Split(chirp, " ")
	fmt.Println(words)

	for i, w := range words {
		if slices.Contains(profane, strings.ToLower(w)) {
			words[i] = "****"
		}
	}

	cleaned := strings.Join(words, " ")
	respondWithJson(w, 200, jsonRes{CleanedBody: cleaned})
}
