package domain

import "time"

// User represents a registered user in the auth system.
type User struct {
	ID             string    `json:"id"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	PasswordHash   string    `json:"-"`
	RoleID         string    `json:"role_id"`
	RoleName       string    `json:"role"`
	EmailConfirmed bool      `json:"email_confirmed"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Role represents an actor type in the system.
type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// EmailOTP represents a 6-digit verification code sent to a user's email.
type EmailOTP struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

// RefreshToken represents a stored refresh token with rotation support.
type RefreshToken struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	TokenHash  string    `json:"token_hash"`
	FamilyID   string    `json:"family_id"`
	ReplacedBy *string   `json:"replaced_by,omitempty"`
	Revoked    bool      `json:"revoked"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// IsExpired returns true if the OTP has passed its expiry time.
func (o *EmailOTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// IsValid returns true if the OTP is unused and not expired.
func (o *EmailOTP) IsValid(code string) bool {
	return o.Code == code && !o.Used && !o.IsExpired()
}
