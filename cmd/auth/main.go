package main

import (
	"fmt"
	"log"
	"net"

	"github.com/datapeice/astolfosplayer-backend/internal/auth"
	"github.com/datapeice/astolfosplayer-backend/internal/config"
	"github.com/datapeice/astolfosplayer-backend/internal/db"
	pb "github.com/datapeice/astolfosplayer-backend/protos/gen/go/auth"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadAuthConfig()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the schema
	if err := database.AutoMigrate(&auth.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &auth.Server{
		DB:     database,
		Config: cfg,
	})

	log.Printf("Auth Service listening on :%s", cfg.Port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
