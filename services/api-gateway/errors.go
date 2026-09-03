package main

import (
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// tailorGRPCError translates raw gRPC and internal backend errors into clean,
// user-friendly messages and appropriate HTTP status codes for the client.
func tailorGRPCError(err error) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}

	st, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError, "An unexpected error occurred. Please try again."
	}

	rawMsg := st.Message()

	// Default HTTP status based on gRPC status code
	httpStatus := http.StatusInternalServerError
	switch st.Code() {
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
	case codes.ResourceExhausted:
		httpStatus = http.StatusTooManyRequests
	case codes.DeadlineExceeded:
		httpStatus = http.StatusGatewayTimeout
	case codes.Unavailable:
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusInternalServerError
	}

	// Strip common RPC prefix strings added by backend services
	cleanMsg := rawMsg
	prefixes := []string{
		"login failed: ",
		"registration failed: ",
		"OTP verification failed: ",
		"failed to resend OTP: ",
		"token refresh failed: ",
		"logout failed: ",
		"failed to get user: ",
		"failed to get user by email: ",
		"failed to call trip preview: ",
		"failed to call trip start: ",
		"failed to call trip cancel: ",
	}
	for _, p := range prefixes {
		cleanMsg = strings.TrimPrefix(cleanMsg, p)
	}

	lower := strings.ToLower(cleanMsg)

	// Tailor specific error messages to friendly client text
	switch {
	case strings.Contains(lower, "invalid email or password"):
		return http.StatusUnauthorized, "Invalid email or password. Please check your credentials."
	case strings.Contains(lower, "email not verified"):
		return http.StatusForbidden, "Email not verified. Please verify your OTP code before signing in."
	case strings.Contains(lower, "user with this email already exists") || (strings.Contains(lower, "duplicate key") && strings.Contains(lower, "email")):
		return http.StatusConflict, "An account with this email address already exists."
	case strings.Contains(lower, "user with this phone number already exists") || (strings.Contains(lower, "duplicate key") && strings.Contains(lower, "phone")):
		return http.StatusConflict, "An account with this phone number already exists."
	case strings.Contains(lower, "invalid otp") || strings.Contains(lower, "invalid code") || strings.Contains(lower, "invalid verification code"):
		return http.StatusBadRequest, "Invalid verification code. Please check and try again."
	case strings.Contains(lower, "expired"):
		return http.StatusBadRequest, "Verification code has expired. Please request a new one."
	case strings.Contains(lower, "already been used") || strings.Contains(lower, "already used"):
		return http.StatusBadRequest, "This verification code has already been used. Please request a new one."
	case strings.Contains(lower, "invalid or expired refresh token") || strings.Contains(lower, "stolen token"):
		return http.StatusUnauthorized, "Your session has expired. Please sign in again."
	case strings.Contains(lower, "not found"):
		return http.StatusNotFound, "Requested resource or user was not found."
	case strings.Contains(lower, "connection refused") || strings.Contains(lower, "unavailable"):
		return http.StatusServiceUnavailable, "Service is temporarily unavailable. Please try again shortly."
	case strings.Contains(lower, "pq:") || strings.Contains(lower, "sql") || strings.Contains(lower, "relation"):
		// Protect backend database details from leaking to the client
		return http.StatusInternalServerError, "A database error occurred. Please try again later."
	default:
		// Capitalize first character of clean string
		if len(cleanMsg) > 0 {
			cleanMsg = strings.ToUpper(cleanMsg[:1]) + cleanMsg[1:]
			return httpStatus, cleanMsg
		}
		return httpStatus, "Unable to complete request. Please try again."
	}
}
