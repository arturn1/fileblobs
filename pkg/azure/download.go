package azure

import (
	"context"
	"fmt"
	"io"
)

func DownloadBlob(blobPath string) ([]byte, error) {
	containerClient, err := GetAzureBlobClient()
	if err != nil {
		return nil, err
	}

	blobClient := containerClient.NewBlobClient(blobPath)

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
