package main

import (
	"fileblobs/config"
	"fileblobs/internal/handlers"
	"log"
	"net/http"
)

func main() {
	config.LoadEnv()

	http.HandleFunc("/", handlers.ListFilesHandler)
	http.HandleFunc("/download", handlers.DownloadHandler)

	http.HandleFunc("/download-folder", handlers.DownloadFolderHandler)
	http.HandleFunc("/download-multiple", handlers.DownloadMultipleHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	log.Println("Servidor rodando em http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
