package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrDictionaryNotFound = errors.New("dictionary not found")
	ErrReviewNotFound     = errors.New("review not found")
	ErrInvalidReviewTrans = errors.New("invalid review status transition")
	ErrReviewReasonReq    = errors.New("reason is required for moderation decisions")
)

type ModerationService struct {
	wordRepo   *repository.SensitiveWordRepository
	reviewRepo *repository.ModerationReviewRepository
	audit      *AuditService
}

func NewModerationService(
	wordRepo *repository.SensitiveWordRepository,
	reviewRepo *repository.ModerationReviewRepository,
	audit *AuditService,
) *ModerationService {
	return &ModerationService{
		wordRepo:   wordRepo,
		reviewRepo: reviewRepo,
		audit:      audit,
	}
}

// Dictionary management

func (s *ModerationService) CreateDictionary(ctx context.Context, req *dto.CreateDictionaryRequest, actorID uuid.UUID) (*models.SensitiveWordDictionary, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	now := time.Now()
	d := &models.SensitiveWordDictionary{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		IsActive:    true,
		CreatedBy:   actorID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.wordRepo.CreateDictionary(ctx, d); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "sensitive_word_dictionary", d.ID, actorID, "created", map[string]interface{}{
		"name": d.Name,
	})

	return d, nil
}

func (s *ModerationService) GetDictionary(ctx context.Context, id uuid.UUID) (*models.SensitiveWordDictionary, error) {
	d, err := s.wordRepo.GetDictionaryByID(ctx, id)
	if err != nil {
		return nil, ErrDictionaryNotFound
	}
	return d, nil
}

func (s *ModerationService) ListDictionaries(ctx context.Context, offset, limit int) ([]models.SensitiveWordDictionary, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.wordRepo.ListDictionaries(ctx, offset, limit)
}

func (s *ModerationService) UpdateDictionary(ctx context.Context, id uuid.UUID, req *dto.UpdateDictionaryRequest, actorID uuid.UUID) (*models.SensitiveWordDictionary, error) {
	d, err := s.wordRepo.GetDictionaryByID(ctx, id)
	if err != nil {
		return nil, ErrDictionaryNotFound
	}

	if req.Name != nil {
		d.Name = *req.Name
	}
	if req.Description != nil {
		d.Description = req.Description
	}
	if req.IsActive != nil {
		d.IsActive = *req.IsActive
	}

	if err := s.wordRepo.UpdateDictionary(ctx, d); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "sensitive_word_dictionary", id, actorID, "updated", map[string]interface{}{
		"name":      d.Name,
		"is_active": d.IsActive,
	})

	return d, nil
}

func (s *ModerationService) DeleteDictionary(ctx context.Context, id, actorID uuid.UUID) error {
	d, err := s.wordRepo.GetDictionaryByID(ctx, id)
	if err != nil {
		return ErrDictionaryNotFound
	}

	if err := s.wordRepo.DeleteDictionary(ctx, id); err != nil {
		return err
	}

	s.audit.Log(ctx, "sensitive_word_dictionary", id, actorID, "deleted", map[string]interface{}{
		"name": d.Name,
	})

	return nil
}

// Word management

func (s *ModerationService) AddWord(ctx context.Context, dictionaryID uuid.UUID, req *dto.AddWordRequest, actorID uuid.UUID) (*models.SensitiveWord, error) {
	if req.Word == "" {
		return nil, fmt.Errorf("word is required")
	}
	if req.Severity == "" {
		req.Severity = "medium"
	}

	if _, err := s.wordRepo.GetDictionaryByID(ctx, dictionaryID); err != nil {
		return nil, ErrDictionaryNotFound
	}

	w := &models.SensitiveWord{
		ID:           uuid.New(),
		DictionaryID: dictionaryID,
		Word:         strings.ToLower(strings.TrimSpace(req.Word)),
		Severity:     req.Severity,
		CreatedAt:    time.Now(),
	}

	if err := s.wordRepo.AddWord(ctx, w); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "sensitive_word", w.ID, actorID, "added", map[string]interface{}{
		"dictionary_id": dictionaryID,
		"word":          w.Word,
		"severity":      w.Severity,
	})

	return w, nil
}

func (s *ModerationService) AddWords(ctx context.Context, dictionaryID uuid.UUID, req *dto.AddWordsRequest, actorID uuid.UUID) ([]models.SensitiveWord, error) {
	if _, err := s.wordRepo.GetDictionaryByID(ctx, dictionaryID); err != nil {
		return nil, ErrDictionaryNotFound
	}

	var added []models.SensitiveWord
	for _, wr := range req.Words {
		if wr.Word == "" {
			continue
		}
		if wr.Severity == "" {
			wr.Severity = "medium"
		}
		w := &models.SensitiveWord{
			ID:           uuid.New(),
			DictionaryID: dictionaryID,
			Word:         strings.ToLower(strings.TrimSpace(wr.Word)),
			Severity:     wr.Severity,
			CreatedAt:    time.Now(),
		}
		if err := s.wordRepo.AddWord(ctx, w); err == nil {
			added = append(added, *w)
		}
	}

	s.audit.Log(ctx, "sensitive_word_dictionary", dictionaryID, actorID, "words_bulk_added", map[string]interface{}{
		"count": len(added),
	})

	return added, nil
}

