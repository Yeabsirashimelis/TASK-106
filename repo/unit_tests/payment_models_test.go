package unit_tests

import (
	"testing"

	"github.com/eaglepoint/authapi/internal/models"
)

func TestPaymentStatusValidation(t *testing.T) {
	valid := []models.PaymentStatus{
		models.PaymentObligation,
		models.PaymentSettled,
		models.PaymentFailed,
	}
	for _, s := range valid {
		if !models.ValidPaymentStatuses[s] {
			t.Errorf("expected %s to be valid", s)
		}
	}

	if models.ValidPaymentStatuses[models.PaymentStatus("Bogus")] {
		t.Error("Bogus should not be valid")
	}
}

func TestPaymentChannelValidation(t *testing.T) {
	valid := []models.PaymentChannel{
		models.ChannelCash,
		models.ChannelCheck,
		models.ChannelWireTransfer,
		models.ChannelInternalTransfer,
		models.ChannelJournalEntry,
	}
	for _, c := range valid {
		if !models.ValidPaymentChannels[c] {
			t.Errorf("expected %s to be valid", c)
		}
	}

	if models.ValidPaymentChannels[models.PaymentChannel("Bitcoin")] {
		t.Error("Bitcoin should not be valid channel")
	}
}

func TestMaxSettlementRetries(t *testing.T) {
	if models.MaxSettlementRetries != 3 {
		t.Errorf("expected MaxSettlementRetries to be 3, got %d", models.MaxSettlementRetries)
	}
}
