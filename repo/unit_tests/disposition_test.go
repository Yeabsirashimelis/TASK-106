package unit_tests

import (
	"context"
	"testing"

	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
)

// Fix #6: Disposition write-back - verify callback registration and execution

func TestDispositionCallbackRegistration(t *testing.T) {
	reviewSvc := service.NewReviewService(nil, service.NewAuditService(nil))

	called := false
	var receivedDecision string
	var receivedEntityID uuid.UUID

	reviewSvc.RegisterDisposition("course", func(ctx context.Context, entityType string, entityID uuid.UUID, decision string) error {
		called = true
		receivedDecision = decision
		receivedEntityID = entityID
		return nil
	})

	// Execute the disposition via the exported method
	testID := uuid.New()
	reviewSvc.ExecuteDispositionPublic(context.Background(), "course", testID, "Approved")

	if !called {
		t.Fatal("disposition callback was not called")
	}
	if receivedDecision != "Approved" {
		t.Errorf("expected Approved, got %s", receivedDecision)
	}
	if receivedEntityID != testID {
		t.Errorf("expected entity ID %s, got %s", testID, receivedEntityID)
	}
}

func TestDispositionCallbackNotRegistered(t *testing.T) {
	reviewSvc := service.NewReviewService(nil, service.NewAuditService(nil))

	// Calling disposition for unregistered type should be a no-op (not panic)
	reviewSvc.ExecuteDispositionPublic(context.Background(), "unknown_type", uuid.New(), "Approved")
	// If we get here without panic, the test passes
}

func TestMultipleDispositionCallbacks(t *testing.T) {
	reviewSvc := service.NewReviewService(nil, service.NewAuditService(nil))

	courseCalled := false
	resourceCalled := false
	matchCalled := false

	reviewSvc.RegisterDisposition("course", func(ctx context.Context, entityType string, entityID uuid.UUID, decision string) error {
		courseCalled = true
		return nil
	})
	reviewSvc.RegisterDisposition("resource", func(ctx context.Context, entityType string, entityID uuid.UUID, decision string) error {
		resourceCalled = true
		return nil
	})
	reviewSvc.RegisterDisposition("match", func(ctx context.Context, entityType string, entityID uuid.UUID, decision string) error {
		matchCalled = true
		return nil
	})

	reviewSvc.ExecuteDispositionPublic(context.Background(), "course", uuid.New(), "Approved")
	reviewSvc.ExecuteDispositionPublic(context.Background(), "resource", uuid.New(), "Approved")

	if !courseCalled {
		t.Error("course callback not called")
	}
	if !resourceCalled {
		t.Error("resource callback not called")
	}
	if matchCalled {
		t.Error("match callback should not have been called")
	}
}
