package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	filepb "github.com/datapeice/astolfosplayer-backend/protos/gen/go/file"
	syncpb "github.com/datapeice/astolfosplayer-backend/protos/gen/go/sync"
)

func main() {
	// Connect to Sync Service
	syncConn, err := grpc.NewClient("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Sync: %v", err)
	}
	defer syncConn.Close()
	syncClient := syncpb.NewSyncServiceClient(syncConn)

	// Connect to File Service
	fileConn, err := grpc.NewClient("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to File: %v", err)
	}
	defer fileConn.Close()
	fileClient := filepb.NewFileServiceClient(fileConn)

	// Get all hashes
	resp, err := syncClient.GetSync(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Fatalf("Failed to get sync: %v", err)
	}

	existingHashes := map[string]bool{
		"1c2e13c44278e45b29aec2e49c8e4d51b5b113fdedb816293d6a25f9fb2b3607": true,
		"220a7cc4fa56a6a7d0859d97073032e4b943c7f44e8d1f9ad4e129dc926ae4a5": true,
	}

	for _, hash := range resp.Hashes {
		if !existingHashes[hash] {
			fmt.Printf("Deleting missing hash: %s\n", hash)
			_, err := fileClient.Delete(context.Background(), &filepb.DeleteRequest{Hash: hash})
			if err != nil {
				log.Printf("Failed to delete %s: %v\n", hash, err)
			} else {
				fmt.Println("Deleted successfully.")
			}
		} else {
			fmt.Printf("Keeping valid hash: %s\n", hash)
		}
	}
}
