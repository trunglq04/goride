package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/trunglq04/goride/services/auth-service/internal/domain"
	"github.com/trunglq04/goride/shared/auth"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 12
	otpLength  = 6
	otpExpiry  = 10 * time.Minute
)

// EventPublisher defines the interface for publishing auth events.
type EventPublisher interface {
	PublishUserRegistered(ctx context.Context, userID, email, otp string) error
	PublishOTPResent(ctx context.Context, userID, email, otp string) error
}

// AuthService implements the core authentication business logic.
type authService struct {
	repo       domain.Repository
	publisher  EventPublisher
	privateKey *rsa.PrivateKey
}

// NewAuthService creates a new auth service instance.
func NewAuthService(repo domain.Repository, publisher EventPublisher, privateKey *rsa.PrivateKey) *authService {
	return &authService{
		repo:       repo,
		publisher:  publisher,
		privateKey: privateKey,
	}
}

// Register creates a new user account and sends an OTP for email verification.
// Uses a database transaction (TX) to atomically insert the user and initial OTP.
func (s *authService) Register(ctx context.Context, fullName, email, phone, password, roleName string) (string, error) {
	// Validate role
	if roleName != "rider" && roleName != "driver" {
		return "", fmt.Errorf("invalid role: %s (must be 'rider' or 'driver')", roleName)
	}

	// Check if email already exists
	existingEmail, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("failed to check existing user email: %w", err)
	}
	if existingEmail != nil {
		return "", fmt.Errorf("email already registered")
	}

	// Check if phone already exists
	existingPhone, err := s.repo.GetUserByPhone(ctx, phone)
	if err != nil {
		return "", fmt.Errorf("failed to check existing user phone: %w", err)
	}
	if existingPhone != nil {
		return "", fmt.Errorf("phone already registered")
	}

	// Get role
	role, err := s.repo.GetRoleByName(ctx, roleName)
	if err != nil {
		return "", fmt.Errorf("failed to get role: %w", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate OTP
	otpCode, err := generateOTPCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	now := time.Now()
	user := &domain.User{
		ID:             uuid.New().String(),
		FullName:       fullName,
		Email:          email,
		Phone:          phone,
		PasswordHash:   string(hashedPassword),
		RoleID:         role.ID,
		RoleName:       roleName,
		EmailConfirmed: false,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	otp := &domain.EmailOTP{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Code:      otpCode,
		ExpiresAt: now.Add(otpExpiry),
		Used:      false,
		CreatedAt: now,
	}

	// Execute User and EmailOTP creation within a single atomic database transaction
	err = s.repo.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.CreateUser(txCtx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
		if err := s.repo.CreateEmailOTP(txCtx, otp); err != nil {
			return fmt.Errorf("failed to create email OTP: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	// Publish event for notification service to send email (outside TX)
	if err := s.publisher.PublishUserRegistered(ctx, user.ID, email, otpCode); err != nil {
		slog.ErrorContext(ctx, "Failed to publish user registered event (email may not be sent)",
			"user_id", user.ID,
			"err", err,
		)
	}

	slog.InfoContext(ctx, "User registered successfully",
		"user_id", user.ID,
		"email", email,
		"phone", phone,
		"role", roleName,
	)

	return user.ID, nil
}

// VerifyOTP validates the OTP code, confirms the user's email, and issues tokens.
// Uses a database transaction (TX) to atomically mark the OTP used, confirm email, and save refresh token.
func (s *authService) VerifyOTP(ctx context.Context, userID, code string) (string, string, *domain.User, error) {
	// Get the latest OTP for this user
	otp, err := s.repo.GetLatestOTPByUserID(ctx, userID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get OTP: %w", err)
	}
	if otp == nil {
		return "", "", nil, fmt.Errorf("no OTP found for user")
	}

	// Validate OTP
	if !otp.IsValid(code) {
		if otp.Used {
			return "", "", nil, fmt.Errorf("OTP has already been used")
		}
		if otp.IsExpired() {
			return "", "", nil, fmt.Errorf("OTP has expired")
		}
		return "", "", nil, fmt.Errorf("invalid OTP code")
	}

	// Prepare refresh token model
	rawRefreshToken := uuid.New().String()
	tokenHash := hashToken(rawRefreshToken)
	now := time.Now()

	refreshToken := &domain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		TokenHash: tokenHash,
		FamilyID:  uuid.New().String(), // new family for each login/verification
		Revoked:   false,
		ExpiresAt: now.Add(auth.RefreshTokenDuration),
		CreatedAt: now,
	}

	// Atomic transaction: mark OTP used, confirm user email, and record refresh token
	err = s.repo.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.MarkOTPUsed(txCtx, otp.ID); err != nil {
			return fmt.Errorf("failed to mark OTP as used: %w", err)
		}
		if err := s.repo.ConfirmUserEmail(txCtx, userID); err != nil {
			return fmt.Errorf("failed to confirm email: %w", err)
		}
		if err := s.repo.CreateRefreshToken(txCtx, refreshToken); err != nil {
			return fmt.Errorf("failed to create refresh token: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", "", nil, err
	}

	// Get updated user profile
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", "", nil, fmt.Errorf("user not found")
	}

	// Sign access token
	accessToken, err := auth.SignAccessToken(user.ID, user.Email, user.Phone, user.RoleName, s.privateKey)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	return accessToken, rawRefreshToken, user, nil
}

// ResendOTP generates a new OTP and publishes an event to resend it.
func (s *authService) ResendOTP(ctx context.Context, userID string) error {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	if user.EmailConfirmed {
		return fmt.Errorf("email already confirmed")
	}

	otp, err := s.generateAndStoreOTP(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	if err := s.publisher.PublishOTPResent(ctx, userID, user.Email, otp); err != nil {
		return fmt.Errorf("failed to publish OTP resent event: %w", err)
	}

	return nil
}

// Login authenticates a user with email and password and issues tokens.
func (s *authService) Login(ctx context.Context, email, password string) (string, string, *domain.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", "", nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", nil, fmt.Errorf("invalid email or password")
	}

	// Check email confirmation
	if !user.EmailConfirmed {
		return "", "", nil, fmt.Errorf("email not confirmed")
	}

	// Issue tokens
	accessToken, refreshToken, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return "", "", nil, err
	}

	slog.InfoContext(ctx, "User logged in",
		"user_id", user.ID,
		"email", user.Email,
		"role", user.RoleName,
	)

	return accessToken, refreshToken, user, nil
}

// RefreshToken validates a refresh token, rotates it, and issues new tokens.
// Uses a database transaction (TX) to atomically store the new token and mark the old token replaced.
// Implements reuse detection: if a revoked/replaced token is used, the entire
// token family is revoked (stolen token detection).
func (s *authService) RefreshToken(ctx context.Context, rawRefreshToken string) (string, string, error) {
	tokenHash := hashToken(rawRefreshToken)

	storedToken, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return "", "", fmt.Errorf("failed to get refresh token: %w", err)
	}
	if storedToken == nil {
		return "", "", fmt.Errorf("invalid refresh token")
	}

	// Check if the token was already used/replaced (reuse detection)
	if storedToken.Revoked || storedToken.ReplacedBy != nil {
		slog.WarnContext(ctx, "Refresh token reuse detected! Revoking entire family",
			"user_id", storedToken.UserID,
			"family_id", storedToken.FamilyID,
		)
		// Revoke the entire token family — potential token theft
		if err := s.repo.RevokeTokenFamily(ctx, storedToken.FamilyID); err != nil {
			slog.ErrorContext(ctx, "Failed to revoke token family", "err", err)
		}
		return "", "", fmt.Errorf("refresh token has been revoked (possible theft detected)")
	}

	// Check expiry
	if time.Now().After(storedToken.ExpiresAt) {
		return "", "", fmt.Errorf("refresh token has expired")
	}

	// Get user
	user, err := s.repo.GetUserByID(ctx, storedToken.UserID)
	if err != nil || user == nil {
		return "", "", fmt.Errorf("user not found")
	}

	// Generate new refresh token
	newRawToken := uuid.New().String()
	newTokenHash := hashToken(newRawToken)
	now := time.Now()

	newToken := &domain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    storedToken.UserID,
		TokenHash: newTokenHash,
		FamilyID:  storedToken.FamilyID, // same family
		Revoked:   false,
		ExpiresAt: now.Add(auth.RefreshTokenDuration),
		CreatedAt: now,
	}

	// Atomic transaction: create new refresh token and mark old token replaced_by
	err = s.repo.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.repo.CreateRefreshToken(txCtx, newToken); err != nil {
			return fmt.Errorf("failed to create new refresh token: %w", err)
		}
		if err := s.repo.MarkTokenReplaced(txCtx, storedToken.ID, newToken.ID); err != nil {
			return fmt.Errorf("failed to mark old token as replaced: %w", err)
		}
		return nil
	})
	if err != nil {
		return "", "", err
	}

	// Sign new access token
	accessToken, err := auth.SignAccessToken(user.ID, user.Email, user.Phone, user.RoleName, s.privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return accessToken, newRawToken, nil
}

