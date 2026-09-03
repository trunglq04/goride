package grpc

import (
	"context"
	"log/slog"

	"github.com/trunglq04/goride/services/auth-service/internal/domain"
	pb "github.com/trunglq04/goride/shared/proto/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCHandler handles gRPC requests for the auth service.
type GRPCHandler struct {
	pb.UnimplementedAuthServiceServer

	service domain.AuthService
}

// NewGRPCHandler registers the auth gRPC handler on the given server.
func NewGRPCHandler(server *grpc.Server, svc domain.AuthService) {
	handler := &GRPCHandler{service: svc}
	pb.RegisterAuthServiceServer(server, handler)
}

func (h *GRPCHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.GetFullName() == "" || req.GetEmail() == "" || req.GetPhone() == "" || req.GetPassword() == "" || req.GetRole() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "full_name, email, phone, password, and role are required")
	}

	if len(req.GetPassword()) < 8 {
		return nil, status.Errorf(codes.InvalidArgument, "password must be at least 8 characters")
	}

	userID, err := h.service.Register(ctx, req.GetFullName(), req.GetEmail(), req.GetPhone(), req.GetPassword(), req.GetRole())
	if err != nil {
		slog.ErrorContext(ctx, "Registration failed",
			"email", req.GetEmail(),
			"phone", req.GetPhone(),
			"role", req.GetRole(),
			"err", err,
		)
		return nil, status.Errorf(codes.Internal, "registration failed: %v", err)
	}

	return &pb.RegisterResponse{
		UserId:  userID,
		Message: "OTP sent to email",
	}, nil
}

func (h *GRPCHandler) VerifyOTP(ctx context.Context, req *pb.VerifyOTPRequest) (*pb.VerifyOTPResponse, error) {
	if req.GetUserId() == "" || req.GetCode() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id and code are required")
	}

	accessToken, refreshToken, user, err := h.service.VerifyOTP(ctx, req.GetUserId(), req.GetCode())
	if err != nil {
		slog.ErrorContext(ctx, "OTP verification failed",
			"user_id", req.GetUserId(),
			"err", err,
		)
		return nil, status.Errorf(codes.Unauthenticated, "OTP verification failed: %v", err)
	}

	return &pb.VerifyOTPResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         domainUserToProto(user),
	}, nil
}

func (h *GRPCHandler) ResendOTP(ctx context.Context, req *pb.ResendOTPRequest) (*pb.ResendOTPResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "user_id is required")
	}

	if err := h.service.ResendOTP(ctx, req.GetUserId()); err != nil {
		slog.ErrorContext(ctx, "Resend OTP failed",
			"user_id", req.GetUserId(),
			"err", err,
		)
		return nil, status.Errorf(codes.Internal, "failed to resend OTP: %v", err)
	}

	return &pb.ResendOTPResponse{
		Message: "OTP resent to email",
	}, nil
}

func (h *GRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "email and password are required")
	}

	accessToken, refreshToken, user, err := h.service.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		slog.ErrorContext(ctx, "Login failed",
			"email", req.GetEmail(),
			"err", err,
		)
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         domainUserToProto(user),
	}, nil
}

func (h *GRPCHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh_token is required")
	}

	accessToken, refreshToken, err := h.service.RefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		slog.ErrorContext(ctx, "Token refresh failed", "err", err)
		return nil, status.Errorf(codes.Unauthenticated, "token refresh failed: %v", err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *GRPCHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	if req.GetRefreshToken() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "refresh_token is required")
	}

	if err := h.service.Logout(ctx, req.GetRefreshToken()); err != nil {
		slog.ErrorContext(ctx, "Logout failed", "err", err)
		return nil, status.Errorf(codes.Internal, "logout failed: %v", err)
	}

	return &pb.LogoutResponse{
		Message: "logged out successfully",
	}, nil
}

func (h *GRPCHandler) GetMe(ctx context.Context, _ *pb.GetMeRequest) (*pb.GetMeResponse, error) {
	// Extract user_id from gRPC metadata (set by API gateway)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	userIDs := md.Get("x-user-id")
	if len(userIDs) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing user ID in metadata")
	}

	user, err := h.service.GetMe(ctx, userIDs[0])
	if err != nil {
		slog.ErrorContext(ctx, "GetMe failed",
			"user_id", userIDs[0],
			"err", err,
		)
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	return &pb.GetMeResponse{
		User: domainUserToProto(user),
	}, nil
}

// domainUserToProto converts a domain User to protobuf User.
func domainUserToProto(u *domain.User) *pb.User {
	if u == nil {
		return nil
	}
	return &pb.User{
		Id:             u.ID,
		FullName:       u.FullName,
		Email:          u.Email,
		Phone:          u.Phone,
		Role:           u.RoleName,
		EmailConfirmed: u.EmailConfirmed,
		CreatedAt:      u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
