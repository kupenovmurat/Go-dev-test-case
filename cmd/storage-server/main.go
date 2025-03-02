package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kupenovmurat/Go-dev-test-case/pkg/common"
)

var (
	port       = flag.String("port", "8081", "Port to listen on")
	serverID   = flag.String("id", "", "Server ID (generated if empty)")
	dataDir    = flag.String("data", "./data", "Directory to store chunks")
	restServer = flag.String("rest", "http://localhost:8080", "REST server URL")
)

func main() {
	flag.Parse()

	if *serverID == "" {
		*serverID = uuid.New().String()
	}

	if err := os.MkdirAll(*dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	if err := registerWithRestServer(); err != nil {
		log.Printf("Warning: Failed to register with REST server: %v", err)
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"serverID": *serverID,
		})
	})

	router.POST("/upload", handleUploadChunk)
	router.GET("/download/:chunkId", handleDownloadChunk)

	serverAddr := fmt.Sprintf(":%s", *port)
	log.Printf("Storage server %s starting on %s", *serverID, serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func registerWithRestServer() error {
	log.Printf("Registering storage server %s with REST server at %s", *serverID, *restServer)
	return nil
}

func handleUploadChunk(c *gin.Context) {
	var req common.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{Error: "Invalid request"})
		return
	}

	chunkPath := filepath.Join(*dataDir, req.ChunkID)
	file, err := os.Create(chunkPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse{Error: "Failed to create chunk file"})
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, c.Request.Body); err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse{Error: "Failed to write chunk data"})
		return
	}

	c.JSON(http.StatusOK, common.UploadResponse{Success: true})
}

func handleDownloadChunk(c *gin.Context) {
	chunkID := c.Param("chunkId")
	if chunkID == "" {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{Error: "Chunk ID is required"})
		return
	}

	chunkPath := filepath.Join(*dataDir, chunkID)
	if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, common.ErrorResponse{Error: "Chunk not found"})
		return
	}

	c.File(chunkPath)
}
