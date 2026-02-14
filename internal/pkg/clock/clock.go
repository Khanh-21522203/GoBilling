package clock

import "time"

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now().UTC()
}

type MockClock struct {
	CurrentTime time.Time
}

func (m *MockClock) Now() time.Time {
	return m.CurrentTime
}

func (m *MockClock) Advance(d time.Duration) {
	m.CurrentTime = m.CurrentTime.Add(d)
}

func NewRealClock() Clock {
	return &RealClock{}
}

func NewMockClock(t time.Time) *MockClock {
	return &MockClock{CurrentTime: t}
}