// Logout revokes the refresh token family, effectively logging the user out.
func (s *authService) Logout(ctx context.Context, rawRefreshToken string) error {
	tokenHash := hashToken(rawRefreshToken)

	storedToken, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to get refresh token: %w", err)
	}
	if storedToken == nil {
		return nil // Token doesn't exist, consider it already logged out
	}

	if err := s.repo.RevokeTokenFamily(ctx, storedToken.FamilyID); err != nil {
		return fmt.Errorf("failed to revoke token family: %w", err)
	}

	slog.InfoContext(ctx, "User logged out",
		"user_id", storedToken.UserID,
		"family_id", storedToken.FamilyID,
	)

	return nil
}

// GetMe returns the current user's information.
func (s *authService) GetMe(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// ---- Internal helpers ----

// generateAndStoreOTP creates a random 6-digit code and stores it in the database.
func (s *authService) generateAndStoreOTP(ctx context.Context, userID string) (string, error) {
	code, err := generateOTPCode()
	if err != nil {
		return "", err
	}

	now := time.Now()
	otp := &domain.EmailOTP{
		ID:        uuid.New().String(),
		UserID:    userID,
		Code:      code,
		ExpiresAt: now.Add(otpExpiry),
		Used:      false,
		CreatedAt: now,
	}

	if err := s.repo.CreateEmailOTP(ctx, otp); err != nil {
		return "", err
	}

	return code, nil
}

// issueTokenPair creates a new access token and refresh token for the user.
func (s *authService) issueTokenPair(ctx context.Context, user *domain.User) (string, string, error) {
	// Sign access token
	accessToken, err := auth.SignAccessToken(user.ID, user.Email, user.Phone, user.RoleName, s.privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	rawRefreshToken := uuid.New().String()
	tokenHash := hashToken(rawRefreshToken)
	now := time.Now()

	refreshToken := &domain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		FamilyID:  uuid.New().String(), // new family for each login
		Revoked:   false,
		ExpiresAt: now.Add(auth.RefreshTokenDuration),
		CreatedAt: now,
	}

	if err := s.repo.CreateRefreshToken(ctx, refreshToken); err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, rawRefreshToken, nil
}

// generateOTPCode generates a cryptographically random 6-digit code.
func generateOTPCode() (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(otpLength)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", fmt.Errorf("failed to generate random OTP: %w", err)
	}
	// Zero-pad to otpLength digits
	return fmt.Sprintf("%0*d", otpLength, n), nil
}

// hashToken creates a SHA-256 hash of a raw token string for secure storage.
func hashToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}
