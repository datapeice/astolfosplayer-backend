package auth

import (
	"context"
	"errors"

	"github.com/datapeice/astolfosplayer-backend/internal/config"
	pb "github.com/datapeice/astolfosplayer-backend/protos/gen/go/auth"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Server struct {
	pb.UnimplementedAuthServiceServer
	DB     *gorm.DB
	Config *config.Config
}

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.SecurityKey != s.Config.SecurityKey {
		return nil, status.Errorf(codes.PermissionDenied, "invalid security key")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password")
	}

	user := User{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	if err := s.DB.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, status.Errorf(codes.AlreadyExists, "username already taken")
		}
		return nil, status.Errorf(codes.Internal, "failed to create user")
	}

	token, err := GenerateToken(user.Username, s.Config.SecretKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	return &pb.RegisterResponse{Token: token}, nil
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var user User
	if err := s.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Errorf(codes.Internal, "database error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}

	token, err := GenerateToken(user.Username, s.Config.SecretKey)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token")
	}

	return &pb.LoginResponse{Token: token}, nil
}
