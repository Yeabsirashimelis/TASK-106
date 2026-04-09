package models

import (
	"time"

	"github.com/google/uuid"
)

type Venue struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Location  *string   `db:"location" json:"location,omitempty"`
	Capacity  *int      `db:"capacity" json:"capacity,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
