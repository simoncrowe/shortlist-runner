package main

import (
	"log"
	"net/http"

	"github.com/simoncrowe/reticle-runner/internal/handlers"
)

func main() {
	http.HandleFunc("/api/v1/profiles", handlers.HandleProfiles)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
