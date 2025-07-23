package azure

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func DownloadFolderAsZip(connectionString, containerName, folderPath string) ([]byte, error) {
	// Cria cliente
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, err
	}
	containerClient := client.ServiceClient().NewContainerClient(containerName)

	// Regex para filtrar blobs
	pattern := "^" + regexp.QuoteMeta(folderPath)
	pattern = regexp.MustCompile(`\\\*`).ReplaceAllString(pattern, "[^/]+")
	pattern = regexp.MustCompile(`\\\/`).ReplaceAllString(pattern, "/")
	pattern += "(/|$)"
	regex, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return nil, err
	}

	// Busca blobs
	pager := containerClient.NewListBlobsFlatPager(nil)
	var files []struct {
		Name string
		Data []byte
	}
	ctx := context.Background()
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, blob := range page.Segment.BlobItems {
			if regex.MatchString(*blob.Name) {
				blobClient := containerClient.NewBlobClient(*blob.Name)
				getResp, err := blobClient.DownloadStream(ctx, nil)
				if err != nil {
					return nil, err
				}
				data, err := io.ReadAll(getResp.Body)
				if err != nil {
					return nil, err
				}
				files = append(files, struct {
					Name string
					Data []byte
				}{*blob.Name, data})
			}
		}
	}

	if len(files) == 0 {
		return nil, os.ErrNotExist
	}

	// Cria ZIP em memória
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	for _, file := range files {
		f, err := zipWriter.Create(file.Name)
		if err != nil {
			return nil, err
		}
		_, err = f.Write(file.Data)
		if err != nil {
			return nil, err
		}
	}
	zipWriter.Close()
	return buf.Bytes(), nil
}
