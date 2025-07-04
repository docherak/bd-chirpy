package main

import (
	"encoding/json"
	"net/http"

	"github.com/docherak/bd-chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaEvents(w http.ResponseWriter, r *http.Request) {
	type data struct {
		UserID string `json:"user_id"`
	}
	type parameters struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't extract apiKey", err)
		return
	}

	if apiKey != cfg.polkApiSecret {
		respondWithError(w, http.StatusUnauthorized, "API key is invalid", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse UUID", err)
		return
	}

	user, err := cfg.db.GrantPremium(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "User not found", err)
		return
	}

	if !user.IsChirpyRed {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to grant premium to user"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
