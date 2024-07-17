package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/simoncrowe/reticle-runner/internal/profiles"
)

func HandleProfiles(w http.ResponseWriter, r *http.Request) {
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

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(profile); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
