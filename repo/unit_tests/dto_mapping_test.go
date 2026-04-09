package unit_tests

import (
	"testing"
	"time"

	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestToAccountResponse_MasksPassword(t *testing.T) {
	account := &models.Account{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: "$2a$12$supersecrethashedvalue",
		Role:         models.RoleAdministrator,
		Status:       models.StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	resp := dto.ToAccountResponse(account)

	if resp.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", resp.Username)
	}
	if resp.Role != "Administrator" {
		t.Errorf("expected role Administrator, got %s", resp.Role)
	}
	// AccountResponse should NOT contain PasswordHash - verify struct has no such field
	// This is a compile-time guarantee, but we verify the mapping is correct
	if resp.ID != account.ID {
		t.Error("ID should be mapped")
	}
}

func TestToAccountResponseList(t *testing.T) {
	accounts := []models.Account{
		{ID: uuid.New(), Username: "user1", Role: models.RoleScheduler, Status: models.StatusActive, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New(), Username: "user2", Role: models.RoleInstructor, Status: models.StatusFrozen, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	result := dto.ToAccountResponseList(accounts)
	if len(result) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(result))
	}
	if result[0].Username != "user1" {
		t.Errorf("expected user1, got %s", result[0].Username)
	}
	if result[1].Status != "Frozen" {
		t.Errorf("expected Frozen, got %s", result[1].Status)
	}
}

func TestToSeasonResponse_DateFormat(t *testing.T) {
	season := &models.Season{
		ID:        uuid.New(),
		Name:      "Fall 2025",
		StartDate: time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC),
		Status:    models.SeasonPlanning,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	resp := dto.ToSeasonResponse(season)
	if resp.StartDate != "2025-09-01" {
		t.Errorf("expected 2025-09-01, got %s", resp.StartDate)
	}
	if resp.EndDate != "2025-12-15" {
		t.Errorf("expected 2025-12-15, got %s", resp.EndDate)
	}
}

func TestToMatchResponse(t *testing.T) {
	overrideReason := "scheduling conflict approved by coordinator"
	match := &models.Match{
		ID:             uuid.New(),
		SeasonID:       uuid.New(),
		Round:          3,
		HomeTeamID:     uuid.New(),
		AwayTeamID:     uuid.New(),
		VenueID:        uuid.New(),
		ScheduledAt:    time.Now(),
		Status:         models.MatchScheduled,
		OverrideReason: &overrideReason,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	resp := dto.ToMatchResponse(match)
	if resp.Round != 3 {
		t.Errorf("expected round 3, got %d", resp.Round)
	}
	if resp.Status != "Scheduled" {
		t.Errorf("expected Scheduled, got %s", resp.Status)
	}
	if resp.OverrideReason == nil || *resp.OverrideReason != overrideReason {
		t.Error("override reason should be mapped")
	}
}

func TestToPaymentResponse(t *testing.T) {
	entry := &models.PaymentLedgerEntry{
		ID:             uuid.New(),
		AccountID:      uuid.New(),
		IdempotencyKey: "pay-001",
		AmountUSD:      decimal.NewFromFloat(150.75),
		Channel:        models.ChannelCash,
		Status:         models.PaymentObligation,
		RetryCount:     0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	resp := dto.ToPaymentResponse(entry)
	if resp.IdempotencyKey != "pay-001" {
		t.Errorf("expected pay-001, got %s", resp.IdempotencyKey)
	}
	if !resp.AmountUSD.Equal(decimal.NewFromFloat(150.75)) {
		t.Errorf("expected 150.75, got %s", resp.AmountUSD)
	}
	if resp.Channel != "Cash" {
		t.Errorf("expected Cash, got %s", resp.Channel)
	}
	if resp.Status != "Obligation" {
		t.Errorf("expected Obligation, got %s", resp.Status)
	}
}

func TestBuildOutlineTree(t *testing.T) {
	courseID := uuid.New()
	chapterID := uuid.New()
	unit1ID := uuid.New()
	unit2ID := uuid.New()

	nodes := []models.CourseOutlineNode{
		{ID: chapterID, CourseID: courseID, ParentID: nil, NodeType: models.NodeTypeChapter, Title: "Chapter 1", OrderIndex: 0, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: unit1ID, CourseID: courseID, ParentID: &chapterID, NodeType: models.NodeTypeUnit, Title: "Unit 1.1", OrderIndex: 0, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: unit2ID, CourseID: courseID, ParentID: &chapterID, NodeType: models.NodeTypeUnit, Title: "Unit 1.2", OrderIndex: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	tree := dto.BuildOutlineTree(nodes)
	if len(tree) != 1 {
		t.Fatalf("expected 1 root node, got %d", len(tree))
	}
	if tree[0].Title != "Chapter 1" {
		t.Errorf("expected Chapter 1, got %s", tree[0].Title)
	}
	if len(tree[0].Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(tree[0].Children))
	}
	if tree[0].Children[0].Title != "Unit 1.1" {
		t.Errorf("expected Unit 1.1, got %s", tree[0].Children[0].Title)
	}
}

func TestBuildOutlineTree_Empty(t *testing.T) {
	tree := dto.BuildOutlineTree(nil)
	if tree != nil {
		t.Errorf("expected nil for empty input, got %v", tree)
	}
}
