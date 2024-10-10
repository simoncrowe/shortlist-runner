package handlers

import (
	"encoding/json"
	"net/http"
)

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	respData := map[string]string{
		"status": "ok",
	}
	if err := json.NewEncoder(w).Encode(respData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
