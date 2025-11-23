package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	filepb "github.com/datapeice/astolfosplayer-backend/protos/gen/go/file"
)

func main() {
	// Connect to File Service
	fileConn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to File: %v", err)
	}
	defer fileConn.Close()
	fileClient := filepb.NewFileServiceClient(fileConn)

	// Known existing hash from previous MinIO list
	// 1c2e13c44278e45b29aec2e49c8e4d51b5b113fdedb816293d6a25f9fb2b3607
	hash := "1c2e13c44278e45b29aec2e49c8e4d51b5b113fdedb816293d6a25f9fb2b3607"

	fmt.Printf("Attempting to download hash: %s\n", hash)
	stream, err := fileClient.Download(context.Background(), &filepb.DownloadRequest{Hash: hash})
	if err != nil {
		log.Fatalf("Failed to start download: %v", err)
	}

	totalBytes := 0
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error receiving chunk: %v", err)
		}
		totalBytes += len(resp.Chunk)
		fmt.Printf("\rReceived %d bytes...", totalBytes)
	}
	fmt.Printf("\nDownload complete! Total bytes: %d\n", totalBytes)
}
