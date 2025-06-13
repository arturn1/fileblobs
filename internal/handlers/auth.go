package handlers

import (
	"encoding/json"
	"fileblobs/internal/repository"
	"fileblobs/pkg/azure"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var loginTmpl = template.Must(template.ParseFiles("web/templates/login.html"))
var storageAccountsTmpl = template.Must(template.ParseFiles("web/templates/storage_accounts.html"))
var addAccountTmpl = template.Must(template.ParseFiles("web/templates/add_account.html"))
var editAccountTmpl = template.Must(template.ParseFiles("web/templates/edit_account.html"))
var accessDeniedTmpl = template.Must(template.ParseFiles("web/templates/access_denied.html"))
var logoutTmpl = template.Must(template.ParseFiles("web/templates/logout.html"))

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is already logged in
	_, authenticated := getSessionUser(r)
	if authenticated {
		http.Redirect(w, r, "/storage-accounts", http.StatusSeeOther)
		return
	}

	// Verificar se há um cookie de acesso negado para evitar loops
	accessDeniedCookie, err := r.Cookie("access_denied")
	if err == nil && accessDeniedCookie.Value != "" {
		// Limpar o cookie e redirecionar para a página de acesso negado
		http.SetCookie(w, &http.Cookie{
			Name:     "access_denied",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})
		http.Redirect(w, r, "/access-denied", http.StatusSeeOther)
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

	// Limpar também o cookie de acesso negado
	http.SetCookie(w, &http.Cookie{
		Name:     "access_denied",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Limpar o cookie de conta selecionada
	http.SetCookie(w, &http.Cookie{
		Name:     "selected_account",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	log.Printf("Sessão encerrada, cookies limpos")

	// Exibir a página intermediária de logout em vez de redirecionar diretamente
	logoutTmpl.Execute(w, nil)
}

func StorageAccountsHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	username, authenticated := getSessionUser(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Verifica se o usuário é administrador
	isAdmin := repository.IsUserAdmin(username)

	// Valor padrão para role
	userRole := ""
	// Se estiver usando autenticação OIDC, verificar as informações do token
	tokenCookie, err := r.Cookie("access_token")
	if err == nil && tokenCookie.Value != "" {
		// Verificar se o token é válido e obter as claims
		claims, err := ParseJWTClaims(tokenCookie.Value)
		if err == nil {
			// Usar a função de diagnóstico para ver todas as roles
			DumpClaimsInfo(claims)

			// Verificar role nas várias possíveis fontes
			// Prioridade: MsRoles > Roles > Role (campo simples)
			if len(claims.MsRoles) > 0 {
				userRole = claims.MsRoles[0] // Pega a primeira role do array MsRoles
				log.Printf("Usando a primeira role de MsRoles: %s", userRole)
			} else if len(claims.Roles) > 0 {
				userRole = claims.Roles[0] // Pega a primeira role do array Roles
				log.Printf("Usando a primeira role de Roles: %s", userRole)
			} else if claims.Role != "" {
				userRole = claims.Role // Usa o campo Role simples
				log.Printf("Usando o campo Role simples: %s", userRole)
			}

			// Se for admin, definir a flag
			if r.Header.Get("X-User-Is-Admin") == "true" ||
				strings.Contains(strings.ToLower(userRole), "admin") ||
				strings.EqualFold(userRole, "Administrator") {
				isAdmin = true
				log.Printf("Usuário definido como administrador baseado na role: %s", userRole)
			}
		}
	}
	// Formatar a role para exibição mais amigável, removendo prefixos e sufixos técnicos
	displayRole := userRole
	if displayRole != "" {
		// Verificar se é "IdentityConsultant" - se for, não mostrar essa role
		if strings.Contains(strings.ToLower(displayRole), "identity") {
			displayRole = ""
			log.Printf("Role IdentityConsultant detectada, não será exibida: %s", userRole)
		} else {
			// Remover prefixos comuns como "http://schemas.microsoft.com/..."
			if strings.Contains(displayRole, "/") {
				parts := strings.Split(displayRole, "/")
				displayRole = parts[len(parts)-1]
			}

			// Substituir caracteres especiais e formatação
			displayRole = strings.ReplaceAll(displayRole, "#", "")
			displayRole = strings.ReplaceAll(displayRole, "_", " ")

			// Capitalizar primeira letra de cada palavra
			words := strings.Fields(displayRole)
			for i, word := range words {
				if len(word) > 0 {
					words[i] = strings.ToUpper(word[:1]) + word[1:]
				}
			}
			displayRole = strings.Join(words, " ")

			log.Printf("Role formatada para exibição: %s", displayRole)
		}
	}

	// Display storage accounts
	accounts := repository.GetStorageAccounts()
	storageAccountsTmpl.Execute(w, map[string]interface{}{
		"Accounts": accounts,
		"IsAdmin":  isAdmin,
		"UserName": username,
		"Role":     displayRole,
		"RawRole":  userRole, // Adicionar a role original para debugging
	})
}

func AddAccountHandler(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	username, authenticated := getSessionUser(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Verifica se o usuário é administrador
	isAdmin := repository.IsUserAdmin(username)

	// Se estiver usando autenticação OIDC, verifica o header X-User-Is-Admin
	if r.Header.Get("X-User-Is-Admin") == "true" {
		isAdmin = true
	}

	if !isAdmin {
		http.Error(w, "Acesso negado. Apenas administradores podem adicionar contas.", http.StatusForbidden)
		return
	}

	// Carregar a chave padrão para exibição
	defaultAccountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	defaultAccountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	defaultContainerName := os.Getenv("AZURE_STORAGE_CONTAINER")

	if r.Method == http.MethodPost {
		// Process the form submission
		name := r.FormValue("name")
		description := r.FormValue("description")
		accountName := r.FormValue("accountName")
		useDefaultKey := r.FormValue("useDefaultKey")
		accountKey := r.FormValue("accountKey")
		containerName := r.FormValue("containerName")

		// Se selecionou usar a chave padrão
		if useDefaultKey == "yes" {
			accountKey = defaultAccountKey
		}

		// Validate inputs
		if name == "" || accountName == "" || accountKey == "" || containerName == "" {
			addAccountTmpl.Execute(w, map[string]interface{}{
				"Error":                "Todos os campos são obrigatórios",
				"DefaultAccountName":   defaultAccountName,
				"DefaultAccountKey":    defaultAccountKey,
				"DefaultContainerName": defaultContainerName,
			})
			return
		}
		// Create new storage account
		newAccount := repository.StorageAccount{
			Name:          name,
			Description:   description,
			AccountName:   accountName,
			AccountKey:    accountKey,
			ContainerName: containerName,
		}

		// Add to repository
		err := repository.AddStorageAccount(newAccount)
		if err != nil {
			addAccountTmpl.Execute(w, map[string]interface{}{
				"Error":                err.Error(),
				"DefaultAccountName":   defaultAccountName,
				"DefaultAccountKey":    defaultAccountKey,
				"DefaultContainerName": defaultContainerName,
			})
			return
		}

		// Redirect to storage accounts list
		http.Redirect(w, r, "/storage-accounts", http.StatusSeeOther)
		return
	}
	// Display add account form
	addAccountTmpl.Execute(w, map[string]interface{}{
		"DefaultAccountName":   defaultAccountName,
		"DefaultAccountKey":    defaultAccountKey,
		"DefaultContainerName": defaultContainerName,
	})
}

// EditAccountHandler lida com a edição de contas de armazenamento
func EditAccountHandler(w http.ResponseWriter, r *http.Request) {
	// Verificar se o usuário está autenticado
	username, authenticated := getSessionUser(r)
	if !authenticated {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Verificar se o usuário é administrador
	isAdmin := repository.IsUserAdmin(username)

	// Se estiver usando autenticação OIDC, verifica o header X-User-Is-Admin
	if r.Header.Get("X-User-Is-Admin") == "true" {
		isAdmin = true
	}

	if !isAdmin {
		http.Error(w, "Acesso negado. Apenas administradores podem editar contas.", http.StatusForbidden)
		return
	}

	// Carregar a chave padrão para exibição e comparação
	defaultAccountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")

	// Obter o nome da conta da URL
	accountName := r.URL.Query().Get("name")

	if r.Method == http.MethodPost {
		// Processar envio do formulário
		originalName := r.FormValue("originalName")
		name := r.FormValue("name")
		description := r.FormValue("description")
		azureAccountName := r.FormValue("accountName")
		useDefaultKey := r.FormValue("useDefaultKey")
		accountKey := r.FormValue("accountKey")
		containerName := r.FormValue("containerName")

		// Se selecionou usar a chave padrão
		if useDefaultKey == "yes" {
			accountKey = defaultAccountKey
		}

		// Validar entradas
		if name == "" || azureAccountName == "" || accountKey == "" || containerName == "" {
			account, found := repository.GetStorageAccountByName(originalName)
			if !found {
				http.Error(w, "Conta não encontrada", http.StatusNotFound)
				return
			}

			editAccountTmpl.Execute(w, map[string]interface{}{
				"Error":             "Todos os campos são obrigatórios",
				"Account":           account,
				"DefaultAccountKey": defaultAccountKey,
			})
			return
		}

		// Verificar se está tentando editar a conta padrão
		if originalName == "Conta Padrão" {
			http.Error(w, "Não é permitido editar a conta padrão", http.StatusForbidden)
			return
		}

		// Criar objeto da conta atualizada
		updatedAccount := repository.StorageAccount{
			Name:          name,
			Description:   description,
			AccountName:   azureAccountName,
			AccountKey:    accountKey,
			ContainerName: containerName,
		}

		// Atualizar no repositório
		err := repository.UpdateStorageAccount(originalName, updatedAccount)
		if err != nil {
			account, _ := repository.GetStorageAccountByName(originalName)
			editAccountTmpl.Execute(w, map[string]interface{}{
				"Error":             err.Error(),
				"Account":           account,
				"DefaultAccountKey": defaultAccountKey,
			})
			return
		}

		// Redirecionar para a lista de contas
		http.Redirect(w, r, "/storage-accounts", http.StatusSeeOther)
		return
	}

	// Obter a conta pelo nome
	account, found := repository.GetStorageAccountByName(accountName)
	if !found {
		http.Error(w, "Conta não encontrada", http.StatusNotFound)
		return
	}

	// Verificar se está tentando editar a conta padrão
	if account.Name == "Conta Padrão" {
		http.Error(w, "Não é permitido editar a conta padrão", http.StatusForbidden)
		return
	}

	// Exibir formulário de edição
	editAccountTmpl.Execute(w, map[string]interface{}{
		"Account":           account,
		"DefaultAccountKey": defaultAccountKey,
	})
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
	} // Set environment variables for Azure storage
	os.Setenv("AZURE_STORAGE_ACCOUNT_NAME", account.AccountName)
	os.Setenv("AZURE_STORAGE_ACCOUNT_KEY", account.AccountKey)
	os.Setenv("AZURE_STORAGE_CONTAINER", account.ContainerName) // Store the selected account in a cookie for better state management
	selectedAccountCookie := &http.Cookie{
		Name:     "selected_account",
		Value:    accountName,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	}

	log.Printf("Selecionando conta: '%s', Cookie definido: '%s', É padrão: %v",
		accountName,
		accountName,
		accountName == "" || strings.Contains(strings.ToLower(accountName), "conta padr") || accountName == "Conta Padrão")

	http.SetCookie(w, selectedAccountCookie)

	// Clear Azure client cache to use new credentials
	azure.ResetClient()

	// Redirect to home page to browse files
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// TokenRequest representa o payload JSON do cliente
type TokenRequest struct {
	Token string `json:"token"`
}

// StoreTokenHandler armazena o token OIDC na sessão
func StoreTokenHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("StoreTokenHandler chamado com método:", r.Method)

	// Configurar cabeçalhos CORS para esta resposta específica
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// Preflight response para OPTIONS
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Aceita apenas solicitações POST
	if r.Method != http.MethodPost {
		log.Println("Método não permitido:", r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Método não permitido"})
		return
	}

	// Lê e analisa o corpo da requisição
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Erro ao ler corpo da requisição:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao ler dados"})
		return
	}

	var tokenReq TokenRequest
	err = json.Unmarshal(body, &tokenReq)
	if err != nil {
		log.Println("Erro ao analisar JSON:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "JSON inválido"})
		return
	}
	if tokenReq.Token == "" {
		log.Println("Token vazio recebido")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Token não fornecido"})
		return
	}

	// Verifica se o usuário tem as permissões necessárias
	claims, err := ParseJWTClaims(tokenReq.Token)
	if err != nil {
		log.Println("Erro ao analisar token:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Token inválido"})
		return
	}
	// Log detalhado das claims para diagnóstico
	log.Printf("Claims detalhadas do token:\n- Sub: %s\n- Name: %s\n- Email: %s\n- Role: %s\n- MsRoles: %v\n- Roles: %v\n- Groups: %v\n- PreferredName: %s",
		claims.Sub, claims.Name, claims.Email, claims.Role, claims.MsRoles, claims.Roles, claims.Groups, claims.PreferredName)

	// Verifica se o usuário tem uma role válida
	hasValidRole := HasValidRole(claims)
	log.Printf("Resultado da verificação de roles para %s: %v", claims.Name, hasValidRole)

	if !hasValidRole {
		log.Println("Acesso negado: usuário não tem permissão. Roles:", claims.Role, claims.Roles)

		// Define um cookie para evitar loops de redirecionamento
		http.SetCookie(w, &http.Cookie{
			Name:     "access_denied",
			Value:    "true",
			Path:     "/",
			MaxAge:   300, // 5 minutos
			HttpOnly: true,
		})
		// Sempre definir o status como 403 Forbidden para que o cliente possa detectar
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)

		// Incluir informação de redirecionamento no JSON
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "access_denied",
			"error_description": "Acesso negado. Você não tem permissão para acessar este aplicativo. Roles necessárias: Administrator ou Consultant. A role IdentityConsultant não é suficiente.",
			"redirect":          "/access-denied",
		})
		return
	}
	// Registra o login bem-sucedido com informações do usuário
	log.Printf("Login autorizado para usuário: %s, email: %s, roles: %v",
		claims.Name,
		claims.Email,
		claims.Roles)

	// Determina o nome de usuário a ser armazenado
	userName := claims.Name
	if userName == "" {
		userName = claims.PreferredName
	}
	if userName == "" {
		userName = claims.Email
	}
	if userName == "" {
		userName = "oidc_user"
	}

	// Armazena o token em um cookie HTTP seguro
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    tokenReq.Token,
		Path:     "/",
		MaxAge:   3600 * 8, // 8 horas
		HttpOnly: true,
		Secure:   r.TLS != nil, // Seguro se usando HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Também definir o nome de usuário na sessão para compatibilidade com o fluxo existente
	setSessionUser(w, userName)

	log.Println("Token armazenado com sucesso")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "redirect": "/storage-accounts"})
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

	// Também limpar o cookie de token
	tokenCookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, tokenCookie)
	// Limpar o cookie de conta selecionada
	selectedAccountCookie := &http.Cookie{
		Name:     "selected_account",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, selectedAccountCookie)

	log.Printf("Sessão encerrada, cookies limpos")
}

