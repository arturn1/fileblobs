package handlers

import (
	"archive/zip"
	"bytes"
	"fileblobs/pkg/azure"
	"io"
	"net/http"
)

func DownloadMultipleHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	files := r.Form["files"]
	if len(files) == 0 {
		http.Error(w, "Nenhum arquivo selecionado", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=arquivos.zip")

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, path := range files {
		data, err := azure.DownloadBlob(path)
		if err != nil {
			continue // ignora erros individuais
		}

		fw, err := zipWriter.Create(path)
		if err != nil {
			continue
		}

		io.Copy(fw, bytes.NewReader(data))
	}
}
