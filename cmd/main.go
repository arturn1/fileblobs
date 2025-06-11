package main

import (
	"fileblobs/config"
	"fileblobs/internal/handlers"
	"log"
	"net/http"
)

func main() {
	config.LoadEnv()
	// Authentication routes
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/storage-accounts", handlers.StorageAccountsHandler)
	http.HandleFunc("/add-account", handlers.AddAccountHandler)
	http.HandleFunc("/edit-account", handlers.EditAccountHandler)
	http.HandleFunc("/select-account", handlers.SelectAccountHandler)

	// File handling routes - protected by auth middleware
	http.HandleFunc("/", handlers.AuthMiddleware(handlers.ListFilesHandler))
	http.HandleFunc("/download", handlers.AuthMiddleware(handlers.DownloadHandler))
	http.HandleFunc("/download-folder", handlers.AuthMiddleware(handlers.DownloadFolderHandler))
	http.HandleFunc("/download-multiple", handlers.AuthMiddleware(handlers.DownloadMultipleHandler))
	http.HandleFunc("/upload", handlers.AuthMiddleware(handlers.UploadHandler))

	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	log.Println("Servidor rodando em http://localhost")
	http.ListenAndServe(":80", nil)
}
