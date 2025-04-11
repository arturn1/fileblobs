package azure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

func UploadBlob(path string, data []byte) error {
	account := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	key := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	containerName := os.Getenv("AZURE_STORAGE_CONTAINER")

	if account == "" || key == "" || containerName == "" {
		return fmt.Errorf("variáveis de ambiente ausentes")
	}

	cred, err := azblob.NewSharedKeyCredential(account, key)
	if err != nil {
		return fmt.Errorf("erro criando credencial: %w", err)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", account)
	serviceClient, err := service.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return fmt.Errorf("erro criando service client: %w", err)
	}

	containerClient := serviceClient.NewContainerClient(containerName)
	blobClient := containerClient.NewBlockBlobClient(path)

	_, err = blobClient.UploadBuffer(context.Background(), data, nil)
	if err != nil {
		return fmt.Errorf("erro ao fazer upload do blob: %w", err)
	}

	return nil
}

func UploadMultipleBlobs(prefix string, files map[string][]byte) error {
	account := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	key := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	containerName := os.Getenv("AZURE_STORAGE_CONTAINER")

	if account == "" || key == "" || containerName == "" {
		return fmt.Errorf("variáveis de ambiente ausentes")
	}

	cred, err := azblob.NewSharedKeyCredential(account, key)
	if err != nil {
		return fmt.Errorf("erro criando credencial: %w", err)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", account)
	serviceClient, err := service.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return fmt.Errorf("erro criando service client: %w", err)
	}

	containerClient := serviceClient.NewContainerClient(containerName)

	for filename, content := range files {
		path := filepath.Join(prefix, filename)
		blobClient := containerClient.NewBlockBlobClient(path)

		_, err := blobClient.UploadBuffer(context.Background(), content, nil)
		if err != nil {
			return fmt.Errorf("erro ao fazer upload de %s: %w", filename, err)
		}
	}

	return nil
}
