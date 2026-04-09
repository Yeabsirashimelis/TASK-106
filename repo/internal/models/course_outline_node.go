package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type OutlineNodeType string

const (
	NodeTypeChapter OutlineNodeType = "Chapter"
	NodeTypeUnit    OutlineNodeType = "Unit"
)

var ValidNodeTypes = map[OutlineNodeType]bool{
	NodeTypeChapter: true,
	NodeTypeUnit:    true,
}

func (n *OutlineNodeType) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("node type cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("node type must be a string")
		}
		sv = string(bv)
	}
	*n = OutlineNodeType(sv)
	return nil
}

func (n OutlineNodeType) Value() (driver.Value, error) {
	return string(n), nil
}

type CourseOutlineNode struct {
	ID          uuid.UUID       `db:"id" json:"id"`
	CourseID    uuid.UUID       `db:"course_id" json:"course_id"`
	ParentID    *uuid.UUID      `db:"parent_id" json:"parent_id,omitempty"`
	NodeType    OutlineNodeType `db:"node_type" json:"node_type"`
	Title       string          `db:"title" json:"title"`
	Description *string         `db:"description" json:"description,omitempty"`
	OrderIndex  int             `db:"order_index" json:"order_index"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
}
