package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/eaglepoint/authapi/internal/config"
	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrAccountLocked      = errors.New("account is temporarily locked due to too many failed login attempts")
	ErrAccountNotActive   = errors.New("account is not active")
	ErrTokenInvalid       = errors.New("invalid or expired refresh token")
	ErrTokenReuse         = errors.New("refresh token reuse detected, all sessions revoked")
	ErrPasswordPolicy     = errors.New("password does not meet policy requirements")
)

type AuthService struct {
	accountRepo      *repository.AccountRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	loginAttemptRepo *repository.LoginAttemptRepository
	tokenService     *TokenService
	deviceService    *DeviceService
	audit            *AuditService
	cfg              *config.Config
}

func NewAuthService(
	accountRepo *repository.AccountRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	loginAttemptRepo *repository.LoginAttemptRepository,
	tokenService *TokenService,
	deviceService *DeviceService,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		accountRepo:      accountRepo,
		refreshTokenRepo: refreshTokenRepo,
		loginAttemptRepo: loginAttemptRepo,
		tokenService:     tokenService,
		deviceService:    deviceService,
		cfg:              cfg,
	}
}

// SetAuditService sets the audit service (called after construction to avoid circular dep)
func (s *AuthService) SetAuditService(audit *AuditService) {
	s.audit = audit
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest, ipAddress string) (*dto.LoginResponse, error) {
	account, err := s.accountRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// Prevent timing-based user enumeration: always hash
		bcrypt.CompareHashAndPassword([]byte("$2a$12$000000000000000000000000000000000000000000000000000000"), []byte(req.Password))
		return nil, ErrInvalidCredentials
	}

	if account.Status != models.StatusActive {
		return nil, ErrAccountNotActive
	}

	// Check lockout
	failCount, err := s.loginAttemptRepo.CountRecentFailures(ctx, account.ID, s.cfg.LockoutDuration)
	if err != nil {
		return nil, fmt.Errorf("auth_service.Login: %w", err)
	}
	if failCount >= s.cfg.MaxLoginAttempts {
		if s.audit != nil {
			s.audit.LogExtended(ctx, &AuditEntry{
				EntityType: "auth",
				EntityID:   account.ID,
				ActorID:    account.ID,
				Action:     "login_locked",
				Tier:       models.TierAccess,
				Source:     strPtr("auth/login"),
				Reason:     strPtr("too many failed attempts"),
				Details:    map[string]interface{}{"username": account.Username, "ip_address": ipAddress},
			})
		}
		return nil, ErrAccountLocked
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Password)); err != nil {
		s.recordAttempt(ctx, account.ID, false, ipAddress)
		if s.audit != nil {
			s.audit.LogExtended(ctx, &AuditEntry{
				EntityType: "auth",
				EntityID:   account.ID,
				ActorID:    account.ID,
				Action:     "login_failure",
				Tier:       models.TierAccess,
				Source:     strPtr("auth/login"),
				Reason:     strPtr("invalid password"),
				Details:    map[string]interface{}{"username": account.Username, "ip_address": ipAddress},
			})
		}
		return nil, ErrInvalidCredentials
	}

	// Record successful attempt
	s.recordAttempt(ctx, account.ID, true, ipAddress)

	// Handle device fingerprint
	var deviceID *uuid.UUID
	if req.DeviceFingerprint != nil {
		fpHash := s.deviceService.ComputeFingerprint(req.DeviceFingerprint.UserAgent, req.DeviceFingerprint.Attributes)
		device, err := s.deviceService.RegisterOrUpdateDevice(ctx, account.ID, fpHash, nil)
		if err == nil {
			deviceID = &device.ID
		}
	}

	// Generate tokens
	accessToken, err := s.tokenService.GenerateAccessToken(account)
	if err != nil {
		return nil, fmt.Errorf("auth_service.Login: %w", err)
	}

	rawRefresh, hashedRefresh, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("auth_service.Login: %w", err)
	}

	now := time.Now()
	refreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		AccountID: account.ID,
		TokenHash: hashedRefresh,
		DeviceID:  deviceID,
		ExpiresAt: now.Add(s.cfg.RefreshTokenTTL),
		Revoked:   false,
		CreatedAt: now,
	}
	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("auth_service.Login: %w", err)
	}

	// Audit: access-tier log for successful login
	if s.audit != nil {
		s.audit.LogExtended(ctx, &AuditEntry{
			EntityType: "auth",
			EntityID:   account.ID,
			ActorID:    account.ID,
			Action:     "login_success",
			Tier:       models.TierAccess,
			Source:     strPtr("auth/login"),
			Details: map[string]interface{}{
				"username":   account.Username,
				"ip_address": ipAddress,
			},
		})
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int(s.cfg.AccessTokenTTL.Seconds()),
		Account:      dto.ToAccountResponse(account),
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string) (*dto.RefreshResponse, error) {
	hashed := HashRefreshToken(rawRefreshToken)
	stored, err := s.refreshTokenRepo.GetByHash(ctx, hashed)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	// Reuse detection: if token was already revoked, revoke all tokens for the account
	if stored.Revoked {
		s.refreshTokenRepo.RevokeAllForAccount(ctx, stored.AccountID)
		if s.audit != nil {
			s.audit.LogExtended(ctx, &AuditEntry{
				EntityType: "auth",
				EntityID:   stored.AccountID,
				ActorID:    stored.AccountID,
				Action:     "refresh_token_reuse",
				Tier:       models.TierAccess,
				Source:     strPtr("auth/refresh"),
				Reason:     strPtr("token reuse detected, all sessions revoked"),
			})
		}
		return nil, ErrTokenReuse
	}

	if time.Now().After(stored.ExpiresAt) {
		if s.audit != nil {
			s.audit.LogExtended(ctx, &AuditEntry{
				EntityType: "auth",
				EntityID:   stored.AccountID,
				ActorID:    stored.AccountID,
				Action:     "refresh_token_expired",
				Tier:       models.TierAccess,
				Source:     strPtr("auth/refresh"),
				Reason:     strPtr("refresh token expired"),
			})
		}
		return nil, ErrTokenInvalid
	}

	// Revoke the current token
	if err := s.refreshTokenRepo.Revoke(ctx, stored.ID); err != nil {
		return nil, fmt.Errorf("auth_service.Refresh: %w", err)
	}

	// Get account for new access token
	account, err := s.accountRepo.GetByID(ctx, stored.AccountID)
	if err != nil {
		return nil, fmt.Errorf("auth_service.Refresh: %w", err)
	}

	if account.Status != models.StatusActive {
		return nil, ErrAccountNotActive
	}

	// Generate new token pair
	accessToken, err := s.tokenService.GenerateAccessToken(account)
	if err != nil {
		return nil, fmt.Errorf("auth_service.Refresh: %w", err)
	}

	newRaw, newHashed, err := s.tokenService.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("auth_service.Refresh: %w", err)
	}

	now := time.Now()
	newRefreshToken := &models.RefreshToken{
		ID:        uuid.New(),
		AccountID: stored.AccountID,
		TokenHash: newHashed,
		DeviceID:  stored.DeviceID,
		ExpiresAt: now.Add(s.cfg.RefreshTokenTTL),
		Revoked:   false,
		CreatedAt: now,
	}
	if err := s.refreshTokenRepo.Create(ctx, newRefreshToken); err != nil {
		return nil, fmt.Errorf("auth_service.Refresh: %w", err)
	}

	if s.audit != nil {
		s.audit.LogExtended(ctx, &AuditEntry{
			EntityType: "auth",
			EntityID:   stored.AccountID,
			ActorID:    stored.AccountID,
			Action:     "token_refresh",
			Tier:       models.TierAccess,
			Source:     strPtr("auth/refresh"),
		})
	}

	return &dto.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRaw,
		ExpiresIn:    int(s.cfg.AccessTokenTTL.Seconds()),
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	hashed := HashRefreshToken(rawRefreshToken)
	stored, err := s.refreshTokenRepo.GetByHash(ctx, hashed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("auth_service.Logout: %w", err)
	}
	if s.audit != nil {
		s.audit.LogExtended(ctx, &AuditEntry{
			EntityType: "auth",
			EntityID:   stored.AccountID,
			ActorID:    stored.AccountID,
			Action:     "logout",
			Tier:       models.TierAccess,
			Source:     strPtr("auth/logout"),
		})
	}
	return s.refreshTokenRepo.Revoke(ctx, stored.ID)
}

func ValidatePassword(password string) error {
	var errs []string

	if len(password) < 12 {
		errs = append(errs, "at least 12 characters")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}

	if !hasUpper {
		errs = append(errs, "at least 1 uppercase letter")
	}
	if !hasLower {
		errs = append(errs, "at least 1 lowercase letter")
	}
	if !hasDigit {
		errs = append(errs, "at least 1 number")
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: password must contain %s", ErrPasswordPolicy, strings.Join(errs, ", "))
	}
	return nil
}

func strPtr(s string) *string { return &s }

func (s *AuthService) recordAttempt(ctx context.Context, accountID uuid.UUID, success bool, ipAddress string) {
	attempt := &models.LoginAttempt{
		ID:          uuid.New(),
		AccountID:   accountID,
		Success:     success,
		IPAddress:   &ipAddress,
		AttemptedAt: time.Now(),
	}
	s.loginAttemptRepo.Record(ctx, attempt)
}
