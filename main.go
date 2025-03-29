package main

import (
	"net/http"
)

func main() {
	sM := http.NewServeMux()
	s := http.Server{
		Addr:                         ":8080",
		Handler:                      sM,
		DisableGeneralOptionsHandler: false,
	}

	s.ListenAndServe()
}
