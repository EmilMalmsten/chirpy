package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/emilmalmsten/chirpy/internal/auth"
	"github.com/emilmalmsten/chirpy/internal/jsonDB"
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

		storedHash, err := auth.HashPassword(params.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't hash password")
			return
		}

		user, err := db.CreateUser(params.Email, storedHash)
		if err != nil {
			if errors.Is(err, jsonDB.ErrAlreadyExists) {
				respondWithError(w, http.StatusConflict, "user already exists")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "failed to create user")
			return
		}

		type returnUser struct {
			Id    int    `json:"id"`
			Email string `json:"email"`
		}

		respondWithJSON(w, http.StatusCreated, returnUser{
			Id:    user.Id,
			Email: user.Email,
		})
	}
}

func postLogin(db *jsonDB.DB) func(http.ResponseWriter, *http.Request) {
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

		user, err := db.GetUserByEmail(params.Email)
		if err != nil {
			if errors.Is(err, jsonDB.ErrDoesNotExists) {
				respondWithError(w, http.StatusNotFound, "user does not exist")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "error retrieving user")
			return
		}

		err = auth.CheckPasswordHash(params.Password, user.Password)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "wrong password")
			return
		}

		type returnUser struct {
			Id    int    `json:"id"`
			Email string `json:"email"`
		}

		respondWithJSON(w, http.StatusOK, returnUser{
			Id:    user.Id,
			Email: user.Email,
		})

	}
}
