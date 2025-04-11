package main

import (
	"fileblobs/pkg/azure"
	"fileblobs/utils"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/joho/godotenv"
)

type PageData struct {
	Folders []string
	Files   []string
	Prefix  string
	Query   string
}

func main() {
	if err := godotenv.Load(); err != nil {
		utils.LogIfDevelopment("⚠️ Arquivo .env não encontrado, usando valores padrão")
	}

	tmpl := template.Must(template.New("index.html").Funcs(template.FuncMap{
		"splitPrefix": func(s string) []string {
			s = strings.TrimSuffix(s, "/")
			if s == "" {
				return []string{}
			}
			return strings.Split(s, "/")
		},
		"joinPath": func(base, seg string) string {
			return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(seg, "/")
		},
		"baseName": func(path string) string {
			parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
			return parts[len(parts)-1]
		},
		"add": func(i int) int {
			return i + 1
		},
		"len": func(arr []string) int {
			return len(arr)
		},
		"joinPrefix": func(parts []string, index int) string {
			return strings.Join(parts[:index+1], "/")
		},
	}).ParseFiles("templates/index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		prefix := r.URL.Query().Get("prefix")
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}

		query := r.URL.Query().Get("q")

		folders, files, err := azure.ListFoldersAndFiles(prefix)
		if err != nil {
			http.Error(w, "Erro ao listar blobs", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		if query != "" {
			folders = filterByQuery(folders, query)
			files = filterByQuery(files, query)
		}

		data := PageData{
			Folders: folders,
			Files:   files,
			Prefix:  prefix,
			Query:   query,
		}

		tmpl.Execute(w, data)
	})

	// Rota de download
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
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
	})

	log.Println("Servidor rodando em http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func filterByQuery(items []string, query string) []string {
	var filtered []string
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), strings.ToLower(query)) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func baseName(path string) string {
	split := strings.Split(strings.TrimSuffix(path, "/"), "/")
	return split[len(split)-1]
}
