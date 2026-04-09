package models

import (
	"time"

	"github.com/google/uuid"
)

type ResourceTag struct {
	ID         uuid.UUID `db:"id" json:"id"`
	ResourceID uuid.UUID `db:"resource_id" json:"resource_id"`
	Tag        string    `db:"tag" json:"tag"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}
