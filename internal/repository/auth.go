package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"` // In a real application, store hashed passwords
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

	// Get the fixed account key from environment variable
	fixedAccountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")

	authFilePath := filepath.Join(dataDir, authFile)

	// Check if auth file exists
	if _, err := os.Stat(authFilePath); os.IsNotExist(err) {
		// Create initial auth data with default admin user
		authData = AuthData{
			Users: []User{
				{
					Username: "admin",
					Password: "admin", // In a real app, use hashed passwords
				},
			},
			StorageAccounts: []StorageAccount{
				{
					Name:          "Conta Padrão",
					Description:   "Conta de armazenamento padrão",
					AccountName:   os.Getenv("AZURE_STORAGE_ACCOUNT_NAME"),
					AccountKey:    fixedAccountKey,
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

func GetStorageAccounts() []StorageAccount {
	authDataOnce.Do(initAuthData)
	authMutex.RLock()
	defer authMutex.RUnlock()

	accounts := make([]StorageAccount, len(authData.StorageAccounts))
	copy(accounts, authData.StorageAccounts)
	return accounts
}

func AddStorageAccount(account StorageAccount) error {
	authDataOnce.Do(initAuthData)
	authMutex.Lock()
	defer authMutex.Unlock()

	// Check if account with same name already exists
	for _, acc := range authData.StorageAccounts {
		if acc.Name == account.Name {
			return fmt.Errorf("uma conta com esse nome já existe")
		}
	}

	authData.StorageAccounts = append(authData.StorageAccounts, account)
	return saveAuthData()
}

func GetStorageAccountByName(name string) (StorageAccount, bool) {
	authDataOnce.Do(initAuthData)
	authMutex.RLock()
	defer authMutex.RUnlock()

	for _, account := range authData.StorageAccounts {
		if account.Name == name {
			return account, true
		}
	}
	return StorageAccount{}, false
}
