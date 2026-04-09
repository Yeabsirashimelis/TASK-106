package models

import (
	"time"

	"github.com/google/uuid"
)

type Device struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	AccountID       uuid.UUID  `db:"account_id" json:"account_id"`
	FingerprintHash string     `db:"fingerprint_hash" json:"-"`
	DeviceName      *string    `db:"device_name" json:"device_name,omitempty"`
	LastSeenAt      time.Time  `db:"last_seen_at" json:"last_seen_at"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
}
