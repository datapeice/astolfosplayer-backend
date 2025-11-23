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
	conn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	client := filepb.NewFileServiceClient(conn)

	content := []byte("Soft Delete Test Content " + time.Now().String())
	hashBytes := sha256.Sum256(content)
	hash := hex.EncodeToString(hashBytes[:])

	fmt.Printf("Testing with hash: %s\n", hash)

	// 1. Upload
	fmt.Println("1. Uploading...")
	upload(client, content, "test.txt")
	fmt.Println("   Uploaded.")

	// 2. Delete
	fmt.Println("2. Deleting...")
	_, err = client.Delete(context.Background(), &filepb.DeleteRequest{Hash: hash})
	if err != nil {
		log.Fatalf("Failed to delete: %v", err)
	}
	fmt.Println("   Deleted.")

	// 3. Upload again (should succeed now)
	fmt.Println("3. Uploading again...")
	upload(client, content, "test.txt")
	fmt.Println("   Uploaded again successfully.")
}

func upload(client filepb.FileServiceClient, content []byte, filename string) {
	stream, err := client.Upload(context.Background())
	if err != nil {
		log.Fatalf("Failed to start upload: %v", err)
	}

	err = stream.Send(&filepb.UploadRequest{
		Data: &filepb.UploadRequest_Metadata{
			Metadata: &filepb.FileMetadata{
				Filename: filename,
			},
		},
	})
	if err != nil {
		log.Fatalf("Failed to send metadata: %v", err)
	}

	err = stream.Send(&filepb.UploadRequest{
		Data: &filepb.UploadRequest_Chunk{
			Chunk: content,
		},
	})
	if err != nil {
		log.Fatalf("Failed to send chunk: %v", err)
	}

	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Failed to complete upload: %v", err)
	}
}
