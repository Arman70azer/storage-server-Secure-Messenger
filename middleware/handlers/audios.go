package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

// Serve un fichier audio stocké sur le serveur avec support du streaming
func ServeAudio(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Range")

	// Récupérer le nom du fichier audio
	fileName := r.URL.Path[len("/audios/"):]
	filePath := "./db/audios/" + fileName

	// Ouvrir le fichier
	file, err := os.Open(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	// Obtenir la taille du fichier
	fileStat, err := file.Stat()
	if err != nil {
		http.Error(w, "Error getting file info", http.StatusInternalServerError)
		return
	}
	fileSize := fileStat.Size()

	// Gérer les requêtes Range pour un streaming audio fluide
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		w.Header().Set("Content-Type", "audio/mpeg") // Modifier selon le type de fichier
		w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
		io.Copy(w, file)
		return
	}

	// Gestion du streaming avec Range
	var start, end int64
	fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
	if end == 0 || end >= fileSize {
		end = fileSize - 1
	}

	// Définir les en-têtes pour le streaming
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	w.WriteHeader(http.StatusPartialContent)

	// Lire et envoyer la partie demandée
	file.Seek(start, io.SeekStart)
	io.CopyN(w, file, end-start+1)
}

// Reçoit et stocke un fichier audio envoyé par un client
func ReceiveAudio(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limite de 50 MB pour l'upload
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Stocker l'audio
	fileName, err := stockeAudio(r)
	if err != nil {
		response := UploadResponse{Message: "Failed to upload audio", Error: err.Error()}
		jsonResponse(w, response, http.StatusInternalServerError)
		return
	}

	// Réponse réussie
	response := UploadResponse{Message: "Audio uploaded successfully", FileName: fileName}
	jsonResponse(w, response, http.StatusOK)
}

// Fonction pour stocker un fichier audio sur le serveur
func stockeAudio(r *http.Request) (string, error) {
	// Récupérer le fichier
	file, handler, err := r.FormFile("audio")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve file: %v", err)
	}
	defer file.Close()

	filePath := "./db/audios/" + handler.Filename

	// Vérifier si le fichier existe déjà
	if _, err := os.Stat(filePath); err == nil {
		return "", fmt.Errorf("file already exists")
	}

	// Créer et écrire dans le fichier
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return "", fmt.Errorf("failed to copy file data: %v", err)
	}

	return handler.Filename, nil
}
