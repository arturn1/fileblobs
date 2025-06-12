package handlers

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// CustomFileServer é um wrapper para http.FileServer que adiciona cabeçalhos de tipo de conteúdo adequados
type CustomFileServer struct {
	handler http.Handler
}

func (fs *CustomFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Log da requisição para ajudar na depuração
	log.Printf("Servindo arquivo estático: %s", r.URL.Path)

	// Define o tipo de conteúdo com base na extensão do arquivo
	ext := strings.ToLower(filepath.Ext(r.URL.Path))
	switch ext {
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	}

	// Adiciona cabeçalhos de cache para melhorar o desempenho
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache por 1 hora

	// Serve o arquivo
	fs.handler.ServeHTTP(w, r)
}

// NewCustomFileServer cria um novo CustomFileServer
func NewCustomFileServer(root http.FileSystem) http.Handler {
	return &CustomFileServer{
		handler: http.FileServer(root),
	}
}
