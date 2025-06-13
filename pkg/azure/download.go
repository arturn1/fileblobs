package azure

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
)

func DownloadBlob(blobPath string) ([]byte, error) {
	containerClient, err := GetAzureBlobClient()
	if err != nil {
		return nil, err
	}
	// Normalizar o caminho do blob removendo barras iniciais
	// para evitar caminhos como "container//path"
	normalizedPath := blobPath
	for len(normalizedPath) > 0 && normalizedPath[0] == '/' {
		normalizedPath = normalizedPath[1:]
	}

	// Substituir barra invertida por barra normal (importante para Windows)
	normalizedPath = filepath.ToSlash(normalizedPath)

	blobClient := containerClient.NewBlobClient(normalizedPath)

	resp, err := blobClient.DownloadStream(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao baixar blob: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro lendo blob: %w", err)
	}

	return data, nil
}
