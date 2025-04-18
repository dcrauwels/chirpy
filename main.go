package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/dcrauwels/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	secret         string
}

func main() {
	// load .env into env variables
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Println(err)
		return
	}
	dbQueries := database.New(db)

	// apiconfig
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		secret:         os.Getenv("SECRET"),
	}

	// servemux
	mux := http.NewServeMux()

	// register handlers
	mux.HandleFunc("GET /api/healthz", readinessHandler)                       //api.go
	mux.HandleFunc("POST /api/users", apiCfg.postUsersHandler)                 //api.go
	mux.HandleFunc("PUT /api/users", apiCfg.putUsersHandler)                   //api.go
	mux.HandleFunc("POST /api/chirps", apiCfg.postChirpsHandler)               //api.go
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirpsHandler)                 //api.go
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getSingleChirpHandler)  //api.go
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirpsHandler) //api.go
	mux.HandleFunc("POST /api/login", apiCfg.loginHandler)                     //api.go
	mux.HandleFunc("POST /api/refresh", apiCfg.refreshHandler)                 //api.go
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeHandler)                   //api.go
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.polkaHandler)            //api.go

	mux.HandleFunc("GET /admin/metrics", apiCfg.hitsHandler) //admin.go
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler) //admin.go

	// fileserver handler
	fS := http.FileServer(http.Dir("."))
	fS = http.StripPrefix("/app/", fS)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fS)) //middleware in admin.go

	// server
	s := http.Server{
		Addr:                         ":8080",
		Handler:                      mux,
		DisableGeneralOptionsHandler: false,
		ReadTimeout:                  30 * time.Second,
		WriteTimeout:                 60 * time.Second,
		IdleTimeout:                  120 * time.Second,
	}

	err = s.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}
