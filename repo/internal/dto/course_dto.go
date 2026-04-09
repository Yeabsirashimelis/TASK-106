package dto

import (
	"time"

	"github.com/eaglepoint/authapi/internal/models"
	"github.com/google/uuid"
)

type CreateCourseRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
}

type UpdateCourseRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type CourseResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	Status      string    `json:"status"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToCourseResponse(c *models.Course) CourseResponse {
	return CourseResponse{
		ID:          c.ID,
		Title:       c.Title,
		Description: c.Description,
		Status:      string(c.Status),
		CreatedBy:   c.CreatedBy,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func ToCourseResponseList(courses []models.Course) []CourseResponse {
	result := make([]CourseResponse, len(courses))
	for i, c := range courses {
		result[i] = ToCourseResponse(&c)
	}
	return result
}

// Outline nodes

type CreateOutlineNodeRequest struct {
	CourseID    string  `json:"course_id"`
	ParentID    *string `json:"parent_id,omitempty"`
	NodeType    string  `json:"node_type"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	OrderIndex  int     `json:"order_index"`
}

type UpdateOutlineNodeRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	OrderIndex  *int    `json:"order_index,omitempty"`
	ParentID    *string `json:"parent_id,omitempty"`
}

type OutlineNodeResponse struct {
	ID          uuid.UUID  `json:"id"`
	CourseID    uuid.UUID  `json:"course_id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	NodeType    string     `json:"node_type"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	OrderIndex  int        `json:"order_index"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func ToOutlineNodeResponse(n *models.CourseOutlineNode) OutlineNodeResponse {
	return OutlineNodeResponse{
		ID:          n.ID,
		CourseID:    n.CourseID,
		ParentID:    n.ParentID,
		NodeType:    string(n.NodeType),
		Title:       n.Title,
		Description: n.Description,
		OrderIndex:  n.OrderIndex,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
}

func ToOutlineNodeResponseList(nodes []models.CourseOutlineNode) []OutlineNodeResponse {
	result := make([]OutlineNodeResponse, len(nodes))
	for i, n := range nodes {
		result[i] = ToOutlineNodeResponse(&n)
	}
	return result
}

// Tree structure for returning hierarchical outline

type OutlineTreeNode struct {
	OutlineNodeResponse
	Children []OutlineTreeNode `json:"children,omitempty"`
}

func BuildOutlineTree(nodes []models.CourseOutlineNode) []OutlineTreeNode {
	nodeMap := make(map[uuid.UUID]*OutlineTreeNode)
	var roots []OutlineTreeNode

	// Create all tree nodes first
	for i := range nodes {
		n := &nodes[i]
		tn := OutlineTreeNode{
			OutlineNodeResponse: ToOutlineNodeResponse(n),
		}
		nodeMap[n.ID] = &tn
	}

	// Build tree
	for i := range nodes {
		n := &nodes[i]
		tn := nodeMap[n.ID]
		if n.ParentID == nil {
			roots = append(roots, *tn)
		} else if parent, ok := nodeMap[*n.ParentID]; ok {
			parent.Children = append(parent.Children, *tn)
		}
	}

	// Update roots with populated children
	for i, root := range roots {
		if populated, ok := nodeMap[root.ID]; ok {
			roots[i] = *populated
		}
	}

	return roots
}

// Course memberships

type AddMemberRequest struct {
	AccountID string `json:"account_id"`
	Role      string `json:"role"`
}

type MembershipResponse struct {
	ID        uuid.UUID `json:"id"`
	CourseID  uuid.UUID `json:"course_id"`
	AccountID uuid.UUID `json:"account_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func ToMembershipResponse(m *models.CourseMembership) MembershipResponse {
	return MembershipResponse{
		ID:        m.ID,
		CourseID:  m.CourseID,
		AccountID: m.AccountID,
		Role:      string(m.Role),
		CreatedAt: m.CreatedAt,
	}
}

func ToMembershipResponseList(memberships []models.CourseMembership) []MembershipResponse {
	result := make([]MembershipResponse, len(memberships))
	for i, m := range memberships {
		result[i] = ToMembershipResponse(&m)
	}
	return result
}
