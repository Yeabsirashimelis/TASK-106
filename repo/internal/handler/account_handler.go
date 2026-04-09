package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{accountService: accountService}
}

func (h *AccountHandler) Create(c echo.Context) error {
	var req dto.CreateAccountRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Username == "" || req.Password == "" || req.Role == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "username, password, and role are required")
	}

	account, err := h.accountService.CreateAccount(c.Request().Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrPasswordPolicy) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if errors.Is(err, service.ErrInvalidRole) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if errors.Is(err, service.ErrDuplicateUsername) {
			return echo.NewHTTPError(http.StatusConflict, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToAccountResponse(account))
}

func (h *AccountHandler) List(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	accounts, err := h.accountService.ListAccounts(c.Request().Context(), offset, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToAccountResponseList(accounts))
}

func (h *AccountHandler) Get(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account id")
	}

	// Allow self-access: any authenticated user can view their own account
	callerID, _ := c.Get("account_id").(uuid.UUID)
	callerRole, _ := c.Get("role").(models.Role)
	if callerID != id && callerRole != models.RoleAdministrator {
		return echo.NewHTTPError(http.StatusForbidden, "insufficient permissions")
	}

	account, err := h.accountService.GetAccount(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "account not found")
	}

	return c.JSON(http.StatusOK, dto.ToAccountResponse(account))
}

func (h *AccountHandler) UpdateStatus(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account id")
	}

	var req dto.UpdateStatusRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Status == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "status is required")
	}

	if err := h.accountService.UpdateStatus(c.Request().Context(), id, models.Status(req.Status)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "status updated"})
}

func (h *AccountHandler) ChangePassword(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid account id")
	}

	// Only the account owner can change their own password
	callerID, _ := c.Get("account_id").(uuid.UUID)
	if callerID != id {
		return echo.NewHTTPError(http.StatusForbidden, "you can only change your own password")
	}

	var req dto.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "old_password and new_password are required")
	}

	if err := h.accountService.ChangePassword(c.Request().Context(), id, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return echo.NewHTTPError(http.StatusUnauthorized, "incorrect current password")
		}
		if errors.Is(err, service.ErrPasswordPolicy) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "password changed successfully"})
}
