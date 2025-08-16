package security

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type DownloadVerifier struct {
	client *http.Client
}

func NewDownloadVerifier() *DownloadVerifier {
	return &DownloadVerifier{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
	}
}

func (dv *DownloadVerifier) VerifyAndDownload(url, outputPath, expectedSHA256 string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	if outputPath == "" {
		return fmt.Errorf("output path cannot be empty")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "ACC-Server-Manager/1.0")

	resp, err := dv.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer file.Close()

	hash := sha256.New()
	writer := io.MultiWriter(file, hash)

	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		os.Remove(outputPath)
		return fmt.Errorf("failed to write file: %v", err)
	}

	if expectedSHA256 != "" {
		actualHash := fmt.Sprintf("%x", hash.Sum(nil))
		if actualHash != expectedSHA256 {
			os.Remove(outputPath)
			return fmt.Errorf("file hash mismatch: expected %s, got %s", expectedSHA256, actualHash)
		}
	}

	return nil
}