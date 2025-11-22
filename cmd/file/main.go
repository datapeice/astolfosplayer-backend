package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/datapeice/astolfosplayer-backend/internal/config"
	"github.com/datapeice/astolfosplayer-backend/internal/db"
	"github.com/datapeice/astolfosplayer-backend/internal/file"
	pb "github.com/datapeice/astolfosplayer-backend/protos/gen/go/file"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadFileConfig()

	// Connect to DB
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate
	if err := database.AutoMigrate(&file.Track{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Connect to MinIO
	minioClient, err := file.NewMinioClient(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3UseSSL)
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
	}

	// Ensure bucket exists
	if err := file.EnsureBucket(context.Background(), minioClient, cfg.S3Bucket); err != nil {
		log.Fatalf("Failed to ensure bucket exists: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterFileServiceServer(s, &file.Server{
		MinioClient: minioClient,
		DB:          database,
		Config:      cfg,
	})

	log.Printf("File Service listening on :%s", cfg.Port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
