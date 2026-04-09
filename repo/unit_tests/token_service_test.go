package unit_tests

import (
	"testing"
	"time"

	"github.com/eaglepoint/authapi/internal/config"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
)

func newTestTokenService() *service.TokenService {
	cfg := &config.Config{
		JWTSecret:      "test-secret-at-least-32-bytes-long!!",
		AccessTokenTTL: 30 * time.Minute,
	}
	return service.NewTokenService(cfg)
}

func TestTokenService_GenerateAndValidate(t *testing.T) {
	ts := newTestTokenService()
	account := &models.Account{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleAdministrator,
	}

	token, err := ts.GenerateAccessToken(account)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	if token == "" {
		t.Fatal("generated token is empty")
	}

	claims, err := ts.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}
	if claims.AccountID != account.ID {
		t.Errorf("expected account ID %s, got %s", account.ID, claims.AccountID)
	}
	if claims.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", claims.Username)
	}
	if claims.Role != models.RoleAdministrator {
		t.Errorf("expected role Administrator, got %s", claims.Role)
	}
}

func TestTokenService_InvalidToken(t *testing.T) {
	ts := newTestTokenService()
	_, err := ts.ValidateAccessToken("invalid-token-string")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestTokenService_ExpiredToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:      "test-secret-at-least-32-bytes-long!!",
		AccessTokenTTL: -1 * time.Hour, // Already expired
	}
	ts := service.NewTokenService(cfg)
	account := &models.Account{
		ID:       uuid.New(),
		Username: "expireduser",
		Role:     models.RoleInstructor,
	}

	token, err := ts.GenerateAccessToken(account)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = ts.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestTokenService_WrongSecret(t *testing.T) {
	ts1 := newTestTokenService()
	account := &models.Account{
		ID:       uuid.New(),
		Username: "testuser",
		Role:     models.RoleScheduler,
	}

	token, _ := ts1.GenerateAccessToken(account)

	// Validate with a different secret
	cfg2 := &config.Config{
		JWTSecret:      "completely-different-secret-32-bytes!!",
		AccessTokenTTL: 30 * time.Minute,
	}
	ts2 := service.NewTokenService(cfg2)

	_, err := ts2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error when validating with wrong secret")
	}
}

func TestTokenService_DifferentRoles(t *testing.T) {
	ts := newTestTokenService()
	roles := []models.Role{
		models.RoleAdministrator,
		models.RoleScheduler,
		models.RoleInstructor,
		models.RoleReviewer,
		models.RoleFinanceClerk,
		models.RoleAuditor,
	}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			account := &models.Account{
				ID:       uuid.New(),
				Username: "user_" + string(role),
				Role:     role,
			}
			token, err := ts.GenerateAccessToken(account)
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}
			claims, err := ts.ValidateAccessToken(token)
			if err != nil {
				t.Fatalf("failed to validate token: %v", err)
			}
			if claims.Role != role {
				t.Errorf("expected role %s, got %s", role, claims.Role)
			}
		})
	}
}

func TestRefreshToken_GenerateAndHash(t *testing.T) {
	ts := newTestTokenService()
	raw, hashed, err := ts.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}
	if raw == "" || hashed == "" {
		t.Fatal("raw or hashed token is empty")
	}
	if raw == hashed {
		t.Error("raw and hashed should be different")
	}

	// Verify deterministic hashing
	rehashed := service.HashRefreshToken(raw)
	if rehashed != hashed {
		t.Errorf("re-hashing should produce same result: got %s, expected %s", rehashed, hashed)
	}
}

func TestRefreshToken_Uniqueness(t *testing.T) {
	ts := newTestTokenService()
	raw1, _, _ := ts.GenerateRefreshToken()
	raw2, _, _ := ts.GenerateRefreshToken()
	if raw1 == raw2 {
		t.Error("two refresh tokens should be unique")
	}
}
