package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emilmalmsten/chirpy/internal/jsonDB"
)

func postUser(db *jsonDB.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email string `json:"email"`
		}

		w.Header().Set("Content-Type", "application/json")

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "couldn't decode parameters")
			return
		}

		user, err := db.CreateUser(params.Email)
		if err != nil {
			fmt.Printf("err with creating user: %s", err)
			respondWithError(w, http.StatusInternalServerError, "failed to create user")
			return
		}

		respondWithJSON(w, http.StatusCreated, user)
	}
}
