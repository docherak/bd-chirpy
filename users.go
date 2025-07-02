package main

import (
	"encoding/json"
	"github.com/docherak/bd-chirpy/internal/auth"
	"github.com/docherak/bd-chirpy/internal/database"
	"github.com/google/uuid"
	"net/http"
	"net/mail"
	"time"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	err = validateEmail(params.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid email format", err)
		return
	}

	// TODO: handle password eval
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	// By passing your handler's http.Request.Context() to the query, the library will automatically cancel the database query if the HTTP request is canceled or times out.
	// TODO: handle diff errors
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user", err)
	}

	apiUser := databaseUserToAPIUser(user)

	respondWithJSON(w, http.StatusCreated, apiUser)
}

func databaseUserToAPIUser(dbUser database.User) User {
	return User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
}

func validateEmail(emailAddress string) error {
	_, err := mail.ParseAddress(emailAddress)
	return err
}
