package main

import (
	"log"
	"net/http"

	"github.com/simoncrowe/shortlist-runner/internal/handlers"
	"github.com/simoncrowe/shortlist-runner/internal/jobs"
)

func main() {
	http.HandleFunc("/health", handlers.HandleHealth)
	http.Handle("/api/v1/profiles", handlers.ProfilesHandler{jobs.K8sRepository{}})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
