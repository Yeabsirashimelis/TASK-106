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

type CourseHandler struct {
	courseService *service.CourseService
}

func NewCourseHandler(courseService *service.CourseService) *CourseHandler {
	return &CourseHandler{courseService: courseService}
}

func (h *CourseHandler) CreateCourse(c echo.Context) error {
	var req dto.CreateCourseRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	course, err := h.courseService.CreateCourse(c.Request().Context(), &req, actorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToCourseResponse(course))
}

func (h *CourseHandler) GetCourse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid course id")
	}

	callerID, _ := c.Get("account_id").(uuid.UUID)
	callerRole, _ := c.Get("role").(models.Role)

	course, err := h.courseService.GetCourse(c.Request().Context(), id, callerID, callerRole)
	if err != nil {
		if errors.Is(err, service.ErrNotCourseMember) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToCourseResponse(course))
}

func (h *CourseHandler) ListCourses(c echo.Context) error {
	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	callerRole, _ := c.Get("role").(models.Role)

	courses, err := h.courseService.ListCourses(c.Request().Context(), offset, limit, callerRole)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToCourseResponseList(courses))
}

func (h *CourseHandler) UpdateCourse(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid course id")
	}

	var req dto.UpdateCourseRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	course, err := h.courseService.UpdateCourse(c.Request().Context(), id, &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrCourseNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrNotCourseStaff) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToCourseResponse(course))
}

// Outline nodes

func (h *CourseHandler) CreateOutlineNode(c echo.Context) error {
	var req dto.CreateOutlineNodeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	node, err := h.courseService.CreateOutlineNode(c.Request().Context(), &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrCourseNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrNotCourseStaff) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToOutlineNodeResponse(node))
}

func (h *CourseHandler) GetOutlineTree(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("course_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid course_id")
	}

	callerID, _ := c.Get("account_id").(uuid.UUID)
	callerRole, _ := c.Get("role").(models.Role)

	tree, err := h.courseService.GetOutlineTree(c.Request().Context(), courseID, callerID, callerRole)
	if err != nil {
		if errors.Is(err, service.ErrNotCourseMember) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	if tree == nil {
		tree = []dto.OutlineTreeNode{}
	}
	return c.JSON(http.StatusOK, tree)
}

func (h *CourseHandler) UpdateOutlineNode(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node id")
	}

	var req dto.UpdateOutlineNodeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	node, err := h.courseService.UpdateOutlineNode(c.Request().Context(), id, &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrNodeNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrNotCourseStaff) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, dto.ToOutlineNodeResponse(node))
}

func (h *CourseHandler) DeleteOutlineNode(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node id")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	if err := h.courseService.DeleteOutlineNode(c.Request().Context(), id, actorID); err != nil {
		if errors.Is(err, service.ErrNodeNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrNotCourseStaff) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// Memberships

func (h *CourseHandler) AddMember(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("course_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid course_id")
	}

	var req dto.AddMemberRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	membership, err := h.courseService.AddMember(c.Request().Context(), courseID, &req, actorID)
	if err != nil {
		if errors.Is(err, service.ErrNotCourseStaff) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.ToMembershipResponse(membership))
}

func (h *CourseHandler) ListMembers(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("course_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid course_id")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	callerRole, _ := c.Get("role").(models.Role)

	memberships, err := h.courseService.ListMembers(c.Request().Context(), courseID, actorID, callerRole)
	if err != nil {
		if errors.Is(err, service.ErrNotCourseStaff) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, dto.ToMembershipResponseList(memberships))
}

func (h *CourseHandler) RemoveMember(c echo.Context) error {
	membershipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid membership id")
	}
	courseID, err := uuid.Parse(c.Param("course_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid course_id")
	}

	actorID, _ := c.Get("account_id").(uuid.UUID)
	callerRole, _ := c.Get("role").(models.Role)

	if err := h.courseService.RemoveMember(c.Request().Context(), membershipID, actorID, callerRole, courseID); err != nil {
		if errors.Is(err, service.ErrNotCourseStaff) {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
