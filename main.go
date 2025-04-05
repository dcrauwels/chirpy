package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) hitsHandler(w http.ResponseWriter, r *http.Request) {
	body := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())
	w.Write([]byte(body))
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	h := make(http.Header)
	h.Add("Content-Type", "text/plain")
	h.Add("charset", "utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = atomic.Int32{}
	w.Write([]byte("Fileserver hits reset succesfully."))
}

func main() {
	// apiconfig
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	// servemux
	mux := http.NewServeMux()

	// register handlers
	mux.HandleFunc("/healthz", readinessHandler)
	mux.HandleFunc("/metrics", apiCfg.hitsHandler)
	mux.HandleFunc("/reset", apiCfg.resetHandler)

	// fileserver handler
	fS := http.FileServer(http.Dir("."))
	fS = http.StripPrefix("/app/", fS)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fS))

	// server
	s := http.Server{
		Addr:                         ":8080",
		Handler:                      mux,
		DisableGeneralOptionsHandler: false,
	}

	s.ListenAndServe()
}
