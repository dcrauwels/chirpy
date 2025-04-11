package main

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) hitsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	body := fmt.Sprintf(`<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`, cfg.fileserverHits.Load())
	w.Write([]byte(body))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	// check if platform in .env is set to dev
	godotenv.Load()
	platform := os.Getenv("PLATFORM")
	if platform != "dev" {
		writeError(w, 403, nil, "access only allowed from development environment")
		return
	}

	// reset fileserverhits
	cfg.fileserverHits = atomic.Int32{}

	// reset users
	err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		writeError(w, 500, err, "error running resetusers query")
		return
	}

	writeJSON(w, 200, "configuration reset succesfully")

}
