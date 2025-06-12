package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"` // In a real application, store hashed passwords
	IsAdmin  bool   `json:"isAdmin"`  // Indica se o usuário é administrador
}

type StorageAccount struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	AccountName   string `json:"accountName"`
	AccountKey    string `json:"accountKey"`
	ContainerName string `json:"containerName"`
}

type AuthData struct {
	Users           []User           `json:"users"`
	StorageAccounts []StorageAccount `json:"storageAccounts"`
}

var (
	authData     AuthData
	authDataOnce sync.Once
	authMutex    sync.RWMutex

	// Array para armazenar contas temporárias na memória
	temporaryAccounts []StorageAccount
	tempMutex         sync.RWMutex
)

const dataDir = "./data"
const authFile = "auth.json"

func initAuthData() {
	// Create data directory if it doesn't exist
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		err := os.MkdirAll(dataDir, 0755)
		if err != nil {
			fmt.Printf("Erro ao criar diretório de dados: %v\n", err)
			return
		}
	}

	authFilePath := filepath.Join(dataDir, authFile)

	// Check if auth file exists
	if _, err := os.Stat(authFilePath); os.IsNotExist(err) {
		// Create initial auth data with default admin user
		authData = AuthData{
			Users: []User{
				{
					Username: "admin",
					Password: "admin", // In a real app, use hashed passwords
					IsAdmin:  true,
				},
			},
			StorageAccounts: []StorageAccount{
				{
					Name:          "Conta Padrão",
					Description:   "Conta de armazenamento padrão",
					AccountName:   os.Getenv("AZURE_STORAGE_ACCOUNT_NAME"),
					AccountKey:    os.Getenv("AZURE_STORAGE_ACCOUNT_KEY"),
					ContainerName: os.Getenv("AZURE_STORAGE_CONTAINER"),
				},
			},
		}

		// Save the initial data
		saveAuthData()
		return
	}

	// Read existing auth data
	data, err := os.ReadFile(authFilePath)
	if err != nil {
		fmt.Printf("Erro ao ler arquivo de autenticação: %v\n", err)
		return
	}

	// Parse the JSON data
	err = json.Unmarshal(data, &authData)
	if err != nil {
		fmt.Printf("Erro ao analisar dados de autenticação: %v\n", err)
		return
	}
}

func saveAuthData() error {
	authMutex.Lock()
	defer authMutex.Unlock()

	authFilePath := filepath.Join(dataDir, authFile)
	data, err := json.MarshalIndent(authData, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar dados de autenticação: %w", err)
	}

	err = os.WriteFile(authFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("erro ao salvar arquivo de autenticação: %w", err)
	}

	return nil
}

func GetAuthData() *AuthData {
	authDataOnce.Do(initAuthData)
	return &authData
}

func ValidateUser(username, password string) bool {
	authDataOnce.Do(initAuthData)
	authMutex.RLock()
	defer authMutex.RUnlock()

	for _, user := range authData.Users {
		if user.Username == username && user.Password == password {
			return true
		}
	}
	return false
}

// IsUserAdmin verifica se um usuário é administrador
func IsUserAdmin(username string) bool {
	authDataOnce.Do(initAuthData)
	authMutex.RLock()
	defer authMutex.RUnlock()

	// Se o username é oidc_user, consideramos que a verificação
	// já foi feita no middleware e depende do header X-User-Is-Admin
	if username == "oidc_user" {
		return true // A verificação real é feita no middleware via token OIDC
	}

	for _, user := range authData.Users {
		if user.Username == username && user.IsAdmin {
			return true
		}
	}
	return false
}

func GetStorageAccounts() []StorageAccount {
	authDataOnce.Do(initAuthData)

	var persistentAccounts []StorageAccount
	var tempAccounts []StorageAccount

	// Get accounts from persistent storage
	authMutex.RLock()
	persistentAccounts = make([]StorageAccount, len(authData.StorageAccounts))
	copy(persistentAccounts, authData.StorageAccounts)
	authMutex.RUnlock()

	// Get accounts from temporary storage
	tempMutex.RLock()
	tempAccounts = make([]StorageAccount, len(temporaryAccounts))
	copy(tempAccounts, temporaryAccounts)
	tempMutex.RUnlock()

	// Combine accounts, avoiding duplicates
	totalSize := len(persistentAccounts) + len(tempAccounts)
	combinedAccounts := make([]StorageAccount, 0, totalSize)

	// Add accounts from persistent storage
	combinedAccounts = append(combinedAccounts, persistentAccounts...)

	// Add accounts from temporary array, avoiding duplicates
	for _, tempAccount := range tempAccounts {
		isDuplicate := false
		for _, persistentAccount := range persistentAccounts {
			if tempAccount.Name == persistentAccount.Name {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			combinedAccounts = append(combinedAccounts, tempAccount)
		}
	}

	return combinedAccounts
}

func AddStorageAccount(account StorageAccount) error {
	authDataOnce.Do(initAuthData)

	// Use separate mutex for temporary accounts to avoid deadlocks
	tempMutex.Lock()
	defer tempMutex.Unlock()

	// Generate a name if empty
	if account.Name == "" {
		account.Name = fmt.Sprintf("Conta %d", time.Now().Unix())
	}

	// Add to temporary array only
	temporaryAccounts = append(temporaryAccounts, account)

	fmt.Printf("Conta adicionada ao array temporário: %s\n", account.Name)
	return nil
}

func GetStorageAccountByName(name string) (StorageAccount, bool) {
	authDataOnce.Do(initAuthData)

	// First look in persistent storage
	authMutex.RLock()
	for _, account := range authData.StorageAccounts {
		if account.Name == name {
			authMutex.RUnlock()
			return account, true
		}
	}
	authMutex.RUnlock()

	// Then look in temporary storage
	tempMutex.RLock()
	defer tempMutex.RUnlock()
	for _, account := range temporaryAccounts {
		if account.Name == name {
			return account, true
		}
	}

	return StorageAccount{}, false
}

// UpdateStorageAccount updates an existing storage account
func UpdateStorageAccount(originalName string, updatedAccount StorageAccount) error {
	authDataOnce.Do(initAuthData)

	// First check if account is in temporary array
	tempMutex.Lock()
	for i, account := range temporaryAccounts {
		if account.Name == originalName {
			// Update in temporary array
			temporaryAccounts[i] = updatedAccount
			tempMutex.Unlock()
			return nil
		}
	}
	tempMutex.Unlock()

	// Then check if account is in persistent storage
	authMutex.Lock()
	defer authMutex.Unlock()

	// Don't allow updating the default account
	if originalName == "Conta Padrão" {
		return fmt.Errorf("não é permitido editar a conta padrão")
	}

	for i, account := range authData.StorageAccounts {
		if account.Name == originalName {
			// Update in persistent storage
			authData.StorageAccounts[i] = updatedAccount
			return nil
		}
	}

	return fmt.Errorf("conta não encontrada")
}
