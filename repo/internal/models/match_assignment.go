package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AssignmentRole string

const (
	AssignmentRoleReferee          AssignmentRole = "Referee"
	AssignmentRoleAssistantReferee AssignmentRole = "Assistant Referee"
	AssignmentRoleStaff            AssignmentRole = "Staff"
)

var ValidAssignmentRoles = map[AssignmentRole]bool{
	AssignmentRoleReferee:          true,
	AssignmentRoleAssistantReferee: true,
	AssignmentRoleStaff:            true,
}

func (r *AssignmentRole) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("assignment role cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("assignment role must be a string")
		}
		sv = string(bv)
	}
	*r = AssignmentRole(sv)
	return nil
}

func (r AssignmentRole) Value() (driver.Value, error) {
	return string(r), nil
}

type MatchAssignment struct {
	ID                  uuid.UUID      `db:"id" json:"id"`
	MatchID             uuid.UUID      `db:"match_id" json:"match_id"`
	AccountID           uuid.UUID      `db:"account_id" json:"account_id"`
	Role                AssignmentRole `db:"role" json:"role"`
	AssignedBy          uuid.UUID      `db:"assigned_by" json:"assigned_by"`
	ReassignmentReason  *string        `db:"reassignment_reason" json:"reassignment_reason,omitempty"`
	CreatedAt           time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at" json:"updated_at"`
}
