package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type MembershipRole string

const (
	MembershipRoleStaff    MembershipRole = "Staff"
	MembershipRoleEnrolled MembershipRole = "Enrolled"
)

var ValidMembershipRoles = map[MembershipRole]bool{
	MembershipRoleStaff:    true,
	MembershipRoleEnrolled: true,
}

func (r *MembershipRole) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("membership role cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("membership role must be a string")
		}
		sv = string(bv)
	}
	*r = MembershipRole(sv)
	return nil
}

func (r MembershipRole) Value() (driver.Value, error) {
	return string(r), nil
}

type CourseMembership struct {
	ID        uuid.UUID      `db:"id" json:"id"`
	CourseID  uuid.UUID      `db:"course_id" json:"course_id"`
	AccountID uuid.UUID      `db:"account_id" json:"account_id"`
	Role      MembershipRole `db:"role" json:"role"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
}
