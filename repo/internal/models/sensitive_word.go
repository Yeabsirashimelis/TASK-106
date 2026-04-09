package models

import (
	"time"

	"github.com/google/uuid"
)

type SensitiveWordDictionary struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description,omitempty"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedBy   uuid.UUID `db:"created_by" json:"created_by"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type SensitiveWord struct {
	ID           uuid.UUID `db:"id" json:"id"`
	DictionaryID uuid.UUID `db:"dictionary_id" json:"dictionary_id"`
	Word         string    `db:"word" json:"word"`
	Severity     string    `db:"severity" json:"severity"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
