package common

import (
	"time"
)

type FileMetadata struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Size        int64     `json:"size"`
	ContentType string    `json:"contentType"`
	CreatedAt   time.Time `json:"createdAt"`
	ChunkCount  int       `json:"chunkCount"`
	Chunks      []Chunk   `json:"chunks"`
}

type Chunk struct {
	ID        string `json:"id"`
	FileID    string `json:"fileId"`
	Index     int    `json:"index"`
	Size      int64  `json:"size"`
	ServerID  string `json:"serverId"`
	ServerURL string `json:"serverUrl"`
}

type StorageServer struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Available bool   `json:"available"`
	UsedSpace int64  `json:"usedSpace"`
}
}

type UploadRequest struct {
	ChunkID string `json:"chunkId"`
	FileID  string `json:"fileId"`
	Index   int    `json:"index"`
}

type UploadResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type DownloadRequest struct {
	ChunkID string `json:"chunkId"`
	FileID  string `json:"fileId"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
