package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emilmalmsten/chirpy/internal/jsonDB"
	"golang.org/x/crypto/bcrypt"
)

func postUser(db *jsonDB.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		w.Header().Set("Content-Type", "application/json")

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "couldn't decode parameters")
			return
		}
		password := []byte(params.Password)

		passwordHash, err := bcrypt.GenerateFromPassword(password, 10)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to create user")
		}

		passwordHashString := base64.StdEncoding.EncodeToString(passwordHash)

		returnUser, err := db.CreateUser(params.Email, passwordHashString)
		if err != nil {
			fmt.Printf("err with creating user: %s", err)
			respondWithError(w, http.StatusInternalServerError, "failed to create user")
			return
		}

		respondWithJSON(w, http.StatusCreated, returnUser)
	}
}
