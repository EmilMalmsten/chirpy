package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/emilmalmsten/chirpy/internal/auth"
	"github.com/emilmalmsten/chirpy/internal/jsonDB"
)

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
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

	user, err := cfg.DB.CreateUser(params.Email, storedHash)
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

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't decode parameters")
		return
	}

	user, err := cfg.DB.GetUserByEmail(params.Email)
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

	token, err := auth.CreateToken(user.Id, []byte(cfg.jwtSecret), params.ExpiresInSeconds)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "jwt token error")
		return
	}

	type returnUser struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, returnUser{
		Id:    user.Id,
		Email: user.Email,
		Token: token,
	})
}

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		if errors.Is(err, jsonDB.ErrAlreadyExists) {
			respondWithError(w, http.StatusInternalServerError, "auth header missing")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "malformed auth header")
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid jwt token")
		return
	}

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't decode parameters")
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't hash password")
		return
	}

	userIDInt, err := strconv.Atoi(userId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't parse user ID")
		return
	}

	user, err := cfg.DB.UpdateUser(userIDInt, params.Email, hashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to update user info")
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
