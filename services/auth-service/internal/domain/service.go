package domain

import (
	"context"
)

type AuthService interface {
	Register(ctx context.Context, fullName, email, phone, password, roleName string) (string, error)
	VerifyOTP(ctx context.Context, userID, code string) (string, string, *User, error)
	ResendOTP(ctx context.Context, userID string) error
	Login(ctx context.Context, email, password string) (string, string, *User, error)
	RefreshToken(ctx context.Context, rawRefreshToken string) (string, string, error)
	Logout(ctx context.Context, rawRefreshToken string) error
	GetMe(ctx context.Context, userID string) (*User, error)
}
