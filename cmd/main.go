package main

import (
	"fileblobs/config"
	"fileblobs/internal/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	config.LoadEnv()

	// Configuração para CORS, permitindo requisições da página de login
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Permite requisições do mesmo origem
			w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Responde imediatamente às requisições OPTIONS (preflight)
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	} // Configura o roteador com middleware CORS
	mux := http.NewServeMux()
	// Authentication routes
	mux.HandleFunc("/login", handlers.LoginHandler)
	mux.HandleFunc("/logout", handlers.LogoutHandler)
	mux.HandleFunc("/access-denied", handlers.AccessDeniedPageHandler)
	mux.HandleFunc("/auth/store-token", handlers.StoreTokenHandler)

	// Páginas protegidas por autenticação
	mux.HandleFunc("/storage-accounts", handlers.AuthMiddleware(handlers.StorageAccountsHandler))
	mux.HandleFunc("/add-account", handlers.AuthMiddleware(handlers.AddAccountHandler))
	mux.HandleFunc("/edit-account", handlers.AuthMiddleware(handlers.EditAccountHandler))
	mux.HandleFunc("/select-account", handlers.AuthMiddleware(handlers.SelectAccountHandler))

	// File handling routes - protected by auth middleware
	mux.HandleFunc("/", handlers.AuthMiddleware(handlers.ListFilesHandler))
	mux.HandleFunc("/download", handlers.AuthMiddleware(handlers.DownloadHandler))
	mux.HandleFunc("/download-folder", handlers.AuthMiddleware(handlers.DownloadFolderHandler))
	mux.HandleFunc("/download-multiple", handlers.AuthMiddleware(handlers.DownloadMultipleHandler))
	mux.HandleFunc("/upload", handlers.AuthMiddleware(handlers.UploadHandler))
	mux.HandleFunc("/download-zip", handlers.AuthMiddleware(handlers.DownloadZipHandler))

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", handlers.NewCustomFileServer(http.Dir("web/static"))))

	// Template JS files
	mux.Handle("/js/", http.StripPrefix("/js/", handlers.NewCustomFileServer(http.Dir("web/templates"))))

	// Aplicar middleware CORS ao roteador completo
	corsHandler := corsMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	log.Printf("Servidor rodando em http://localhost:%s", port)
	http.ListenAndServe(":"+port, corsHandler)
}
