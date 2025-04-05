package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dcrauwels/chirpy/strutils"
)

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	h := make(http.Header)
	h.Add("Content-Type", "text/plain")
	h.Add("charset", "utf-8")
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
	type respErrParameters struct {
		Error string `json:"error"`
	}
	type respValidParameters struct {
		Valid bool `json:"valid"`
	}

	// receive request

	decoder := json.NewDecoder(r.Body)
	params := reqParameters{}
	errParams := respErrParameters{}
	err := decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(400)
		errParams.Error = fmt.Sprintf("JSON request invalid: %s", err)
		dat, err := json.Marshal(errParams)
		// double error case
		if err != nil {
			log.Printf("Error marshalling data: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Write([]byte(dat))
		return
	}

	// other possible error checks

	// chirp length
	err = strutils.ChirpLength(params.Body, maxChirpLength)
	if err != nil {
		errParams.Error = fmt.Sprint(err)
		w.WriteHeader(400)
		dat, err := json.Marshal(errParams)
		if err != nil {
			log.Printf("Error marshalling data: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Write([]byte(dat))
		return
	}

	// send response

	validParams := respValidParameters{Valid: true}
	dat, err := json.Marshal(validParams)
	if err != nil {
		log.Printf("Error marshalling data: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)

}