// AccessDeniedPageHandler exibe a página de acesso negado com uma mensagem personalizada
// Esta função pode ser chamada diretamente como um handler de rota
func AccessDeniedPageHandler(w http.ResponseWriter, r *http.Request) {
	// Obter a mensagem da query string se existir
	message := r.URL.Query().Get("message")
	if message == "" {
		message = "Seu usuário não possui as permissões necessárias para acessar esta aplicação."
	}

	// Verifica se temos um token para extrair mais informações
	tokenCookie, err := r.Cookie("access_token")
	roleInfo := ""
	if err == nil && tokenCookie.Value != "" {
		claims, err := ParseJWTClaims(tokenCookie.Value)
		if err == nil { // Obter informações sobre as roles do usuário para diagnóstico
			roleInfo = "Suas roles: "
			if claims.Role != "" {
				roleInfo += fmt.Sprintf("Role: %s, ", claims.Role)
			}
			if len(claims.MsRoles) > 0 {
				roleInfo += fmt.Sprintf("MsRoles: %v, ", claims.MsRoles)
			}
			if len(claims.Roles) > 0 {
				roleInfo += fmt.Sprintf("Roles: %v, ", claims.Roles)
			}
			if len(claims.Groups) > 0 {
				roleInfo += fmt.Sprintf("Groups: %v", claims.Groups)
			}

			// Se não encontramos nenhuma role
			if claims.Role == "" && len(claims.MsRoles) == 0 && len(claims.Roles) == 0 && len(claims.Groups) == 0 {
				roleInfo += "Nenhuma role encontrada no token."
			}

			// Verificar especificamente por IdentityConsultant
			hasIdentityConsultant := false
			if strings.Contains(strings.ToLower(claims.Role), "identity") {
				hasIdentityConsultant = true
			}
			for _, role := range claims.MsRoles {
				if strings.Contains(strings.ToLower(role), "identity") {
					hasIdentityConsultant = true
					break
				}
			}
			for _, role := range claims.Roles {
				if strings.Contains(strings.ToLower(role), "identity") {
					hasIdentityConsultant = true
					break
				}
			}

			if hasIdentityConsultant {
				roleInfo += " (Detectada a role IdentityConsultant, que não é suficiente para acesso.)"
			}
		}
	}

	if roleInfo != "" {
		message += " " + roleInfo
	}

	// Limpar todos os cookies de autenticação para evitar loops de redirecionamento
	clearSession(w)

	// Limpar especificamente o cookie de acesso negado
	http.SetCookie(w, &http.Cookie{
		Name:     "access_denied",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Exibir a página de acesso negado
	w.WriteHeader(http.StatusForbidden)
	accessDeniedTmpl.Execute(w, map[string]interface{}{
		"Message": message,
	})
}

// AccessDeniedHandler exibe uma página de acesso negado com uma mensagem personalizada
func AccessDeniedHandler(w http.ResponseWriter, r *http.Request, message string) {
	w.WriteHeader(http.StatusForbidden)
	accessDeniedTmpl.Execute(w, map[string]interface{}{
		"Message": message,
	})
}

// Middleware to check if user is authenticated
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Exclude login page, token storage and access-denied from authentication check
		if r.URL.Path == "/login" || r.URL.Path == "/auth/store-token" || r.URL.Path == "/access-denied" {
			next(w, r)
			return
		}

		// Verificar se há um cookie de acesso negado para evitar loops
		accessDeniedCookie, err := r.Cookie("access_denied")
		if err == nil && accessDeniedCookie.Value != "" {
			// Limpar o cookie e redirecionar para a página de acesso negado
			http.SetCookie(w, &http.Cookie{
				Name:     "access_denied",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
			})
			http.Redirect(w, r, "/access-denied", http.StatusSeeOther)
			return
		}

		// First check if user is authenticated via OIDC token
		tokenCookie, err := r.Cookie("access_token")
		if err == nil && tokenCookie.Value != "" {
			// Verificar se o token é válido e tem as permissões corretas
			claims, err := ParseJWTClaims(tokenCookie.Value)
			if err != nil {
				log.Printf("Token inválido: %v", err)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			// Verificar permissões
			if !HasValidRole(claims) {
				log.Printf("Acesso negado: usuário %s não tem permissão. Roles: %v, %v",
					claims.Name, claims.Role, claims.Roles)

				// Redirecionar para a página de acesso negado em vez de chamar o handler diretamente
				http.Redirect(w, r, "/access-denied", http.StatusSeeOther)
				return
			}

			// Token found and valid, user has required roles
			log.Printf("Usuário autenticado via token OIDC: %s (%s)", claims.Name, claims.Email) // Se o usuário tem role de Administrator, configurar como admin
			isAdmin := false
			for _, role := range claims.Roles {
				if strings.EqualFold(role, "Administrator") || strings.EqualFold(role, "Admin") {
					isAdmin = true
					log.Printf("Usuário %s é admin baseado em role: %s", claims.Name, role)
					break
				}
			}
			if strings.EqualFold(claims.Role, "Administrator") || strings.EqualFold(claims.Role, "Admin") {
				isAdmin = true
				log.Printf("Usuário %s é admin baseado em claims.Role: %s", claims.Name, claims.Role)
			}

			// Verificar nas MsRoles processadas
			for _, msRole := range claims.MsRoles {
				if strings.EqualFold(msRole, "Administrator") || strings.EqualFold(msRole, "Admin") {
					isAdmin = true
					log.Printf("Usuário %s é admin baseado em claims.MsRoles: %s", claims.Name, msRole)
					break
				}
			}

			// Definir informações do usuário na sessão
			userName := claims.Name
			if userName == "" {
				userName = claims.PreferredName
			}
			if userName == "" {
				userName = claims.Email
			}
			if userName == "" {
				userName = "oidc_user"
			}

			// Armazenar o nome de usuário na sessão
			setSessionUser(w, userName)

			// Definir se o usuário é admin baseado em suas roles
			if isAdmin {
				r.Header.Set("X-User-Is-Admin", "true")
			}

			next(w, r)
			return
		}

		// Then check if user is authenticated via traditional session
		_, authenticated := getSessionUser(r)
		if authenticated {
			next(w, r)
			return
		}

		// Not authenticated with either method
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
