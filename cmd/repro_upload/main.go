package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	filepb "github.com/datapeice/astolfosplayer-backend/protos/gen/go/file"
)

func main() {
	// Connect to File Service
	conn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to File Service: %v", err)
	}
	defer conn.Close()
	client := filepb.NewFileServiceClient(conn)

	// Create dummy file content
	content := []byte("Hello, World! This is a test file for reproduction.")
	hashBytes := sha256.Sum256(content)
	hash := hex.EncodeToString(hashBytes[:])

	fmt.Printf("Uploading file with hash: %s\n", hash)

	// Start Upload
	stream, err := client.Upload(context.Background())
	if err != nil {
		log.Fatalf("Failed to start upload: %v", err)
	}

	// Send Metadata
	err = stream.Send(&filepb.UploadRequest{
		Data: &filepb.UploadRequest_Metadata{
			Metadata: &filepb.FileMetadata{
				Filename: "test_repro.txt",
				Title:    "Test Repro",
				Artist:   "Tester",
				Album:    "Reproduction",
				Duration: 10,
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to send metadata: %v", err)
	}

	// Send Chunk
	err = stream.Send(&filepb.UploadRequest{
		Data: &filepb.UploadRequest_Chunk{
			Chunk: content,
		},
	})
	if err != nil {
		log.Fatalf("Failed to send chunk: %v", err)
	}

	// Close and Recv
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Failed to complete upload: %v", err)
	}

	fmt.Printf("Upload complete. Server returned hash: %s\n", resp.Hash)

	if resp.Hash != hash {
		log.Fatalf("Hash mismatch! Expected %s, got %s", hash, resp.Hash)
	}

	// Wait a bit
	time.Sleep(1 * time.Second)

	// Try to Download
	fmt.Println("Attempting to download...")
	downStream, err := client.Download(context.Background(), &filepb.DownloadRequest{Hash: hash})
	if err != nil {
		log.Fatalf("Failed to start download: %v", err)
	}

	recvContent := []byte{}
	for {
		chunk, err := downStream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Fatalf("Failed to download chunk: %v", err)
		}
		recvContent = append(recvContent, chunk.Chunk...)
	}

	if string(recvContent) != string(content) {
		log.Fatalf("Content mismatch! Expected %s, got %s", string(content), string(recvContent))
	}

	fmt.Println("Download successful! content matches.")
}
