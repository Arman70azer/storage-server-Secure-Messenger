package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

// Serve une vidéo stockée sur le serveur avec prise en charge du streaming
func ServeVideo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Range")

	// Récupérer le nom du fichier vidéo
	fileName := r.URL.Path[len("/videos/"):]
	filePath := "./db/videos/" + fileName

	// Ouvrir le fichier
	file, err := os.Open(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	// Obtenir les informations du fichier
	fileStat, err := file.Stat()
	if err != nil {
		http.Error(w, "Error getting file info", http.StatusInternalServerError)
		return
	}
	fileSize := fileStat.Size()

	// Gérer la requête HTTP Range pour permettre le streaming
	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		// Si pas de requête Range, on envoie toute la vidéo
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
		io.Copy(w, file)
		return
	}

	// Gestion du streaming avec l'en-tête Range
	var start, end int64
	fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
	if end == 0 || end >= fileSize {
		end = fileSize - 1
	}

	// Définir la réponse partielle
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	w.WriteHeader(http.StatusPartialContent)

	// Lire la partie demandée et l'envoyer
	file.Seek(start, io.SeekStart)
	io.CopyN(w, file, end-start+1)
}

// Reçoit et stocke une vidéo envoyée par un client
func ReceiveVideo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limite de 100 MB pour l'upload
	err := r.ParseMultipartForm(100 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Stocker la vidéo
	fileName, err := stockeVideo(r)
	if err != nil {
		response := UploadResponse{Message: "Failed to upload video", Error: err.Error()}
		jsonResponse(w, response, http.StatusInternalServerError)
		return
	}

	// Réponse réussie
	response := UploadResponse{Message: "Video uploaded successfully", FileName: fileName}
	jsonResponse(w, response, http.StatusOK)
}

// Fonction pour stocker une vidéo sur le serveur
func stockeVideo(r *http.Request) (string, error) {
	// Récupérer le fichier
	file, handler, err := r.FormFile("video")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve file: %v", err)
	}
	defer file.Close()

	filePath := "./db/videos/" + handler.Filename

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
