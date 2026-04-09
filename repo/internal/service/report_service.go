package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/eaglepoint/authapi/internal/config"
	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/google/uuid"
)

const maxReportsPerDay = 10

var (
	ErrReportNotFound    = errors.New("report not found")
	ErrReportLimitExc    = errors.New("daily report limit exceeded (max 10 per day)")
	ErrInvalidReportTrans = errors.New("invalid report status transition")
	ErrEvidenceNotFound  = errors.New("evidence not found")
)

type ReportService struct {
	reportRepo *repository.ReportRepository
	audit      *AuditService
	cfg        *config.Config
}

func NewReportService(
	reportRepo *repository.ReportRepository,
	audit *AuditService,
	cfg *config.Config,
) *ReportService {
	return &ReportService{
		reportRepo: reportRepo,
		audit:      audit,
		cfg:        cfg,
	}
}

func (s *ReportService) CreateReport(ctx context.Context, req *dto.CreateReportRequest, reporterID uuid.UUID) (*models.Report, error) {
	if req.TargetType == "" || req.TargetID == "" {
		return nil, fmt.Errorf("target_type and target_id are required")
	}
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}

	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		return nil, fmt.Errorf("invalid target_id")
	}

	category := models.ReportCategory(req.Category)
	if !models.ValidReportCategories[category] {
		return nil, fmt.Errorf("invalid category: %s", req.Category)
	}

	// Enforce daily report limit
	count, err := s.reportRepo.CountReporterToday(ctx, reporterID)
	if err != nil {
		return nil, fmt.Errorf("report_service.CreateReport: %w", err)
	}
	if count >= maxReportsPerDay {
		return nil, ErrReportLimitExc
	}

	now := time.Now()
	report := &models.Report{
		ID:          uuid.New(),
		ReporterID:  reporterID,
		TargetType:  req.TargetType,
		TargetID:    targetID,
		Category:    category,
		Description: req.Description,
		Status:      models.ReportOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.reportRepo.Create(ctx, report); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "report", report.ID, reporterID, "created", map[string]interface{}{
		"target_type": req.TargetType,
		"target_id":   req.TargetID,
		"category":    req.Category,
	})

	return report, nil
}

func (s *ReportService) GetReport(ctx context.Context, id uuid.UUID) (*models.Report, error) {
	r, err := s.reportRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrReportNotFound
	}
	return r, nil
}

func (s *ReportService) ListReports(ctx context.Context, statusFilter *string, offset, limit int) ([]models.Report, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	var status *models.ReportStatus
	if statusFilter != nil {
		st := models.ReportStatus(*statusFilter)
		if models.ValidReportStatuses[st] {
			status = &st
		}
	}
	return s.reportRepo.List(ctx, status, offset, limit)
}

func (s *ReportService) UpdateReportStatus(ctx context.Context, id uuid.UUID, req *dto.UpdateReportStatusRequest, actorID uuid.UUID) (*models.Report, error) {
	report, err := s.reportRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrReportNotFound
	}

	newStatus := models.ReportStatus(req.Status)
	if !models.ValidReportStatuses[newStatus] {
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	if !models.CanTransitionReport(report.Status, newStatus) {
		return nil, fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidReportTrans, report.Status, newStatus)
	}

	report.Status = newStatus
	if req.Resolution != nil {
		report.Resolution = req.Resolution
	}
	if newStatus == models.ReportResolved || newStatus == models.ReportDismissed {
		now := time.Now()
		report.ResolvedAt = &now
	}

	if err := s.reportRepo.Update(ctx, report); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "report", id, actorID, "status_changed", map[string]interface{}{
		"new_status": req.Status,
		"resolution": req.Resolution,
	})

	return report, nil
}

func (s *ReportService) AssignReport(ctx context.Context, id uuid.UUID, req *dto.AssignReportRequest, actorID uuid.UUID) (*models.Report, error) {
	report, err := s.reportRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrReportNotFound
	}

	assignedTo, err := uuid.Parse(req.AssignedTo)
	if err != nil {
		return nil, fmt.Errorf("invalid assigned_to")
	}

	report.AssignedTo = &assignedTo
	if err := s.reportRepo.Update(ctx, report); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "report", id, actorID, "assigned", map[string]interface{}{
		"assigned_to": req.AssignedTo,
	})

	return report, nil
}

// Evidence management

func (s *ReportService) UploadEvidence(ctx context.Context, reportID uuid.UUID, fileName, mimeType string, sizeBytes int64, fileContent io.Reader, actorID uuid.UUID) (*models.ReportEvidence, error) {
	if _, err := s.reportRepo.GetByID(ctx, reportID); err != nil {
		return nil, ErrReportNotFound
	}

	storageDir := filepath.Join(s.cfg.StoragePath, "reports", reportID.String())
	if err := os.MkdirAll(storageDir, 0750); err != nil {
		return nil, fmt.Errorf("creating storage directory: %w", err)
	}

	evidenceID := uuid.New()
	storagePath := filepath.Join(storageDir, evidenceID.String())

	f, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("creating file: %w", err)
	}

	hasher := sha256.New()
	written, err := io.Copy(f, io.TeeReader(fileContent, hasher))
	f.Close()
	if err != nil {
		os.Remove(storagePath)
		return nil, fmt.Errorf("writing file: %w", err)
	}

	sha256Hash := hex.EncodeToString(hasher.Sum(nil))

	evidence := &models.ReportEvidence{
		ID:          evidenceID,
		ReportID:    reportID,
		FileName:    fileName,
		MimeType:    mimeType,
		SizeBytes:   written,
		SHA256Hash:  sha256Hash,
		StoragePath: storagePath,
		UploadedBy:  actorID,
		CreatedAt:   time.Now(),
	}

	if err := s.reportRepo.CreateEvidence(ctx, evidence); err != nil {
		os.Remove(storagePath)
		return nil, err
	}

	s.audit.Log(ctx, "report_evidence", evidenceID, actorID, "uploaded", map[string]interface{}{
		"report_id":   reportID,
		"file_name":   fileName,
		"sha256_hash": sha256Hash,
	})

	return evidence, nil
}

func (s *ReportService) ListEvidence(ctx context.Context, reportID uuid.UUID) ([]models.ReportEvidence, error) {
	return s.reportRepo.ListEvidence(ctx, reportID)
}

func (s *ReportService) GetEvidenceFile(ctx context.Context, evidenceID uuid.UUID) (*models.ReportEvidence, error) {
	e, err := s.reportRepo.GetEvidenceByID(ctx, evidenceID)
	if err != nil {
		return nil, ErrEvidenceNotFound
	}
	return e, nil
}

// Notes

func (s *ReportService) AddNote(ctx context.Context, reportID uuid.UUID, req *dto.AddNoteRequest, authorID uuid.UUID) (*models.ReportNote, error) {
	if req.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	if _, err := s.reportRepo.GetByID(ctx, reportID); err != nil {
		return nil, ErrReportNotFound
	}

	note := &models.ReportNote{
		ID:        uuid.New(),
		ReportID:  reportID,
		AuthorID:  authorID,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	if err := s.reportRepo.CreateNote(ctx, note); err != nil {
		return nil, err
	}

	s.audit.Log(ctx, "report_note", note.ID, authorID, "added", map[string]interface{}{
		"report_id": reportID,
	})

	return note, nil
}

func (s *ReportService) ListNotes(ctx context.Context, reportID uuid.UUID) ([]models.ReportNote, error) {
	return s.reportRepo.ListNotes(ctx, reportID)
}