func (s *ModerationService) ListWords(ctx context.Context, dictionaryID uuid.UUID, offset, limit int) ([]models.SensitiveWord, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	return s.wordRepo.ListWords(ctx, dictionaryID, offset, limit)
}

func (s *ModerationService) DeleteWord(ctx context.Context, wordID, actorID uuid.UUID) error {
	if err := s.wordRepo.DeleteWord(ctx, wordID); err != nil {
		return err
	}

	s.audit.Log(ctx, "sensitive_word", wordID, actorID, "deleted", nil)
	return nil
}

// Content checking with match highlighting

func (s *ModerationService) CheckContent(ctx context.Context, text string) (*dto.CheckContentResponse, error) {
	words, err := s.wordRepo.GetAllActiveWords(ctx)
	if err != nil {
		return nil, fmt.Errorf("moderation_service.CheckContent: %w", err)
	}

	lowerText := strings.ToLower(text)
	var matches []dto.ContentMatch

	for _, w := range words {
		lowerWord := strings.ToLower(w.Word)
		searchFrom := 0
		for {
			idx := strings.Index(lowerText[searchFrom:], lowerWord)
			if idx == -1 {
				break
			}
			absIdx := searchFrom + idx
			endIdx := absIdx + len(w.Word)

			// Extract context: up to 30 chars before and after
			ctxStart := absIdx - 30
			if ctxStart < 0 {
				ctxStart = 0
			}
			ctxEnd := endIdx + 30
			if ctxEnd > len(text) {
				ctxEnd = len(text)
			}

			// Build highlighted context with markers
			contextStr := text[ctxStart:ctxEnd]

			matches = append(matches, dto.ContentMatch{
				Word:     w.Word,
				Severity: w.Severity,
				Start:    absIdx,
				End:      endIdx,
				Context:  contextStr,
			})
			searchFrom = endIdx
		}
	}

	return &dto.CheckContentResponse{
		Clean:   len(matches) == 0,
		Matches: matches,
	}, nil
}

// Review management

func (s *ModerationService) CreateReview(ctx context.Context, req *dto.CreateReviewRequest, actorID uuid.UUID) (*models.ModerationReview, error) {
	if req.ContentType == "" || req.ContentID == "" {
		return nil, fmt.Errorf("content_type and content_id are required")
	}

	contentID, err := uuid.Parse(req.ContentID)
	if err != nil {
		return nil, fmt.Errorf("invalid content_id")
	}

	now := time.Now()
	review := &models.ModerationReview{
		ID:             uuid.New(),
		ContentType:    req.ContentType,
		ContentID:      contentID,
		ContentSnippet: req.ContentSnippet,
		Status:         models.ReviewPending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "moderation_review", review.ID, actorID, "created", map[string]interface{}{
		"content_type": req.ContentType,
		"content_id":   req.ContentID,
	})

	return review, nil
}

func (s *ModerationService) GetReview(ctx context.Context, id uuid.UUID) (*models.ModerationReview, error) {
	r, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrReviewNotFound
	}
	return r, nil
}

func (s *ModerationService) ListReviews(ctx context.Context, statusFilter *string, offset, limit int) ([]models.ModerationReview, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var status *models.ReviewStatus
	if statusFilter != nil {
		s := models.ReviewStatus(*statusFilter)
		if models.ValidReviewStatuses[s] {
			status = &s
		}
	}
	return s.reviewRepo.List(ctx, status, offset, limit)
}

func (s *ModerationService) DecideReview(ctx context.Context, id uuid.UUID, req *dto.DecideReviewRequest, moderatorID uuid.UUID) (*models.ModerationReview, error) {
	if req.Reason == "" {
		return nil, ErrReviewReasonReq
	}

	review, err := s.reviewRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrReviewNotFound
	}

	newStatus := models.ReviewStatus(req.Status)
	if !models.ValidReviewStatuses[newStatus] {
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	if !models.CanTransitionReview(review.Status, newStatus) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidReviewTrans, review.Status, newStatus)
	}

	now := time.Now()
	review.Status = newStatus
	review.ModeratorID = &moderatorID
	review.Reason = &req.Reason
	review.DecidedAt = &now

	if err := s.reviewRepo.Update(ctx, review); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "moderation_review", id, moderatorID, "decided", map[string]interface{}{
		"status":       req.Status,
		"reason":       req.Reason,
		"content_type": review.ContentType,
		"content_id":   review.ContentID,
	})

	return review, nil
}
