package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// Serve une image stockée sur le serveur
func ServeImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Récupérer le nom de fichier depuis l'URL
	fileName := r.URL.Path[len("/images/"):]
	filePath := "./db/images/" + fileName

	// Ouvrir le fichier
	file, err := os.Open(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	// Détecter le type MIME et l'envoyer dans l'en-tête
	w.Header().Set("Content-Type", "image/jpeg") // Par défaut, ajuster selon besoin
	io.Copy(w, file)
}

// Reçoit et stocke une image envoyée par un client
func ReceiveImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		fmt.Println("Method not allowed")
		return
	}

	// Limite de 10 MB pour l'upload
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		fmt.Println("Unable to parse form")
		return
	}

	// Stocker l'image
	err = stockeImage(r)
	if err != nil {
		response := UploadResponse{Message: "Failed to upload image", Error: err.Error()}
		fmt.Println("stockeImage:")
		fmt.Println(response)
		fmt.Println("end.")
		jsonResponse(w, response, http.StatusInternalServerError)
		return
	}

	// Réponse réussie
	response := UploadResponse{Message: "Image uploaded successfully"}
	fmt.Println(response)
	jsonResponse(w, response, http.StatusOK)
}

// Fonction pour stocker une image sur le serveur
func stockeImage(r *http.Request) error {
	// Récupérer le fichier
	file, handler, err := r.FormFile("image")
	if err != nil {
		return fmt.Errorf("failed to retrieve file: %v", err)
	}
	defer file.Close()

	filePath := "./db/images/" + handler.Filename

	// Vérifier si le fichier existe déjà
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("file already exists")
	}

	// Créer et écrire dans le fichier
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return fmt.Errorf("failed to copy file data: %v", err)
	}

	return nil
}
