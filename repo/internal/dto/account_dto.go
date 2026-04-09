package dto

import (
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
)

type CreateAccountRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type AccountResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToAccountResponse(a *models.Account) AccountResponse {
	return AccountResponse{
		ID:        a.ID,
		Username:  a.Username,
		Role:      string(a.Role),
		Status:    string(a.Status),
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

func ToAccountResponseList(accounts []models.Account) []AccountResponse {
	result := make([]AccountResponse, len(accounts))
	for i, a := range accounts {
		result[i] = ToAccountResponse(&a)
	}
	return result
}
