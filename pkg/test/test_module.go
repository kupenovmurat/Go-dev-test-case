package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TestModule struct {
	RestServerURL    string
	StorageServerURL string
	TestDataDir      string
}

func NewTestModule(restServerURL, storageServerURL, testDataDir string) *TestModule {
	return &TestModule{
		RestServerURL:    restServerURL,
		StorageServerURL: storageServerURL,
		TestDataDir:      testDataDir,
	}
}

func (tm *TestModule) RegisterStorageServer(serverURL string) error {
	resp, err := http.Post(
		fmt.Sprintf("%s/storage/register", tm.RestServerURL),
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"url": "%s"}`, serverURL)),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register storage server: %s", resp.Status)
	}

	return nil
}

func (tm *TestModule) GenerateTestFile(filename string, size int64) (string, error) {
	if err := os.MkdirAll(tm.TestDataDir, 0755); err != nil {
		return "", err
	}

	filePath := filepath.Join(tm.TestDataDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	remaining := size
	for remaining > 0 {
		writeSize := remaining
		if writeSize > int64(len(data)) {
			writeSize = int64(len(data))
		}
		if _, err := file.Write(data[:writeSize]); err != nil {
			return "", err
		}
		remaining -= writeSize
	}

	return filePath, nil
}

func (tm *TestModule) UploadFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	fileField, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(fileField, file); err != nil {
		return "", err
	}

	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/upload", tm.RestServerURL), &requestBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		FileID string `json:"fileId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	return response.FileID, nil
}

func (tm *TestModule) DownloadFile(fileID, outputPath string) error {
	resp, err := http.Get(fmt.Sprintf("%s/download/%s", tm.RestServerURL, fileID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	if _, err := io.Copy(outputFile, resp.Body); err != nil {
		return err
	}

	return nil
}

func (tm *TestModule) CompareFiles(file1, file2 string) (bool, error) {
	f1, err := os.Open(file1)
	if err != nil {
		return false, err
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return false, err
	}
	defer f2.Close()

	const bufferSize = 64 * 1024
	buf1 := make([]byte, bufferSize)
	buf2 := make([]byte, bufferSize)

	for {
		n1, err1 := f1.Read(buf1)
		n2, err2 := f2.Read(buf2)

		if n1 != n2 {
			return false, nil
		}

		if !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}

		if err1 == io.EOF && err2 == io.EOF {
			return true, nil
		}

		if err1 != nil && err1 != io.EOF {
			return false, err1
		}

		if err2 != nil && err2 != io.EOF {
			return false, err2
		}
	}
}

// RunTest runs a complete test of the distributed storage system
func (tm *TestModule) RunTest(fileSize int64) error {
	log.Println("Starting test...")

	filename := fmt.Sprintf("test_file_%d.dat", time.Now().Unix())
	filePath, err := tm.GenerateTestFile(filename, fileSize)
	if err != nil {
		return fmt.Errorf("failed to generate test file: %v", err)
	}
	log.Printf("Generated test file: %s (%d bytes)", filePath, fileSize)

	fileID, err := tm.UploadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}
	log.Printf("Uploaded file with ID: %s", fileID)

	downloadPath := filepath.Join(tm.TestDataDir, fmt.Sprintf("downloaded_%s", filename))
	if err := tm.DownloadFile(fileID, downloadPath); err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	log.Printf("Downloaded file to: %s", downloadPath)

	identical, err := tm.CompareFiles(filePath, downloadPath)
	if err != nil {
		return fmt.Errorf("failed to compare files: %v", err)
	}

	if !identical {
		return fmt.Errorf("downloaded file does not match the original")
	}

	log.Println("Test completed successfully!")
	return nil
}
