package azure

import (
	"fmt"
	"os"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

var (
	clientMutex  sync.RWMutex
	clientCache  *container.Client
	clientCached bool
)

// ResetClient clears the client cache to force creating a new client with updated credentials
func ResetClient() {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	clientCache = nil
	clientCached = false
}

func GetAzureBlobClient() (*container.Client, error) {
	clientMutex.RLock()
	if clientCached && clientCache != nil {
		defer clientMutex.RUnlock()
		return clientCache, nil
	}
	clientMutex.RUnlock()

	clientMutex.Lock()
	defer clientMutex.Unlock()

	// Check again after acquiring the write lock
	if clientCached && clientCache != nil {
		return clientCache, nil
	}

	containerName := os.Getenv("AZURE_STORAGE_CONTAINER")
	if containerName == "" {
		return nil, fmt.Errorf("variável de ambiente AZURE_STORAGE_CONTAINER não definida")
	}

	account := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	key := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")

	if containerName == "" || account == "" || key == "" {
		return nil, fmt.Errorf("variáveis de ambiente ausentes")
	}

	cred, err := azblob.NewSharedKeyCredential(account, key)
	if err != nil {
		return nil, fmt.Errorf("erro criando credencial: %w", err)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", account)
	serviceClient, err := service.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("erro criando service client: %w", err)
	}

	containerClient := serviceClient.NewContainerClient(containerName)

	// Cache the client
	clientCache = containerClient
	clientCached = true

	return containerClient, nil
}
