package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dcrauwels/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) polkaHandler(w http.ResponseWriter, r *http.Request) {
	// read request
	type requestParameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	reqParams := requestParameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqParams)
	if err != nil {
		writeError(w, 400, err, "request has incorrect JSON structure")
		return
	}

	// some checks
	//event != user.upgraded -> 204 get out
	if reqParams.Event != "user.upgraded" {
		writeJSON(w, 204, nil)
		return
	}
	//apikey
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		writeError(w, 401, err, "no API key provided")
		return
	}
	if apiKey != cfg.polkaKey {
		writeError(w, 401, errors.New("incorrect API key"), "API key provided does not match known key")
		return
	}

	// if that check passed then event == user.upgraded -> we upgrade user
	//query SetChirpyRedByID
	_, err = cfg.db.SetChirpyRedByID(r.Context(), reqParams.Data.UserID)
	//err == sql.ErrNoRows > 404
	if err == sql.ErrNoRows {
		writeError(w, 404, err, "user not found")
		return
	} else if err != nil {
		writeError(w, 500, err, "error querying database")
		return
	}

	// return 204 > u get out
	writeJSON(w, 204, nil)

}
