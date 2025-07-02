package main

import (
	"encoding/json"
	"net/http"

	"errors"
	"github.com/docherak/bd-chirpy/internal/database"
	"github.com/google/uuid"
	"strings"
	"time"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsGetSingle(w http.ResponseWriter, r *http.Request) {

	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse UUID", err)
		return
	}
	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	apiChirp := databaseChirpToAPIChirp(dbChirp)
	respondWithJSON(w, http.StatusOK, apiChirp)
}

func (cfg *apiConfig) handlerChirpsGetAll(w http.ResponseWriter, r *http.Request) {
	type returnVals struct {
		Users []User
	}

	chirpIDString := r.PathValue("chirpID")
	if len(chirpIDString) > 0 {
		chirpID, err := uuid.Parse(chirpIDString)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Couldn't parse UUID", err)
			return
		}
		dbChirp, err := cfg.db.GetChirp(r.Context(), chirpID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "Chirp not found", err)
			return
		}

		apiChirp := databaseChirpToAPIChirp(dbChirp)
		respondWithJSON(w, http.StatusOK, apiChirp)
		return
	}

	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting chirps", err)
	}

	apiChirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		apiChirps = append(apiChirps, databaseChirpToAPIChirp(dbChirp))
	}

	respondWithJSON(w, http.StatusOK, apiChirps)
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	cleanedBody, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	userID, err := uuid.Parse(params.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse UUID", err)
		return
	}

	chirpParams := database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: userID,
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp", err)
	}

	apiChirp := databaseChirpToAPIChirp(chirp)

	respondWithJSON(w, http.StatusCreated, apiChirp)
}

func databaseChirpToAPIChirp(dbChirp database.Chirp) Chirp {
	return Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
}

func validateChirp(body string) (string, error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := getCleanedBody(body, badWords)
	return cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}
