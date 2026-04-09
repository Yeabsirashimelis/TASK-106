package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SeasonStatus string

const (
	SeasonPlanning  SeasonStatus = "Planning"
	SeasonActive    SeasonStatus = "Active"
	SeasonCompleted SeasonStatus = "Completed"
	SeasonArchived  SeasonStatus = "Archived"
)

var ValidSeasonStatuses = map[SeasonStatus]bool{
	SeasonPlanning:  true,
	SeasonActive:    true,
	SeasonCompleted: true,
	SeasonArchived:  true,
}

func (s *SeasonStatus) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("season status cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("season status must be a string")
		}
		sv = string(bv)
	}
	*s = SeasonStatus(sv)
	return nil
}

func (s SeasonStatus) Value() (driver.Value, error) {
	return string(s), nil
}

type Season struct {
	ID        uuid.UUID    `db:"id" json:"id"`
	Name      string       `db:"name" json:"name"`
	StartDate time.Time    `db:"start_date" json:"start_date"`
	EndDate   time.Time    `db:"end_date" json:"end_date"`
	Status    SeasonStatus `db:"status" json:"status"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
}
