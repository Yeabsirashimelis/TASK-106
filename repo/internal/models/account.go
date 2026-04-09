package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdministrator Role = "Administrator"
	RoleScheduler     Role = "Scheduler"
	RoleInstructor    Role = "Instructor"
	RoleReviewer      Role = "Reviewer"
	RoleFinanceClerk  Role = "Finance Clerk"
	RoleAuditor       Role = "Auditor"
)

var ValidRoles = map[Role]bool{
	RoleAdministrator: true,
	RoleScheduler:     true,
	RoleInstructor:    true,
	RoleReviewer:      true,
	RoleFinanceClerk:  true,
	RoleAuditor:       true,
}

func (r *Role) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("role cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("role must be a string")
		}
		sv = string(bv)
	}
	*r = Role(sv)
	return nil
}

func (r Role) Value() (driver.Value, error) {
	return string(r), nil
}

type Status string

const (
	StatusActive      Status = "Active"
	StatusFrozen      Status = "Frozen"
	StatusDeactivated Status = "Deactivated"
)

var ValidStatuses = map[Status]bool{
	StatusActive:      true,
	StatusFrozen:      true,
	StatusDeactivated: true,
}

func (s *Status) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("status cannot be null")
	}
	sv, ok := value.(string)
	if !ok {
		bv, ok := value.([]byte)
		if !ok {
			return fmt.Errorf("status must be a string")
		}
		sv = string(bv)
	}
	*s = Status(sv)
	return nil
}

func (s Status) Value() (driver.Value, error) {
	return string(s), nil
}

type Account struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Role         Role      `db:"role" json:"role"`
	Status       Status    `db:"status" json:"status"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
