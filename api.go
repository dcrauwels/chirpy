package main

import "net/http"

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	h := make(http.Header)
	h.Add("Content-Type", "text/plain")
	h.Add("charset", "utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
