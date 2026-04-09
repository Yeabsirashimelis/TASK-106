package dto

import (
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
)

// Dictionary DTOs

type CreateDictionaryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type UpdateDictionaryRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

type DictionaryResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToDictionaryResponse(d *models.SensitiveWordDictionary) DictionaryResponse {
	return DictionaryResponse{
		ID:          d.ID,
		Name:        d.Name,
		Description: d.Description,
		IsActive:    d.IsActive,
		CreatedBy:   d.CreatedBy,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func ToDictionaryResponseList(dicts []models.SensitiveWordDictionary) []DictionaryResponse {
	result := make([]DictionaryResponse, len(dicts))
	for i, d := range dicts {
		result[i] = ToDictionaryResponse(&d)
	}
	return result
}

// Word DTOs

type AddWordRequest struct {
	Word     string `json:"word"`
	Severity string `json:"severity"`
}

type AddWordsRequest struct {
	Words []AddWordRequest `json:"words"`
}

type WordResponse struct {
	ID           uuid.UUID `json:"id"`
	DictionaryID uuid.UUID `json:"dictionary_id"`
	Word         string    `json:"word"`
	Severity     string    `json:"severity"`
	CreatedAt    time.Time `json:"created_at"`
}

func ToWordResponse(w *models.SensitiveWord) WordResponse {
	return WordResponse{
		ID:           w.ID,
		DictionaryID: w.DictionaryID,
		Word:         w.Word,
		Severity:     w.Severity,
		CreatedAt:    w.CreatedAt,
	}
}

func ToWordResponseList(words []models.SensitiveWord) []WordResponse {
	result := make([]WordResponse, len(words))
	for i, w := range words {
		result[i] = ToWordResponse(&w)
	}
	return result
}

// Content check DTOs

type CheckContentRequest struct {
	Text string `json:"text"`
}

type ContentMatch struct {
	Word     string `json:"word"`
	Severity string `json:"severity"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Context  string `json:"context"`
}

type CheckContentResponse struct {
	Clean   bool           `json:"clean"`
	Matches []ContentMatch `json:"matches"`
}

// Review DTOs

type CreateReviewRequest struct {
	ContentType    string  `json:"content_type"`
	ContentID      string  `json:"content_id"`
	ContentSnippet *string `json:"content_snippet,omitempty"`
}

type DecideReviewRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

type ReviewResponse struct {
	ID             uuid.UUID  `json:"id"`
	ContentType    string     `json:"content_type"`
	ContentID      uuid.UUID  `json:"content_id"`
	ContentSnippet *string    `json:"content_snippet,omitempty"`
	Status         string     `json:"status"`
	ModeratorID    *uuid.UUID `json:"moderator_id,omitempty"`
	Reason         *string    `json:"reason,omitempty"`
	DecidedAt      *time.Time `json:"decided_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func ToReviewResponse(r *models.ModerationReview) ReviewResponse {
	return ReviewResponse{
		ID:             r.ID,
		ContentType:    r.ContentType,
		ContentID:      r.ContentID,
		ContentSnippet: r.ContentSnippet,
		Status:         string(r.Status),
		ModeratorID:    r.ModeratorID,
		Reason:         r.Reason,
		DecidedAt:      r.DecidedAt,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

func ToReviewResponseList(reviews []models.ModerationReview) []ReviewResponse {
	result := make([]ReviewResponse, len(reviews))
	for i, r := range reviews {
		result[i] = ToReviewResponse(&r)
	}
	return result
}

// Report DTOs

type CreateReportRequest struct {
	TargetType  string `json:"target_type"`
	TargetID    string `json:"target_id"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

type UpdateReportStatusRequest struct {
	Status     string  `json:"status"`
	Resolution *string `json:"resolution,omitempty"`
}

type AssignReportRequest struct {
	AssignedTo string `json:"assigned_to"`
}

type AddNoteRequest struct {
	Content string `json:"content"`
}

type ReportResponse struct {
	ID          uuid.UUID  `json:"id"`
	ReporterID  uuid.UUID  `json:"reporter_id"`
	TargetType  string     `json:"target_type"`
	TargetID    uuid.UUID  `json:"target_id"`
	Category    string     `json:"category"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	AssignedTo  *uuid.UUID `json:"assigned_to,omitempty"`
	Resolution  *string    `json:"resolution,omitempty"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func ToReportResponse(r *models.Report) ReportResponse {
	return ReportResponse{
		ID:          r.ID,
		ReporterID:  r.ReporterID,
		TargetType:  r.TargetType,
		TargetID:    r.TargetID,
		Category:    string(r.Category),
		Description: r.Description,
		Status:      string(r.Status),
		AssignedTo:  r.AssignedTo,
		Resolution:  r.Resolution,
		ResolvedAt:  r.ResolvedAt,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func ToReportResponseList(reports []models.Report) []ReportResponse {
	result := make([]ReportResponse, len(reports))
	for i, r := range reports {
		result[i] = ToReportResponse(&r)
	}
	return result
}

type EvidenceResponse struct {
	ID         uuid.UUID `json:"id"`
	ReportID   uuid.UUID `json:"report_id"`
	FileName   string    `json:"file_name"`
	MimeType   string    `json:"mime_type"`
	SizeBytes  int64     `json:"size_bytes"`
	SHA256Hash string    `json:"sha256_hash"`
	UploadedBy uuid.UUID `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

func ToEvidenceResponse(e *models.ReportEvidence) EvidenceResponse {
	return EvidenceResponse{
		ID:         e.ID,
		ReportID:   e.ReportID,
		FileName:   e.FileName,
		MimeType:   e.MimeType,
		SizeBytes:  e.SizeBytes,
		SHA256Hash: e.SHA256Hash,
		UploadedBy: e.UploadedBy,
		CreatedAt:  e.CreatedAt,
	}
}

func ToEvidenceResponseList(evidence []models.ReportEvidence) []EvidenceResponse {
	result := make([]EvidenceResponse, len(evidence))
	for i, e := range evidence {
		result[i] = ToEvidenceResponse(&e)
	}
	return result
}

type NoteResponse struct {
	ID        uuid.UUID `json:"id"`
	ReportID  uuid.UUID `json:"report_id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func ToNoteResponse(n *models.ReportNote) NoteResponse {
	return NoteResponse{
		ID:        n.ID,
		ReportID:  n.ReportID,
		AuthorID:  n.AuthorID,
		Content:   n.Content,
		CreatedAt: n.CreatedAt,
	}
}

func ToNoteResponseList(notes []models.ReportNote) []NoteResponse {
	result := make([]NoteResponse, len(notes))
	for i, n := range notes {
		result[i] = ToNoteResponse(&n)
	}
	return result
}
