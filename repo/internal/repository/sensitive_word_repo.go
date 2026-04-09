package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type SensitiveWordRepository struct {
	db *sqlx.DB
}

func NewSensitiveWordRepository(db *sqlx.DB) *SensitiveWordRepository {
	return &SensitiveWordRepository{db: db}
}

// Dictionary CRUD

func (r *SensitiveWordRepository) CreateDictionary(ctx context.Context, d *models.SensitiveWordDictionary) error {
	query := `INSERT INTO sensitive_word_dictionaries (id, name, description, is_active, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query, d.ID, d.Name, d.Description, d.IsActive, d.CreatedBy, d.CreatedAt, d.UpdatedAt)
	if err != nil {
		return fmt.Errorf("sensitive_word_repo.CreateDictionary: %w", err)
	}
	return nil
}

func (r *SensitiveWordRepository) GetDictionaryByID(ctx context.Context, id uuid.UUID) (*models.SensitiveWordDictionary, error) {
	var d models.SensitiveWordDictionary
	query := `SELECT id, name, description, is_active, created_by, created_at, updated_at
		FROM sensitive_word_dictionaries WHERE id = $1`
	if err := r.db.QueryRowxContext(ctx, query, id).StructScan(&d); err != nil {
		return nil, fmt.Errorf("sensitive_word_repo.GetDictionaryByID: %w", err)
	}
	return &d, nil
}

func (r *SensitiveWordRepository) ListDictionaries(ctx context.Context, offset, limit int) ([]models.SensitiveWordDictionary, error) {
	var dicts []models.SensitiveWordDictionary
	query := `SELECT id, name, description, is_active, created_by, created_at, updated_at
		FROM sensitive_word_dictionaries ORDER BY name LIMIT $1 OFFSET $2`
	if err := r.db.SelectContext(ctx, &dicts, query, limit, offset); err != nil {
		return nil, fmt.Errorf("sensitive_word_repo.ListDictionaries: %w", err)
	}
	return dicts, nil
}

func (r *SensitiveWordRepository) UpdateDictionary(ctx context.Context, d *models.SensitiveWordDictionary) error {
	query := `UPDATE sensitive_word_dictionaries SET name = $1, description = $2, is_active = $3, updated_at = NOW()
		WHERE id = $4`
	result, err := r.db.ExecContext(ctx, query, d.Name, d.Description, d.IsActive, d.ID)
	if err != nil {
		return fmt.Errorf("sensitive_word_repo.UpdateDictionary: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("sensitive_word_repo.UpdateDictionary: dictionary not found")
	}
	return nil
}

func (r *SensitiveWordRepository) DeleteDictionary(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sensitive_word_dictionaries WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("sensitive_word_repo.DeleteDictionary: %w", err)
	}
	return nil
}

// Word CRUD

func (r *SensitiveWordRepository) AddWord(ctx context.Context, w *models.SensitiveWord) error {
	query := `INSERT INTO sensitive_words (id, dictionary_id, word, severity, created_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, w.ID, w.DictionaryID, w.Word, w.Severity, w.CreatedAt)
	if err != nil {
		return fmt.Errorf("sensitive_word_repo.AddWord: %w", err)
	}
	return nil
}

func (r *SensitiveWordRepository) ListWords(ctx context.Context, dictionaryID uuid.UUID, offset, limit int) ([]models.SensitiveWord, error) {
	var words []models.SensitiveWord
	query := `SELECT id, dictionary_id, word, severity, created_at
		FROM sensitive_words WHERE dictionary_id = $1 ORDER BY word LIMIT $2 OFFSET $3`
	if err := r.db.SelectContext(ctx, &words, query, dictionaryID, limit, offset); err != nil {
		return nil, fmt.Errorf("sensitive_word_repo.ListWords: %w", err)
	}
	return words, nil
}

func (r *SensitiveWordRepository) DeleteWord(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sensitive_words WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("sensitive_word_repo.DeleteWord: %w", err)
	}
	return nil
}

// GetAllActiveWords loads all words from active dictionaries for matching
func (r *SensitiveWordRepository) GetAllActiveWords(ctx context.Context) ([]models.SensitiveWord, error) {
	var words []models.SensitiveWord
	query := `SELECT w.id, w.dictionary_id, w.word, w.severity, w.created_at
		FROM sensitive_words w
		JOIN sensitive_word_dictionaries d ON d.id = w.dictionary_id
		WHERE d.is_active = TRUE
		ORDER BY LENGTH(w.word) DESC`
	if err := r.db.SelectContext(ctx, &words, query); err != nil {
		return nil, fmt.Errorf("sensitive_word_repo.GetAllActiveWords: %w", err)
	}
	return words, nil
}
