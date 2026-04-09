package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	SeasonID  uuid.UUID `db:"season_id" json:"season_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
