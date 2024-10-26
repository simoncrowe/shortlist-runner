package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	schemav1 "github.com/simoncrowe/shortlist-schema/lib/v1"
)

type profileResp struct {
	Id string `json:"id"`
}

type JobsRepo interface {
	Create(ctx context.Context, profile schemav1.Profile) (string, error)
}

type ProfilesHandler struct {
	Jobs JobsRepo
}

func (h ProfilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	profile, err := schemav1.DecodeProfileJSON(r.Body)
	if err != nil {
		msg := strings.Join([]string{"Deserialization error", err.Error()}, ": ")
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	errMsgs := schemav1.ValidateProfile(profile)
	if len(errMsgs) > 0 {
		http.Error(w, strings.Join(errMsgs, ", "), http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	jobId, err := h.Jobs.Create(ctx, profile)
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
