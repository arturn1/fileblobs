package azure

import (
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

func GetAzureBlobClient() (*container.Client, error) {
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
	return containerClient, nil
}
