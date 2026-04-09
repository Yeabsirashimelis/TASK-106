package unit_tests

import (
	"strings"
	"testing"

	"github.com/eaglepoint/authapi/internal/service"
)

func TestValidatePassword_ValidPasswords(t *testing.T) {
	valid := []string{
		"Abcdefghij1!",
		"MyStr0ngPass!",
		"A1aaaaaaaaaa",
		"Zzzzzzzzzz9z",
		"Test12345678",
		"UpperLower1xxxx",
	}
	for _, pw := range valid {
		if err := service.ValidatePassword(pw); err != nil {
			t.Errorf("expected password %q to be valid, got error: %v", pw, err)
		}
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	err := service.ValidatePassword("Ab1cdefgh")
	if err == nil {
		t.Fatal("expected error for short password")
	}
	if !strings.Contains(err.Error(), "12 characters") {
		t.Errorf("expected '12 characters' in error, got: %v", err)
	}
}

func TestValidatePassword_NoUppercase(t *testing.T) {
	err := service.ValidatePassword("abcdefghij1x")
	if err == nil {
		t.Fatal("expected error for no uppercase")
	}
	if !strings.Contains(err.Error(), "uppercase") {
		t.Errorf("expected 'uppercase' in error, got: %v", err)
	}
}

func TestValidatePassword_NoLowercase(t *testing.T) {
	err := service.ValidatePassword("ABCDEFGHIJ1X")
	if err == nil {
		t.Fatal("expected error for no lowercase")
	}
	if !strings.Contains(err.Error(), "lowercase") {
		t.Errorf("expected 'lowercase' in error, got: %v", err)
	}
}

func TestValidatePassword_NoDigit(t *testing.T) {
	err := service.ValidatePassword("Abcdefghijkl")
	if err == nil {
		t.Fatal("expected error for no digit")
	}
	if !strings.Contains(err.Error(), "number") {
		t.Errorf("expected 'number' in error, got: %v", err)
	}
}

func TestValidatePassword_Empty(t *testing.T) {
	err := service.ValidatePassword("")
	if err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestValidatePassword_MultipleViolations(t *testing.T) {
	err := service.ValidatePassword("short")
	if err == nil {
		t.Fatal("expected error for multiple violations")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, "12 characters") {
		t.Error("expected '12 characters' in error")
	}
	if !strings.Contains(errMsg, "uppercase") {
		t.Error("expected 'uppercase' in error")
	}
	if !strings.Contains(errMsg, "number") {
		t.Error("expected 'number' in error")
	}
}

func TestValidatePassword_ExactlyMinLength(t *testing.T) {
	// Exactly 12 characters with all requirements
	err := service.ValidatePassword("Abcdefghij1k")
	if err != nil {
		t.Errorf("expected 12-char password to be valid, got: %v", err)
	}
}

func TestValidatePassword_OnlyElevenChars(t *testing.T) {
	err := service.ValidatePassword("Abcdefghi1k")
	if err == nil {
		t.Fatal("expected error for 11-char password")
	}
}
