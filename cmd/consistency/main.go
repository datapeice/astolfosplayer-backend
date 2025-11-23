package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/datapeice/astolfosplayer-backend/internal/config"
	"github.com/datapeice/astolfosplayer-backend/internal/db"
	"github.com/datapeice/astolfosplayer-backend/internal/file"
	"github.com/minio/minio-go/v7"
)

func main() {
	fix := flag.Bool("fix", false, "Delete metadata for missing files")
	flag.Parse()

	cfg := config.LoadFileConfig()

	// Connect to DB
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Connect to MinIO
	minioClient, err := file.NewMinioClient(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3UseSSL)
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
	}

	var tracks []file.Track
	if err := database.Find(&tracks).Error; err != nil {
		log.Fatalf("Failed to fetch tracks: %v", err)
	}

	fmt.Printf("Checking %d tracks...\n", len(tracks))

	missingCount := 0
	for _, track := range tracks {
		_, err := minioClient.StatObject(context.Background(), cfg.S3Bucket, track.Hash, minio.StatObjectOptions{})
		if err != nil {
			errResponse := minio.ToErrorResponse(err)
			if errResponse.Code == "NoSuchKey" {
				fmt.Printf("❌ Missing file for hash: %s (%s)\n", track.Hash, track.Filename)
				missingCount++

				if *fix {
					fmt.Printf("   Deleting from DB...\n")
					if err := database.Delete(&track).Error; err != nil {
						log.Printf("   Failed to delete: %v\n", err)
					} else {
						fmt.Printf("   Deleted.\n")
					}
				}
			} else {
				log.Printf("⚠️ Error checking %s: %v\n", track.Hash, err)
			}
		} else {
			// fmt.Printf("✅ Found: %s\n", track.Hash)
		}
	}

	fmt.Printf("Done. Found %d missing files.\n", missingCount)
	if missingCount > 0 && !*fix {
		fmt.Println("Run with -fix to delete missing entries from DB.")
	}
}
