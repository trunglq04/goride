package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/trunglq04/goride/services/auth-service/internal/domain"
)

type txKey struct{}

// queryExecutor abstracts *sql.DB and *sql.Tx
type queryExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// PostgresRepository implements domain.Repository using raw SQL with transaction support.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL-backed repository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) getExecutor(ctx context.Context) queryExecutor {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok && tx != nil {
		return tx
	}
	return r.db
}

// WithTx executes the given function within a database transaction.
// If fn returns an error, the transaction is automatically rolled back.
// If fn returns nil, the transaction is committed.
// Reuses an existing transaction if one is already in ctx.
func (r *PostgresRepository) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if existingTx, ok := ctx.Value(txKey{}).(*sql.Tx); ok && existingTx != nil {
		return fn(ctx)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txCtx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// ---------- Roles ----------

func (r *PostgresRepository) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	query := `SELECT id, name FROM roles WHERE name = $1`
	role := &domain.Role{}
	err := r.getExecutor(ctx).QueryRowContext(ctx, query, name).Scan(&role.ID, &role.Name)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("role not found: %s", name)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}
	return role, nil
}

// ---------- Users ----------

func (r *PostgresRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, full_name, email, phone, password_hash, role_id, email_confirmed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.getExecutor(ctx).ExecContext(ctx, query,
		user.ID, user.FullName, user.Email, user.Phone, user.PasswordHash, user.RoleID,
		user.EmailConfirmed, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT u.id, u.full_name, u.email, u.phone, u.password_hash, u.role_id, r.name, u.email_confirmed, u.created_at, u.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1
	`
	user := &domain.User{}
	err := r.getExecutor(ctx).QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.FullName, &user.Email, &user.Phone, &user.PasswordHash, &user.RoleID,
		&user.RoleName, &user.EmailConfirmed, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		SELECT u.id, u.full_name, u.email, u.phone, u.password_hash, u.role_id, r.name, u.email_confirmed, u.created_at, u.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.phone = $1
	`
	user := &domain.User{}
	err := r.getExecutor(ctx).QueryRowContext(ctx, query, phone).Scan(
		&user.ID, &user.FullName, &user.Email, &user.Phone, &user.PasswordHash, &user.RoleID,
		&user.RoleName, &user.EmailConfirmed, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by phone: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT u.id, u.full_name, u.email, u.phone, u.password_hash, u.role_id, r.name, u.email_confirmed, u.created_at, u.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1
	`
	user := &domain.User{}
	err := r.getExecutor(ctx).QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.FullName, &user.Email, &user.Phone, &user.PasswordHash, &user.RoleID,
		&user.RoleName, &user.EmailConfirmed, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return user, nil
}

func (r *PostgresRepository) ConfirmUserEmail(ctx context.Context, userID string) error {
	query := `UPDATE users SET email_confirmed = TRUE, updated_at = NOW() WHERE id = $1`
	result, err := r.getExecutor(ctx).ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to confirm user email: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user not found: %s", userID)
	}
	return nil
}

// ---------- Email OTPs ----------

func (r *PostgresRepository) CreateEmailOTP(ctx context.Context, otp *domain.EmailOTP) error {
	query := `
		INSERT INTO email_otps (id, user_id, code, expires_at, used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.getExecutor(ctx).ExecContext(ctx, query,
		otp.ID, otp.UserID, otp.Code, otp.ExpiresAt, otp.Used, otp.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create email OTP: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetLatestOTPByUserID(ctx context.Context, userID string) (*domain.EmailOTP, error) {
	query := `
		SELECT id, user_id, code, expires_at, used, created_at
		FROM email_otps
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	otp := &domain.EmailOTP{}
	err := r.getExecutor(ctx).QueryRowContext(ctx, query, userID).Scan(
		&otp.ID, &otp.UserID, &otp.Code, &otp.ExpiresAt, &otp.Used, &otp.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest OTP: %w", err)
	}
	return otp, nil
}

func (r *PostgresRepository) MarkOTPUsed(ctx context.Context, otpID string) error {
	query := `UPDATE email_otps SET used = TRUE WHERE id = $1`
	_, err := r.getExecutor(ctx).ExecContext(ctx, query, otpID)
	if err != nil {
		return fmt.Errorf("failed to mark OTP as used: %w", err)
	}
	return nil
}

// ---------- Refresh Tokens ----------

func (r *PostgresRepository) CreateRefreshToken(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, family_id, revoked, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.getExecutor(ctx).ExecContext(ctx, query,
		token.ID, token.UserID, token.TokenHash, token.FamilyID,
		token.Revoked, token.ExpiresAt, token.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, family_id, replaced_by, revoked, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	token := &domain.RefreshToken{}
	err := r.getExecutor(ctx).QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.FamilyID,
		&token.ReplacedBy, &token.Revoked, &token.ExpiresAt, &token.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token by hash: %w", err)
	}
	return token, nil
}

func (r *PostgresRepository) RevokeTokenFamily(ctx context.Context, familyID string) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE family_id = $1`
	_, err := r.getExecutor(ctx).ExecContext(ctx, query, familyID)
	if err != nil {
		return fmt.Errorf("failed to revoke token family: %w", err)
	}
	return nil
}

func (r *PostgresRepository) MarkTokenReplaced(ctx context.Context, tokenID string, replacedByID string) error {
	query := `UPDATE refresh_tokens SET replaced_by = $1 WHERE id = $2`
	_, err := r.getExecutor(ctx).ExecContext(ctx, query, replacedByID, tokenID)
	if err != nil {
		return fmt.Errorf("failed to mark token as replaced: %w", err)
	}
	return nil
}
