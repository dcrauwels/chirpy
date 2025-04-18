package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dcrauwels/chirpy/internal/auth"
	"github.com/dcrauwels/chirpy/internal/database"
	"github.com/dcrauwels/chirpy/strutils"
	"github.com/google/uuid"
)

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
		writeError(w, 401, err, "user not authorized")
		return
	}

	// validate token
	tokenUserID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		writeError(w, 401, err, "user not authorized")
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
	responseChirp := database.Chirp{
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
	responseChirps := []database.Chirp{}
	for _, chirp := range chirps {
		responseChirps = append(responseChirps, database.Chirp{
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
	responseChirp := database.Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	writeJSON(w, 200, responseChirp)
}

func (cfg *apiConfig) deleteChirpsHandler(w http.ResponseWriter, r *http.Request) {
	// read request
	req := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(req)
	if err != nil {
		writeError(w, 400, err, "endpoint is not a valid uuid")
		return
	}

	// tokenomics
	//get token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeError(w, 401, err, "no authorization header found in request")
		return
	}
	//validate token
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		writeError(w, 401, err, "no valid access token found")
		return
	}

	// querynomics
	//check if chirp exists
	chirp, err := cfg.db.GetSingleChirp(r.Context(), chirpID)
	if err == sql.ErrNoRows {
		writeError(w, 404, err, "chirp not found")
		return
	} else if err != nil {
		writeError(w, 500, err, "error querying database")
		return
	} else if chirp.UserID != userID {
		writeError(w, 403, errors.New("wrong user ID"), "user not authorized to delete chirp")
	}
	//delete query
	delParams := database.DeleteSingleChirpParams{
		ID:     chirpID,
		UserID: userID,
	}
	_, err = cfg.db.DeleteSingleChirp(r.Context(), delParams)
	if err != nil {
		writeError(w, 500, err, "error querying database")
	}

	// return 204
	writeJSON(w, 204, nil)
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
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}
	writeJSON(w, 201, responseParams)
}

func (cfg *apiConfig) putUsersHandler(w http.ResponseWriter, r *http.Request) {
	// read request header
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeError(w, 401, err, "no authorization field in header")
		return
	}
	userID, err := auth.ValidateJWT(accessToken, cfg.secret)

	if err != nil {
		writeError(w, 401, err, "access token invalid")
		return
	}

	// read request body
	reqParams := struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&reqParams)
	if err != nil {
		writeError(w, 400, err, "incorrect JSON structure in request")
		return
	}

	// hash password
	hashedPassword, err := auth.HashPassword(reqParams.Password)
	if err != nil {
		writeError(w, 500, err, "error hashing password")
		return
	}

	// run uupdateemailpassword query
	updateParams := database.UpdateEmailPasswordParams{
		ID:             userID,
		Email:          reqParams.Email,
		HashedPassword: hashedPassword,
	}
	updatedUser, err := cfg.db.UpdateEmailPassword(r.Context(), updateParams)
	if err != nil {
		writeError(w, 500, err, "error updating email and password")
		return
	}

	// return user values with 200 code
	updatedUserWithoutPassword := struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}{
		ID:          updatedUser.ID,
		CreatedAt:   updatedUser.CreatedAt,
		UpdatedAt:   updatedUser.UpdatedAt,
		Email:       updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
	}
	writeJSON(w, 200, updatedUserWithoutPassword)

}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	// read request
	reqParams := struct {
		Password string `json:"password"`
		Email    string `json:"email"`
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

	// time to make access token
	token, err := auth.MakeJWT(user.ID, cfg.secret)
	if err != nil {
		writeError(w, 500, err, "error creating JWT")
		return
	}

	// time to make refresh token
	// 1. hi im daisy refresh token
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		writeError(w, 500, err, "error creating refresh token")
		return
	}
	// 2. you put the refresh token in the database bag
	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60), // 60 days
	}
	_, err = cfg.db.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		writeError(w, 500, err, "error adding refresh token to database")
		return
	}

	// write response
	respParams := struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		IsChirpyRed:  user.IsChirpyRed,
		Token:        token,
		RefreshToken: refreshToken,
	}

	writeJSON(w, 200, respParams)
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	// read request
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeError(w, 401, err, "no authorization field in header")
		return
	}

	// get refresh token from db
	refreshToken, err := cfg.db.GetRefreshTokenByToken(r.Context(), token)
	if err != nil {
		writeError(w, 401, err, "refresh token not found in DB")
		return
	}
	// check if expired
	if refreshToken.ExpiresAt.Before(time.Now()) {
		writeError(w, 401, err, "refresh token expired")
		return
	}
	// check if revoked
	if refreshToken.RevokedAt.Valid {
		writeError(w, 401, err, "refresh token revoked")
		return
	}

	// return access token
	accessToken, err := auth.MakeJWT(refreshToken.UserID, cfg.secret)
	if err != nil {
		writeError(w, 500, err, "error creating access token")
	}

	respParams := struct {
		Token string `json:"token"`
	}{
		Token: accessToken,
	}

	writeJSON(w, 200, respParams)
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	// get token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		writeError(w, 401, err, "no authorization field in header")
		return
	}

	// run revoke query
	err = cfg.db.RevokeRefreshTokenByToken(r.Context(), token)
	if err != nil {
		writeError(w, 401, err, "refresh token not found in database")
	}

	writeJSON(w, 204, nil)
}
