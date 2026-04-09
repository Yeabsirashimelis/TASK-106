package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eaglepoint/authapi/internal/config"
	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidRole      = errors.New("invalid role")
	ErrDuplicateUsername = errors.New("username already exists")
)

type AccountService struct {
	accountRepo      *repository.AccountRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	cfg              *config.Config
}

func NewAccountService(
	accountRepo *repository.AccountRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	cfg *config.Config,
) *AccountService {
	return &AccountService{
		accountRepo:      accountRepo,
		refreshTokenRepo: refreshTokenRepo,
		cfg:              cfg,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, req *dto.CreateAccountRequest) (*models.Account, error) {
	if err := ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	role := models.Role(req.Role)
	if !models.ValidRoles[role] {
		return nil, fmt.Errorf("%w: %s", ErrInvalidRole, req.Role)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.cfg.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("account_service.CreateAccount: %w", err)
	}

	now := time.Now()
	account := &models.Account{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: string(hash),
		Role:         role,
		Status:       models.StatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.accountRepo.Create(ctx, account); err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateUsername, req.Username)
		}
		return nil, err
	}
	return account, nil
}

func (s *AccountService) GetAccount(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	return s.accountRepo.GetByID(ctx, id)
}

func (s *AccountService) ListAccounts(ctx context.Context, offset, limit int) ([]models.Account, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.accountRepo.List(ctx, offset, limit)
}

func (s *AccountService) UpdateStatus(ctx context.Context, id uuid.UUID, newStatus models.Status) error {
	if !models.ValidStatuses[newStatus] {
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	if err := s.accountRepo.UpdateStatus(ctx, id, newStatus); err != nil {
		return err
	}

	// When deactivating, revoke all sessions
	if newStatus == models.StatusDeactivated || newStatus == models.StatusFrozen {
		if err := s.refreshTokenRepo.RevokeAllForAccount(ctx, id); err != nil {
			return fmt.Errorf("account_service.UpdateStatus: %w", err)
		}
	}
	return nil
}

func (s *AccountService) ChangePassword(ctx context.Context, accountID uuid.UUID, oldPassword, newPassword string) error {
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("account_service.ChangePassword: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.cfg.BcryptCost)
	if err != nil {
		return fmt.Errorf("account_service.ChangePassword: %w", err)
	}

	if err := s.accountRepo.UpdatePassword(ctx, accountID, string(hash)); err != nil {
		return err
	}

	// Revoke all refresh tokens to force re-login on all devices
	return s.refreshTokenRepo.RevokeAllForAccount(ctx, accountID)
}
