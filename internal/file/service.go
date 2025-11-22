package file

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"github.com/datapeice/astolfosplayer-backend/internal/config"
	pb "github.com/datapeice/astolfosplayer-backend/protos/gen/go/file"
	"github.com/minio/minio-go/v7"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Server struct {
	pb.UnimplementedFileServiceServer
	MinioClient *minio.Client
	DB          *gorm.DB
	Config      *config.FileConfig
}

func (s *Server) Upload(stream pb.FileService_UploadServer) error {
	tempFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	hasher := sha256.New()
	var metadata *pb.FileMetadata

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "failed to receive chunk: %v", err)
		}

		switch payload := req.Data.(type) {
		case *pb.UploadRequest_Metadata:
			metadata = payload.Metadata
		case *pb.UploadRequest_Chunk:
			if _, err := tempFile.Write(payload.Chunk); err != nil {
				return status.Errorf(codes.Internal, "failed to write to temp file: %v", err)
			}
			if _, err := hasher.Write(payload.Chunk); err != nil {
				return status.Errorf(codes.Internal, "failed to update hash: %v", err)
			}
		}
	}

	hash := hex.EncodeToString(hasher.Sum(nil))

	// Reset temp file pointer
	if _, err := tempFile.Seek(0, 0); err != nil {
		return status.Errorf(codes.Internal, "failed to seek temp file: %v", err)
	}

	// Upload to MinIO
	_, err = s.MinioClient.PutObject(context.Background(), s.Config.S3Bucket, hash, tempFile, -1, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to upload to S3: %v", err)
	}

	// Save metadata
	track := Track{
		Hash: hash,
	}
	if metadata != nil {
		track.Filename = metadata.Filename
		track.Title = metadata.Title
		track.Artist = metadata.Artist
		track.Album = metadata.Album
		track.Duration = metadata.Duration
	}

	// Upsert metadata
	if err := s.DB.Where(Track{Hash: hash}).Assign(track).FirstOrCreate(&track).Error; err != nil {
		return status.Errorf(codes.Internal, "failed to save metadata: %v", err)
	}

	return stream.SendAndClose(&pb.UploadResponse{Hash: hash})
}

func (s *Server) Download(req *pb.DownloadRequest, stream pb.FileService_DownloadServer) error {
	object, err := s.MinioClient.GetObject(context.Background(), s.Config.S3Bucket, req.Hash, minio.GetObjectOptions{})
	if err != nil {
		return status.Errorf(codes.NotFound, "file not found: %v", err)
	}
	defer object.Close()

	buffer := make([]byte, 64*1024) // 64KB chunks
	for {
		n, err := object.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to read from S3: %v", err)
		}

		if err := stream.Send(&pb.DownloadResponse{Chunk: buffer[:n]}); err != nil {
			return status.Errorf(codes.Unknown, "failed to send chunk: %v", err)
		}
	}

	return nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	// Delete from MinIO
	err := s.MinioClient.RemoveObject(context.Background(), s.Config.S3Bucket, req.Hash, minio.RemoveObjectOptions{})
	if err != nil {
		return &pb.DeleteResponse{Success: false}, status.Errorf(codes.Internal, "failed to delete from S3: %v", err)
	}

	// Delete from DB
	result := s.DB.Where("hash = ?", req.Hash).Delete(&Track{})
	if result.Error != nil {
		return &pb.DeleteResponse{Success: false}, status.Errorf(codes.Internal, "failed to delete metadata: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		// Even if not in DB, if deleted from S3 (or didn't exist), we might consider it success?
		// But user asked for status.
		// If MinIO delete didn't error, it's "success" even if file didn't exist (MinIO behavior).
		// But if DB didn't have it, maybe it wasn't there.
		// Let's return true if no error occurred.
	}

	return &pb.DeleteResponse{Success: true}, nil
}
