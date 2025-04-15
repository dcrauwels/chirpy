package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dcrauwels/chirpy/internal/auth"
	"github.com/dcrauwels/chirpy/internal/database"
	"github.com/dcrauwels/chirpy/strutils"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) postChirpsHandler(w http.ResponseWriter, r *http.Request) {
	// define types
	type requestParameters struct {
		Body string `json:"body"`
	}

	// retrieve token from request
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeError(w, 401, err, "1 user not authorized")
		return
	}

	// validate token
	tokenUserID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		writeError(w, 401, err, "2 user not authorized")
		return
	}

	// constants
	const maxChirpLength int = 140
	var invalidWords = [3]string{"kerfuffle", "sharbert", "fornax"} //used as const but cannot use const with arrays

	// receive request
	decoder := json.NewDecoder(r.Body)
	rParams := requestParameters{}
	err = decoder.Decode(&rParams)
	if err != nil {
		writeError(w, 400, err, "request has incorrect JSON structure") //json.go
		return
	}
	chirpParams := database.CreateChirpParams{
		Body:   rParams.Body,
		UserID: tokenUserID,
	}

	// other possible checks
	// 1. chirp length
	err = strutils.ChirpLength(chirpParams.Body, maxChirpLength)
	if err != nil {
		writeError(w, 400, err, fmt.Sprintf("chirp cannot exceed %d characters", maxChirpLength)) //json.go
		return
	}

	// clean body
	chirpParams.Body = strutils.ReplaceWord(chirpParams.Body, invalidWords[:], "****")

	// create chirp
	chirp, err := cfg.db.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		writeError(w, 500, err, "server error creating chirp")
		return
	}

	// write response
	responseChirp := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	writeJSON(w, 201, responseChirp) //json.go
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	// query DB
	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		writeError(w, 500, err, "error querying database when getting chirps")
		return
	}

	// write response
	responseChirps := []Chirp{}
	for _, chirp := range chirps {
		responseChirps = append(responseChirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}
	writeJSON(w, 200, responseChirps) //json.go
}

func (cfg *apiConfig) getSingleChirpHandler(w http.ResponseWriter, r *http.Request) {
	// define types
	// receive request
	req := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(req)

	if err != nil {
		writeError(w, 400, err, "endpoint is not a valid uuid")
		return
	}

	// query DB
	chirp, err := cfg.db.GetSingleChirp(r.Context(), chirpID)
	if err != nil {
		writeError(w, 404, err, "chirp not found")
		return
	}

	// write response
	responseChirp := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	writeJSON(w, 200, responseChirp)

}

func (cfg *apiConfig) postUsersHandler(w http.ResponseWriter, r *http.Request) {
	// receive request
	decoder := json.NewDecoder(r.Body)
	params := struct { // anonymous as I'm only using this once
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	err := decoder.Decode(&params)
	if err != nil {
		writeError(w, 400, err, "request has incorrect JSON structure") // json.go
		return
	}

	// check request for validity
	if err = strutils.ValidateEmail(params.Email); err != nil {
		writeError(w, 400, err, "not a valid email address")
		return
	}

	// hash password
	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		writeError(w, 500, err, "error hashing password")
		return
	}

	// query DB
	queryParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.db.CreateUser(r.Context(), queryParams)
	if err != nil {
		writeError(w, 500, err, "error querying database when creating user")
		return
	}

	// write response
	responseParams := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	writeJSON(w, 201, responseParams)
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	// read request
	reqParams := struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds *int   `json:"expires_in_seconds,omitempty"`
	}{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqParams)
	if err != nil {
		writeError(w, 400, err, "incorrect JSON structure in request")
		return
	}

	// check if email present in db
	user, err := cfg.db.GetUserByEmail(r.Context(), reqParams.Email)
	if err != nil {
		writeError(w, 401, err, "Incorrect email or password")
		return
	}
	// check if password matches
	err = auth.CheckPasswordHash(user.HashedPassword, reqParams.Password)
	if err != nil {
		writeError(w, 401, err, "Incorrect email or password") //  not perfectly DRY but I think the DRY solution would be less legible
		return
	}

	// time to make a token
	// 1. deal with expiresinseconds value
	expiresInSeconds := 3600
	if reqParams.ExpiresInSeconds != nil {
		if *reqParams.ExpiresInSeconds <= 3600 {
			expiresInSeconds = *reqParams.ExpiresInSeconds
		}
	}
	// 2. make token
	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Second*time.Duration(expiresInSeconds))
	if err != nil {
		writeError(w, 500, err, "error creating JWT")
		return
	}

	// write response
	respParams := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	}

	writeJSON(w, 200, respParams)
}
