package models

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID  `db:"id"`
	AccountID uuid.UUID  `db:"account_id"`
	TokenHash string     `db:"token_hash"`
	DeviceID  *uuid.UUID `db:"device_id"`
	ExpiresAt time.Time  `db:"expires_at"`
	Revoked   bool       `db:"revoked"`
	CreatedAt time.Time  `db:"created_at"`
}
