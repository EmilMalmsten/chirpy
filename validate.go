package main

import (
	"encoding/json"
	"net/http"
)

func validationHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type errorReturnVals struct {
		Error string `json:"error"`
	}

	type returnVals struct {
		Valid bool `json:"valid"`
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		http.Error(w, "Couldn't decode parameters", http.StatusBadRequest)
		return
	}

	if len(params.Body) > 140 {
		errorObj := errorReturnVals{
			Error: "Chirp is too long",
		}
		errorJSON, _ := json.Marshal(errorObj)
		w.WriteHeader(http.StatusBadRequest)

		w.Write(errorJSON)
		return
	}

	w.WriteHeader(http.StatusOK)
	responseObj := returnVals{
		Valid: true,
	}
	responseJSON, _ := json.Marshal(responseObj)
	w.Write(responseJSON)
}
