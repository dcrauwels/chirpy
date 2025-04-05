package main

import (
	"net/http"
	"sync/atomic"
	"time"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	// apiconfig
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	// servemux
	mux := http.NewServeMux()

	// register handlers
	mux.HandleFunc("GET /api/healthz", readinessHandler)        //api.go
	mux.HandleFunc("POST /api/validate_chirp", validateHandler) //api.go

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

	err := s.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			panic(err)
		}
	}
}
