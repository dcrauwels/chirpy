package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dcrauwels/chirpy/strutils"
)

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	// constants
	const maxChirpLength int = 140
	var invalidWords = [3]string{"kerfuffle", "sharbert", "fornax"} //hardly a constant but used as one

	// define types
	type reqParameters struct {
		Body string `json:"body"`
	}
	type respValidParameters struct {
		CleanedBody string `json:"cleaned_body"`
	}

	// receive request

	decoder := json.NewDecoder(r.Body)
	params := reqParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		writeError(w, 400, err, "request has incorrect JSON structure") //json.go
	}

	// other possible checks

	// chirp length
	err = strutils.ChirpLength(params.Body, maxChirpLength)
	if err != nil {
		writeError(w, 400, err, fmt.Sprintf("chirp cannot exceed %d characters", maxChirpLength)) //json.go
		return
	}

	// clean body
	cleanedBody := strutils.ReplaceWord(params.Body, invalidWords[:], "****")

	// send response

	validParams := respValidParameters{CleanedBody: cleanedBody}
	writeJSON(w, 200, validParams) //json.go
}
