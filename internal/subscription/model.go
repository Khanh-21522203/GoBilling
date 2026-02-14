package subscription

import "time"

type Status string

const (
	StatusTrialing Status = "trialing"
	StatusActive   Status = "active"
	StatusPastDue  Status = "past_due"
	StatusPaused   Status = "paused"
	StatusCanceled Status = "canceled"
	StatusExpired  Status = "expired"
)

func (s Status) Valid() bool {
	switch s {
	case StatusTrialing, StatusActive, StatusPastDue, StatusPaused, StatusCanceled, StatusExpired:
		return true
	}
	return false
}

type Subscription struct {
	ID                 string
	CustomerID         string
	PlanID             string
	Status             Status
	Quantity           int
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	TrialStart         *time.Time
	TrialEnd           *time.Time
	CancelAtPeriodEnd  bool
	CanceledAt         *time.Time
	EndedAt            *time.Time
	Metadata           map[string]string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (s *Subscription) IsActive() bool {
	return s.Status == StatusActive || s.Status == StatusTrialing
}

func (s *Subscription) IsCanceled() bool {
	return s.Status == StatusCanceled || s.Status == StatusExpired
}

func (s *Subscription) CanBeModified() bool {
	return s.Status == StatusActive
}

func (s *Subscription) InTrial() bool {
	return s.Status == StatusTrialing
}

func (s *Subscription) Activate() error {
	if s.Status != StatusTrialing && s.Status != StatusPastDue && s.Status != StatusPaused {
		return ErrInvalidStatusTransition
	}
	s.Status = StatusActive
	s.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *Subscription) MarkPastDue() error {
	if s.Status != StatusActive {
		return ErrInvalidStatusTransition
	}
	s.Status = StatusPastDue
	s.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *Subscription) Pause() error {
	if s.Status != StatusActive {
		return ErrInvalidStatusTransition
	}
	s.Status = StatusPaused
	s.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *Subscription) Resume() error {
	if s.Status != StatusPaused {
		return ErrInvalidStatusTransition
	}
	s.Status = StatusActive
	s.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *Subscription) ScheduleCancellation() error {
	if s.Status != StatusActive && s.Status != StatusTrialing {
		return ErrInvalidStatusTransition
	}
	s.CancelAtPeriodEnd = true
	now := time.Now().UTC()
	s.CanceledAt = &now
	s.UpdatedAt = now
	return nil
}

func (s *Subscription) CancelImmediately() error {
	if s.IsCanceled() {
		return ErrAlreadyCanceled
	}
	s.Status = StatusCanceled
	now := time.Now().UTC()
	s.CanceledAt = &now
	s.EndedAt = &now
	s.UpdatedAt = now
	return nil
}

func (s *Subscription) ExtendPeriod(newEnd time.Time) {
	s.CurrentPeriodStart = s.CurrentPeriodEnd
	s.CurrentPeriodEnd = newEnd
	s.UpdatedAt = time.Now().UTC()
}

func (s *Subscription) ChangePlan(newPlanID string) error {
	if !s.CanBeModified() {
		return ErrCannotModify
	}
	s.PlanID = newPlanID
	s.UpdatedAt = time.Now().UTC()
	return nil
}
