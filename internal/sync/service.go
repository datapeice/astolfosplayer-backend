package sync

import (
	"context"

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
	var hashes []string

	// We need to access the tracks table. Since we are in a separate microservice,
	// we share the database schema/models. Ideally, models should be in a shared package.
	// For now, we import the model from internal/file since they share the same DB (sqlite_data volume).

	// Query only hashes
	if err := s.DB.Model(&file.Track{}).Pluck("hash", &hashes).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch hashes: %v", err)
	}

	return &pb.GetSyncResponse{Hashes: hashes}, nil
}
