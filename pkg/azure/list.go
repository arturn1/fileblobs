package azure

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

func ListFoldersAndFiles(prefix string) (folders []string, files []string, err error) {
	account := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	key := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	containerName := os.Getenv("AZURE_STORAGE_CONTAINER")

	if account == "" || key == "" || containerName == "" {
		return nil, nil, fmt.Errorf("variáveis de ambiente ausentes")
	}

	cred, err := azblob.NewSharedKeyCredential(account, key)
	if err != nil {
		return nil, nil, fmt.Errorf("erro criando credencial: %w", err)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", account)
	serviceClient, err := service.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("erro criando service client: %w", err)
	}

	containerClient := serviceClient.NewContainerClient(containerName)

	pager := containerClient.NewListBlobsHierarchyPager("/", &container.ListBlobsHierarchyOptions{
		Prefix: &prefix,
	})

	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, nil, fmt.Errorf("erro ao paginar blobs: %w", err)
		}

		for _, p := range page.Segment.BlobPrefixes {
			if p.Name != nil {
				folders = append(folders, strings.TrimSuffix(*p.Name, "/"))
			}
		}

		for _, blob := range page.Segment.BlobItems {
			if blob.Name != nil && !strings.HasSuffix(*blob.Name, "/") {
				name := strings.TrimPrefix(*blob.Name, prefix)
				if !strings.Contains(name, "/") {
					files = append(files, name)
				}
			}
		}
	}

	return folders, files, nil
}

// Lista todos os arquivos recursivamente dentro de uma pasta (para gerar ZIP)
func ListBlobsFromFolder(prefix string) ([]string, error) {
	var files []string

	account := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	key := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	containerName := os.Getenv("AZURE_STORAGE_CONTAINER")

	if account == "" || key == "" || containerName == "" {
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

	pager := containerClient.NewListBlobsFlatPager(&container.ListBlobsFlatOptions{
		Prefix: &prefix,
	})

	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("erro listando blobs: %w", err)
		}

		for _, blob := range page.Segment.BlobItems {
			if blob.Name != nil && !strings.HasSuffix(*blob.Name, "/") {
				files = append(files, *blob.Name)
			}
		}
	}

	return files, nil
}
