package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/eaglepoint/authapi/internal/dto"
	"github.com/eaglepoint/authapi/internal/models"
	"github.com/eaglepoint/authapi/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ResourceHandler struct {
	resourceService *service.ResourceService
}

func NewResourceHandler(resourceService *service.ResourceService) *ResourceHandler {
	return &ResourceHandler{resourceService: resourceService}
}

func (h *ResourceHandler) CreateResource(c echo.Context) error {
	var req dto.CreateResourceRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	callerRole, _ := c.Get("role").(models.Role)
	resource, tags, err := h.resourceService.CreateResource(c.Request().Context(), &req, actorID, callerRole)
	if err != nil {
		if errors.Is(err, service.ErrResourceAccessDenied) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		if errors.Is(err, service.ErrTooManyTags) || errors.Is(err, service.ErrTagTooLong) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToResourceResponse(resource, tags))
}

func (h *ResourceHandler) GetResource(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid resource id")
	}

	callerID, _ := c.Get("account_id").(uuid.UUID)
	callerRole, _ := c.Get("role").(models.Role)

	resource, tags, err := h.resourceService.GetResource(c.Request().Context(), id, callerID, callerRole)
	if err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrResourceAccessDenied) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToResourceResponse(resource, tags))
}

func (h *ResourceHandler) ListResources(c echo.Context) error {
	courseID, err := uuid.Parse(c.QueryParam("course_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "course_id query parameter is required")
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	callerID, _ := c.Get("account_id").(uuid.UUID)
	callerRole, _ := c.Get("role").(models.Role)

	resources, err := h.resourceService.ListResources(c.Request().Context(), courseID, callerID, callerRole, offset, limit)
	if err != nil {
		if errors.Is(err, service.ErrNotCourseMember) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	// Build responses (without tags for list view for performance)
	result := make([]dto.ResourceResponse, len(resources))
	for i, r := range resources {
		result[i] = dto.ToResourceResponse(&r, nil)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *ResourceHandler) UpdateResource(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid resource id")
	}

	var req dto.UpdateResourceRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	callerRole, _ := c.Get("role").(models.Role)
	resource, tags, err := h.resourceService.UpdateResource(c.Request().Context(), id, &req, actorID, callerRole)
	if err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToResourceResponse(resource, tags))
}

func (h *ResourceHandler) SearchResources(c echo.Context) error {
	courseID, err := uuid.Parse(c.QueryParam("course_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "course_id query parameter is required")
	}

	query := c.QueryParam("q")
	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "q query parameter is required")
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	callerID, _ := c.Get("account_id").(uuid.UUID)
	callerRole, _ := c.Get("role").(models.Role)

	resources, err := h.resourceService.SearchResources(c.Request().Context(), courseID, query, callerID, callerRole, offset, limit)
	if err != nil {
		if errors.Is(err, service.ErrNotCourseMember) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result := make([]dto.ResourceResponse, len(resources))
	for i, r := range resources {
		result[i] = dto.ToResourceResponse(&r, nil)
	}
	return c.JSON(http.StatusOK, result)
}

func (h *ResourceHandler) UploadVersion(c echo.Context) error {
	resourceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid resource id")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "file is required")
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to read uploaded file")
	}
	defer src.Close()

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Extract text from request if provided (for PDF/DOCX)
	extractedText := c.FormValue("extracted_text")
	var extractedTextPtr *string
	if extractedText != "" {
		extractedTextPtr = &extractedText
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	callerRole2, _ := c.Get("role").(models.Role)
	version, err := h.resourceService.UploadVersion(c.Request().Context(), resourceID, file.Filename, mimeType, file.Size, src, extractedTextPtr, actorID, callerRole2)
	if err != nil {
		if errors.Is(err, service.ErrResourceNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrMimeTypeNotAllowed) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if errors.Is(err, service.ErrExtractedTextNotAllowed) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if errors.Is(err, service.ErrResourceAccessDenied) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToVersionResponse(version))
}

func (h *ResourceHandler) ListVersions(c echo.Context) error {
	resourceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid resource id")
	}

	callerID, _ := c.Get("account_id").(uuid.UUID)
	callerRole, _ := c.Get("role").(models.Role)

	versions, err := h.resourceService.ListVersions(c.Request().Context(), resourceID, callerID, callerRole)
	if err != nil {
		if errors.Is(err, service.ErrResourceAccessDenied) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToVersionResponseList(versions))
}

func (h *ResourceHandler) DownloadVersion(c echo.Context) error {
	versionID, err := uuid.Parse(c.Param("version_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid version id")
	}

	callerID, _ := c.Get("account_id").(uuid.UUID)
	callerRole, _ := c.Get("role").(models.Role)

	version, err := h.resourceService.GetDownloadPath(c.Request().Context(), versionID, callerID, callerRole)
	if err != nil {
		if errors.Is(err, service.ErrVersionNotFound) || errors.Is(err, service.ErrResourceNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrResourceAccessDenied) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	// Verify file integrity before serving
	f, err := os.Open(version.StoragePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "file not found on disk")
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to verify file integrity")
	}
	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != version.SHA256Hash {
		return echo.NewHTTPError(http.StatusInternalServerError, "file integrity check failed")
	}

	// Seek back to start for serving
	f.Seek(0, io.SeekStart)

	c.Response().Header().Set("Content-Type", version.MimeType)
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, version.FileName))
	c.Response().Header().Set("X-Content-SHA256", version.SHA256Hash)
	return c.Stream(http.StatusOK, version.MimeType, f)
}

func (h *ResourceHandler) PreviewVersion(c echo.Context) error {
	versionID, err := uuid.Parse(c.Param("version_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid version id")
	}

	callerID, _ := c.Get("account_id").(uuid.UUID)
	callerRole, _ := c.Get("role").(models.Role)

	version, err := h.resourceService.GetDownloadPath(c.Request().Context(), versionID, callerID, callerRole)
	if err != nil {
		if errors.Is(err, service.ErrVersionNotFound) || errors.Is(err, service.ErrResourceNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrResourceAccessDenied) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	// Inline display instead of download
	f, err := os.Open(version.StoragePath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "file not found on disk")
	}
	defer f.Close()

	c.Response().Header().Set("Content-Type", version.MimeType)
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, version.FileName))
	return c.Stream(http.StatusOK, version.MimeType, f)
}
