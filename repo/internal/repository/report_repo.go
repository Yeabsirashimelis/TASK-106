package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ReportRepository struct {
	db *sqlx.DB
}

func NewReportRepository(db *sqlx.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Create(ctx context.Context, report *models.Report) error {
	query := `INSERT INTO reports (id, reporter_id, target_type, target_id, category, description, status, assigned_to, resolution, resolved_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.ExecContext(ctx, query,
		report.ID, report.ReporterID, report.TargetType, report.TargetID,
		report.Category, report.Description, report.Status, report.AssignedTo,
		report.Resolution, report.ResolvedAt, report.CreatedAt, report.UpdatedAt)
	if err != nil {
		return fmt.Errorf("report_repo.Create: %w", err)
	}
	return nil
}

func (r *ReportRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Report, error) {
	var report models.Report
	query := `SELECT id, reporter_id, target_type, target_id, category, description, status, assigned_to, resolution, resolved_at, created_at, updated_at
		FROM reports WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&report); err != nil {
		return nil, fmt.Errorf("report_repo.GetByID: %w", err)
	}
	return &report, nil
}

func (r *ReportRepository) List(ctx context.Context, status *models.ReportStatus, offset, limit int) ([]models.Report, error) {
	var reports []models.Report
	if status != nil {
		query := `SELECT id, reporter_id, target_type, target_id, category, description, status, assigned_to, resolution, resolved_at, created_at, updated_at
			FROM reports WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		if err := r.db.SelectContext(ctx, &reports, query, *status, limit, offset); err != nil {
			return nil, fmt.Errorf("report_repo.List: %w", err)
		}
	} else {
		query := `SELECT id, reporter_id, target_type, target_id, category, description, status, assigned_to, resolution, resolved_at, created_at, updated_at
			FROM reports ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		if err := r.db.SelectContext(ctx, &reports, query, limit, offset); err != nil {
			return nil, fmt.Errorf("report_repo.List: %w", err)
		}
	}
	return reports, nil
}

func (r *ReportRepository) Update(ctx context.Context, report *models.Report) error {
	query := `UPDATE reports SET status = $1, assigned_to = $2, resolution = $3, resolved_at = $4, updated_at = NOW()
		WHERE id = $5`
	result, err := r.db.ExecContext(ctx, query, report.Status, report.AssignedTo, report.Resolution, report.ResolvedAt, report.ID)
	if err != nil {
		return fmt.Errorf("report_repo.Update: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("report_repo.Update: report not found")
	}
	return nil
}

func (r *ReportRepository) CountReporterToday(ctx context.Context, reporterID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM reports WHERE reporter_id = $1 AND created_at >= $2`
	today := time.Now().Truncate(24 * time.Hour)
	if err := r.db.QueryRowxContext(ctx, query, reporterID, today).Scan(&count); err != nil {
		return 0, fmt.Errorf("report_repo.CountReporterToday: %w", err)
	}
	return count, nil
}

// Evidence

func (r *ReportRepository) CreateEvidence(ctx context.Context, e *models.ReportEvidence) error {
	query := `INSERT INTO report_evidence (id, report_id, file_name, mime_type, size_bytes, sha256_hash, storage_path, uploaded_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, query,
		e.ID, e.ReportID, e.FileName, e.MimeType, e.SizeBytes,
		e.SHA256Hash, e.StoragePath, e.UploadedBy, e.CreatedAt)
	if err != nil {
		return fmt.Errorf("report_repo.CreateEvidence: %w", err)
	}
	return nil
}

func (r *ReportRepository) ListEvidence(ctx context.Context, reportID uuid.UUID) ([]models.ReportEvidence, error) {
	var evidence []models.ReportEvidence
	query := `SELECT id, report_id, file_name, mime_type, size_bytes, sha256_hash, storage_path, uploaded_by, created_at
		FROM report_evidence WHERE report_id = $1 ORDER BY created_at`
	if err := r.db.SelectContext(ctx, &evidence, query, reportID); err != nil {
		return nil, fmt.Errorf("report_repo.ListEvidence: %w", err)
	}
	return evidence, nil
}

func (r *ReportRepository) GetEvidenceByID(ctx context.Context, id uuid.UUID) (*models.ReportEvidence, error) {
	var e models.ReportEvidence
	query := `SELECT id, report_id, file_name, mime_type, size_bytes, sha256_hash, storage_path, uploaded_by, created_at
		FROM report_evidence WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&e); err != nil {
		return nil, fmt.Errorf("report_repo.GetEvidenceByID: %w", err)
	}
	return &e, nil
}

// Notes

func (r *ReportRepository) CreateNote(ctx context.Context, n *models.ReportNote) error {
	query := `INSERT INTO report_notes (id, report_id, author_id, content, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, n.ID, n.ReportID, n.AuthorID, n.Content, n.CreatedAt)
	if err != nil {
		return fmt.Errorf("report_repo.CreateNote: %w", err)
	}
	return nil
}

func (r *ReportRepository) ListNotes(ctx context.Context, reportID uuid.UUID) ([]models.ReportNote, error) {
	var notes []models.ReportNote
	query := `SELECT id, report_id, author_id, content, created_at
		FROM report_notes WHERE report_id = $1 ORDER BY created_at`
	if err := r.db.SelectContext(ctx, &notes, query, reportID); err != nil {
		return nil, fmt.Errorf("report_repo.ListNotes: %w", err)
	}
	return notes, nil
}
