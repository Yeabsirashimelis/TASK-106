package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/eaglepoint/authapi/internal/config"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/repository"
	"github.com/google/uuid"
)

type DeviceService struct {
	repo *repository.DeviceRepository
	cfg  *config.Config
}

func NewDeviceService(repo *repository.DeviceRepository, cfg *config.Config) *DeviceService {
	return &DeviceService{repo: repo, cfg: cfg}
}

func (s *DeviceService) ComputeFingerprint(userAgent string, attributes map[string]string) string {
	var parts []string
	parts = append(parts, s.cfg.DeviceFingerprintSalt)
	parts = append(parts, userAgent)

	keys := make([]string, 0, len(attributes))
	for k := range attributes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, attributes[k]))
	}

	data := strings.Join(parts, "|")
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

func (s *DeviceService) RegisterOrUpdateDevice(ctx context.Context, accountID uuid.UUID, fingerprintHash string, deviceName *string) (*models.Device, error) {
	now := time.Now()
	device := &models.Device{
		ID:              uuid.New(),
		AccountID:       accountID,
		FingerprintHash: fingerprintHash,
		DeviceName:      deviceName,
		LastSeenAt:      now,
		CreatedAt:       now,
	}
	return s.repo.Upsert(ctx, device)
}

func (s *DeviceService) ListDevices(ctx context.Context, accountID uuid.UUID) ([]models.Device, error) {
	return s.repo.ListByAccount(ctx, accountID)
}
