package customer

import (
	"time"
)

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusDeleted   Status = "deleted"
)

func (s Status) Valid() bool {
	switch s {
	case StatusActive, StatusSuspended, StatusDeleted:
		return true
	}
	return false
}

type Customer struct {
	ID         string
	Email      string
	Name       string
	ExternalID *string
	Status     Status
	Metadata   map[string]string
	Version    int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

func (c *Customer) IsActive() bool {
	return c.Status == StatusActive
}

func (c *Customer) IsSuspended() bool {
	return c.Status == StatusSuspended
}

func (c *Customer) IsDeleted() bool {
	return c.Status == StatusDeleted
}

func (c *Customer) CanBeModified() bool {
	return c.Status == StatusActive || c.Status == StatusSuspended
}

func (c *Customer) Suspend() error {
	if c.Status != StatusActive {
		return ErrInvalidStatusTransition
	}
	c.Status = StatusSuspended
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *Customer) Reactivate() error {
	if c.Status != StatusSuspended {
		return ErrInvalidStatusTransition
	}
	c.Status = StatusActive
	c.UpdatedAt = time.Now().UTC()
	return nil
}

func (c *Customer) Delete() error {
	if c.Status == StatusDeleted {
		return ErrAlreadyDeleted
	}
	c.Status = StatusDeleted
	now := time.Now().UTC()
	c.DeletedAt = &now
	c.UpdatedAt = now
	return nil
}
