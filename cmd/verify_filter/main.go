package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	syncpb "github.com/datapeice/astolfosplayer-backend/protos/gen/go/sync"
)

func main() {
	conn, err := grpc.NewClient("localhost:50053", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Sync: %v", err)
	}
	defer conn.Close()
	client := syncpb.NewSyncServiceClient(conn)

	resp, err := client.GetSync(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Fatalf("Failed to get sync: %v", err)
	}

	foundTxt := false
	for _, file := range resp.Files {
		fmt.Printf("File: %s\n", file.Filename)
		if strings.HasSuffix(file.Filename, ".txt") {
			foundTxt = true
		}
	}

	if foundTxt {
		log.Fatalf("❌ Found .txt file in sync response!")
	} else {
		fmt.Println("✅ No .txt files found.")
	}
}
