package utils

import (
	"archive/zip"
	"net/http"
	"path/filepath"

	"fileblobs/pkg/azure"
)

func StreamZip(w http.ResponseWriter, prefix string, files []string) {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, file := range files {
		content, err := azure.DownloadBlob(prefix + file)
		if err != nil {
			continue
		}
		f, _ := zipWriter.Create(file)
		f.Write(content)
	}
}

func StreamMultipleZip(w http.ResponseWriter, data map[string][]string) {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for folder, files := range data {
		for _, file := range files {
			content, err := azure.DownloadBlob(folder + "/" + file)
			if err != nil {
				continue
			}
			// Inclui o nome da pasta no caminho
			zipPath := filepath.Join(filepath.Base(folder), filepath.Base(file))
			f, _ := zipWriter.Create(zipPath)
			f.Write(content)
		}
	}
}
