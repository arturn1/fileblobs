package handlers

import (
	"archive/zip"
	"bytes"
	"fileblobs/pkg/azure"
	"io"
	"log"
	"net/http"
	"strings"
)

func DownloadFolderHandler(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("path")
	if prefix == "" {
		respondWithError(w, r, "Caminho não informado", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	files, err := azure.ListBlobsFromFolder(prefix)
	if err != nil {
		log.Printf("Erro ao listar arquivos da pasta %s: %v", prefix, err)
		respondWithError(w, r, "Erro ao listar arquivos", http.StatusInternalServerError)
		return
	}

	if len(files) == 0 {
		respondWithError(w, r, "Pasta vazia ou não encontrada", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=pasta.zip")

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, path := range files {
		data, err := azure.DownloadBlob(path)
		if err != nil {
			log.Printf("Erro ao baixar arquivo %s: %v", path, err)
			continue
		}

		relative := strings.TrimPrefix(path, prefix)
		if relative == "" {
			continue
		}

		fw, err := zipWriter.Create(relative)
		if err != nil {
			log.Printf("Erro ao criar entrada no ZIP para %s: %v", relative, err)
			continue
		}

		io.Copy(fw, bytes.NewReader(data))
	}
}
