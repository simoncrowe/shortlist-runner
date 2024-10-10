package main

import (
	"log"
	"net/http"

	"github.com/simoncrowe/shortlist-runner/internal/handlers"
)

func main() {
	http.HandleFunc("/health", handlers.HandleHealth)
	http.HandleFunc("/api/v1/profiles", handlers.HandleProfiles)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
