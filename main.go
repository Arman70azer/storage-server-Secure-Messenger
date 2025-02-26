package main

import (
	handlers "back-end/middleware/handlers"
	"fmt"
	"net/http"
)

func main() {

	fmt.Println("Server is running at http://localhost:8080")

	http.HandleFunc("/images/", handlers.ServeImage)
	http.HandleFunc("/imagesToSave/", handlers.ReceiveImage)

	http.HandleFunc("/videos/", handlers.ServeVideo)
	http.HandleFunc("/videosToSave/", handlers.ReceiveVideo)

	http.HandleFunc("/audios/", handlers.ServeImage)
	http.HandleFunc("/audiosToSave/", handlers.ReceiveImage)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting the server:", err)
	}
	fmt.Println("Server served")
}
