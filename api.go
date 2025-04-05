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

	// define types
	type reqParameters struct {
		Body string `json:"body"`
	}
	type respValidParameters struct {
		Valid bool `json:"valid"`
	}

	// receive request

	decoder := json.NewDecoder(r.Body)
	params := reqParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		writeError(w, 400, err, "request has incorrect JSON structure")
	}

	// other possible error checks

	// chirp length
	err = strutils.ChirpLength(params.Body, maxChirpLength)
	if err != nil {
		writeError(w, 400, err, fmt.Sprintf("chirp cannot exceed %d characters", maxChirpLength))
	}

	// send response

	validParams := respValidParameters{Valid: true}
	writeJSON(w, 200, validParams)
}
