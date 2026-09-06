package main

import (
	"encoding/json"
	"net/http"

	"github.com/trunglq04/goride/services/api-gateway/grpc_clients"
	"github.com/trunglq04/goride/shared/auth"
	"github.com/trunglq04/goride/shared/contracts"
	"github.com/trunglq04/goride/shared/logger"
	pb "github.com/trunglq04/goride/shared/proto/auth"
	"github.com/trunglq04/goride/shared/util"

	"google.golang.org/grpc/metadata"
)

// ---- Request/Response types ----

type registerRequest struct {
	FullName string `json:"fullName"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type verifyOTPRequest struct {
	UserID string `json:"userID"`
	Code   string `json:"code"`
}

type resendOTPRequest struct {
	UserID string `json:"userID"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type logoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// ---- HTTP Handlers ----

func handleRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.L()

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WarnContext(ctx, "Failed to parse register request", "err", err)
		util.WriteError(w, http.StatusBadRequest, "Please fill in all required registration fields.")
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		util.WriteError(w, http.StatusBadGateway, "Unable to connect to auth service. Please try again.")
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
		util.WriteError(w, status, msg)
		return
	}

	log.InfoContext(ctx, "User registered", "user_id", resp.UserId, "email", req.Email)
	util.WriteJSON(w, http.StatusCreated, contracts.APIResponse{Data: resp})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.L()

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WarnContext(ctx, "Failed to parse login request", "err", err)
		util.WriteError(w, http.StatusBadRequest, "Please provide both email and password.")
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		util.WriteError(w, http.StatusBadGateway, "Unable to connect to auth service. Please try again.")
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
		util.WriteError(w, status, msg)
		return
	}

	log.InfoContext(ctx, "User logged in", "email", req.Email)
	util.WriteJSON(w, http.StatusOK, contracts.APIResponse{Data: resp})
}

func handleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.L()

	var req verifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WarnContext(ctx, "Failed to parse verify OTP request", "err", err)
		util.WriteError(w, http.StatusBadRequest, "Please enter the 6-digit verification code.")
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		util.WriteError(w, http.StatusBadGateway, "Unable to connect to auth service. Please try again.")
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
		util.WriteError(w, status, msg)
		return
	}

	log.InfoContext(ctx, "OTP verified", "user_id", req.UserID)
	util.WriteJSON(w, http.StatusOK, contracts.APIResponse{Data: resp})
}

func handleResendOTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.L()

	var req resendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WarnContext(ctx, "Failed to parse resend OTP request", "err", err)
		util.WriteError(w, http.StatusBadRequest, "Invalid request parameters.")
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		util.WriteError(w, http.StatusBadGateway, "Unable to connect to auth service. Please try again.")
		return
	}
	defer authService.Close()

	resp, err := authService.Client.ResendOTP(ctx, &pb.ResendOTPRequest{
		UserId: req.UserID,
	})
	if err != nil {
		log.ErrorContext(ctx, "Resend OTP failed", "user_id", req.UserID, "err", err)
		status, msg := tailorGRPCError(err)
		util.WriteError(w, status, msg)
		return
	}

	util.WriteJSON(w, http.StatusOK, contracts.APIResponse{Data: resp})
}

func handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.L()

	var req refreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WarnContext(ctx, "Failed to parse refresh token request", "err", err)
		util.WriteError(w, http.StatusBadRequest, "Refresh token is required.")
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		util.WriteError(w, http.StatusBadGateway, "Unable to connect to auth service. Please try again.")
		return
	}
	defer authService.Close()

	resp, err := authService.Client.RefreshToken(ctx, &pb.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		log.ErrorContext(ctx, "Token refresh failed", "err", err)
		status, msg := tailorGRPCError(err)
		util.WriteError(w, status, msg)
		return
	}

	util.WriteJSON(w, http.StatusOK, contracts.APIResponse{Data: resp})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.L()

	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.WarnContext(ctx, "Failed to parse logout request", "err", err)
		util.WriteError(w, http.StatusBadRequest, "Refresh token is required for logout.")
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		util.WriteError(w, http.StatusBadGateway, "Unable to connect to auth service. Please try again.")
		return
	}
	defer authService.Close()

	resp, err := authService.Client.Logout(ctx, &pb.LogoutRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		log.ErrorContext(ctx, "Logout failed", "err", err)
		status, msg := tailorGRPCError(err)
		util.WriteError(w, status, msg)
		return
	}

	util.WriteJSON(w, http.StatusOK, contracts.APIResponse{Data: resp})
}

func handleGetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.L()

	// Get user ID from JWT middleware context
	userID, ok := auth.GetUserIDFromContext(r)
	if !ok {
		util.WriteError(w, http.StatusUnauthorized, "Authentication required. Please sign in.")
		return
	}

	authService, err := grpc_clients.NewAuthServiceClient()
	if err != nil {
		log.ErrorContext(ctx, "Failed to create auth service client", "err", err)
		util.WriteError(w, http.StatusBadGateway, "Unable to connect to auth service. Please try again.")
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
		util.WriteError(w, status, msg)
		return
	}

	util.WriteJSON(w, http.StatusOK, contracts.APIResponse{Data: resp})
}
