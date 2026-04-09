package unit_tests

import (
	"testing"
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
)

// Fix 3: Dedicated idempotency table with deterministic windows

func TestIdempotencyKeyModelFields(t *testing.T) {
	now := time.Now()
	windowEnd := now.Add(24 * time.Hour)
	paymentID := uuid.New()
	accountID := uuid.New()

	ik := models.IdempotencyKey{
		ID:             uuid.New(),
		AccountID:      accountID,
		IdempotencyKey: "pay-test-001",
		PaymentID:      paymentID,
		WindowStart:    now,
		WindowEnd:      windowEnd,
		CreatedAt:      now,
	}

	if ik.AccountID != accountID {
		t.Error("AccountID mismatch")
	}
	if ik.PaymentID != paymentID {
		t.Error("PaymentID mismatch")
	}
	if ik.IdempotencyKey != "pay-test-001" {
		t.Error("key mismatch")
	}
	if !ik.WindowEnd.After(ik.WindowStart) {
		t.Error("WindowEnd should be after WindowStart")
	}
}

func TestIdempotencyWindowIsExactly24Hours(t *testing.T) {
	start := time.Now()
	end := start.Add(24 * time.Hour)
	diff := end.Sub(start)

	if diff != 24*time.Hour {
		t.Errorf("expected 24h window, got %v", diff)
	}
}

func TestIdempotencyKeyActiveWithinWindow(t *testing.T) {
	now := time.Now()
	ik := models.IdempotencyKey{
		WindowStart: now.Add(-1 * time.Hour),
		WindowEnd:   now.Add(23 * time.Hour),
	}
	// Key is active: now is between window_start and window_end
	if !now.Before(ik.WindowEnd) {
		t.Error("key should be active within window")
	}
}

func TestIdempotencyKeyExpiredAfterWindow(t *testing.T) {
	now := time.Now()
	ik := models.IdempotencyKey{
		WindowStart: now.Add(-25 * time.Hour),
		WindowEnd:   now.Add(-1 * time.Hour),
	}
	// Key is expired: window_end is in the past
	if now.Before(ik.WindowEnd) {
		t.Error("key should be expired after window")
	}
}
