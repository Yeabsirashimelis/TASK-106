package repository

import (
	"context"
	"fmt"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type DeviceRepository struct {
	db *sqlx.DB
}

func NewDeviceRepository(db *sqlx.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) Upsert(ctx context.Context, device *models.Device) (*models.Device, error) {
	query := `INSERT INTO devices (id, account_id, fingerprint_hash, device_name, last_seen_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (account_id, fingerprint_hash) DO UPDATE SET
			last_seen_at = EXCLUDED.last_seen_at,
			device_name = COALESCE(EXCLUDED.device_name, devices.device_name)
		RETURNING id, account_id, fingerprint_hash, device_name, last_seen_at, created_at`
	var result models.Device
	if err := r.db.QueryRowxContext(ctx, query,
		device.ID, device.AccountID, device.FingerprintHash,
		device.DeviceName, device.LastSeenAt, device.CreatedAt,
	).StructScan(&result); err != nil {
		return nil, fmt.Errorf("device_repo.Upsert: %w", err)
	}
	return &result, nil
}

func (r *DeviceRepository) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]models.Device, error) {
	var devices []models.Device
	query := `SELECT id, account_id, fingerprint_hash, device_name, last_seen_at, created_at
		FROM devices WHERE account_id = $1 ORDER BY last_seen_at DESC`
	if err := r.db.SelectContext(ctx, &devices, query, accountID); err != nil {
		return nil, fmt.Errorf("device_repo.ListByAccount: %w", err)
	}
	return devices, nil
}

func (r *DeviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM devices WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("device_repo.Delete: %w", err)
	}
	return nil
}
