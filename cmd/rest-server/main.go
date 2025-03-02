package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kupenovmurat/Go-dev-test-case/pkg/common"
)

var (
	port = flag.String("port", "8080", "Port to listen on")
)

// In-memory database for file metadata and storage servers
// In a production environment, this would be a persistent database
var (
	filesMutex     sync.RWMutex
	files          = make(map[string]common.FileMetadata)
	serversMutex   sync.RWMutex
	storageServers = make(map[string]common.StorageServer)
	uploadsMutex   sync.Mutex
	activeUploads  = make(map[string]bool)
	chunkSize      = int64(1024 * 1024) // 1MB default chunk size, will be adjusted based on file size
	minChunks      = 6                  // Minimum number of chunks per file
)

func main() {
	flag.Parse()

	// Set up the HTTP server
	router := gin.Default()

	// API routes
	router.POST("/upload", handleUpload)
	router.GET("/download/:fileId", handleDownload)
	router.GET("/files", listFiles)
	router.DELETE("/files/:fileId", deleteFile)

	// Storage server management
	router.POST("/storage/register", registerStorageServer)
	router.GET("/storage/servers", listStorageServers)

	// Start the server
	serverAddr := fmt.Sprintf(":%s", *port)
	log.Printf("REST server starting on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleUpload(c *gin.Context) {
	// Get the file from the request
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{Error: "No file provided"})
		return
	}
	defer file.Close()

	// Check if we have enough storage servers
	serversMutex.RLock()
	numServers := len(storageServers)
	serversMutex.RUnlock()

	if numServers < minChunks {
		c.JSON(http.StatusServiceUnavailable, common.ErrorResponse{
			Error: fmt.Sprintf("Not enough storage servers available. Need at least %d, have %d", minChunks, numServers),
		})
		return
	}

	// Generate a unique ID for the file
	fileID := uuid.New().String()

	// Mark this upload as active
	uploadsMutex.Lock()
	activeUploads[fileID] = true
	uploadsMutex.Unlock()

	// Create metadata for the file
	metadata := common.FileMetadata{
		ID:          fileID,
		Name:        header.Filename,
		Size:        header.Size,
		ContentType: header.Header.Get("Content-Type"),
		CreatedAt:   time.Now(),
		ChunkCount:  minChunks,
		Chunks:      make([]common.Chunk, minChunks),
	}

	// Calculate chunk size
	chunkSize := calculateChunkSize(header.Size, minChunks)

	// Select storage servers for each chunk
	selectedServers, err := selectStorageServers(minChunks)
	if err != nil {
		uploadsMutex.Lock()
		delete(activeUploads, fileID)
		uploadsMutex.Unlock()
		c.JSON(http.StatusInternalServerError, common.ErrorResponse{Error: "Failed to select storage servers"})
		return
	}

	// Create a buffer to read the file in chunks
	buffer := make([]byte, chunkSize)

	// Upload each chunk to a storage server
	var wg sync.WaitGroup
	errChan := make(chan error, minChunks)

	for i := 0; i < minChunks; i++ {
		wg.Add(1)
		go func(index int, server common.StorageServer) {
			defer wg.Done()

			// Read a chunk from the file
			n, err := io.ReadFull(file, buffer)
			if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
				errChan <- fmt.Errorf("failed to read chunk %d: %v", index, err)
				return
			}

			// If we read some data, upload it
			if n > 0 {
				chunkID := uuid.New().String()
				chunk := common.Chunk{
					ID:        chunkID,
					FileID:    fileID,
					Index:     index,
					Size:      int64(n),
					ServerID:  server.ID,
					ServerURL: server.URL,
				}

				// Upload the chunk to the storage server
				if err := uploadChunk(chunk, buffer[:n], server.URL); err != nil {
					errChan <- fmt.Errorf("failed to upload chunk %d: %v", index, err)
					return
				}

				// Add the chunk to the metadata
				metadata.Chunks[index] = chunk

				// Update the server's used space
				serversMutex.Lock()
				server.UsedSpace += int64(n)
				storageServers[server.ID] = server
				serversMutex.Unlock()
			}
		}(i, selectedServers[i])
	}

	// Wait for all uploads to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		uploadsMutex.Lock()
		delete(activeUploads, fileID)
		uploadsMutex.Unlock()
		c.JSON(http.StatusInternalServerError, common.ErrorResponse{Error: err.Error()})
		return
	}

	// Save the file metadata
	filesMutex.Lock()
	files[fileID] = metadata
	filesMutex.Unlock()

	// Mark the upload as complete
	uploadsMutex.Lock()
	delete(activeUploads, fileID)
	uploadsMutex.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"fileId": fileID,
		"name":   metadata.Name,
		"size":   metadata.Size,
	})
}

