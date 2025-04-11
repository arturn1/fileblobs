package handlers

import (
	"fileblobs/pkg/azure"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		http.Error(w, "Erro ao ler arquivos", http.StatusBadRequest)
		return
	}

	prefix := r.FormValue("prefix")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	files := r.MultipartForm.File["files"]
	fileMap := make(map[string][]byte)

	for _, f := range files {
		src, err := f.Open()
		if err != nil {
			continue
		}
		defer src.Close()

		data, err := io.ReadAll(src)
		if err != nil {
			continue
		}

		filename := filepath.Base(f.Filename)
		fileMap[filename] = data
	}

	if len(fileMap) > 0 {
		err = azure.UploadMultipleBlobs(prefix, fileMap)
		if err != nil {
			http.Error(w, "Erro ao fazer upload dos arquivos", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/?prefix="+prefix, http.StatusSeeOther)
}
