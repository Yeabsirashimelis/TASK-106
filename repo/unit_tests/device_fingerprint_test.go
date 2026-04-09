package unit_tests

import (
	"testing"

	"github.com/eaglepoint/authapi/internal/config"
	"github.com/eaglepoint/authapi/internal/service"
)

func TestComputeFingerprint_Deterministic(t *testing.T) {
	cfg := &config.Config{DeviceFingerprintSalt: "test-salt"}
	ds := service.NewDeviceService(nil, cfg)

	attrs := map[string]string{
		"screen": "1920x1080",
		"lang":   "en-US",
	}

	fp1 := ds.ComputeFingerprint("Mozilla/5.0", attrs)
	fp2 := ds.ComputeFingerprint("Mozilla/5.0", attrs)

	if fp1 != fp2 {
		t.Error("same inputs should produce same fingerprint")
	}
	if len(fp1) != 64 {
		t.Errorf("expected 64-char hex hash, got %d chars", len(fp1))
	}
}

func TestComputeFingerprint_DifferentInputs(t *testing.T) {
	cfg := &config.Config{DeviceFingerprintSalt: "test-salt"}
	ds := service.NewDeviceService(nil, cfg)

	attrs := map[string]string{"screen": "1920x1080"}

	fp1 := ds.ComputeFingerprint("Mozilla/5.0", attrs)
	fp2 := ds.ComputeFingerprint("Chrome/120", attrs)

	if fp1 == fp2 {
		t.Error("different user agents should produce different fingerprints")
	}
}

func TestComputeFingerprint_DifferentSalt(t *testing.T) {
	cfg1 := &config.Config{DeviceFingerprintSalt: "salt-1"}
	cfg2 := &config.Config{DeviceFingerprintSalt: "salt-2"}
	ds1 := service.NewDeviceService(nil, cfg1)
	ds2 := service.NewDeviceService(nil, cfg2)

	attrs := map[string]string{"screen": "1920x1080"}

	fp1 := ds1.ComputeFingerprint("Mozilla/5.0", attrs)
	fp2 := ds2.ComputeFingerprint("Mozilla/5.0", attrs)

	if fp1 == fp2 {
		t.Error("different salts should produce different fingerprints")
	}
}

func TestComputeFingerprint_EmptyAttributes(t *testing.T) {
	cfg := &config.Config{DeviceFingerprintSalt: "test-salt"}
	ds := service.NewDeviceService(nil, cfg)

	fp := ds.ComputeFingerprint("Mozilla/5.0", nil)
	if fp == "" {
		t.Error("fingerprint should not be empty even with nil attributes")
	}
	if len(fp) != 64 {
		t.Errorf("expected 64-char hex hash, got %d", len(fp))
	}
}

func TestComputeFingerprint_AttributeOrderIndependent(t *testing.T) {
	cfg := &config.Config{DeviceFingerprintSalt: "test-salt"}
	ds := service.NewDeviceService(nil, cfg)

	// Go maps are unordered, but the function sorts keys, so same content = same hash
	attrs1 := map[string]string{"a": "1", "b": "2", "c": "3"}
	attrs2 := map[string]string{"c": "3", "a": "1", "b": "2"}

	fp1 := ds.ComputeFingerprint("Mozilla/5.0", attrs1)
	fp2 := ds.ComputeFingerprint("Mozilla/5.0", attrs2)

	if fp1 != fp2 {
		t.Error("attribute order should not affect fingerprint")
	}
}
