package handlers

import (
	"fileblobs/pkg/azure"
	"log"
	"net/http"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	blobPath := r.URL.Query().Get("path")
	if blobPath == "" {
		http.Error(w, "Caminho do arquivo não especificado", http.StatusBadRequest)
		return
	}

	data, err := azure.DownloadBlob(blobPath)
	if err != nil {
		http.Error(w, "Erro ao baixar arquivo", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+baseName(blobPath))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
