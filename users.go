package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/emilmalmsten/chirpy/internal/auth"
	"github.com/emilmalmsten/chirpy/internal/jsonDB"
)

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

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
		Id            int    `json:"id"`
		Email         string `json:"email"`
		Is_chirpy_red bool   `json:"is_chirpy_red"`
	}

	respondWithJSON(w, http.StatusCreated, returnUser{
		Id:            user.Id,
		Email:         user.Email,
		Is_chirpy_red: false,
	})

}

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	type response struct {
		Id            int    `json:"id"`
		Email         string `json:"email"`
		Is_chirpy_red bool   `json:"is_chirpy_red"`
		Token         string `json:"token"`
		RefreshToken  string `json:"refresh_token"`
	}

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

	accessToken, err := auth.CreateJWT(user.Id, []byte(cfg.jwtSecret), time.Hour, auth.TokenTypeAccess)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "jwt accessToken error")
		return
	}

	refreshToken, err := auth.CreateJWT(user.Id, []byte(cfg.jwtSecret), time.Hour*24*60, auth.TokenTypeAccess)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "jwt refreshToken error")
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Id:            user.Id,
		Email:         user.Email,
		Is_chirpy_red: user.Is_chirpy_red,
		Token:         accessToken,
		RefreshToken:  refreshToken,
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
		Id            int    `json:"id"`
		Email         string `json:"email"`
		Is_chirpy_red bool   `json:"is_chirpy_red"`
	}

	respondWithJSON(w, http.StatusOK, returnUser{
		Id:            user.Id,
		Email:         user.Email,
		Is_chirpy_red: user.Is_chirpy_red,
	})

}

func (cfg *apiConfig) handlerUpgradeMembership(w http.ResponseWriter, r *http.Request) {

	requestKey, err := auth.GetApiKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "could not verify api key")
	}

	if requestKey != cfg.polkaApiKey {
		respondWithError(w, http.StatusUnauthorized, "wrong api key")
	}

	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't decode parameters")
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusOK, struct{}{})
		return
	}

	user, err := cfg.DB.UpgradeUser(params.Data.UserID)
	if err != nil {
		if errors.Is(err, jsonDB.ErrDoesNotExists) {
			respondWithError(w, http.StatusNotFound, "user not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "failed to upgrade user")
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}
