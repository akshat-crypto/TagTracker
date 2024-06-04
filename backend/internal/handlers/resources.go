package handlers

import (
	"encoding/json"
	"net/http"
)

func ResourcesHandler(w http.ResponseWriter, r *http.Request) {
	// Mock data, replace with actual implementation
	resources := []string{"Resource1", "Resource2", "Resource3"}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string][]string{"resources": resources})
}
