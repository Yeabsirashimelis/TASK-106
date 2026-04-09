package models

import (
	"time"

	"github.com/google/uuid"
)

type LoginAttempt struct {
	ID         uuid.UUID `db:"id"`
	AccountID  uuid.UUID `db:"account_id"`
	Success    bool      `db:"success"`
	IPAddress  *string   `db:"ip_address"`
	AttemptedAt time.Time `db:"attempted_at"`
}
