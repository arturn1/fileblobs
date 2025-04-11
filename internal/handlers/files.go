package handlers

import (
	"fileblobs/pkg/azure"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type PageData struct {
	Folders      []string
	Files        []string
	Prefix       string
	Query        string
	DownloadMode bool
}

var tmpl = template.Must(template.New("index.html").Funcs(template.FuncMap{
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
}).ParseFiles("web/templates/index.html"))

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	query := r.URL.Query().Get("q")
	downloadMode := r.URL.Query().Get("downloadMode") == "1"

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
		Folders:      folders,
		Files:        files,
		Prefix:       prefix,
		Query:        query,
		DownloadMode: downloadMode,
	}

	tmpl.Execute(w, data)
}
