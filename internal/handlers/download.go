package handlers

import (
	"encoding/json"
	"fileblobs/pkg/azure"
	"log"
	"net/http"
	"strings"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	blobPath := r.URL.Query().Get("path")
	if blobPath == "" {
		respondWithError(w, r, "Caminho do arquivo não especificado", http.StatusBadRequest)
		return
	}

	data, err := azure.DownloadBlob(blobPath)
	if err != nil {
		log.Printf("Erro ao baixar arquivo %s: %v", blobPath, err)
		respondWithError(w, r, "Erro ao baixar arquivo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+baseName(blobPath))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// respondWithError returns an appropriate error response based on the Accept header
func respondWithError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	acceptHeader := r.Header.Get("Accept")

	// If the client accepts JSON, return a JSON error
	if strings.Contains(acceptHeader, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]string{"error": message})
		return
	}

	// Otherwise return a plain text error
	http.Error(w, message, statusCode)
}
