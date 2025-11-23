package sync

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/datapeice/astolfosplayer-backend/internal/config"
	"github.com/datapeice/astolfosplayer-backend/internal/file"
	pb "github.com/datapeice/astolfosplayer-backend/protos/gen/go/sync"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type Server struct {
	pb.UnimplementedSyncServiceServer
	DB     *gorm.DB
	Config *config.SyncConfig
}

func (s *Server) GetSync(ctx context.Context, req *emptypb.Empty) (*pb.GetSyncResponse, error) {
	var tracks []file.Track

	// We need to access the tracks table. Since we are in a separate microservice,
	// we share the database schema/models. Ideally, models should be in a shared package.
	// For now, we import the model from internal/file since they share the same DB (sqlite_data volume).

	// Query all tracks
	if err := s.DB.Find(&tracks).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch tracks: %v", err)
	}

	supportedExtensions := map[string]bool{
		".mp3":  true,
		".flac": true,
		".wav":  true,
		".ogg":  true,
		".m4a":  true,
	}

	var files []*pb.FileInfo
	for _, t := range tracks {
		ext := strings.ToLower(filepath.Ext(t.Filename))
		if supportedExtensions[ext] {
			files = append(files, &pb.FileInfo{
				Hash:     t.Hash,
				Filename: t.Filename,
			})
		}
	}

	return &pb.GetSyncResponse{Files: files}, nil
}
