package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/simoncrowe/reticle-runner/internal/jobs"
	"github.com/simoncrowe/reticle-runner/internal/profiles"
)

type profileResp struct {
	Id string `json:"id"`
}

func HandleProfiles(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	profile, err := profiles.DecodeProfile(r.Body)
	if err != nil {
		msg := strings.Join([]string{"Deserialization error", err.Error()}, ": ")
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	errMsgs := profiles.ValidateProfile(profile)
	if len(errMsgs) > 0 {
		http.Error(w, strings.Join(errMsgs, ", "), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	jobId, err := jobs.CreateJob(ctx, profile)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Error creating Kubernetes Job", http.StatusInternalServerError)
		return
	}
	respData := profileResp{jobId}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(respData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
