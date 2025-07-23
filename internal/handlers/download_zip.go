package handlers

import (
	"encoding/json"
	"errors"
	"fileblobs/pkg/azure"
	"net/http"
	"os"
)

type DownloadZipRequest struct {
	ConnectionString string `json:"connectionString"`
	ContainerName    string `json:"containerName"`
	FolderPath       string `json:"folderPath"`
}

func DownloadZipHandler(w http.ResponseWriter, r *http.Request) {
	var req DownloadZipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	zipBytes, err := azure.DownloadFolderAsZip(req.ConnectionString, req.ContainerName, req.FolderPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "Nenhum arquivo encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to download zip: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=download.zip")
	w.Write(zipBytes)
}
