package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trunglq04/goride/services/api-gateway/grpc_clients"
	"github.com/trunglq04/goride/shared/auth"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/logger"
	pb "github.com/trunglq04/goride/shared/proto/auth"

	"google.golang.org/grpc/metadata"
)

// ---- Request/Response types ----

type registerRequest struct {
	FullName string `json:"fullName" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type verifyOTPRequest struct {
	UserID string `json:"userID" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

type resendOTPRequest struct {
	UserID string `json:"userID" binding:"required"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// ---- HTTP Handlers ----

func handleRegister(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.L()

	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WarnContext(ctx, "Failed to parse register request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please fill in all required registration fields."})
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Unable to connect to auth service. Please try again."})
		return
	}
	defer authService.Close()

	resp, err := authService.Client.Register(ctx, &pb.RegisterRequest{
		FullName: req.FullName,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		log.ErrorContext(ctx, "Registration failed", "email", req.Email, "err", err)
		status, msg := tailorGRPCError(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	log.InfoContext(ctx, "User registered", "user_id", resp.UserId, "email", req.Email)
	c.JSON(http.StatusCreated, contracts.APIResponse{Data: resp})
}

func handleLogin(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.L()

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WarnContext(ctx, "Failed to parse login request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide both email and password."})
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Unable to connect to auth service. Please try again."})
		return
	}
	defer authService.Close()

	resp, err := authService.Client.Login(ctx, &pb.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		log.ErrorContext(ctx, "Login failed", "email", req.Email, "err", err)
		status, msg := tailorGRPCError(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	log.InfoContext(ctx, "User logged in", "email", req.Email)
	c.JSON(http.StatusOK, contracts.APIResponse{Data: resp})
}

func handleVerifyOTP(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.L()

	var req verifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WarnContext(ctx, "Failed to parse verify OTP request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please enter the 6-digit verification code."})
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Unable to connect to auth service. Please try again."})
		return
	}
	defer authService.Close()

	resp, err := authService.Client.VerifyOTP(ctx, &pb.VerifyOTPRequest{
		UserId: req.UserID,
		Code:   req.Code,
	})
	if err != nil {
		log.ErrorContext(ctx, "OTP verification failed", "user_id", req.UserID, "err", err)
		status, msg := tailorGRPCError(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	log.InfoContext(ctx, "OTP verified", "user_id", req.UserID)
	c.JSON(http.StatusOK, contracts.APIResponse{Data: resp})
}

func handleResendOTP(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.L()

	var req resendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WarnContext(ctx, "Failed to parse resend OTP request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters."})
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Unable to connect to auth service. Please try again."})
		return
	}
	defer authService.Close()

	resp, err := authService.Client.ResendOTP(ctx, &pb.ResendOTPRequest{
		UserId: req.UserID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Resend OTP failed", "user_id", req.UserID, "err", err)
		status, msg := tailorGRPCError(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, contracts.APIResponse{Data: resp})
}

func handleRefreshToken(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.L()

	var req refreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WarnContext(ctx, "Failed to parse refresh token request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required."})
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Unable to connect to auth service. Please try again."})
		return
	}
	defer authService.Close()

	resp, err := authService.Client.RefreshToken(ctx, &pb.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		log.ErrorContext(ctx, "Token refresh failed", "err", err)
		status, msg := tailorGRPCError(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, contracts.APIResponse{Data: resp})
}

func handleLogout(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.L()

	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WarnContext(ctx, "Failed to parse logout request", "err", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token is required for logout."})
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Unable to connect to auth service. Please try again."})
		return
	}
	defer authService.Close()

	resp, err := authService.Client.Logout(ctx, &pb.LogoutRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		log.ErrorContext(ctx, "Logout failed", "err", err)
		status, msg := tailorGRPCError(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, contracts.APIResponse{Data: resp})
}

func handleGetMe(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.L()

	// Get user ID from JWT middleware context
	userID, ok := auth.GetUserIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required. Please sign in."})
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Unable to connect to auth service. Please try again."})
		return
	}
	defer authService.Close()

	// Forward user ID via gRPC metadata
	md := metadata.Pairs("x-user-id", userID)
	grpcCtx := metadata.NewOutgoingContext(ctx, md)

	resp, err := authService.Client.GetMe(grpcCtx, &pb.GetMeRequest{})
	if err != nil {
		log.ErrorContext(ctx, "GetMe failed", "user_id", userID, "err", err)
		status, msg := tailorGRPCError(err)
		c.JSON(status, gin.H{"error": msg})
		return
	}

	c.JSON(http.StatusOK, contracts.APIResponse{Data: resp})
}