func handleDownload(c *gin.Context) {
	fileID := c.Param("fileId")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{Error: "File ID is required"})
		return
	}

	// Get the file metadata
	filesMutex.RLock()
	metadata, exists := files[fileID]
	filesMutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, common.ErrorResponse{Error: "File not found"})
		return
	}

	// Set the response headers
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", metadata.Name))
	c.Header("Content-Type", metadata.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", metadata.Size))

	// Download and stream each chunk
	for i := 0; i < metadata.ChunkCount; i++ {
		chunk := metadata.Chunks[i]

		// Download the chunk from the storage server
		data, err := downloadChunk(chunk)
		if err != nil {
			c.JSON(http.StatusInternalServerError, common.ErrorResponse{Error: fmt.Sprintf("Failed to download chunk %d: %v", i, err)})
			return
		}

		// Write the chunk to the response
		if _, err := c.Writer.Write(data); err != nil {
			log.Printf("Error writing chunk %d to response: %v", i, err)
			return
		}
	}
}

func listFiles(c *gin.Context) {
	filesMutex.RLock()
	defer filesMutex.RUnlock()

	fileList := make([]common.FileMetadata, 0, len(files))
	for _, file := range files {
		fileList = append(fileList, file)
	}

	c.JSON(http.StatusOK, fileList)
}

func deleteFile(c *gin.Context) {
	fileID := c.Param("fileId")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{Error: "File ID is required"})
		return
	}

	// Get the file metadata
	filesMutex.Lock()
	metadata, exists := files[fileID]
	if !exists {
		filesMutex.Unlock()
		c.JSON(http.StatusNotFound, common.ErrorResponse{Error: "File not found"})
		return
	}

	// Delete the file metadata
	delete(files, fileID)
	filesMutex.Unlock()

	// Delete each chunk from its storage server
	for _, chunk := range metadata.Chunks {
		// In a real implementation, we would send a delete request to the storage server
		log.Printf("Deleting chunk %s from server %s", chunk.ID, chunk.ServerID)
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func registerStorageServer(c *gin.Context) {
	var server common.StorageServer
	if err := c.ShouldBindJSON(&server); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{Error: "Invalid server data"})
		return
	}

	// Generate a server ID if not provided
	if server.ID == "" {
		server.ID = uuid.New().String()
	}

	// Set the server as available
	server.Available = true

	// Add the server to our list
	serversMutex.Lock()
	storageServers[server.ID] = server
	serversMutex.Unlock()

	c.JSON(http.StatusOK, server)
}

func listStorageServers(c *gin.Context) {
	serversMutex.RLock()
	defer serversMutex.RUnlock()

	serverList := make([]common.StorageServer, 0, len(storageServers))
	for _, server := range storageServers {
		serverList = append(serverList, server)
	}

	c.JSON(http.StatusOK, serverList)
}

func calculateChunkSize(fileSize int64, numChunks int) int64 {
	return int64(math.Ceil(float64(fileSize) / float64(numChunks)))
}

func selectStorageServers(count int) ([]common.StorageServer, error) {
	serversMutex.RLock()
	defer serversMutex.RUnlock()

	if len(storageServers) < count {
		return nil, fmt.Errorf("not enough storage servers available")
	}

	// In a real implementation, we would select servers based on their load, availability, etc.
	// For now, we'll just select the first 'count' servers
	servers := make([]common.StorageServer, 0, count)
	for _, server := range storageServers {
		if server.Available {
			servers = append(servers, server)
			if len(servers) == count {
				break
			}
		}
	}

	if len(servers) < count {
		return nil, fmt.Errorf("not enough available storage servers")
	}

	return servers, nil
}

func uploadChunk(chunk common.Chunk, data []byte, serverURL string) error {
	// Create the upload request
	uploadReq := common.UploadRequest{
		ChunkID: chunk.ID,
		FileID:  chunk.FileID,
		Index:   chunk.Index,
	}

	// Convert the request to JSON
	reqBody, err := json.Marshal(uploadReq)
	if err != nil {
		return err
	}

	// Create the HTTP request
	url := fmt.Sprintf("%s/upload", serverURL)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-Chunk-Metadata", string(reqBody))

	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("storage server returned status %d", resp.StatusCode)
	}

	return nil
}

func downloadChunk(chunk common.Chunk) ([]byte, error) {
	// Create the HTTP request
	url := fmt.Sprintf("%s/download/%s", chunk.ServerURL, chunk.ID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("storage server returned status %d", resp.StatusCode)
	}

	// Read the response body
	return io.ReadAll(resp.Body)
}
