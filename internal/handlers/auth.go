package handlers

import (
	"fileblobs/internal/repository"
	"fileblobs/pkg/azure"
	"html/template"
	"net/http"
	"os"
)

var loginTmpl = template.Must(template.ParseFiles("web/templates/login.html"))
var storageAccountsTmpl = template.Must(template.ParseFiles("web/templates/storage_accounts.html"))
var addAccountTmpl = template.Must(template.ParseFiles("web/templates/add_account.html"))
var editAccountTmpl = template.Must(template.ParseFiles("web/templates/edit_account.html"))

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is already logged in
	_, authenticated := getSessionUser(r)
	if authenticated {
		http.Redirect(w, r, "/storage-accounts", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		// Process login form
		username := r.FormValue("username")
		password := r.FormValue("password")

		if repository.ValidateUser(username, password) {
			// Set session cookie
			setSessionUser(w, username)
			http.Redirect(w, r, "/storage-accounts", http.StatusSeeOther)
			return
		}

		// Invalid credentials
		loginTmpl.Execute(w, map[string]interface{}{
			"Error": "Nome de usuário ou senha inválidos",
		})
		return
	}

	// Display login form
	loginTmpl.Execute(w, nil)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear session cookie
	clearSession(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func StorageAccountsHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	_, authenticated := getSessionUser(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Display storage accounts
	accounts := repository.GetStorageAccounts()
	storageAccountsTmpl.Execute(w, map[string]interface{}{
		"Accounts": accounts,
	})
}

func AddAccountHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	_, authenticated := getSessionUser(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodPost { // Process the form submission
		name := r.FormValue("name")
		description := r.FormValue("description")
		accountName := r.FormValue("accountName")
		accountKey := r.FormValue("accountKey")
		containerName := r.FormValue("containerName")

		// Create new storage account (nome será gerado no repository se vazio)
		newAccount := repository.StorageAccount{
			Name:          name,
			Description:   description,
			AccountName:   accountName,
			AccountKey:    accountKey,
			ContainerName: containerName,
		}

		// Add to repository (array simples)
		_ = repository.AddStorageAccount(newAccount)

		// Redirect to storage accounts list
		http.Redirect(w, r, "/storage-accounts", http.StatusSeeOther)
		return
	}

	// Display add account form
	addAccountTmpl.Execute(w, nil)
}

func SelectAccountHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	_, authenticated := getSessionUser(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get account name from query parameters
	accountName := r.URL.Query().Get("name")
	if accountName == "" {
		http.Redirect(w, r, "/storage-accounts", http.StatusSeeOther)
		return
	}

	// Find account in repository
	account, found := repository.GetStorageAccountByName(accountName)
	if !found {
		http.Redirect(w, r, "/storage-accounts", http.StatusSeeOther)
		return
	}

	// Set environment variables for Azure storage
	os.Setenv("AZURE_STORAGE_ACCOUNT_NAME", account.AccountName)
	os.Setenv("AZURE_STORAGE_ACCOUNT_KEY", account.AccountKey)
	os.Setenv("AZURE_STORAGE_CONTAINER", account.ContainerName)

	// Clear Azure client cache to use new credentials
	azure.ResetClient()

	// Redirect to home page to browse files
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// EditAccountHandler handles the editing of a storage account
func EditAccountHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	_, authenticated := getSessionUser(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get account name from query parameters
	accountName := r.URL.Query().Get("name")

	// Prevent editing of default account
	if accountName == "Conta Padrão" {
		http.Redirect(w, r, "/storage-accounts", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		// Process the form submission
		originalName := r.FormValue("originalName")
		name := r.FormValue("name")
		description := r.FormValue("description")
		accountName := r.FormValue("accountName")
		accountKey := r.FormValue("accountKey")
		containerName := r.FormValue("containerName")

		// Prevent editing of default account (extra safety check)
		if originalName == "Conta Padrão" {
			http.Redirect(w, r, "/storage-accounts", http.StatusSeeOther)
			return
		}

		// Create updated storage account
		updatedAccount := repository.StorageAccount{
			Name:          name,
			Description:   description,
			AccountName:   accountName,
			AccountKey:    accountKey,
			ContainerName: containerName,
		}

		// Update account in repository
		err := repository.UpdateStorageAccount(originalName, updatedAccount)
		if err != nil {
			// Handle error
			editAccountTmpl.Execute(w, map[string]interface{}{
				"Error":   "Erro ao atualizar conta: " + err.Error(),
				"Account": updatedAccount,
			})
			return
		}

		// Redirect to storage accounts list
		http.Redirect(w, r, "/storage-accounts", http.StatusSeeOther)
		return
	}

	// Get account from repository
	account, found := repository.GetStorageAccountByName(accountName)
	if !found {
		http.Redirect(w, r, "/storage-accounts", http.StatusSeeOther)
		return
	}

	// Display edit account form
	editAccountTmpl.Execute(w, map[string]interface{}{
		"Account": account,
	})
}

// Session management
func setSessionUser(w http.ResponseWriter, username string) {
	cookie := &http.Cookie{
		Name:     "session_user",
		Value:    username,
		Path:     "/",
		MaxAge:   3600, // 1 hour
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

func getSessionUser(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session_user")
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

func clearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "session_user",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

// Middleware to check if user is authenticated
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Exclude login page from authentication check
		if r.URL.Path == "/login" {
			next(w, r)
			return
		}

		// Check if user is authenticated
		_, authenticated := getSessionUser(r)
		if !authenticated {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next(w, r)
	}
}
