package domain

import "context"

type Repository interface {
	// Transaction
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error

	// Roles
	GetRoleByName(ctx context.Context, name string) (*Role, error)

	// Users
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	ConfirmUserEmail(ctx context.Context, userID string) error

	// Email OTPs
	CreateEmailOTP(ctx context.Context, otp *EmailOTP) error
	GetLatestOTPByUserID(ctx context.Context, userID string) (*EmailOTP, error)
	MarkOTPUsed(ctx context.Context, otpID string) error

	// Refresh Tokens
	CreateRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeTokenFamily(ctx context.Context, familyID string) error
	MarkTokenReplaced(ctx context.Context, tokenID string, replacedByID string) error
}
