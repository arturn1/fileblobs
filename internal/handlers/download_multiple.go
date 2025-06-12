package handlers

import (
	"archive/zip"
	"bytes"
	"fileblobs/pkg/azure"
	"io"
	"log"
	"net/http"
)

func DownloadMultipleHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		respondWithError(w, r, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	files := r.Form["files"]
	if len(files) == 0 {
		respondWithError(w, r, "Nenhum arquivo selecionado", http.StatusBadRequest)
		return
	}

	prefix := r.FormValue("prefix")

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=arquivos.zip")

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	successCount := 0

	for _, path := range files {
		data, err := azure.DownloadBlob(path)
		if err != nil {
			log.Printf("Erro ao baixar arquivo %s: %v", path, err)
			continue
		}

		relativePath := path
		if prefix != "" && len(path) > len(prefix) {
			relativePath = path[len(prefix):]
		}

		fw, err := zipWriter.Create(relativePath)
		if err != nil {
			log.Printf("Erro ao criar entrada no ZIP para %s: %v", relativePath, err)
			continue
		}

		io.Copy(fw, bytes.NewReader(data))
		successCount++
	}

	if successCount == 0 {
		// If no files were successfully processed, return an error
		// This will only work if no data has been written to the response yet
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Erro ao baixar todos os arquivos selecionados"}`))
	}
}
