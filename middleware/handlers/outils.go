package handlers

import (
	"encoding/json"
	"net/http"
)

// Structure de réponse JSON pour ReceiveImage
type UploadResponse struct {
	Message  string `json:"message"`
	FileName string `json:"fileName,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Fonction utilitaire pour envoyer des réponses JSON
func jsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
