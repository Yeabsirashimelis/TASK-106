package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentStatus string

const (
	PaymentObligation PaymentStatus = "Obligation"
	PaymentSettled    PaymentStatus = "Settled"
	PaymentFailed     PaymentStatus = "Failed"
)

var ValidPaymentStatuses = map[PaymentStatus]bool{
	PaymentObligation: true,
	PaymentSettled:    true,
	PaymentFailed:     true,
}

func (s *PaymentStatus) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("payment status cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("payment status must be a string")
		}
		sv = string(bv)
	}
	*s = PaymentStatus(sv)
	return nil
}

func (s PaymentStatus) Value() (driver.Value, error) {
	return string(s), nil
}

type PaymentChannel string

const (
	ChannelCash             PaymentChannel = "Cash"
	ChannelCheck            PaymentChannel = "Check"
	ChannelWireTransfer     PaymentChannel = "Wire Transfer"
	ChannelInternalTransfer PaymentChannel = "Internal Transfer"
	ChannelJournalEntry     PaymentChannel = "Journal Entry"
)

var ValidPaymentChannels = map[PaymentChannel]bool{
	ChannelCash:             true,
	ChannelCheck:            true,
	ChannelWireTransfer:     true,
	ChannelInternalTransfer: true,
	ChannelJournalEntry:     true,
}

func (c *PaymentChannel) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("payment channel cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("payment channel must be a string")
		}
		sv = string(bv)
	}
	*c = PaymentChannel(sv)
	return nil
}

func (c PaymentChannel) Value() (driver.Value, error) {
	return string(c), nil
}

type PaymentLedgerEntry struct {
	ID             uuid.UUID       `db:"id" json:"id"`
	AccountID      uuid.UUID       `db:"account_id" json:"account_id"`
	IdempotencyKey string          `db:"idempotency_key" json:"idempotency_key"`
	AmountUSD      decimal.Decimal `db:"amount_usd" json:"amount_usd"`
	Description    *string         `db:"description" json:"description,omitempty"`
	Channel        PaymentChannel  `db:"channel" json:"channel"`
	Status         PaymentStatus   `db:"status" json:"status"`
	FinanceClerkID *uuid.UUID      `db:"finance_clerk_id" json:"finance_clerk_id,omitempty"`
	RetryCount     int             `db:"retry_count" json:"retry_count"`
	ReferenceType  *string         `db:"reference_type" json:"reference_type,omitempty"`
	ReferenceID    *uuid.UUID      `db:"reference_id" json:"reference_id,omitempty"`
	CreatedAt            time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time       `db:"updated_at" json:"updated_at"`
	SettledAt            *time.Time      `db:"settled_at" json:"settled_at,omitempty"`
	IdempotencyExpiresAt *time.Time      `db:"idempotency_expires_at" json:"-"`
}

const MaxSettlementRetries = 3

type IdempotencyKey struct {
	ID             uuid.UUID `db:"id"`
	AccountID      uuid.UUID `db:"account_id"`
	IdempotencyKey string    `db:"idempotency_key"`
	PaymentID      uuid.UUID `db:"payment_id"`
	WindowStart    time.Time `db:"window_start"`
	WindowEnd      time.Time `db:"window_end"`
	CreatedAt      time.Time `db:"created_at"`
}
