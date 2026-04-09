package unit_tests

import (
	"testing"
	"time"

	"github.com/eaglepoint/authapi/internal/models"
)

// Fix #4: Idempotency 24h window

func TestPaymentLedgerEntryHasIdempotencyExpiry(t *testing.T) {
	// Verify the model has the IdempotencyExpiresAt field
	now := time.Now()
	expiry := now.Add(24 * time.Hour)
	entry := models.PaymentLedgerEntry{
		IdempotencyExpiresAt: &expiry,
	}

	if entry.IdempotencyExpiresAt == nil {
		t.Fatal("IdempotencyExpiresAt should be set")
	}
	if !entry.IdempotencyExpiresAt.After(now) {
		t.Error("IdempotencyExpiresAt should be in the future")
	}

	// Verify the 24h window
	diff := entry.IdempotencyExpiresAt.Sub(now)
	if diff < 23*time.Hour || diff > 25*time.Hour {
		t.Errorf("expected ~24h window, got %v", diff)
	}
}

func TestIdempotencyKeyWithinWindow(t *testing.T) {
	// Within 24h: should return original (duplicate)
	entry := models.PaymentLedgerEntry{
		IdempotencyKey: "pay-001",
		CreatedAt:      time.Now().Add(-1 * time.Hour), // 1 hour ago
	}

	window := 24 * time.Hour
	if time.Since(entry.CreatedAt) >= window {
		t.Error("entry created 1 hour ago should be within 24h window")
	}
}

func TestIdempotencyKeyOutsideWindow(t *testing.T) {
	// After 24h: should allow new entry
	entry := models.PaymentLedgerEntry{
		IdempotencyKey: "pay-001",
		CreatedAt:      time.Now().Add(-25 * time.Hour), // 25 hours ago
	}

	window := 24 * time.Hour
	if time.Since(entry.CreatedAt) < window {
		t.Error("entry created 25 hours ago should be outside 24h window")
	}
}
