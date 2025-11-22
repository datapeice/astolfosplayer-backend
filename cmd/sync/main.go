package main

import (
	"fmt"
	"log"
	"net"

	"github.com/datapeice/astolfosplayer-backend/internal/config"
	"github.com/datapeice/astolfosplayer-backend/internal/db"
	"github.com/datapeice/astolfosplayer-backend/internal/sync"
	pb "github.com/datapeice/astolfosplayer-backend/protos/gen/go/sync"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadSyncConfig()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterSyncServiceServer(s, &sync.Server{
		DB:     database,
		Config: cfg,
	})

	log.Printf("Sync Service listening on :%s", cfg.Port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
