package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/emilmalmsten/chirpy/internal/auth"
	"github.com/emilmalmsten/chirpy/internal/jsonDB"
	"github.com/go-chi/chi"
)

func (cfg apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

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

	userIDInt, err := strconv.Atoi(userId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't parse user ID")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't decode parameters")
		return
	}

	const maxChirpLength = 140

	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "chirp is too long")
		return
	}

	cleanChirp := filterProfanity(params.Body)

	chirp, err := cfg.DB.CreateChirp(cleanChirp, userIDInt)
	if err != nil {
		fmt.Printf("err with create chirp: %s", err)
		respondWithError(w, http.StatusInternalServerError, "failed to create Chirp")
		return
	}

	respondWithJSON(w, http.StatusCreated, chirp)
}

func (cfg apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.DB.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to fetch chirps")
		return
	}

	authorId := -1
	authorIdString := r.URL.Query().Get("author_id")
	if authorIdString != "" {
		authorId, err = strconv.Atoi(authorIdString)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid author id")
		}
	}

	sortBy := "asc"
	sortByString := r.URL.Query().Get("sort")
	if sortByString == "desc" {
		sortBy = "desc"
	}

	chirps := []jsonDB.Chirp{}
	for _, chirp := range dbChirps {
		if authorId != -1 && chirp.AuthorId != authorId {
			continue
		}
		chirps = append(chirps, chirp)
	}

	sort.Slice(chirps, func(i, j int) bool {
		if sortBy == "desc" {
			return chirps[i].Id > chirps[j].Id
		}
		return chirps[i].Id < chirps[j].Id
	})

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg apiConfig) handlerGetChirpById(w http.ResponseWriter, r *http.Request) {

	chirpIdString := chi.URLParam(r, "chirpID")
	chirpID, err := strconv.Atoi(chirpIdString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chirp ID")
		return
	}

	chirp, err := cfg.DB.GetChirp(chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "chirp not found")
		return
	}

	respondWithJSON(w, http.StatusOK, chirp)

}

func (cfg apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
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

	userIDInt, err := strconv.Atoi(userId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't parse user ID")
		return
	}

	chirpId := chi.URLParam(r, "chirpID")
	chirpIDInt, err := strconv.Atoi(chirpId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid chirp ID")
		return
	}

	err = cfg.DB.DeleteChirp(chirpIDInt, userIDInt)
	if err != nil {
		if errors.Is(err, jsonDB.ErrDoesNotExists) {
			respondWithError(w, http.StatusNotFound, "chirp not found")
			return
		} else if errors.Is(err, jsonDB.ErrNotAuthorized) {
			respondWithError(w, http.StatusForbidden, "unauthorized to delete chirp")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, "failed to delete chirp")
			return
		}
	}

	type response struct {
		Body string `json:"body"`
	}

	respondWithJSON(w, http.StatusOK, response{})
}
