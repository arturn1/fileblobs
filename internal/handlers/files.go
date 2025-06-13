package handlers

import (
	"fileblobs/pkg/azure"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

type PageData struct {
	Folders          []string
	Files            []string
	Prefix           string
	Query            string
	DownloadMode     bool
	IsDefaultAccount bool
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
	"fileIcon": func(filename string) string {
		ext := strings.ToLower(filepath.Ext(filename))
		switch ext {
		case ".jpg", ".jpeg":
			return "/static/icons/jpg.png"
		case ".png":
			return "/static/icons/png.png"
		case ".pdf":
			return "/static/icons/pdf.png"
		case ".doc", ".docx":
			return "/static/icons/docx.png"
		case ".xls", ".xlsx":
			return "/static/icons/xls.png"
		case ".zip", ".rar":
			return "/static/icons/zip.png"
		case ".txt":
			return "/static/icons/txt.png"
		default:
			return "/static/icons/file.png"
		}
	},
}).ParseFiles("web/templates/index.html"))

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	// Verificar o cookie selected_account logo no início
	selectedAccountCookie, cookieErr := r.Cookie("selected_account")
	selectedAccountName := ""

	if cookieErr == nil {
		selectedAccountName = selectedAccountCookie.Value
		log.Printf("Cookie 'selected_account' encontrado com valor: '%s'", selectedAccountName)
	} else {
		log.Printf("Cookie 'selected_account' não encontrado: %v", cookieErr)
	}

	prefix := r.URL.Query().Get("prefix")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	query := r.URL.Query().Get("q")
	downloadMode := r.URL.Query().Get("downloadMode") == "1"

	folders, files, err := azure.ListFoldersAndFiles(prefix)
	if err != nil {
		log.Printf("Erro ao listar blobs: %v", err)

		// Set an error message in a cookie
		errorCookie := http.Cookie{
			Name:     "blob_list_error",
			Value:    "Erro ao listar blobs. A conta pode estar inválida ou inacessível.",
			Path:     "/",
			MaxAge:   60,
			HttpOnly: false,
		}
		http.SetCookie(w, &errorCookie)

		// Redirect to storage accounts page
		http.Redirect(w, r, "/storage-accounts", http.StatusSeeOther)
		return
	}

	if query != "" {
		folders = filterByQuery(folders, query)
		files = filterByQuery(files, query)
	}
	// Verificar se é a conta padrão - verificando várias formas do nome para ser mais robusto
	isDefaultAccount := selectedAccountName == "" ||
		strings.Contains(strings.ToLower(selectedAccountName), "conta padr") ||
		selectedAccountName == "Conta Padrão"

	// Verificamos se estamos na raiz da conta padrão
	isRootOfDefaultAccount := isDefaultAccount && prefix == ""

	// Determinar se devemos mostrar os botões de ação
	showActionButtons := !isRootOfDefaultAccount // Inverso da condição para esconder

	// Log para debug detalhado com mais informações
	log.Printf("Verificação de botões: Account='%s', IsDefault=%v, Prefix='%s', IsRootOfDefault=%v, ShowButtons=%v",
		selectedAccountName, isDefaultAccount, prefix, isRootOfDefaultAccount, showActionButtons)

	// Garantir que o cookie esteja definido como vazio para a conta padrão
	// Isso ajuda a evitar inconsistências no estado
	if isDefaultAccount && selectedAccountName == "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "selected_account",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})
	}

	// Log para debug
	log.Printf("Selected Account: '%s', Show Buttons: %v, Prefix: '%s', Is Default: %v, Root of Default: %v",
		selectedAccountName, showActionButtons, prefix, isDefaultAccount, isRootOfDefaultAccount)
	data := PageData{
		Folders:          folders,
		Files:            files,
		Prefix:           prefix,
		Query:            query,
		DownloadMode:     downloadMode,
		IsDefaultAccount: isRootOfDefaultAccount, // True quando estamos na raiz da conta padrão (onde queremos esconder botões)
	}

	// Log extra para debug na renderização do template
	log.Printf("Renderizando template com IsDefaultAccount=%v (botões %s)",
		isRootOfDefaultAccount,
		map[bool]string{true: "escondidos", false: "visíveis"}[isRootOfDefaultAccount])

	tmpl.Execute(w, data)
}
