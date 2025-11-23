package main

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	endpoint := "localhost:9000"
	accessKeyID := "minioadmin"
	secretAccessKey := "minioadmin"
	useSSL := false

	// Initialize MinIO client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	bucketName := "music"

	fmt.Println("Checking MinIO bucket:", bucketName)

	exists, err := minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		log.Println("Error checking bucket:", err)
	}
	if !exists {
		log.Println("Bucket does not exist! Creating it...")
		err = minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalln("Failed to create bucket:", err)
		}
		log.Println("Bucket created.")
	} else {
		log.Println("Bucket exists.")
	}

	ctx := context.Background()
	objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{})

	fmt.Println("\nObjects in MinIO:")
	for object := range objectCh {
		if object.Err != nil {
			log.Println(object.Err)
			return
		}
		fmt.Printf("- %s (%d bytes)\n", object.Key, object.Size)
	}
}
