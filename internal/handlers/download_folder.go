package handlers

import (
	"archive/zip"
	"bytes"
	"fileblobs/pkg/azure"
	"io"
	"net/http"
	"strings"
)

func DownloadFolderHandler(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("path")
	if prefix == "" {
		http.Error(w, "Caminho não informado", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	files, err := azure.ListBlobsFromFolder(prefix)
	if err != nil {
		http.Error(w, "Erro ao listar arquivos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=pasta.zip")

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, path := range files {
		data, err := azure.DownloadBlob(path)
		if err != nil {
			continue
		}

		relative := strings.TrimPrefix(path, prefix)
		if relative == "" {
			continue
		}

		fw, err := zipWriter.Create(relative)
		if err != nil {
			continue
		}

		io.Copy(fw, bytes.NewReader(data))
	}
}
