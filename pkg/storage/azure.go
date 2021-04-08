package storage

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

type AzureReplayStorage struct {
	AccountName    string
	AccountKey     string
	ContainerName  string
	EndpointSuffix string
}

func (a AzureReplayStorage) Upload(gZipFilePath, target string) (err error) {
	file, err := os.Open(gZipFilePath)
	if err != nil {
		return
	}
	defer file.Close()
	credential, err := azblob.NewSharedKeyCredential(a.AccountName, a.AccountKey)
	if err != nil {
		log.Println("Invalid credentials with error: " + err.Error())
		return err
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	endpoint := fmt.Sprintf("https://%s.blob.%s/%s", a.AccountName, a.EndpointSuffix, a.ContainerName)
	URL, _ := url.Parse(endpoint)
	containerURL := azblob.NewContainerURL(*URL, p)
	blobURL := containerURL.NewBlockBlobURL(target)

	commonResp, err := azblob.UploadFileToBlockBlob(context.TODO(), file, blobURL, azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16})
	if err != nil {
		log.Printf("Azure upload file %s failed: %s", gZipFilePath, err)
		return err
	}
	if httpResp := commonResp.Response(); httpResp != nil {
		defer httpResp.Body.Close()
	}
	return
}

func (a AzureReplayStorage) TypeName() string {
	return "azure"
}
